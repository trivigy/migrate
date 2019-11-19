package releases

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
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

// Describe represents the cluster release describe command object.
type Describe struct {
	Driver interface {
		driver.WithNamespace
		driver.WithReleases
		driver.WithSource
	} `json:"driver" yaml:"driver"`
}

// describeOptions is used for executing the run() command.
type describeOptions struct {
	Name     string         `json:"name" yaml:"name"`
	Version  semver.Version `json:"version" yaml:"version"`
	Resource string         `json:"resource" yaml:"resource"`
	Filter   string         `json:"filter" yaml:"filter"`
}

var _ interface {
	types.Resource
	types.Command
} = new(Describe)

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r Describe) NewCommand(ctx context.Context, name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name + " [NAME[:VERSION]] [RESOURCE] [FILTER]",
		Short: "Describe registered releases with states information.",
		Long:  "Describe registered releases with states information",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			var filter string
			if len(args) > 2 {
				filter = args[2]
			}

			opts := describeOptions{
				Name:     name,
				Version:  version,
				Resource: resource,
				Filter:   filter,
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
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Describe) Execute(name string, out io.Writer, args []string) error {
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
func (r Describe) validation(cmd *cobra.Command, args []string) error {
	if err := require.MaxArgs(args, 3); err != nil {
		return err
	}
	return nil
}

// run is a starting point method for executing the create command.
func (r Describe) run(ctx context.Context, out io.Writer, opts describeOptions) error {
	kubectl, err := GetK8sClientset(ctx, r.Driver, *r.Driver.Namespace())
	if err != nil {
		return err
	}

	var render string
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Version", "Kind", "Results"})
	table.SetAutoWrapText(false)

	sort.Sort(*r.Driver.Releases())
	for _, rel := range *r.Driver.Releases() {
		if opts.Name != "" && rel.Name != opts.Name ||
			(!opts.Version.EQ(semver.Version{}) &&
				!rel.Version.Equals(opts.Version)) {
			continue
		}

		tables := map[string]*EmbeddedTable{}
		for _, manifest := range rel.Manifests {
			if m, ok := manifest.(runtime.Object); ok && opts.Resource != "" {
				resource := m.GetObjectKind().GroupVersionKind().Kind
				if strings.ToLower(resource) != strings.ToLower(opts.Resource) {
					continue
				}
			}

			switch manifest := manifest.(type) {
			case *v1core.Namespace:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAME", "STATUS", "AGE"})
					tables[manifest.Kind] = tbl
				}

				result, err := kubectl.CoreV1().
					Namespaces().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Table.Append([]string{
					"✓",
					result.ObjectMeta.Name,
					string(result.Status.Phase),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
			case *v1core.Pod:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAMESPACE", "NAME", "READY", "STATUS", "RESTARTS", "AGE"})
					tables[manifest.Kind] = tbl
				}

				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				result, err := kubectl.CoreV1().
					Pods(namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", namespace, manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

				readyCount := 0
				restartCount := 0
				for _, stat := range result.Status.ContainerStatuses {
					restartCount += int(stat.RestartCount)
					if stat.Ready {
						readyCount++
					}
				}

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Table.Append([]string{
					"✓",
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					fmt.Sprintf("%d/%d", readyCount, len(result.Status.ContainerStatuses)),
					string(result.Status.Phase),
					fmt.Sprintf("%d", restartCount),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
			case *v1core.ServiceAccount:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAMESPACE", "NAME", "SECRETS", "AGE"})
					tables[manifest.Kind] = tbl
				}

				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				result, err := kubectl.CoreV1().
					ServiceAccounts(namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", namespace, manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Table.Append([]string{
					"✓",
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					fmt.Sprintf("%d", len(result.Secrets)),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
			case *v1core.ConfigMap:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAMESPACE", "NAME", "DATA", "AGE"})
					tables[manifest.Kind] = tbl
				}

				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				result, err := kubectl.CoreV1().
					ConfigMaps(namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", namespace, manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

				var rbytes []byte
				for _, data := range result.Data {
					rbytes = append(rbytes, []byte(data)...)
				}
				for _, binary := range result.BinaryData {
					rbytes = append(rbytes, binary...)
				}

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Table.Append([]string{
					"✓",
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					fmt.Sprintf("%dB", len(rbytes)),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
			case *v1core.Endpoints:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAMESPACE", "NAME", "ENDPOINTS", "AGE"})
					tables[manifest.Kind] = tbl
				}

				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				result, err := kubectl.CoreV1().
					Endpoints(namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", namespace, manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

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
				tbl.Table.Append([]string{
					"✓",
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					strings.Join(endpoints, ","),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
			case *v1core.Service:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAMESPACE", "NAME", "TYPE", "CLUSTER-IP", "EXTERNAL-IP", "PORT(s)", "AGE"})
					tables[manifest.Kind] = tbl
				}

				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				result, err := kubectl.CoreV1().
					Services(namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", namespace, manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

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
				tbl.Table.Append([]string{
					"✓",
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					string(result.Spec.Type),
					result.Spec.ClusterIP,
					externalIP,
					strings.Join(ports, ","),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
			case *v1core.Secret:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAMESPACE", "NAME", "TYPE", "DATA", "AGE"})
					tables[manifest.Kind] = tbl
				}

				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				result, err := kubectl.CoreV1().
					Secrets(namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", namespace, manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

				var rbytes []byte
				for _, data := range result.StringData {
					rbytes = append(rbytes, []byte(data)...)
				}
				for _, binary := range result.Data {
					rbytes = append(rbytes, binary...)
				}

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Table.Append([]string{
					"✓",
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					string(result.Type),
					fmt.Sprintf("%dB", len(rbytes)),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
			case *v1rbac.Role:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAMESPACE", "NAME", "AGE"})
					tables[manifest.Kind] = tbl
				}

				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				result, err := kubectl.RbacV1().
					Roles(namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", namespace, manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Table.Append([]string{
					"✓",
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
			case *v1rbac.RoleBinding:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAMESPACE", "NAME", "AGE"})
					tables[manifest.Kind] = tbl
				}

				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				result, err := kubectl.RbacV1().
					RoleBindings(namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", namespace, manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Table.Append([]string{
					"✓",
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
			case *v1rbac.ClusterRole:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAME", "AGE"})
					tables[manifest.Kind] = tbl
				}

				result, err := kubectl.RbacV1().
					ClusterRoles().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Table.Append([]string{
					"✓",
					result.ObjectMeta.Name,
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
			case *v1rbac.ClusterRoleBinding:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAME", "AGE"})
					tables[manifest.Kind] = tbl
				}

				result, err := kubectl.RbacV1().
					ClusterRoleBindings().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Table.Append([]string{
					"✓",
					result.ObjectMeta.Name,
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
			case *v1policy.PodSecurityPolicy:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAME", "PRIV", "CAPS", "VOLUMES"})
					tables[manifest.Kind] = tbl
				}

				result, err := kubectl.PolicyV1beta1().
					PodSecurityPolicies().
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

				var caps []string
				for _, capability := range result.Spec.AllowedCapabilities {
					caps = append(caps, string(capability))
				}

				var volumes []string
				for _, volume := range result.Spec.Volumes {
					volumes = append(volumes, string(volume))
				}

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Table.Append([]string{
					"✓",
					result.ObjectMeta.Name,
					strconv.FormatBool(result.Spec.Privileged),
					strings.Join(caps, ","),
					strings.Join(volumes, ","),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
			case *v1apps.DaemonSet:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAMESPACE", "NAME", "DESIRED", "CURRENT", "READY", "UP-TO-DATE", "AVAILABLE", "AGE"})
					tables[manifest.Kind] = tbl
				}

				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				result, err := kubectl.AppsV1().
					DaemonSets(namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", namespace, manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Table.Append([]string{
					"✓",
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					fmt.Sprintf("%d", result.Status.DesiredNumberScheduled),
					fmt.Sprintf("%d", result.Status.CurrentNumberScheduled),
					fmt.Sprintf("%d", result.Status.NumberReady),
					fmt.Sprintf("%d", result.Status.UpdatedNumberScheduled),
					fmt.Sprintf("%d", result.Status.NumberAvailable),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
			case *v1apps.Deployment:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAMESPACE", "NAME", "READY", "UP-TO-DATE", "AVAILABLE", "AGE"})
					tables[manifest.Kind] = tbl
				}

				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				result, err := kubectl.AppsV1().
					Deployments(namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", namespace, manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				tbl.Table.Append([]string{
					"✓",
					result.ObjectMeta.Namespace,
					result.ObjectMeta.Name,
					fmt.Sprintf("%d/%d", result.Status.ReadyReplicas, result.Status.Replicas),
					fmt.Sprintf("%d", result.Status.UpdatedReplicas),
					fmt.Sprintf("%d", result.Status.AvailableReplicas),
					fmt.Sprintf("%.2fm", diffTime.Minutes()),
				})
			case *v1ext.Ingress:
				tbl, ok := tables[manifest.Kind]
				if !ok {
					tbl = NewEmbeddedTable()
					tbl.Table.SetHeader([]string{"", "NAMESPACE", "NAME", "HOST", "PATH", "PORT", "AGE"})
					tables[manifest.Kind] = tbl
				}

				namespace := FallBackNS(manifest.Namespace, *r.Driver.Namespace())
				result, err := kubectl.ExtensionsV1beta1().
					Ingresses(namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					tbl.Table.Append([]string{"✗", namespace, manifest.Name, ""})
					break
				} else if err != nil {
					return err
				}
				tbl.Results = append(tbl.Results, result)

				diffTime := time.Since(result.ObjectMeta.CreationTimestamp.Time)
				name := result.ObjectMeta.Name
				age := fmt.Sprintf("%.2fm", diffTime.Minutes())
				check := "✓"
				for i := range result.Spec.Rules {
					if i != 0 {
						check = ""
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
						tbl.Table.Append([]string{check, namespace, name, host, path, port, age})
					}
				}
			default:
				return fmt.Errorf("unsupported manifest type %T", manifest)
			}
		}

		var j int
		for kind, tbl := range tables {
			if opts.Filter != "" {
				render, err = FilterResult(map[string]interface{}{
					"name":    rel.Name,
					"version": rel.Version,
					"kind":    kind,
					"results": tbl.Results,
				}, opts.Filter)
				if err != nil {
					return err
				}
			} else {
				tbl.Table.Render()
				lines := strings.Split(tbl.Buffer.String(), "\n")
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
				j++
			}
		}
	}

	if len(*r.Driver.Releases()) > 0 {
		if opts.Filter != "" {
			fmt.Fprintf(out, "%s", render)
		} else {
			table.Render()
		}

	}

	return nil
}
