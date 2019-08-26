package gcloud

import (
	"context"
	"encoding/base64"
	"io"

	container "cloud.google.com/go/container/apiv1"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"

	// gcp auth client driver for kubectl
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// Kubernetes represents the driver for creating and destroying gke cluster.
type Kubernetes struct {
	Config
	MachineType string `json:"machineType" yaml:"machineType"`
	ImageType   string `json:"imageType" yaml:"imageType"`
	NodeCount   int    `json:"nodeCount" yaml:"nodeCount"`
}

// Create executes the resource creation process.
func (r Kubernetes) Create(out io.Writer) error {
	if err := r.EnsureCluster(out); err != nil {
		return err
	}
	return nil
}

// EnsureCluster runs the steps for constructing a gke cluster.
func (r Kubernetes) EnsureCluster(out io.Writer) error {
	ctx := context.Background()
	service, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return err
	}

	ctx = context.Background()
	req := &containerpb.GetClusterRequest{
		Name: r.BasePath() + "/clusters/" + r.Name,
	}
	_, err = service.GetCluster(ctx, req)
	if err == nil {
		return nil
	}

	stat, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch stat.Code() {
	case codes.NotFound:
		req := &containerpb.CreateClusterRequest{
			Parent: r.BasePath(),
			Cluster: &containerpb.Cluster{
				Name:             r.Name,
				InitialNodeCount: int32(r.NodeCount),
				NodeConfig: &containerpb.NodeConfig{
					MachineType: r.MachineType,
					ImageType:   r.ImageType,
					OauthScopes: container.DefaultAuthScopes(),
				},
			},
		}

		ctx := context.Background()
		op, err := service.CreateCluster(ctx, req)
		if err != nil {
			return err
		}

		ctx = context.Background()
		if err := r.WaitForOp(ctx, service, op.Name); err != nil {
			return err
		}
	default:
		return err
	}

	return nil
}

// Destroy executes the resource destruction process.
func (r Kubernetes) Destroy(out io.Writer) error {
	if err := r.DestroyCluster(out); err != nil {
		return err
	}
	return nil
}

// DestroyCluster destroys a cluster and waits for the completion of the
// operation.
func (r Kubernetes) DestroyCluster(out io.Writer) error {
	ctx := context.Background()
	service, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return err
	}

	ctx = context.Background()
	getReq := &containerpb.GetClusterRequest{
		Name: r.BasePath() + "/clusters/" + r.Name,
	}
	_, err = service.GetCluster(ctx, getReq)
	if err != nil {
		stat, ok := status.FromError(err)
		if !ok {
			return err
		}

		switch stat.Code() {
		case codes.NotFound:
			return nil
		default:
			return err
		}
	}

	delReq := &containerpb.DeleteClusterRequest{
		Name: r.BasePath() + "/clusters/" + r.Name,
	}
	op, err := service.DeleteCluster(context.Background(), delReq)
	if err != nil {
		return err
	}

	ctx = context.Background()
	if err := r.WaitForOp(ctx, service, op.Name); err != nil {
		return err
	}
	return nil
}

// KubeConfig returns the content of kubeconfig file.
func (r Kubernetes) KubeConfig() (*rest.Config, error) {
	ctx := context.Background()
	gcloud, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return nil, err
	}

	req := &containerpb.GetClusterRequest{
		Name: r.BasePath() + "/clusters/" + r.Name,
	}
	resp, err := gcloud.GetCluster(context.Background(), req)
	if err != nil {
		return nil, err
	}

	decCAData, err := base64.StdEncoding.DecodeString(resp.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return nil, err
	}

	kubeConfig := &rest.Config{
		Host: "https://" + resp.Endpoint,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: false,
			CAData:   decCAData,
		},
		AuthConfigPersister: &InMemoryPersister{make(map[string]string)},
		AuthProvider:        &api.AuthProviderConfig{Name: "gcp"},
	}

	return kubeConfig, nil
}
