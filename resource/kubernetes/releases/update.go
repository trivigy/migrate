package releases

import (
	"bytes"
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
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/trivigy/migrate/v2/driver"
	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Update represents the cluster release upgrade command object.
type Update struct {
	Driver interface {
		driver.WithNamespace
		driver.WithReleases
		driver.WithSource
	} `json:"driver" yaml:"driver"`
}

// updateOptions is used for executing the run() command.
type updateOptions struct {
	Name     string         `json:"name" yaml:"name"`
	Version  semver.Version `json:"version" yaml:"version"`
	Resource string         `json:"resource" yaml:"resource"`
}

var _ interface {
	types.Resource
	types.Command
} = new(Update)

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r Update) NewCommand(ctx context.Context, name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name[strings.LastIndex(name, ".")+1:] + " [NAME[:VERSION]] [RESOURCE]",
		Short: "Redeploy a modified release and track revision version.",
		Long:  "Redeploy a modified release and track revision version",
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

			opts := updateOptions{
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
func (r Update) Execute(name string, out io.Writer, args []string) error {
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
func (r Update) validation(cmd *cobra.Command, args []string) error {
	if err := require.MaxArgs(args, 2); err != nil {
		return err
	}
	return nil
}

// run is a starting point method for executing the create command.
func (r Update) run(ctx context.Context, out io.Writer, opts updateOptions) error {
	kubectl, err := GetK8sClientset(ctx, r.Driver, *r.Driver.Namespace())
	if err != nil {
		return err
	}

	decoder := scheme.Codecs.UniversalDeserializer()
	encoder := json.NewSerializerWithOptions(
		json.DefaultMetaFactory, nil, nil,
		json.SerializerOptions{Yaml: true, Strict: true},
	)

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
				instance, err := kubectl.CoreV1().
					Namespaces().
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.CoreV1().
					Namespaces().
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1core.Pod:
				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				instance, err := kubectl.CoreV1().
					Pods(namespace).
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.CoreV1().
					Pods(namespace).
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1core.ServiceAccount:
				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				instance, err := kubectl.CoreV1().
					ServiceAccounts(namespace).
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.CoreV1().
					ServiceAccounts(namespace).
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1core.ConfigMap:
				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				instance, err := kubectl.CoreV1().
					ConfigMaps(namespace).
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.CoreV1().
					ConfigMaps(namespace).
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1core.Endpoints:
				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				instance, err := kubectl.CoreV1().
					Endpoints(namespace).
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.CoreV1().
					Endpoints(namespace).
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1core.Service:
				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				instance, err := kubectl.CoreV1().
					Services(namespace).
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.CoreV1().
					Services(namespace).
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1core.Secret:
				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				instance, err := kubectl.CoreV1().
					Secrets(namespace).
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.CoreV1().
					Secrets(namespace).
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1rbac.Role:
				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				instance, err := kubectl.RbacV1().
					Roles(namespace).
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.RbacV1().
					Roles(namespace).
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1rbac.RoleBinding:
				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				instance, err := kubectl.RbacV1().
					RoleBindings(namespace).
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.RbacV1().
					RoleBindings(namespace).
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1rbac.ClusterRole:
				instance, err := kubectl.RbacV1().
					ClusterRoles().
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.RbacV1().
					ClusterRoles().
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1rbac.ClusterRoleBinding:
				instance, err := kubectl.RbacV1().
					ClusterRoleBindings().
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.RbacV1().
					ClusterRoleBindings().
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1policy.PodSecurityPolicy:
				instance, err := kubectl.PolicyV1beta1().
					PodSecurityPolicies().
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.PolicyV1beta1().
					PodSecurityPolicies().
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1apps.DaemonSet:
				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				instance, err := kubectl.AppsV1().
					DaemonSets(namespace).
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.AppsV1().
					DaemonSets(namespace).
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1apps.Deployment:
				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				instance, err := kubectl.AppsV1().
					Deployments(namespace).
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.AppsV1().
					Deployments(namespace).
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1apps.StatefulSet:
				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				instance, err := kubectl.AppsV1().
					StatefulSets(namespace).
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.AppsV1().
					StatefulSets(namespace).
					Update(ctx, instance, v1meta.UpdateOptions{})
				if err != nil {
					return err
				}
			case *v1ext.Ingress:
				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				instance, err := kubectl.ExtensionsV1beta1().
					Ingresses(namespace).
					Get(ctx, manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					continue
				} else if err != nil {
					return err
				}

				buffer := bytes.NewBuffer(nil)
				if err := encoder.Encode(manifest, buffer); err != nil {
					return err
				}
				_, _, err = decoder.Decode(buffer.Bytes(), nil, instance)
				if err != nil {
					return err
				}

				_, err = kubectl.ExtensionsV1beta1().
					Ingresses(namespace).
					Update(ctx, instance, v1meta.UpdateOptions{})
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
