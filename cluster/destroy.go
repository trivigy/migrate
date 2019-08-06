package cluster

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/internal/nub"
	"github.com/trivigy/migrate/v2/internal/require"
	"github.com/trivigy/migrate/v2/types"
)

// Destroy represents the cluster destroy command.
type Destroy struct {
	config map[string]config.Cluster
}

// NewDestroy instantiates and returns a destroy command object.
func NewDestroy(config map[string]config.Cluster) types.Command {
	return &Destroy{config: config}
}

// DestroyOptions is used for executing the Run() command.
type DestroyOptions struct {
	Env string `json:"env" yaml:"env"`
}

// NewCommand returns a new cobra.Command destroy command object.
func (r *Destroy) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Stops an existing running kubernetes cluster.",
		Long:  "Stops an existing running kubernetes cluster",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return errors.WithStack(err)
			}

			opts := CreateOptions{Env: env}
			return r.Run(cmd.OutOrStdout(), opts)
		},
	}

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
func (r *Destroy) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r *Destroy) validation(args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the destroy command.
func (r *Destroy) Run(out io.Writer, opts CreateOptions) error {
	cfg, ok := r.config[opts.Env]
	if !ok {
		return fmt.Errorf("missing %q environment configuration", opts.Env)
	}

	if err := cfg.Driver.TearDown(out); err != nil {
		return err
	}
	return nil
}

// func (r *DestroyCommand) run(cmd *cobra.Command, args []string) error {
// 	config, clusterName, err := buildClusterCommandConfig(cmd, r.config)
// 	if err != nil {
// 		return err
// 	}
//
// 	kubectl, err := container.NewClusterManagerClient(context.Background())
// 	if err != nil {
// 		return errors.WithStack(err)
// 	}
//
// 	req := &containerpb.GetClusterRequest{
// 		Name: config.Gcloud.BasePath() + "/clusters/" + clusterName,
// 	}
// 	_, err = kubectl.GetCluster(context.Background(), req)
// 	if err == nil {
// 		err := destroyCluster(kubectl, config, clusterName)
// 		if err != nil {
// 			return err
// 		}
// 		return nil
// 	}
//
// 	s, ok := status.FromError(err)
// 	if !ok {
// 		return errors.WithStack(err)
// 	}
//
// 	switch s.Code() {
// 	case codes.NotFound:
// 		return nil
// 	default:
// 		return errors.WithStack(err)
// 	}
// }
//
// func destroyCluster(
// 	client *container.ClusterManagerClient,
// 	config Cluster,
// 	clusterName string,
// ) error {
// 	req := &containerpb.DeleteClusterRequest{
// 		Name: config.Gcloud.BasePath() + "/clusters/" + clusterName,
// 	}
//
// 	resp, err := client.DeleteCluster(context.Background(), req)
// 	if err != nil {
// 		return errors.WithStack(err)
// 	}
//
// 	resp, err = waitOperationDone(client, config, resp.Name)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
