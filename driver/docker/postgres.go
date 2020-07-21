// Package docker implements migrate drivers that operate on top of docker.
package docker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"

	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"

	// postgres driver
	_ "github.com/lib/pq"

	"github.com/trivigy/migrate/v2/driver"
	"github.com/trivigy/migrate/v2/internal/retry"
	"github.com/trivigy/migrate/v2/types"
)

// Postgres represents a driver for a docker based postgres database.
type Postgres struct {
	Name         string `json:"name" yaml:"name"`
	Version      string `json:"version" yaml:"version"`
	Password     string `json:"password,omitempty" yaml:"password,omitempty"`
	User         string `json:"user,omitempty" yaml:"user,omitempty"`
	DBName       string `json:"dbName,omitempty" yaml:"dbName,omitempty"`
	InitDBArgs   string `json:"initDBArgs,omitempty" yaml:"initDBArgs,omitempty"`
	InitDBWalDir string `json:"initDBWalDir,omitempty" yaml:"initDBWalDir,omitempty"`
	PGData       string `json:"pgData,omitempty" yaml:"pgData,omitempty"`
	Network      string `json:"network,omitempty" yaml:"network,omitempty"`
}

var _ interface {
	driver.WithCreate
	driver.WithDestroy
	driver.WithSource
} = new(Postgres)

// Create executes the resource creation process.
func (r Postgres) Create(ctx context.Context, out io.Writer) error {
	docker, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.38"),
	)
	if err != nil {
		return err
	}
	defer docker.Close()

	filter := filters.NewArgs()
	filter.Add("name", r.Name)
	listOpts := dtypes.ContainerListOptions{Filters: filter}
	containers, err := docker.ContainerList(ctx, listOpts)
	if err != nil {
		return err
	}

	if len(containers) != 0 {
		fmt.Fprintf(out, "database %q already exists\n", r.Name)
		return nil
	}

	refStr := "postgres:" + r.Version
	pullOpts := dtypes.ImagePullOptions{}
	reader, err := docker.ImagePull(ctx, refStr, pullOpts)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(reader)
	for {
		var jm jsonmessage.JSONMessage
		if err := decoder.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err := jm.Display(out, false); err != nil {
			return err
		}
	}

	envVars := make([]string, 0)
	if r.Password != "" {
		envVars = append(envVars, "POSTGRES_PASSWORD="+r.Password)
	}
	if r.User != "" {
		envVars = append(envVars, "POSTGRES_USER="+r.User)
	}
	if r.DBName != "" {
		envVars = append(envVars, "POSTGRES_DB="+r.DBName)
	}
	if r.InitDBArgs != "" {
		envVars = append(envVars, "POSTGRES_INITDB_ARGS="+r.InitDBArgs)
	}
	if r.InitDBWalDir != "" {
		envVars = append(envVars, "POSTGRES_INITDB_WALDIR="+r.InitDBWalDir)
	}
	if r.PGData != "" {
		envVars = append(envVars, "PGDATA="+r.PGData)
	}
	createCfg := &container.Config{Image: refStr, Tty: true, Env: envVars}
	hostCfg := &container.HostConfig{
		AutoRemove: true,
	}
	networkCfg := &network.NetworkingConfig{}
	if r.Network != "" {
		_, err := docker.NetworkInspect(ctx, r.Network, dtypes.NetworkInspectOptions{})
		if err != nil {
			switch {
			case client.IsErrNotFound(err):
				_, err = docker.NetworkCreate(ctx, r.Network, dtypes.NetworkCreate{Driver: "bridge"})
				if err != nil {
					return err
				}
			default:
				return err
			}
		}
		networkCfg.EndpointsConfig = map[string]*network.EndpointSettings{r.Network: {}}
	}

	resp, err := docker.ContainerCreate(ctx, createCfg, hostCfg, networkCfg, r.Name)
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

	var address string
	if r.Network != "" {
		address = info.NetworkSettings.Networks[r.Network].IPAddress
	} else {
		address = info.NetworkSettings.IPAddress
	}

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

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	if err := retry.Do(ctx, 2*time.Second, func() (bool, error) {
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
func (r Postgres) Destroy(ctx context.Context, out io.Writer) error {
	docker, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.38"),
	)
	if err != nil {
		return err
	}
	defer docker.Close()

	filter := filters.NewArgs()
	filter.Add("name", r.Name)
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

// Source returns the data source name for the driver.
func (r Postgres) Source(ctx context.Context, out io.Writer) error {
	docker, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.38"),
	)
	if err != nil {
		return err
	}
	defer docker.Close()

	filter := filters.NewArgs()
	filter.Add("name", r.Name)
	listOpts := dtypes.ContainerListOptions{Filters: filter}
	containers, err := docker.ContainerList(ctx, listOpts)
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return fmt.Errorf("container %q not found", r.Name)
	}

	info, err := docker.ContainerInspect(ctx, containers[0].ID)
	if err != nil {
		return err
	}

	var address string
	if r.Network != "" {
		address = info.NetworkSettings.Networks[r.Network].IPAddress
	} else {
		address = info.NetworkSettings.IPAddress
	}

	url := types.PsqlDSN{Host: address}
	if r.Password != "" {
		url.Password = r.Password
	}
	if r.User != "" {
		url.User = r.User
	} else {
		url.User = "postgres"
	}
	if r.DBName != "" {
		url.DBName = r.DBName
	} else {
		url.DBName = url.User
	}

	if _, err := out.Write([]byte(url.Source())); err != nil {
		return err
	}
	return nil
}
