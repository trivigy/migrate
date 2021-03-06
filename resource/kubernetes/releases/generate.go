package releases

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/blang/semver"
	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/driver"
	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

const templateContent = `
package migrations

import (
	"github.com/blang/semver"
	"github.com/trivigy/migrate"
	"github.com/trivigy/migrate/database"
)

func init() {
	tag := "{{ .Tag }}"
	name := "{{ .Name }}"
	migrate.Registry.Store(database.Path(name, tag), database.Migration{
		Name: name,		
		Tag: semver.MustParse(tag),
		Up: []database.Operation{},
		Down: []database.Operation{},
	})
}
`

// Generate represents the generate command which allows for generating new
// templates of the cluster migrations file.
type Generate struct {
	Driver interface {
		driver.WithReleases
		driver.WithSource
	} `json:"driver" yaml:"driver"`
}

// GenerateOptions is used for executing the run() method.
type GenerateOptions struct {
	Dir  string `json:"dir" yaml:"dir"`
	Name string `json:"name" yaml:"name"`
	Tag  string `json:"tag" yaml:"tag"`
}

var _ interface {
	types.Resource
	types.Command
} = new(Generate)

// NewCommand returns a new cobra.Command generate command object.
func (r Generate) NewCommand(ctx context.Context, name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name[strings.LastIndex(name, ".")+1:] + " NAME[:TAG]",
		Short: "Adds a new release template.",
		Long:  "Adds a new release template",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cmd.Flags().GetString("dir")
			if err != nil {
				return err
			}

			parts := strings.Split(args[0], ":")
			parts = append(parts, "")
			name, tag := parts[0], parts[1]

			opts := GenerateOptions{Dir: dir, Name: name, Tag: tag}
			return r.run(cmd.OutOrStdout(), opts)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.SetUsageTemplate(global.DefaultUsageTemplate)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringP(
		"dir", "d", ".",
		"Specify directory `PATH` where to generate miration file.",
	)
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Generate) Execute(name string, out io.Writer, args []string) error {
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
func (r Generate) validation(cmd *cobra.Command, args []string) error {
	if err := require.ExactArgs(args, 1); err != nil {
		return err
	}

	parts := strings.Split(args[0], ":")
	if len(parts) > 2 {
		return fmt.Errorf("invalid argument %q", args[0])
	}

	// padding with empty string to check if tag is present.
	parts = append(parts, "")
	if parts[1] != "" {
		v, err := semver.Make(parts[1])
		if err != nil || len(v.Build) > 0 {
			return fmt.Errorf("invalid argument %q", args[0])
		}
	}
	return nil
}

// run is a starting point method for executing the generate command.
func (r Generate) run(out io.Writer, opts GenerateOptions) error {
	base, err := filepath.Abs(opts.Dir)
	if err != nil {
		return err
	}

	if fi, err := os.Stat(base); os.IsNotExist(err) || !fi.IsDir() {
		return fmt.Errorf("directory %q not found", opts.Dir)
	}

	sort.Sort(*r.Driver.Releases())
	tags := semver.Versions{semver.Version{}}
	for _, rgMig := range *r.Driver.Releases() {
		tags = append(tags, rgMig.Version)
	}

	if opts.Tag == "" {
		v := tags[len(tags)-1]
		v.Patch++
		opts.Tag = v.String()
	}

	for _, v := range tags {
		if opts.Tag == v.String() {
			return fmt.Errorf("release tag %q exists", opts.Tag)
		}
	}

	filename := fmt.Sprintf("%s_%s.go", opts.Tag, opts.Name)
	fullpath := path.Join(base, filename)
	file, err := os.Create(fullpath)
	if err != nil {
		return err
	}
	defer file.Close()

	tpl := template.Must(template.New("releases").Parse(templateContent))
	if err := tpl.Execute(file, opts); err != nil {
		return err
	}

	fmt.Fprintf(out, "Created release %q\n", fullpath)
	return nil
}
