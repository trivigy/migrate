package migrations

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

const templateContent = `package migrations

import (
	"github.com/blang/semver"
	"github.com/trivigy/migrate/v2"
	"github.com/trivigy/migrate/v2/types"
)

func init() {
	migrate.Registry.Store(&types.Migration{
		Name: "{{ .Name }}",
		Tag:  semver.MustParse("{{ .Tag }}"),
		Up: []types.Operation{},
		Down: []types.Operation{},
	})
}

`

// Generate represents the generate command which allows for generating new
// templates of the database migrations file.
type Generate struct {
	Driver interface {
		driver.WithMigrations
		driver.WithSource
	} `json:"driver" yaml:"driver"`
}

// generateOptions is used for executing the run() method.
type generateOptions struct {
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
		Short: "Adds a new blank migration with increasing version.",
		Long:  "Adds a new blank migration with increasing version",
		Args:  require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cmd.Flags().GetString("dir")
			if err != nil {
				return err
			}

			parts := strings.Split(args[0], ":")
			parts = append(parts, "")
			name, tag := parts[0], parts[1]

			opts := generateOptions{Dir: dir, Name: name, Tag: tag}
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
func (r Generate) run(out io.Writer, opts generateOptions) error {
	base, err := filepath.Abs(opts.Dir)
	if err != nil {
		return err
	}

	if fi, err := os.Stat(base); os.IsNotExist(err) || !fi.IsDir() {
		return fmt.Errorf("directory %q not found", opts.Dir)
	}

	sort.Sort(*r.Driver.Migrations())
	tags := semver.Versions{semver.Version{}}
	for _, rgMig := range *r.Driver.Migrations() {
		tags = append(tags, rgMig.Tag)
	}

	if opts.Tag == "" {
		v := tags[len(tags)-1]
		v.Patch++
		opts.Tag = v.String()
	}

	for _, v := range tags {
		if opts.Tag == v.String() {
			return fmt.Errorf("migration tag %q exists", opts.Tag)
		}
	}

	filename := fmt.Sprintf("%s_%s.go", opts.Tag, opts.Name)
	fullpath := path.Join(base, filename)
	file, err := os.Create(fullpath)
	if err != nil {
		return err
	}
	defer file.Close()

	tpl := template.Must(template.New("migration").Parse(templateContent))
	if err := tpl.Execute(file, opts); err != nil {
		return err
	}

	fmt.Fprintf(out, "Created migration %q\n", fullpath)
	return nil
}
