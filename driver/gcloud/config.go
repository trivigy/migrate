// Package gcloud implements gcloud related migrate drivers.
package gcloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	container "cloud.google.com/go/container/apiv1"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

// Config represents the shared configuration information for all gcloud
// resource drivers.
type Config struct {
	Name      string `json:"name" yaml:"name"`
	ProjectID string `json:"projectID" yaml:"projectID"`
	Location  string `json:"location" yaml:"location"`
}

// BasePath returns the base location path for kubernetes resources.
func (r Config) BasePath() string {
	return fmt.Sprintf("projects/%s/locations/%s", r.ProjectID, r.Location)
}

// Region extracts and returns the cluster region from the Location field.
func (r Config) Region() string {
	parts := strings.Split(r.Location, "-")
	return strings.Join(parts[:2], "-")
}

// WaitForOp allows for waiting of completion of an api Operation.
func (r Config) WaitForOp(ctx context.Context, service interface{}, opName string) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			switch svc := service.(type) {
			case *container.ClusterManagerClient:
				req := &containerpb.GetOperationRequest{
					Name: r.BasePath() + "/operations/" + opName,
				}

				ctx := context.Background()
				op, err := svc.GetOperation(ctx, req)
				if err != nil {
					return err
				}

				if op.Status == containerpb.Operation_DONE {
					return nil
				}
			case *sqladmin.Service:
				op, err := svc.Operations.Get(r.ProjectID, opName).Do()
				if err != nil {
					return err
				}

				if op.Status == "DONE" {
					if op.Error != nil {
						var errs []string
						for _, e := range op.Error.Errors {
							errs = append(errs, e.Message)
						}
						return fmt.Errorf(
							"operation %q failed with error(s): %s",
							op.Name, strings.Join(errs, ","),
						)
					}
					return nil
				}
			default:
				return fmt.Errorf("unsupported service type %T", service)
			}
		}
	}
}

// InMemoryPersister is a helper method for defining `rest.Config` object.
type InMemoryPersister struct {
	store map[string]string
}

// Persist is a visitor function which does the storing operation.
func (i *InMemoryPersister) Persist(config map[string]string) error {
	i.store = make(map[string]string)
	for k, v := range config {
		i.store[k] = v
	}
	return nil
}
