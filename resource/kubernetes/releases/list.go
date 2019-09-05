package releases

import (
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

	"github.com/trivigy/migrate/v2/nub"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// List represents the cluster release list command object.
type List struct {
	common
	Namespace string          `json:"namespace" yaml:"namespace"`
	Releases  *types.Releases `json:"releases" yaml:"releases"`
	Driver    interface {
		types.KubeConfiged
	} `json:"driver" yaml:"driver"`
}

// ListOptions is used for executing the Run() command.
type ListOptions struct {
	Env string `json:"env" yaml:"env"`
}

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r List) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "List registered releases with states information.",
		Long:  "List registered releases with states information",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return err
			}

			opts := ListOptions{Env: env}
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
func (r List) validation(args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the create command.
func (r List) Run(out io.Writer, opts ListOptions) error {
	kubectl, err := r.GetKubeCtl(r.Driver, r.Namespace)
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

				tbl, buf := r.NewEmbeddedTable()
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
					Pods(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := r.NewEmbeddedTable()
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
					ServiceAccounts(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := r.NewEmbeddedTable()
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
					ConfigMaps(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := r.NewEmbeddedTable()
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
					Endpoints(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := r.NewEmbeddedTable()
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
					Services(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := r.NewEmbeddedTable()
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
					Secrets(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := r.NewEmbeddedTable()
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
					Roles(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := r.NewEmbeddedTable()
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
					RoleBindings(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := r.NewEmbeddedTable()
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

				tbl, buf := r.NewEmbeddedTable()
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

				tbl, buf := r.NewEmbeddedTable()
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

				tbl, buf := r.NewEmbeddedTable()
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
					DaemonSets(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := r.NewEmbeddedTable()
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
					Deployments(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := r.NewEmbeddedTable()
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
					Ingresses(r.FallBackNS(manifest.Namespace, r.Namespace)).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				tbl, buf := r.NewEmbeddedTable()
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
