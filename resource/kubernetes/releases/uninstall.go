package releases

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	v1apps "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	v1ext "k8s.io/api/extensions/v1beta1"
	v1policy "k8s.io/api/policy/v1beta1"
	v1rbac "k8s.io/api/rbac/v1"
	v1err "k8s.io/apimachinery/pkg/api/errors"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/trivigy/migrate/v2/driver"
	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Uninstall represents the cluster release uninstall command object.
type Uninstall struct {
	Driver interface {
		driver.WithNamespace
		driver.WithReleases
		driver.WithSource
	} `json:"driver" yaml:"driver"`
}

// uninstallOptions is used for executing the run() command.
type uninstallOptions struct {
	Name     string         `json:"name" yaml:"name"`
	Version  semver.Version `json:"version" yaml:"version"`
	Resource string         `json:"resource" yaml:"resource"`
}

var _ interface {
	types.Resource
	types.Command
} = new(Uninstall)

// NewCommand creates a new cobra.Command, configures it, and returns it.
func (r Uninstall) NewCommand(ctx context.Context, name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name[strings.LastIndex(name, ".")+1:] + " [NAME[:VERSION]] [RESOURCE]",
		Short: "Stops a running release and removes the resources.",
		Long:  "Stops a running release and removes the resources",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			if patches, ok := r.Driver.(driver.WithPatches); ok {
				for _, patch := range *patches.Patches(name) {
					if err := patch.Do(ctx, cmd.OutOrStdout()); err != nil {
						return err
					}
				}
			}

			if try, _ := cmd.Flags().GetBool("try"); try {
				rbytes, err := yaml.Marshal(r.Driver)
				if err != nil {
					return err
				}

				fmt.Fprintf(cmd.OutOrStdout(), "%+v\n", string(rbytes))
				return nil
			}

			var name string
			var version semver.Version
			if len(args) > 0 {
				parts := strings.Split(args[0], ":")
				parts = append(parts, "")
				name = parts[0]

				if parts[1] != "" {
					var err error
					version, err = semver.Parse(parts[1])
					if err != nil {
						return err
					}
				}
			}

			var resource string
			if len(args) > 1 {
				resource = args[1]
			}

			opts := uninstallOptions{
				Name:     name,
				Version:  version,
				Resource: resource,
			}

			return r.run(ctx, cmd.OutOrStdout(), opts)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.SetUsageTemplate(global.DefaultUsageTemplate)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Bool(
		"try", false,
		"Simulates and prints resource execution parameters.",
	)
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Uninstall) Execute(name string, out io.Writer, args []string) error {
	wrap := types.Executor{Name: name, Command: r}
	ctx := context.WithValue(context.Background(), global.RefRoot, wrap)
	cmd := r.NewCommand(ctx, name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Uninstall) validation(cmd *cobra.Command, args []string) error {
	if err := require.MaxArgs(args, 2); err != nil {
		return err
	}
	return nil
}

// run is a starting point method for executing the cluster release uninstall
// command.
func (r Uninstall) run(ctx context.Context, out io.Writer, opts uninstallOptions) error {
	kubectl, err := GetK8sClientset(ctx, r.Driver, *r.Driver.Namespace())
	if err != nil {
		return err
	}

	sort.Sort(*r.Driver.Releases())
	for _, rel := range *r.Driver.Releases() {
		if opts.Name != "" && rel.Name != opts.Name ||
			(!opts.Version.EQ(semver.Version{}) &&
				!rel.Version.Equals(opts.Version)) {
			continue
		}

		for _, manifest := range rel.Manifests {
			if m, ok := manifest.(runtime.Object); ok && opts.Resource != "" {
				resource := m.GetObjectKind().GroupVersionKind().Kind
				if !strings.EqualFold(resource, opts.Resource) {
					continue
				}
			}

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
					Pods(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					Pods(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1core.ServiceAccount:
				_, err := kubectl.CoreV1().
					ServiceAccounts(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					ServiceAccounts(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1core.ConfigMap:
				_, err := kubectl.CoreV1().
					ConfigMaps(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					ConfigMaps(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1core.Endpoints:
				_, err := kubectl.CoreV1().
					Endpoints(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					Endpoints(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1core.Service:
				_, err := kubectl.CoreV1().
					Services(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					Services(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1core.Secret:
				_, err := kubectl.CoreV1().
					Secrets(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.CoreV1().
					Secrets(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1rbac.Role:
				_, err := kubectl.RbacV1().
					Roles(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.RbacV1().
					Roles(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1rbac.RoleBinding:
				_, err := kubectl.RbacV1().
					RoleBindings(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.RbacV1().
					RoleBindings(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
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
					DaemonSets(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.AppsV1().
					DaemonSets(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1apps.Deployment:
				_, err := kubectl.AppsV1().
					Deployments(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.AppsV1().
					Deployments(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Delete(manifest.Name, &v1meta.DeleteOptions{})
				if err != nil {
					return err
				}
			case *v1ext.Ingress:
				_, err := kubectl.ExtensionsV1beta1().
					Ingresses(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				err = kubectl.ExtensionsV1beta1().
					Ingresses(FallBackNS(manifest.Namespace, *r.Driver.Namespace())).
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
