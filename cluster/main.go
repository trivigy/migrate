package cluster

import (
	"io"
	"path"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/cluster/release"
	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/internal/nub"
	"github.com/trivigy/migrate/v2/types"
)

const namespace = "cluster"

// Path generates a key path for a migrations stored in a shared map.
func Path(name string, tag string) string {
	_, caller, _, _ := runtime.Caller(1)
	group := path.Base(path.Dir(caller))
	return strings.Join([]string{namespace, group, tag + "_" + name}, "/")
}

// Filter iterates over all registered cluster migrations.
func Filter(fn func(release types.Release)) types.RangeFunc {
	return func(key, value interface{}) bool {
		fullname := strings.Split(key.(string), "/")
		if fullname[0] == namespace {
			fn(value.(types.Release))
		}
		return true
	}
}

// Collect iterates over all regirstered cluster migrations and adds them to
// the specified migration.
func Collect(releases *types.Releases) types.RangeFunc {
	return Filter(func(release types.Release) {
		*releases = append(*releases, release)
	})
}

// Cluster represents a cluster root command.
type Cluster struct {
	config map[string]config.Cluster
}

// NewCluster instantiates a new cluster command and returns it.
func NewCluster(config map[string]config.Cluster) types.Command {
	return &Cluster{config: config}
}

// NewCommand returns a new cobra.Command object.
func (r *Cluster) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          name,
		Short:        "Kubernetes cluster release and deployment controller.",
		Long:         "Kubernetes cluster release and deployment controller",
		SilenceUsage: true,
	}

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.AddCommand(
		NewCreate(r.config).(*Create).NewCommand("create"),
		NewDestroy(r.config).(*Destroy).NewCommand("destroy"),
		release.NewRelease(r.config).(*release.Release).NewCommand("release"),
	)

	pflags := cmd.PersistentFlags()
	pflags.Bool("help", false, "Show help information.")
	pflags.StringP(
		"env", "e", nub.DefaultEnvironment,
		"Run with env `ENV` configurations.",
	)

	flags := cmd.Flags()
	flags.SortFlags = false
	return cmd
}

// Execute runs the command.
func (r *Cluster) Execute(name string, output io.Writer, args []string) error {
	main := r.NewCommand(name)
	main.SetOut(output)
	main.SetArgs(args)
	if err := main.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// func buildClusterCommandConfig(cmd *cobra.Command, config *viper.Viper) (Cluster, string, error) {
// 	environment, err := cmd.Flags().GetString("env")
// 	if err != nil {
// 		return Cluster{}, "", errors.WithStack(err)
// 	}
//
// 	subtree := config.Sub(environment + ".cluster")
// 	if subtree == nil {
// 		return Cluster{}, "", errors.Errorf("missing %q configurations", environment)
// 	}
//
// 	cfg := Cluster{}
// 	if err := cfg.UnmarshalMap(subtree.AllSettings()); err != nil {
// 		return Cluster{}, "", errors.WithStack(err)
// 	}
//
// 	if environment == nub.DefaultEnvironment {
// 		environment, err = getCurrentBranchName()
// 		if err != nil {
// 			return Cluster{}, "", err
// 		}
// 	}
//
// 	return cfg, environment, nil
// }
//

// func NewKubeCtl(
// 	gcloud *container.ClusterManagerClient,
// 	config Cluster,
// 	clusterName string,
// ) (*kubernetes.Clientset, error) {
// 	req := &containerpb.GetClusterRequest{
// 		Name: config.Gcloud.BasePath() + "/clusters/" + clusterName,
// 	}
// 	resp, err := gcloud.GetCluster(context.Background(), req)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
//
// 	decCAData, err := base64.StdEncoding.DecodeString(resp.MasterAuth.ClusterCaCertificate)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
//
// 	kubeConfig := &rest.Cluster{
// 		Host: "https://" + resp.Endpoint,
// 		TLSClientConfig: rest.TLSClientConfig{
// 			Insecure: false,
// 			CAData:   decCAData,
// 		},
// 		AuthConfigPersister: &persister{},
// 		AuthProvider:        &api.AuthProviderConfig{Name: "gcp"},
// 	}
//
// 	kubectl, err := kubernetes.NewForConfig(kubeConfig)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
// 	return kubectl, nil
// }
