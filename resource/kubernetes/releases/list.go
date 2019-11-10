package releases

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
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

// List represents the cluster release list command object.
type List struct {
	Namespace *string         `json:"namespace" yaml:"namespace"`
	Releases  *types.Releases `json:"releases" yaml:"releases"`
	Driver    interface {
		types.Sourcer
	} `json:"driver" yaml:"driver"`
}

// ListOptions is used for executing the run() command.
type ListOptions struct{}

var _ interface {
	types.Resource
	types.Command
} = new(List)

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r List) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "List registered releases with states information.",
		Long:  "List registered releases with states information",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := ListOptions{}
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
func (r List) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r List) validation(cmd *cobra.Command, args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// run is a starting point method for executing the create command.
func (r List) run(ctx context.Context, out io.Writer, opts ListOptions) error {
	kubectl, err := GetK8sClientset(ctx, r.Driver, *r.Namespace)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Version", "Kind", "Status"})
	table.SetAutoWrapText(false)

	sort.Sort(*r.Releases)
	for _, rel := range *r.Releases {
		for j, manifest := range rel.Manifests {
			var kind string
			var status string
			switch manifest := manifest.(type) {
			case *v1core.Namespace:
				kind = manifest.Kind
				result, err := kubectl.CoreV1().
					Namespaces().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAME", "STATUS", "AGE"})

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Append([]string{
					result.ObjectMeta.Name,
					string(result.Status.Phase),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
				tbl.Render()
				status = buf.String()
			case *v1core.Pod:
				kind = manifest.Kind
				result, err := kubectl.CoreV1().
					Pods(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAMESPACE", "NAME", "READY", "STATUS", "RESTARTS", "AGE"})

				readyCount := 0
				restartCount := 0
				for _, stat := range result.Status.ContainerStatuses {
					restartCount += int(stat.RestartCount)
					if stat.Ready {
						readyCount++
					}
				}

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Append([]string{
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					fmt.Sprintf("%d/%d", readyCount, len(result.Status.ContainerStatuses)),
					string(result.Status.Phase),
					fmt.Sprintf("%d", restartCount),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
				tbl.Render()
				status = buf.String()
			case *v1core.ServiceAccount:
				kind = manifest.Kind
				result, err := kubectl.CoreV1().
					ServiceAccounts(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAMESPACE", "NAME", "SECRETS", "AGE"})

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Append([]string{
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					fmt.Sprintf("%d", len(result.Secrets)),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
				tbl.Render()
				status = buf.String()
			case *v1core.ConfigMap:
				kind = manifest.Kind
				result, err := kubectl.CoreV1().
					ConfigMaps(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAMESPACE", "NAME", "DATA", "AGE"})

				var rbytes []byte
				for _, data := range result.Data {
					rbytes = append(rbytes, []byte(data)...)
				}
				for _, binary := range result.BinaryData {
					rbytes = append(rbytes, binary...)
				}

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Append([]string{
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					fmt.Sprintf("%dB", len(rbytes)),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
				tbl.Render()
				status = buf.String()
			case *v1core.Endpoints:
				kind = manifest.Kind
				result, err := kubectl.CoreV1().
					Endpoints(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAMESPACE", "NAME", "ENDPOINTS", "AGE"})

				endpoints := make([]string, 0)
				for _, subset := range result.Subsets {
					for _, address := range subset.Addresses {
						for _, ports := range subset.Ports {
							endpoints = append(endpoints, fmt.Sprintf("%s:%d", address.IP, ports.Port))
						}
					}
				}
				if len(endpoints) > 3 {
					endpoints = append(endpoints[:3], fmt.Sprintf("+ %d more...", len(endpoints[3:])))
				}

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Append([]string{
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					strings.Join(endpoints, ","),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
				tbl.Render()
				status = buf.String()
			case *v1core.Service:
				kind = manifest.Kind
				result, err := kubectl.CoreV1().
					Services(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAMESPACE", "NAME", "TYPE", "CLUSTER-IP", "EXTERNAL-IP", "PORT(s)", "AGE"})

				externalIP := "<none>"
				if len(result.Status.LoadBalancer.Ingress) > 0 {
					externalIP = result.Status.LoadBalancer.Ingress[0].IP
				}

				var ports []string
				for _, port := range result.Spec.Ports {
					portStr := fmt.Sprintf("%d", port.Port)
					if port.NodePort != 0 {
						portStr += fmt.Sprintf(":%d", port.NodePort)
					}
					portStr += "/" + string(port.Protocol)
					ports = append(ports, portStr)
				}

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Append([]string{
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					string(result.Spec.Type),
					result.Spec.ClusterIP,
					externalIP,
					strings.Join(ports, ","),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
				tbl.Render()
				status = buf.String()
			case *v1core.Secret:
				kind = manifest.Kind
				result, err := kubectl.CoreV1().
					Secrets(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAMESPACE", "NAME", "TYPE", "DATA", "AGE"})

				var rbytes []byte
				for _, data := range result.StringData {
					rbytes = append(rbytes, []byte(data)...)
				}
				for _, binary := range result.Data {
					rbytes = append(rbytes, binary...)
				}

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Append([]string{
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					string(result.Type),
					fmt.Sprintf("%dB", len(rbytes)),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
				tbl.Render()
				status = buf.String()
			case *v1rbac.Role:
				kind = manifest.Kind
				result, err := kubectl.RbacV1().
					Roles(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAMESPACE", "NAME", "AGE"})

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Append([]string{
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
				tbl.Render()
				status = buf.String()
			case *v1rbac.RoleBinding:
				kind = manifest.Kind
				result, err := kubectl.RbacV1().
					RoleBindings(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAMESPACE", "NAME", "AGE"})

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Append([]string{
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
				tbl.Render()
				status = buf.String()
			case *v1rbac.ClusterRole:
				kind = manifest.Kind
				result, err := kubectl.RbacV1().
					ClusterRoles().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAME", "AGE"})

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Append([]string{
					result.ObjectMeta.Name,
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
				tbl.Render()
				status = buf.String()
			case *v1rbac.ClusterRoleBinding:
				kind = manifest.Kind
				result, err := kubectl.RbacV1().
					ClusterRoleBindings().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAME", "AGE"})

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Append([]string{
					result.ObjectMeta.Name,
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
				tbl.Render()
				status = buf.String()
			case *v1policy.PodSecurityPolicy:
				kind = manifest.Kind
				result, err := kubectl.PolicyV1beta1().
					PodSecurityPolicies().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAME", "PRIV", "CAPS", "VOLUMES"})

				var caps []string
				for _, capability := range result.Spec.AllowedCapabilities {
					caps = append(caps, string(capability))
				}

				var volumes []string
				for _, volume := range result.Spec.Volumes {
					volumes = append(volumes, string(volume))
				}

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Append([]string{
					result.ObjectMeta.Name,
					strconv.FormatBool(result.Spec.Privileged),
					strings.Join(caps, ","),
					strings.Join(volumes, ","),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
				tbl.Render()
				status = buf.String()
			case *v1apps.DaemonSet:
				kind = manifest.Kind
				result, err := kubectl.AppsV1().
					DaemonSets(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAMESPACE", "NAME", "DESIRED", "CURRENT", "READY", "UP-TO-DATE", "AVAILABLE", "AGE"})

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Append([]string{
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					fmt.Sprintf("%d", result.Status.DesiredNumberScheduled),
					fmt.Sprintf("%d", result.Status.CurrentNumberScheduled),
					fmt.Sprintf("%d", result.Status.NumberReady),
					fmt.Sprintf("%d", result.Status.UpdatedNumberScheduled),
					fmt.Sprintf("%d", result.Status.NumberAvailable),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
				tbl.Render()
				status = buf.String()
			case *v1apps.Deployment:
				kind = manifest.Kind
				result, err := kubectl.AppsV1().
					Deployments(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAMESPACE", "NAME", "READY", "UP-TO-DATE", "AVAILABLE", "AGE"})
				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Append([]string{
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					fmt.Sprintf("%d/%d", result.Status.ReadyReplicas, result.Status.Replicas),
					fmt.Sprintf("%d", result.Status.UpdatedReplicas),
					fmt.Sprintf("%d", result.Status.AvailableReplicas),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
				tbl.Render()
				status = buf.String()
			case *v1ext.Ingress:
				kind = manifest.Kind
				result, err := kubectl.ExtensionsV1beta1().
					Ingresses(FallBackNS(manifest.Namespace, *r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := NewEmbeddedTable()
				tbl.SetHeader([]string{"NAMESPACE", "NAME", "HOST", "PATH", "PORT", "AGE"})
				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)

				namespace := result.ObjectMeta.Namespace
				name := result.ObjectMeta.Name
				age := fmt.Sprintf("%.2fm", diffTime.Minutes())
				for i := range result.Spec.Rules {
					if i != 0 {
						namespace = ""
						name = ""
						age = ""
					}
					host := result.Spec.Rules[i].Host
					for j := range result.Spec.Rules[i].HTTP.Paths {
						if j != 0 {
							host = ""
						}
						path := result.Spec.Rules[i].HTTP.Paths[j].Path
						if path == "" {
							path = "/"
						}
						port := result.Spec.Rules[i].HTTP.Paths[j].Backend.ServicePort.String()
						tbl.Append([]string{namespace, name, host, path, port, age})
					}
				}
				tbl.Render()
				status = buf.String()
			default:
				return fmt.Errorf("unsupported manifest type %T", manifest)
			}

			lines := strings.Split(status, "\n")
			if j == 0 {
				for k, line := range lines {
					if k == 0 {
						table.Append([]string{
							rel.Name,
							rel.Version.String(),
							kind,
							line,
						})
					} else {
						table.Append([]string{"", "", "", line})
					}
				}
			} else {
				for k, line := range lines {
					if k == 0 {
						table.Append([]string{"", "", kind, line})
					} else {
						table.Append([]string{"", "", "", line})
					}
				}
			}
		}
	}

	if len(*r.Releases) > 0 {
		table.Render()
	}

	return nil
}
