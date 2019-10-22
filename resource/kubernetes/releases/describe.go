package releases

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/spf13/cobra"
	v1apps "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	v1ext "k8s.io/api/extensions/v1beta1"
	v1policy "k8s.io/api/policy/v1beta1"
	v1rbac "k8s.io/api/rbac/v1"
	v1err "k8s.io/apimachinery/pkg/api/errors"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Describe represents the cluster release inspect command object.
type Describe struct {
	common
	Namespace string          `json:"namespace" yaml:"namespace"`
	Releases  *types.Releases `json:"releases" yaml:"releases"`
	Driver    interface {
		types.KubeConfiged
	} `json:"driver" yaml:"driver"`
}

// InspectOptions is used for executing the Run() command.
type InspectOptions struct {
	Env      string         `json:"env" yaml:"env"`
	Name     string         `json:"name" yaml:"name"`
	Version  semver.Version `json:"version" yaml:"version"`
	Resource string         `json:"resource" yaml:"resource"`
	Filter   string         `json:"filter" yaml:"filter"`
}

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r Describe) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name + " NAME[:VERSION] RESOURCE FILTER",
		Short: "Prints release resources detail information.",
		Long:  "Prints release resources detail information",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return err
			}

			parts := strings.Split(args[0], ":")
			parts = append(parts, "")
			name := parts[0]

			version := semver.Version{}
			if parts[1] != "" {
				version, err = semver.Parse(parts[1])
				if err != nil {
					return err
				}
			}

			resource := args[1]
			filter := args[2]

			opts := InspectOptions{
				Env:      env,
				Name:     name,
				Version:  version,
				Resource: resource,
				Filter:   filter,
			}
			return r.Run(cmd.OutOrStdout(), opts)
		},
		SilenceUsage: true,
	}

	pflags := cmd.PersistentFlags()
	pflags.Bool("help", false, "Show help information.")
	pflags.StringP(
		"env", "e", global.DefaultEnvironment,
		"Run with env `ENV` configurations.",
	)

	flags := cmd.Flags()
	flags.SortFlags = false
	return cmd
}

// Execute runs the command.
func (r Describe) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Describe) validation(args []string) error {
	if err := require.ExactArgs(args, 3); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the cluster release inspect
// command.
func (r Describe) Run(out io.Writer, opts InspectOptions) error {
	kubectl, err := r.GetKubeCtl(r.Driver, r.Namespace)
	if err != nil {
		return err
	}

	sort.Sort(*r.Releases)
	for _, rel := range *r.Releases {
		if rel.Name != opts.Name ||
			(!opts.Version.EQ(semver.Version{}) &&
				!rel.Version.Equals(opts.Version)) {
			continue
		}

		for _, manifest := range rel.Manifests {
			if m, ok := manifest.(runtime.Object); ok {
				r := strings.ToLower(m.GetObjectKind().GroupVersionKind().Kind)
				if r != strings.ToLower(opts.Resource) {
					continue
				}
			}

			switch manifest := manifest.(type) {
			case *v1core.Namespace:
				result, err := kubectl.CoreV1().
					Namespaces().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					fmt.Fprintf(out, "%s", string(v1meta.StatusReasonNotFound))
					break
				} else if err != nil {
					return err
				}

				output, err := r.FilterResult(result, opts.Filter)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", output)
			case *v1core.Pod:
				result, err := kubectl.CoreV1().
					Pods(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					fmt.Fprintf(out, "%s", string(v1meta.StatusReasonNotFound))
					break
				} else if err != nil {
					return err
				}

				output, err := r.FilterResult(result, opts.Filter)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", output)
			case *v1core.ServiceAccount:
				result, err := kubectl.CoreV1().
					ServiceAccounts(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					fmt.Fprintf(out, "%s", string(v1meta.StatusReasonNotFound))
					break
				} else if err != nil {
					return err
				}

				output, err := r.FilterResult(result, opts.Filter)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", output)
			case *v1core.ConfigMap:
				result, err := kubectl.CoreV1().
					ConfigMaps(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					fmt.Fprintf(out, "%s", string(v1meta.StatusReasonNotFound))
					break
				} else if err != nil {
					return err
				}

				output, err := r.FilterResult(result, opts.Filter)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", output)
			case *v1core.Endpoints:
				result, err := kubectl.CoreV1().
					Endpoints(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					fmt.Fprintf(out, "%s", string(v1meta.StatusReasonNotFound))
					break
				} else if err != nil {
					return err
				}

				output, err := r.FilterResult(result, opts.Filter)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", output)
			case *v1core.Service:
				result, err := kubectl.CoreV1().
					Services(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					fmt.Fprintf(out, "%s", string(v1meta.StatusReasonNotFound))
					break
				} else if err != nil {
					return err
				}

				output, err := r.FilterResult(result, opts.Filter)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", output)
			case *v1rbac.Role:
				result, err := kubectl.RbacV1().
					Roles(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					fmt.Fprintf(out, "%s", string(v1meta.StatusReasonNotFound))
					break
				} else if err != nil {
					return err
				}

				output, err := r.FilterResult(result, opts.Filter)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", output)
			case *v1rbac.RoleBinding:
				result, err := kubectl.RbacV1().
					RoleBindings(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					fmt.Fprintf(out, "%s", string(v1meta.StatusReasonNotFound))
					break
				} else if err != nil {
					return err
				}

				output, err := r.FilterResult(result, opts.Filter)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", output)
			case *v1rbac.ClusterRole:
				result, err := kubectl.RbacV1().
					ClusterRoles().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					fmt.Fprintf(out, "%s", string(v1meta.StatusReasonNotFound))
					break
				} else if err != nil {
					return err
				}

				output, err := r.FilterResult(result, opts.Filter)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", output)
			case *v1rbac.ClusterRoleBinding:
				result, err := kubectl.RbacV1().
					ClusterRoleBindings().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					fmt.Fprintf(out, "%s", string(v1meta.StatusReasonNotFound))
					break
				} else if err != nil {
					return err
				}

				output, err := r.FilterResult(result, opts.Filter)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", output)
			case *v1policy.PodSecurityPolicy:
				result, err := kubectl.PolicyV1beta1().
					PodSecurityPolicies().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					fmt.Fprintf(out, "%s", string(v1meta.StatusReasonNotFound))
					break
				} else if err != nil {
					return err
				}

				output, err := r.FilterResult(result, opts.Filter)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", output)
			case *v1apps.DaemonSet:
				result, err := kubectl.AppsV1().
					DaemonSets(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					fmt.Fprintf(out, "%s", string(v1meta.StatusReasonNotFound))
					break
				} else if err != nil {
					return err
				}

				output, err := r.FilterResult(result, opts.Filter)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", output)
			case *v1apps.Deployment:
				result, err := kubectl.AppsV1().
					Deployments(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					fmt.Fprintf(out, "%s", string(v1meta.StatusReasonNotFound))
					break
				} else if err != nil {
					return err
				}

				output, err := r.FilterResult(result, opts.Filter)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", output)
			case *v1ext.Ingress:
				result, err := kubectl.ExtensionsV1beta1().
					Ingresses(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					fmt.Fprintf(out, "%s", string(v1meta.StatusReasonNotFound))
					break
				} else if err != nil {
					return err
				}

				output, err := r.FilterResult(result, opts.Filter)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", output)
			default:
				return fmt.Errorf("unsupported manifest type %T", manifest)
			}
		}
	}

	return nil
}
