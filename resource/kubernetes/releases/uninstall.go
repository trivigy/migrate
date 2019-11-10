package releases

import (
	"context"
	"fmt"
	"io"
	"sort"

	"github.com/spf13/cobra"
	v1apps "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	v1ext "k8s.io/api/extensions/v1beta1"
	v1policy "k8s.io/api/policy/v1beta1"
	v1rbac "k8s.io/api/rbac/v1"
	v1err "k8s.io/apimachinery/pkg/api/errors"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Uninstall represents the cluster release uninstall command object.
type Uninstall struct {
	Namespace *string         `json:"namespace" yaml:"namespace"`
	Releases  *types.Releases `json:"releases" yaml:"releases"`
	Driver    interface {
		types.Sourcer
	} `json:"driver" yaml:"driver"`
}

// UninstallOptions is used for executing the run() command.
type UninstallOptions struct{}

var _ interface {
	types.Resource
	types.Command
} = new(Uninstall)

// NewCommand creates a new cobra.Command, configures it, and returns it.
func (r Uninstall) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Stops a running release and removes the resources.",
		Long:  "Stops a running release and removes the resources",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := UninstallOptions{}
			return r.run(context.Background(), cmd.OutOrStdout(), opts)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.SetUsageTemplate(global.DefaultUsageTemplate)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Uninstall) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Uninstall) validation(cmd *cobra.Command, args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// run is a starting point method for executing the cluster release uninstall
// command.
func (r Uninstall) run(ctx context.Context, out io.Writer, opts UninstallOptions) error {
	kubectl, err := GetK8sClientset(ctx, r.Driver, *r.Namespace)
	if err != nil {
		return err
	}

	sort.Sort(*r.Releases)
	for _, rel := range *r.Releases {
		for _, manifest := range rel.Manifests {
			switch manifest := manifest.(type) {
			case *v1core.Namespace:
				_, err := kubectl.CoreV1().
					Namespaces().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					Namespaces().
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1core.Pod:
				_, err := kubectl.CoreV1().
					Pods(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					Pods(FallBackNS(manifest.Namespace, *r.Namespace)).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1core.ServiceAccount:
				_, err := kubectl.CoreV1().
					ServiceAccounts(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					ServiceAccounts(FallBackNS(manifest.Namespace, *r.Namespace)).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1core.ConfigMap:
				_, err := kubectl.CoreV1().
					ConfigMaps(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					ConfigMaps(FallBackNS(manifest.Namespace, *r.Namespace)).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1core.Endpoints:
				_, err := kubectl.CoreV1().
					Endpoints(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					Endpoints(FallBackNS(manifest.Namespace, *r.Namespace)).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1core.Service:
				_, err := kubectl.CoreV1().
					Services(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					Services(FallBackNS(manifest.Namespace, *r.Namespace)).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1core.Secret:
				_, err := kubectl.CoreV1().
					Secrets(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					Secrets(FallBackNS(manifest.Namespace, *r.Namespace)).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1rbac.Role:
				_, err := kubectl.RbacV1().
					Roles(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.RbacV1().
					Roles(FallBackNS(manifest.Namespace, *r.Namespace)).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1rbac.RoleBinding:
				_, err := kubectl.RbacV1().
					RoleBindings(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.RbacV1().
					RoleBindings(FallBackNS(manifest.Namespace, *r.Namespace)).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1rbac.ClusterRole:
				_, err := kubectl.RbacV1().
					ClusterRoles().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.RbacV1().
					ClusterRoles().
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1rbac.ClusterRoleBinding:
				_, err := kubectl.RbacV1().
					ClusterRoleBindings().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.RbacV1().
					ClusterRoleBindings().
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1policy.PodSecurityPolicy:
				_, err := kubectl.PolicyV1beta1().
					PodSecurityPolicies().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.PolicyV1beta1().
					PodSecurityPolicies().
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1apps.DaemonSet:
				_, err := kubectl.AppsV1().
					DaemonSets(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.AppsV1().
					DaemonSets(FallBackNS(manifest.Namespace, *r.Namespace)).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1apps.Deployment:
				_, err := kubectl.AppsV1().
					Deployments(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.AppsV1().
					Deployments(FallBackNS(manifest.Namespace, *r.Namespace)).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1ext.Ingress:
				_, err := kubectl.ExtensionsV1beta1().
					Ingresses(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.ExtensionsV1beta1().
					Ingresses(FallBackNS(manifest.Namespace, *r.Namespace)).
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
