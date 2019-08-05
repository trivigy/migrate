package provider

import (
	"context"
	"fmt"
	"io"
	"time"

	container "cloud.google.com/go/container/apiv1"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// GCP represents the driver for creating and destroying gke cluster.
type GCP struct {
	Name        string `json:"name" yaml:"name"`
	ProjectID   string `json:"projectID" yaml:"projectID"`
	Location    string `json:"location" yaml:"location"`
	MachineType string `json:"machineType" yaml:"machineType"`
	ImageType   string `json:"imageType" yaml:"imageType"`
}

// BasePath returns the base location path for kubernetes resources.
func (r GCP) BasePath() string {
	return fmt.Sprintf("projects/%s/locations/%s", r.ProjectID, r.Location)
}

// Setup executes the resource creation process.
func (r GCP) Setup(out io.Writer) error {
	ctx := context.Background()
	gcloud, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return err
	}

	ctx = context.Background()
	req := &containerpb.GetClusterRequest{
		Name: r.BasePath() + "/clusters/" + r.Name,
	}
	_, err = gcloud.GetCluster(ctx, req)
	if err == nil {
		return nil
	}

	stat, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch stat.Code() {
	case codes.NotFound:
		err := r.createCluster(out, gcloud)
		if err != nil {
			return err
		}
	default:
		return err
	}

	return nil
}

// createCluster runs the steps for constructing a gke cluster.
func (r GCP) createCluster(out io.Writer, gcloud *container.ClusterManagerClient) error {
	req := &containerpb.CreateClusterRequest{
		Parent: r.BasePath(),
		Cluster: &containerpb.Cluster{
			Name:             r.Name,
			InitialNodeCount: 1,
			NodeConfig: &containerpb.NodeConfig{
				MachineType: r.MachineType,
				ImageType:   r.ImageType,
				OauthScopes: container.DefaultAuthScopes(),
			},
		},
	}

	ctx := context.Background()
	resp, err := gcloud.CreateCluster(ctx, req)
	if err != nil {
		return err
	}

	_, err = r.waitOpDone(out, gcloud, resp.Name)
	if err != nil {
		return err
	}
	return nil
}

// func configureCluster(
// 	gcloud *container.ClusterManagerClient,
// 	config Config,
// 	clusterName string,
// ) error {
// 	kubectl, err := NewKubeCtl(gcloud, config, clusterName)
// 	if err != nil {
// 		return err
// 	}
//
// 	_, err = kubectl.CoreV1().
// 		ServiceAccounts("kube-system").
// 		Create(&corev1.ServiceAccount{
// 			ObjectMeta: metav1.ObjectMeta{Name: "tiller"},
// 		})
// 	if err != nil {
// 		return errors.WithStack(err)
// 	}
//
// 	_, err = kubectl.RbacV1().
// 		ClusterRoleBindings().
// 		Create(&rbacv1.ClusterRoleBinding{
// 			ObjectMeta: metav1.ObjectMeta{Name: "tiller-cluster-rule"},
// 			Subjects: []rbacv1.Subject{
// 				{Kind: "ServiceAccount", Namespace: "kube-system", Name: "tiller"},
// 			},
// 			RoleRef: rbacv1.RoleRef{Kind: "ClusterRole", Name: "cluster-admin"},
// 		})
// 	if err != nil {
// 		return errors.WithStack(err)
// 	}
//
// 	_, err = kubectl.AppsV1().
// 		Deployments("kube-system").
// 		Patch(
// 			"tiller-deploy",
// 			types.StrategicMergePatchType,
// 			[]byte(`{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}`),
// 		)
// 	if err != nil {
// 		return errors.WithStack(err)
// 	}
//
// 	return nil
// }
//

// TearDown executes the resource destruction process.
func (r GCP) TearDown(out io.Writer) error {
	ctx := context.Background()
	gcloud, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return err
	}

	ctx = context.Background()
	req := &containerpb.GetClusterRequest{
		Name: r.BasePath() + "/clusters/" + r.Name,
	}
	_, err = gcloud.GetCluster(ctx, req)
	if err == nil {
		err := r.destroyCluster(out, gcloud)
		if err != nil {
			return err
		}
		return nil
	}

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

// destroyCluster destroys a cluster and waits for the completion of the
// operation.
func (r GCP) destroyCluster(out io.Writer, gcloud *container.ClusterManagerClient) error {
	req := &containerpb.DeleteClusterRequest{
		Name: r.BasePath() + "/clusters/" + r.Name,
	}
	resp, err := gcloud.DeleteCluster(context.Background(), req)
	if err != nil {
		return err
	}

	_, err = r.waitOpDone(out, gcloud, resp.Name)
	if err != nil {
		return err
	}
	return nil
}

// waitOpDone allows for waiting of completion of an containerpb.Operation.
func (r GCP) waitOpDone(out io.Writer, gcloud *container.ClusterManagerClient, opName string) (*containerpb.Operation, error) {
	var resp *containerpb.Operation
	for {
		time.Sleep(1000 * time.Millisecond)
		if _, err := out.Write([]byte(".")); err != nil {
			return nil, err
		}

		req := &containerpb.GetOperationRequest{
			Name: r.BasePath() + "/operations/" + opName,
		}

		var err error
		resp, err = gcloud.GetOperation(context.Background(), req)
		if err != nil {
			return nil, err
		}

		if resp.GetStatus() == containerpb.Operation_DONE {
			if _, err := out.Write([]byte("OK\n")); err != nil {
				return nil, err
			}
			break
		}
	}
	return resp, nil
}
