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
	"github.com/trivigy/migrate/v2/nub"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Uninstall represents the cluster release uninstall command object.
type Uninstall struct {
	common
	config map[string]config.Cluster
}

// NewUninstall initializes a new cluster release uninstall command.
func NewUninstall(config map[string]config.Cluster) types.Command {
	return &Uninstall{config: config}
}

// UninstallOptions is used for executing the Run() command.
type UninstallOptions struct {
	Env string `json:"env" yaml:"env"`
}

// NewCommand creates a new cobra.Command, configures it, and returns it.
func (r *Uninstall) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Stops a running release and removes the resources.",
		Long:  "Stops a running release and removes the resources",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return errors.WithStack(err)
			}

			opts := UninstallOptions{Env: env}
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
func (r *Uninstall) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r *Uninstall) validation(args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the cluster release uninstall
// command.
func (r *Uninstall) Run(out io.Writer, opts UninstallOptions) error {
	cfg, ok := r.config[opts.Env]
	if !ok {
		return fmt.Errorf("missing %q environment configuration", opts.Env)
	}

	kubectl, err := r.GetKubeCtl(&cfg)
	if err != nil {
		return err
	}

	sort.Sort(*cfg.Releases)
	for _, rel := range *cfg.Releases {
		for _, manifest := range rel.Manifests {
			switch manifest := manifest.(type) {
			case *v1core.ConfigMap:
				_, err := kubectl.CoreV1().
					ConfigMaps(cfg.Namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					ConfigMaps(cfg.Namespace).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1core.Endpoints:
				_, err := kubectl.CoreV1().
					Endpoints(cfg.Namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					Endpoints(cfg.Namespace).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1core.Service:
				_, err := kubectl.CoreV1().
					Services(cfg.Namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					Services(cfg.Namespace).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1apps.Deployment:
				_, err := kubectl.AppsV1().
					Deployments(cfg.Namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.AppsV1().
					Deployments(cfg.Namespace).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported manifest type %T", manifest)
			}
		}
	}
	return nil
}