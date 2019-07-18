package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const templateContent = `
package migrations

import (
	"github.com/trivigy/migrate"
)

func init() {
	migrate.Append(migrate.Migration{
		Tag: "0.0.3",
		Up: []migrate.Operation{},
		Down: []migrate.Operation{},
	})
}

`

type createCommand struct {
	cobra.Command
}

func newCreateCommand() *createCommand {
	cmd := &createCommand{}
	cmd.Run = cmd.run
	cmd.Args = cmd.args
	cmd.Use = "create NAME[:TAG]"
	cmd.Short = "Create a newly versioned migration template file"

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringP(
		"dir", "d", ".",
		"Specify directory `PATH` to create miration file.",
	)
	flags.Bool("help", false, "Show help information.")
	return cmd
}

func (r *createCommand) run(cmd *cobra.Command, args []string) {
	dir, _ := cmd.Flags().GetString("dir")
	parts := strings.Split(args[0], ":")
	parts = append(parts, "")
	name, tag := parts[0], parts[1]

	base, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(cmd.OutOrStderr(), "%+v\n", errors.WithStack(err))
		return
	}

	if fi, err := os.Stat(base); os.IsNotExist(err) || !fi.IsDir() {
		fmt.Printf("error: directory %q not found\n", base)
		if err := cmd.Usage(); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "%+v\n", errors.WithStack(err))
			return
		}
		fmt.Println()
		return
	}

	tags := semver.Versions{semver.Version{}}
	for _, rgMig := range getSortedRegistryMigrations() {
		tags = append(tags, rgMig.Tag)
	}

	if tag == "" {
		v := tags[len(tags)-1]
		v.Patch++
		tag = v.String()
	}

	for _, v := range tags {
		if tag == "0.0.0" {
			fmt.Printf("error: invalid migration tag %q\n", tag)
			if err := cmd.Usage(); err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "%+v\n", errors.WithStack(err))
				return
			}
			fmt.Println()
			return
		}

		if tag == v.String() {
			fmt.Printf("error: migration tag %q exists\n", tag)
			if err := cmd.Usage(); err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "%+v\n", errors.WithStack(err))
				return
			}
			fmt.Println()
			return
		}
	}

	filename := fmt.Sprintf("%s_%s.go", tag, name)
	fullpath := path.Join(base, filename)
	file, err := os.Create(fullpath)
	if err != nil {
		fmt.Fprintf(cmd.OutOrStderr(), "%+v\n", errors.WithStack(err))
		return
	}
	defer file.Close()

	tpl := template.Must(template.New("migration").Parse(templateContent))
	if err := tpl.Execute(file, nil); err != nil {
		fmt.Fprintf(cmd.OutOrStderr(), "%+v\n", errors.WithStack(err))
		return
	}

	fmt.Printf("Created migration %q\n", fullpath)
}

func (r createCommand) args(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	parts := strings.Split(args[0], ":")
	if len(parts) > 2 {
		return fmt.Errorf("invalid argument %q for NAME[:TAG]", args[0])
	}

	// makes sure at least 2 values
	parts = append(parts, "")
	if parts[1] != "" {
		v, err := semver.Make(parts[1])
		if err != nil {
			return err
		}

		if len(v.Build) > 0 {
			return fmt.Errorf("migration tag %q contains build", parts[1])
		}
	}
	return nil
}
