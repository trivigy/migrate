package gcloud

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os/user"
	"time"

	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	// postgres driver
	_ "github.com/lib/pq"

	"github.com/trivigy/migrate/v2/types"

	"github.com/trivigy/migrate/v2/internal/retry"
)

// CloudSQL represents a driver for gcloud CloudSQL service.
type CloudSQL struct {
	Profile
	User     string `json:"user" yaml:"user" validate:"required"`
	Password string `json:"password" yaml:"password" validate:"required"`
	DBName   string `json:"dbName" yaml:"dbName" validate:"required"`
}

var _ interface {
	types.Creator
	types.Destroyer
	types.Sourcer
} = new(CloudSQL)

// Create executes the resource creation process.
func (r CloudSQL) Create(ctx context.Context, out io.Writer) error {
	ts, err := google.DefaultTokenSource(ctx, iam.CloudPlatformScope)
	if err != nil {
		return err
	}

	if err := r.EnsureInstance(ctx, out, ts); err != nil {
		return err
	}
	if err := r.EnsureUsers(ctx, out, ts); err != nil {
		return err
	}
	if err := r.EnsureDatabase(ctx, out, ts); err != nil {
		return err
	}
	if err := r.EnsureConnection(ctx, out); err != nil {
		return err
	}
	return nil
}

// EnsureInstance ensures that a cloudsql instance has been created.
func (r CloudSQL) EnsureInstance(ctx context.Context, out io.Writer, ts oauth2.TokenSource) error {
	service, err := sqladmin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return err
	}

	inst, err := service.Instances.Get(r.ProjectID, r.Name).Do()
	if err == nil {
		inst.Settings.BackupConfiguration = &sqladmin.BackupConfiguration{
			Enabled:   true,
			StartTime: "08:00",
		}
		op, err := service.Instances.Patch(r.ProjectID, r.Name, inst).Do()
		if err != nil {
			return err
		}

		if err := r.WaitForOp(ctx, service, op.Name); err != nil {
			return err
		}
		return nil
	}

	switch err.(*googleapi.Error).Code {
	case http.StatusNotFound:
		op, err := service.Instances.
			Insert(r.ProjectID, &sqladmin.DatabaseInstance{
				Name:            r.Name,
				Region:          "us-east4",
				DatabaseVersion: "POSTGRES_9_6",
				Settings: &sqladmin.Settings{
					Tier: "db-custom-1-3840",
					LocationPreference: &sqladmin.LocationPreference{
						Zone: r.Location,
					},
					BackupConfiguration: &sqladmin.BackupConfiguration{
						Enabled:   true,
						StartTime: "08:00",
					},
				},
			}).Do()
		if err != nil {
			return err
		}

		if err := r.WaitForOp(ctx, service, op.Name); err != nil {
			return err
		}
	default:
		return err
	}
	return nil
}

// EnsureUsers ensures that a specific user is setup correctly.
func (r CloudSQL) EnsureUsers(ctx context.Context, out io.Writer, ts oauth2.TokenSource) error {
	service, err := sqladmin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return err
	}

	users, err := service.Users.List(r.ProjectID, r.Name).Do()
	if err != nil {
		return err
	}

	found := false
	for _, u := range users.Items {
		if u.Name == r.User {
			found = true
			u.Password = r.Password
			op, err := service.Users.Update(r.ProjectID, r.Name, u.Name, u).Do()
			if err != nil {
				return err
			}

			if err := r.WaitForOp(ctx, service, op.Name); err != nil {
				return err
			}
		}
	}

	if !found {
		op, err := service.Users.Insert(r.ProjectID, r.Name, &sqladmin.User{
			Name:     r.User,
			Password: r.Password,
		}).Do()
		if err != nil {
			return err
		}

		if err := r.WaitForOp(ctx, service, op.Name); err != nil {
			return err
		}
	}

	return nil
}

// EnsureDatabase ensures that a specific database on the cloudsql instance has
// been created.
func (r CloudSQL) EnsureDatabase(ctx context.Context, out io.Writer, ts oauth2.TokenSource) error {
	service, err := sqladmin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return err
	}

	inst, err := service.Instances.Get(r.ProjectID, r.Name).Do()
	if err != nil {
		return err
	}

	_, err = service.Databases.Get(r.ProjectID, r.Name, r.DBName).Do()
	if err == nil {
		return nil
	}

	switch err.(*googleapi.Error).Code {
	case http.StatusNotFound:
		op, err := service.Databases.
			Insert(r.ProjectID, inst.Name, &sqladmin.Database{
				Name: r.DBName,
			}).Do()
		if err != nil {
			return err
		}

		if err := r.WaitForOp(ctx, service, op.Name); err != nil {
			return err
		}
	default:
		return err
	}
	return nil
}

// EnsureConnection starts a gce-proxy container and ensures that a connection
// to the cloudsql database may be established.
func (r CloudSQL) EnsureConnection(ctx context.Context, out io.Writer) error {
	docker, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.38"),
	)
	if err != nil {
		return err
	}
	defer docker.Close()

	filter := filters.NewArgs()
	filter.Add("name", r.Name+"-gce-proxy")
	listOpts := dtypes.ContainerListOptions{Filters: filter}
	containers, err := docker.ContainerList(ctx, listOpts)
	if err != nil {
		return err
	}

	if len(containers) != 0 {
		return nil
	}

	refStr := "gcr.io/cloudsql-docker/gce-proxy:1.14"
	pullOpts := dtypes.ImagePullOptions{}
	reader, err := docker.ImagePull(ctx, refStr, pullOpts)
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, reader); err != nil {
		return err
	}

	containerPort, err := nat.NewPort("tcp", "5432")
	if err != nil {
		return err
	}
	createCfg := &container.Config{
		Image: refStr,
		Tty:   true,
		ExposedPorts: nat.PortSet{
			containerPort: struct{}{},
		},
		Cmd: []string{
			"/cloud_sql_proxy",
			fmt.Sprintf("-instances=%s=tcp:0.0.0.0:5432",
				r.ProjectID+":"+r.Region()+":"+r.Name),
		},
	}
	u, err := user.Current()
	if err != nil {
		return err
	}
	hostCfg := &container.HostConfig{
		AutoRemove: true,
		Mounts: []mount.Mount{
			{Type: mount.TypeBind, Source: u.HomeDir + "/.config/gcloud", Target: "/root/.config/gcloud"},
		},
	}
	networkCfg := &network.NetworkingConfig{}
	resp, err := docker.ContainerCreate(
		ctx,
		createCfg,
		hostCfg,
		networkCfg,
		r.Name+"-gce-proxy",
	)
	if err != nil {
		return err
	}

	startOpts := dtypes.ContainerStartOptions{}
	if err := docker.ContainerStart(ctx, resp.ID, startOpts); err != nil {
		return err
	}

	info, err := docker.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return err
	}

	address := info.NetworkSettings.IPAddress
	url := types.PsqlDSN{Host: address}
	if r.User != "" {
		url.User = r.User
	}
	if r.Password != "" {
		url.Password = r.Password
	}
	if r.DBName != "" {
		url.DBName = r.DBName
	}
	db, err := sql.Open(url.Driver(), url.Source())
	if err != nil {
		return err
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := retry.Do(ctx, 1*time.Second, func() (bool, error) {
		err := db.Ping()
		if err == nil {
			return false, nil
		}
		return true, err
	}); err != nil {
		return err
	}

	if err := db.Close(); err != nil {
		return err
	}
	return nil
}

// Destroy executes the resource destruction process.
func (r CloudSQL) Destroy(ctx context.Context, out io.Writer) error {
	ts, err := google.DefaultTokenSource(ctx, iam.CloudPlatformScope)
	if err != nil {
		return err
	}

	if err := r.DestroyConnection(ctx, out); err != nil {
		return err
	}
	if err := r.DestroyDatabase(ctx, out, ts); err != nil {
		return err
	}
	return nil
}

// DestroyConnection deletes the gce-proxy container.
func (r CloudSQL) DestroyConnection(ctx context.Context, out io.Writer) error {
	docker, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.38"),
	)
	if err != nil {
		return err
	}
	defer docker.Close()

	filter := filters.NewArgs()
	filter.Add("name", r.Name+"-gce-proxy")
	listOpts := dtypes.ContainerListOptions{Filters: filter}
	containers, err := docker.ContainerList(ctx, listOpts)
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return nil
	}

	logsOpts := dtypes.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}
	reader, err := docker.ContainerLogs(ctx, containers[0].ID, logsOpts)
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, reader); err != nil {
		return err
	}

	if err := docker.ContainerKill(ctx, containers[0].ID, "KILL"); err != nil {
		return err
	}
	return nil
}

// DestroyDatabase removes the specific database from the cloudsql instance.
func (r CloudSQL) DestroyDatabase(ctx context.Context, out io.Writer, ts oauth2.TokenSource) error {
	service, err := sqladmin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return err
	}

	_, err = service.Instances.Get(r.ProjectID, r.Name).Do()
	if err != nil {
		switch err.(*googleapi.Error).Code {
		case http.StatusNotFound:
			return nil
		default:
			return err
		}
	}

	_, err = service.Databases.Get(r.ProjectID, r.Name, r.DBName).Do()
	if err != nil {
		switch err.(*googleapi.Error).Code {
		case http.StatusNotFound:
			return nil
		default:
			return err
		}
	}

	op, err := service.Databases.Delete(r.ProjectID, r.Name, r.DBName).Do()
	if err != nil {
		return err
	}

	if err := r.WaitForOp(ctx, service, op.Name); err != nil {
		return err
	}
	return nil
}

// Source returns the data source name for the driver.
func (r CloudSQL) Source(ctx context.Context, out io.Writer) error {
	docker, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.38"),
	)
	if err != nil {
		return err
	}
	defer docker.Close()

	filter := filters.NewArgs()
	filter.Add("name", r.Name+"-gce-proxy")
	listOpts := dtypes.ContainerListOptions{Filters: filter}
	containers, err := docker.ContainerList(ctx, listOpts)
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return fmt.Errorf("container %q not found", r.Name+"-gce-proxy")
	}

	info, err := docker.ContainerInspect(ctx, containers[0].ID)
	if err != nil {
		return err
	}

	address := info.NetworkSettings.IPAddress
	url := types.PsqlDSN{Host: address}
	if r.User != "" {
		url.User = r.User
	}
	if r.Password != "" {
		url.Password = r.Password
	}
	if r.DBName != "" {
		url.DBName = r.DBName
	}

	if _, err := out.Write([]byte(url.Source())); err != nil {
		return err
	}
	return nil
}
