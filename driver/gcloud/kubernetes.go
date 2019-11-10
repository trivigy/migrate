package gcloud

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	container "cloud.google.com/go/container/apiv1"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api/v1"

	// gcp auth client driver for kubectl
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/trivigy/migrate/v2/types"
)

// Kubernetes represents the driver for creating and destroying gke cluster.
type Kubernetes struct {
	Profile
	MachineType string `json:"machineType" yaml:"machineType"`
	ImageType   string `json:"imageType" yaml:"imageType"`
	NodeCount   int    `json:"nodeCount" yaml:"nodeCount"`
}

var _ interface {
	types.Creator
	types.Destroyer
	types.Sourcer
} = new(Kubernetes)

// Create executes the resource creation process.
func (r Kubernetes) Create(ctx context.Context, out io.Writer) error {
	ts, err := google.DefaultTokenSource(ctx, iam.CloudPlatformScope)
	if err != nil {
		return err
	}

	if err := r.EnsureCluster(ctx, out, ts); err != nil {
		return err
	}
	return nil
}

// EnsureCluster runs the steps for constructing a gke cluster.
func (r Kubernetes) EnsureCluster(ctx context.Context, out io.Writer, ts oauth2.TokenSource) error {
	service, err := container.NewClusterManagerClient(ctx, option.WithTokenSource(ts))
	if err != nil {
		return err
	}

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

		op, err := service.CreateCluster(ctx, req)
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

// Destroy executes the resource destruction process.
func (r Kubernetes) Destroy(ctx context.Context, out io.Writer) error {
	ts, err := google.DefaultTokenSource(ctx, iam.CloudPlatformScope)
	if err != nil {
		return err
	}

	if err := r.DestroyCluster(ctx, out, ts); err != nil {
		return err
	}
	return nil
}

// DestroyCluster destroys a cluster and waits for the completion of the
// operation.
func (r Kubernetes) DestroyCluster(ctx context.Context, out io.Writer, ts oauth2.TokenSource) error {
	service, err := container.NewClusterManagerClient(ctx, option.WithTokenSource(ts))
	if err != nil {
		return err
	}

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
	op, err := service.DeleteCluster(ctx, delReq)
	if err != nil {
		return err
	}

	if err := r.WaitForOp(ctx, service, op.Name); err != nil {
		return err
	}
	return nil
}

// Source returns the data source name for the driver.
func (r Kubernetes) Source(ctx context.Context, out io.Writer) error {
	ts, err := google.DefaultTokenSource(ctx, iam.CloudPlatformScope)
	if err != nil {
		return err
	}

	clusterInfo, err := r.getClusterInfo(ctx, ts)
	if err != nil {
		return err
	}

	token, err := ts.Token()
	if err != nil {
		return err
	}

	clusterName := fmt.Sprintf("gke_%s_%s_%s", r.ProjectID, r.Location, r.Name)
	caDec, _ := base64.StdEncoding.DecodeString(clusterInfo.MasterAuth.ClusterCaCertificate)
	gcloudPath, err := exec.LookPath("gcloud")
	if err != nil {
		return err
	}

	config := &clientcmdapi.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: []clientcmdapi.NamedCluster{
			{
				Name: clusterName,
				Cluster: clientcmdapi.Cluster{
					Server:                   "https://" + clusterInfo.Endpoint,
					CertificateAuthorityData: caDec,
				},
			},
		},
		AuthInfos: []clientcmdapi.NamedAuthInfo{
			{
				Name: clusterName,
				AuthInfo: clientcmdapi.AuthInfo{
					Token: token.AccessToken,
					AuthProvider: &clientcmdapi.AuthProviderConfig{
						Name: "gcp",
						Config: map[string]string{
							"cmd-path":   gcloudPath,
							"cmd-args":   "config config-helper --format=json",
							"expiry":     token.Expiry.UTC().Format(time.RFC3339),
							"token-key":  "{.credential.access_token}",
							"expiry-key": "{.credential.token_expiry}",
						},
					},
				},
			},
		},
		Contexts: []clientcmdapi.NamedContext{
			{
				Name: clusterName,
				Context: clientcmdapi.Context{
					Cluster:  clusterName,
					AuthInfo: clusterName,
				},
			},
		},
		CurrentContext: clusterName,
	}

	rbytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	if _, err := out.Write([]byte(string(rbytes) + "\n")); err != nil {
		return err
	}
	return nil
}

func (r Kubernetes) getClusterInfo(ctx context.Context, ts oauth2.TokenSource) (*containerpb.Cluster, error) {
	gcloud, err := container.NewClusterManagerClient(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}

	req := &containerpb.GetClusterRequest{
		Name: r.BasePath() + "/clusters/" + r.Name,
	}
	resp, err := gcloud.GetCluster(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
