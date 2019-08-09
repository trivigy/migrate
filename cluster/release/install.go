package release

import (
	"fmt"
	"io"
	"sort"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1apps "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	v1err "k8s.io/apimachinery/pkg/api/errors"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/internal/nub"
	"github.com/trivigy/migrate/v2/internal/require"
	"github.com/trivigy/migrate/v2/types"
)

// Install represents the cluster release install command object.
type Install struct {
	common
	config map[string]config.Cluster
}

// NewInstall initializes a new cluster release install command.
func NewInstall(config map[string]config.Cluster) types.Command {
	return &Install{config: config}
}

// InstallOptions is used for executing the Run() command.
type InstallOptions struct {
	Env string `json:"env" yaml:"env"`
}

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r *Install) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Deploys release resources on running cluster.",
		Long:  "Deploys release resources on running cluster",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return errors.WithStack(err)
			}

			opts := InstallOptions{Env: env}
			return r.Run(cmd.OutOrStdout(), opts)
		},
		SilenceUsage: true,
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
func (r *Install) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r *Install) validation(args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the cluster release install
// command.
func (r *Install) Run(out io.Writer, opts InstallOptions) error {
	cfg, ok := r.config[opts.Env]
	if !ok {
		return fmt.Errorf("missing %q environment configuration", opts.Env)
	}

	kubectl, err := r.GetKubeCtl(cfg)
	if err != nil {
		return err
	}

	sort.Sort(cfg.Releases)
	for _, rel := range cfg.Releases {
		for _, manifest := range rel.Manifests {
			switch manifest := manifest.(type) {
			case *v1core.ConfigMap:
				_, err := kubectl.CoreV1().
					ConfigMaps(cfg.Namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.CoreV1().
						ConfigMaps(cfg.Namespace).
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1core.Service:
				_, err := kubectl.CoreV1().
					Services(cfg.Namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.CoreV1().
						Services(cfg.Namespace).
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1apps.Deployment:
				_, err := kubectl.AppsV1().
					Deployments(cfg.Namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.AppsV1().
						Deployments(cfg.Namespace).
						Create(manifest)
					if err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("unsupported manifest type %T", manifest)
			}
		}
	}
	return nil
}
