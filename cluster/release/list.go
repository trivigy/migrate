package release

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
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

// List represents the cluster release list command object.
type List struct {
	common
	config map[string]config.Cluster
}

// NewList initializes a new cluster create command.
func NewList(config map[string]config.Cluster) types.Command {
	return &List{config: config}
}

// ListOptions is used for executing the Run() command.
type ListOptions struct {
	Env string `json:"env" yaml:"env"`
}

// NewCommand creates a new cobra.Command, configures it and returns it.
func (r *List) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "List registered releases with states information.",
		Long:  "List registered releases with states information",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return errors.WithStack(err)
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
func (r *List) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r *List) validation(args []string) error {
	if err := require.NoArgs(args); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the create command.
func (r *List) Run(out io.Writer, opts ListOptions) error {
	cfg, ok := r.config[opts.Env]
	if !ok {
		return fmt.Errorf("missing %q environment configuration", opts.Env)
	}

	kubectl, err := r.GetKubeCtl(&cfg)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"#", "Name", "Version", "Kind", "Status"})
	table.SetAutoWrapText(false)

	sort.Sort(*cfg.Releases)
	for i, rel := range *cfg.Releases {
		for j, manifest := range rel.Manifests {
			var kind string
			var status string
			switch manifest := manifest.(type) {
			case *v1core.ConfigMap:
				kind = manifest.Kind
				result, err := kubectl.CoreV1().
					ConfigMaps(cfg.Namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				status, err = r.TrimmedYAML(result)
				if err != nil {
					return err
				}
			case *v1core.Endpoints:
				kind = manifest.Kind
				result, err := kubectl.CoreV1().
					Endpoints(cfg.Namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				status, err = r.TrimmedYAML(result)
				if err != nil {
					return err
				}
			case *v1core.Service:
				kind = manifest.Kind
				result, err := kubectl.CoreV1().
					Services(cfg.Namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				status, err = r.TrimmedYAML(result)
				if err != nil {
					return err
				}
			case *v1apps.Deployment:
				kind = manifest.Kind
				result, err := kubectl.AppsV1().
					Deployments(cfg.Namespace).
					Get(manifest.Name, v1meta.GetOptions{})
				if v1err.IsNotFound(err) {
					status = string(v1meta.StatusReasonNotFound)
					break
				} else if err != nil {
					return err
				}

				status, err = r.TrimmedYAML(result)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported manifest type %T", manifest)
			}

			// Chunk values greater than 64 chars and adds them back
			// as extra status lines. The chunks are aligned under the original
			// value start index. Like a hanging styled text.
			chunkSize := 80
			lines := strings.Split(status, "\n")
			for i, line := range lines {
				keyVal := strings.SplitN(line, ":", 2)
				if len(keyVal) < 2 {
					continue
				}
				value := strings.TrimSpace(keyVal[1])
				if len(value) > chunkSize {
					chunks := r.ChunkString(value, chunkSize)
					lines[i] = keyVal[0] + ": " + chunks[0]
					for x := len(chunks[1:]) - 1; x >= 0; x-- {
						lines = append(lines[:i+1], append([]string{
							fmt.Sprintf("%s%s",
								strings.Repeat(" ", len(keyVal[0])+2),
								chunks[1:][x],
							),
						}, lines[i+1:]...)...)
					}
				}
			}

			// Append the styled lines to the table.
			if j == 0 {
				for k, line := range lines {
					if k == 0 {
						table.Append([]string{
							strconv.Itoa(i + 1),
							rel.Name,
							rel.Version.String(),
							kind,
							line,
						})
					} else {
						table.Append([]string{"", "", "", "", line})
					}
				}
			} else {
				for k, line := range lines {
					if k == 0 {
						table.Append([]string{"", "", "", kind, line})
					} else {
						table.Append([]string{"", "", "", "", line})
					}
				}
			}
		}
	}

	if len(*cfg.Releases) > 0 {
		table.Render()
	}

	return nil
}
