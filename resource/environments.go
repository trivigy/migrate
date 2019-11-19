package resource

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	"github.com/trivigy/migrate/v2/global"
	"github.com/trivigy/migrate/v2/require"
	"github.com/trivigy/migrate/v2/types"
)

// Environments represents a environments aggregator command.
type Environments map[string]Environment

var _ interface {
	types.Resource
	types.Command
} = new(Environments)

// NewCommand returns a new cobra.Command object.
func (r Environments) NewCommand(ctx context.Context, name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:  name[strings.LastIndex(name, ".")+1:] + " [flags] COMMAND",
		Args: require.Args(r.validation),
		RunE: func(cmd *cobra.Command, args []string) error {
			rbytes, err := ioutil.ReadAll(cmd.InOrStdin())
			if err != nil {
				return err
			}
			tail := make([]string, 0)
			_ = json.Unmarshal(rbytes, &tail)

			out := cmd.OutOrStdout()
			args = append(args, tail...)
			env, _ := cmd.Flags().GetString("env")
			return r.run(ctx, out, env, name, args)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.SetHelpFunc(r.helpFunc(cmd.HelpFunc()))
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringP(
		"env", "e", global.DefaultEnvironment,
		"Run with env `ENV` configurations.",
	)
	flags.Bool("help", false, "Show help information.")
	return cmd
}

// Execute runs the command.
func (r Environments) Execute(name string, out io.Writer, args []string) error {
	wrap := types.Executor{Name: name, Command: r}
	ctx := context.WithValue(context.Background(), global.RefRoot, wrap)
	cmd := r.NewCommand(ctx, name)
	cmd.SetOut(out)

	// horible work around because spf13/pflag does not respect flag post
	// target positioning. e.g. parent --help command --help will mean that
	// command is calling help and not parent.
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.String("env", global.DefaultEnvironment, "")
	flags.String("e", global.DefaultEnvironment, "")
	flags.Bool("help", false, "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	tail := flags.Args()
	if len(tail) > 0 {
		tail = tail[1:]
	}
	rbytes, err := json.Marshal(tail)
	if err != nil {
		return err
	}
	cmd.SetIn(bytes.NewBuffer(rbytes))
	i := len(flags.Args())
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
func (r Environments) validation(cmd *cobra.Command, args []string) error {
	if err := require.ExactArgs(args, 1); err != nil {
		r.helpFunc(nil)(cmd, args)
		return err
	}
	return nil
}

func (r Environments) run(ctx context.Context, out io.Writer, env, name string, args []string) error {
	cmd := r[env].NewCommand(ctx, name)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

func (r Environments) helpFunc(helpFunc func(*cobra.Command, []string)) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		cmd.SetUsageTemplate(global.DefaultUsageTemplate)

		// This is just for generating documentation therefore context does not
		// matter here, therefore context.Background() is good enough.
		env, _ := cmd.Flags().GetString("env")
		tmp := r[env].NewCommand(context.Background(), env)
		cmd.AddCommand(tmp.Commands()...)

		if helpFunc != nil {
			helpFunc(cmd, args)
		}
	}
}
