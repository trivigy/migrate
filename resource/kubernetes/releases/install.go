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

// Install represents the cluster release install command object.
type Install struct {
	Namespace *string         `json:"namespace" yaml:"namespace"`
	Releases  *types.Releases `json:"releases" yaml:"releases"`
	Driver    interface {
		types.Sourcer
	} `json:"driver" yaml:"driver"`
}

// InstallOptions is used for executing the run() command.
type InstallOptions struct {
	Seq int `json:"seq" yaml:"seq"`
}

var _ interface {
	types.Resource
	types.Command
} = new(Install)

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r Install) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Deploys release resources on running cluster.",
		Long:  "Deploys release resources on running cluster",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			seq, err := cmd.Flags().GetInt("seq")
			if err != nil {
				return err
			}

			opts := InstallOptions{Seq: seq}
			return r.run(context.Background(), cmd.OutOrStdout(), opts)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.SetUsageTemplate(global.DefaultUsageTemplate)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.IntP(
		"seq", "s", 0,
		"Indicate `ID` of the release to apply. All applied by default.",
	)
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Install) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Install) validation(cmd *cobra.Command, args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// run is a starting point method for executing the cluster release install
// command.
func (r Install) run(ctx context.Context, out io.Writer, opts InstallOptions) error {
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
					_, err := kubectl.CoreV1().
						Namespaces().
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1core.Pod:
				_, err := kubectl.CoreV1().
					Pods(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.CoreV1().
						Pods(FallBackNS(manifest.Namespace, *r.Namespace)).
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1core.ServiceAccount:
				_, err := kubectl.CoreV1().
					ServiceAccounts(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.CoreV1().
						ServiceAccounts(FallBackNS(manifest.Namespace, *r.Namespace)).
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1core.ConfigMap:
				_, err := kubectl.CoreV1().
					ConfigMaps(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.CoreV1().
						ConfigMaps(FallBackNS(manifest.Namespace, *r.Namespace)).
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1core.Endpoints:
				_, err := kubectl.CoreV1().
					ConfigMaps(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.CoreV1().
						Endpoints(FallBackNS(manifest.Namespace, *r.Namespace)).
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1core.Service:
				_, err := kubectl.CoreV1().
					Services(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.CoreV1().
						Services(FallBackNS(manifest.Namespace, *r.Namespace)).
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1core.Secret:
				_, err := kubectl.CoreV1().
					Secrets(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.CoreV1().
						Secrets(FallBackNS(manifest.Namespace, *r.Namespace)).
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1rbac.Role:
				_, err := kubectl.RbacV1().
					Roles(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.RbacV1().
						Roles(FallBackNS(manifest.Namespace, *r.Namespace)).
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1rbac.RoleBinding:
				_, err := kubectl.RbacV1().
					RoleBindings(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.RbacV1().
						RoleBindings(FallBackNS(manifest.Namespace, *r.Namespace)).
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1rbac.ClusterRole:
				_, err := kubectl.RbacV1().
					ClusterRoles().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.RbacV1().
						ClusterRoles().
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1rbac.ClusterRoleBinding:
				_, err := kubectl.RbacV1().
					ClusterRoleBindings().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.RbacV1().
						ClusterRoleBindings().
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1policy.PodSecurityPolicy:
				_, err := kubectl.PolicyV1beta1().
					PodSecurityPolicies().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.PolicyV1beta1().
						PodSecurityPolicies().
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1apps.DaemonSet:
				_, err := kubectl.AppsV1().
					DaemonSets(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.AppsV1().
						DaemonSets(FallBackNS(manifest.Namespace, *r.Namespace)).
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1apps.Deployment:
				_, err := kubectl.AppsV1().
					Deployments(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.AppsV1().
						Deployments(FallBackNS(manifest.Namespace, *r.Namespace)).
						Create(manifest)
					if err != nil {
						return err
					}
				}
			case *v1ext.Ingress:
				_, err := kubectl.ExtensionsV1beta1().
					Ingresses(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					_, err := kubectl.ExtensionsV1beta1().
						Ingresses(FallBackNS(manifest.Namespace, *r.Namespace)).
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
