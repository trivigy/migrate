package testutils

import (
	"context"
	"io"

	"github.com/trivigy/migrate/v2/driver"
	"github.com/trivigy/migrate/v2/types"
)

type Kubernetes struct {
	Namespace *string         `json:"namespace" yaml:"namespace"`
	Releases  *types.Releases `json:"releases" yaml:"releases"`
	Driver    interface {
		driver.WithCreate
		driver.WithDestroy
		driver.WithSource
	} `json:"driver" yaml:"driver"`
}

func (r Kubernetes) Build() *kubernetesImpl {
	return &kubernetesImpl{
		namespace: r.Namespace,
		releases:  r.Releases,
		driver:    r.Driver,
	}
}

type kubernetesImpl struct {
	namespace *string
	releases  *types.Releases
	driver    interface {
		driver.WithCreate
		driver.WithDestroy
		driver.WithSource
	}
}

var _ interface {
	driver.WithCreate
	driver.WithDestroy
	driver.WithNamespace
	driver.WithReleases
	driver.WithSource
} = new(kubernetesImpl)

func (r kubernetesImpl) Namespace() *string {
	return r.namespace
}

func (r kubernetesImpl) Releases() *types.Releases {
	return r.releases
}

// Create executes the resource creation process.
func (r kubernetesImpl) Create(ctx context.Context, out io.Writer) error {
	return r.driver.Create(ctx, out)
}

// Destroy executes the resource destruction process.
func (r kubernetesImpl) Destroy(ctx context.Context, out io.Writer) error {
	return r.driver.Destroy(ctx, out)
}

// Source returns the data source name for the driver.
func (r kubernetesImpl) Source(ctx context.Context, out io.Writer) error {
	return r.driver.Source(ctx, out)
}
