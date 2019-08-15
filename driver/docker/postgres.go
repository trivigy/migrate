package docker

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"

	// postgres driver
	_ "github.com/lib/pq"

	"github.com/trivigy/migrate/v2/internal/retry"
	mg8types "github.com/trivigy/migrate/v2/types"
)

// Postgres represents a driver for a docker based postgres database.
type Postgres struct {
	Name         string
	Version      string
	Password     string
	User         string
	DBName       string
	InitDBArgs   string
	InitDBWalDir string
	PGData       string
}

// Setup executes the resource creation process.
func (r Postgres) Setup(out io.Writer) error {
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
	filter.Add("name", r.Name)
	listOpts := types.ContainerListOptions{Filters: filter}
	containers, err := docker.ContainerList(ctx, listOpts)
	if err != nil {
		return err
	}

	if len(containers) != 0 {
		return nil
	}

	ctx = context.Background()
	refStr := "postgres:" + r.Version
	pullOpts := types.ImagePullOptions{}
	reader, err := docker.ImagePull(ctx, refStr, pullOpts)
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, reader); err != nil {
		return err
	}

	ctx = context.Background()
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
	resp, err := docker.ContainerCreate(ctx, createCfg, hostCfg, networkCfg, r.Name)
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

// TearDown executes the resource destruction process.
func (r Postgres) TearDown(out io.Writer) error {
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
	filter.Add("name", r.Name)
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

// Source returns the data source name for the driver.
func (r Postgres) Source() (string, error) {
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
	filter.Add("name", r.Name)
	listOpts := types.ContainerListOptions{Filters: filter}
	containers, err := docker.ContainerList(ctx, listOpts)
	if err != nil {
		return "", err
	}

	if len(containers) == 0 {
		return "", fmt.Errorf("container %q not found", r.Name)
	}

	ctx = context.Background()
	info, err := docker.ContainerInspect(ctx, containers[0].ID)
	if err != nil {
		return "", err
	}

	address := info.NetworkSettings.IPAddress
	url := mg8types.PsqlDSN{Host: address}
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
	return url.Source(), nil
}
