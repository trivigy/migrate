package resource

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Environments represents a environments aggregator command.
type Environments map[string]interface {
	types.Resource
	types.Command
}

// EnvironmentsOptions is used for executing the Run() command.
type EnvironmentsOptions struct {
	Env  string   `json:"env" yaml:"env"`
	Args []string `json:"args" yaml:"args"`
	Name string   `json:"name" yaml:"name"`
}

// NewCommand returns a new cobra.Command object.
func (r Environments) NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:  name + " [flags] command",
		Args: require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			rbytes, err := ioutil.ReadAll(cmd.InOrStdin())
			if err != nil {
				return err
			}
			tail := make([]string, 0)
			_ = json.Unmarshal(rbytes, &tail)
			cmd.SetIn(os.Stdin)

			env, err := cmd.Flags().GetString("env")
			if err != nil {
				return err
			}

			opts := EnvironmentsOptions{
				Env:  env,
				Name: name,
				Args: append(args, tail...),
			}
			return r.Run(cmd.OutOrStdout(), opts)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	helpFunc := cmd.HelpFunc()
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.SetUsageTemplate(`Usage:
  {{.UseLine}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)

		env, _ := cmd.Flags().GetString("env")
		tmp := r[env].NewCommand(env)
		cmd.AddCommand(tmp.Commands()...)
		helpFunc(cmd, args)
	})

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringP(
		"env", "e", "default",
		"Run with env `ENV` configurations.",
	)
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Environments) Execute(name string, out io.Writer, args []string) error {
	cmd := r.NewCommand(name)
	cmd.SetOut(out)

	// horible work around because spf13/pflag does not respect flag post
	// target positioning. e.g. parent --help command --help will mean that
	// command is calling help and not parent.
	pflags := flag.NewFlagSet("", flag.ContinueOnError)
	pflags.String("env", "default", "")
	pflags.String("e", "default", "")
	pflags.Bool("help", false, "")
	if err := pflags.Parse(args); err != nil {
		return err
	}
	tail := pflags.Args()
	if len(tail) > 0 {
		tail = tail[1:]
	}
	rbytes, err := json.Marshal(tail)
	if err != nil {
		return err
	}
	cmd.SetIn(bytes.NewBuffer(rbytes))
	i := len(pflags.Args())
	if i >= 1 {
		i--
	}

	cmd.SetArgs(args[:len(args)-i])
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// validation represents a sequence of positional argument validation steps.
func (r Environments) validation(args []string) error {
	if err := require.ExactArgs(args, 1); err != nil {
		return err
	}
	return nil
}

// Run is a starting point method for executing the create command.
func (r Environments) Run(out io.Writer, opts EnvironmentsOptions) error {
	return r[opts.Env].Execute(opts.Name, out, opts.Args)
}
