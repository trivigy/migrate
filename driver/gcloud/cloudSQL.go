package gcloud

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os/user"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"google.golang.org/api/googleapi"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"

	// postgres driver
	_ "github.com/lib/pq"

	mg8types "github.com/trivigy/migrate/v2/types"

	"github.com/trivigy/migrate/v2/internal/retry"
)

// CloudSQL represents a driver for gcloud CloudSQL service.
type CloudSQL struct {
	Config
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	DBName   string `json:"dbName" yaml:"dbName"`
}

// Create executes the resource creation process.
func (r CloudSQL) Create(out io.Writer) error {
	if err := r.EnsureInstance(out); err != nil {
		return err
	}
	if err := r.EnsureUsers(out); err != nil {
		return err
	}
	if err := r.EnsureDatabase(out); err != nil {
		return err
	}
	if err := r.EnsureConnection(out); err != nil {
		return err
	}
	return nil
}

// EnsureInstance ensures that a cloudsql instance has been created.
func (r CloudSQL) EnsureInstance(out io.Writer) error {
	ctx := context.Background()
	service, err := sqladmin.NewService(ctx)
	if err != nil {
		return err
	}

	_, err = service.Instances.Get(r.ProjectID, r.Name).Do()
	if err == nil {
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
				},
			}).Do()
		if err != nil {
			return err
		}

		ctx := context.Background()
		if err := r.WaitForOp(ctx, service, op.Name); err != nil {
			return err
		}
	default:
		return err
	}
	return nil
}

// EnsureUsers ensures that a specific user is setup correctly.
func (r CloudSQL) EnsureUsers(out io.Writer) error {
	ctx := context.Background()
	service, err := sqladmin.NewService(ctx)
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

			ctx := context.Background()
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

		ctx := context.Background()
		if err := r.WaitForOp(ctx, service, op.Name); err != nil {
			return err
		}
	}

	return nil
}

// EnsureDatabase ensures that a specific database on the cloudsql instance has
// been created.
func (r CloudSQL) EnsureDatabase(out io.Writer) error {
	ctx := context.Background()
	service, err := sqladmin.NewService(ctx)
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

		ctx := context.Background()
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
func (r CloudSQL) EnsureConnection(out io.Writer) error {
	docker, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.38"),
	)
	if err != nil {
		return err
	}
	defer docker.Close()

	ctx := context.Background()
	filter := filters.NewArgs()
	filter.Add("name", r.Name+"-gce-proxy")
	listOpts := types.ContainerListOptions{Filters: filter}
	containers, err := docker.ContainerList(ctx, listOpts)
	if err != nil {
		return err
	}

	if len(containers) != 0 {
		return nil
	}

	ctx = context.Background()
	refStr := "gcr.io/cloudsql-docker/gce-proxy:1.14"
	pullOpts := types.ImagePullOptions{}
	reader, err := docker.ImagePull(ctx, refStr, pullOpts)
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, reader); err != nil {
		return err
	}

	ctx = context.Background()
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

	ctx = context.Background()
	startOpts := types.ContainerStartOptions{}
	if err := docker.ContainerStart(ctx, resp.ID, startOpts); err != nil {
		return err
	}

	ctx = context.Background()
	info, err := docker.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return err
	}

	address := info.NetworkSettings.IPAddress
	url := mg8types.PsqlDSN{Host: address}
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

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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
func (r CloudSQL) Destroy(out io.Writer) error {
	if err := r.DestroyConnection(out); err != nil {
		return err
	}
	if err := r.DestroyDatabase(out); err != nil {
		return err
	}
	return nil
}

// DestroyConnection deletes the gce-proxy container.
func (r CloudSQL) DestroyConnection(out io.Writer) error {
	docker, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.38"),
	)
	if err != nil {
		return err
	}
	defer docker.Close()

	ctx := context.Background()
	filter := filters.NewArgs()
	filter.Add("name", r.Name+"-gce-proxy")
	listOpts := types.ContainerListOptions{Filters: filter}
	containers, err := docker.ContainerList(ctx, listOpts)
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return nil
	}

	ctx = context.Background()
	logsOpts := types.ContainerLogsOptions{
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

	ctx = context.Background()
	if err := docker.ContainerKill(ctx, containers[0].ID, "KILL"); err != nil {
		return err
	}
	return nil
}

// DestroyDatabase removes the specific database from the cloudsql instance.
func (r CloudSQL) DestroyDatabase(out io.Writer) error {
	ctx := context.Background()
	service, err := sqladmin.NewService(ctx)
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

	ctx = context.Background()
	if err := r.WaitForOp(ctx, service, op.Name); err != nil {
		return err
	}
	return nil
}

// Source returns the data source name for the driver.
func (r CloudSQL) Source() (string, error) {
	docker, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.38"),
	)
	if err != nil {
		return "", err
	}
	defer docker.Close()

	ctx := context.Background()
	filter := filters.NewArgs()
	filter.Add("name", r.Name+"-gce-proxy")
	listOpts := types.ContainerListOptions{Filters: filter}
	containers, err := docker.ContainerList(ctx, listOpts)
	if err != nil {
		return "", err
	}

	if len(containers) == 0 {
		return "", fmt.Errorf("container %q not found", r.Name+"-gce-proxy")
	}

	ctx = context.Background()
	info, err := docker.ContainerInspect(ctx, containers[0].ID)
	if err != nil {
		return "", err
	}

	address := info.NetworkSettings.IPAddress
	url := mg8types.PsqlDSN{Host: address}
	if r.User != "" {
		url.User = r.User
	}
	if r.Password != "" {
		url.Password = r.Password
	}
	if r.DBName != "" {
		url.DBName = r.DBName
	}
	return url.Source(), nil
}
