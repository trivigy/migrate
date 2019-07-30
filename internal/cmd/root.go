package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type rootCommand struct {
	cobra.Command
}

func newRootCommand() *rootCommand {
	cmd := &rootCommand{}
	cmd.Version = "1.0.4"
	cmd.Use = filepath.Base(os.Args[0])
	cmd.Long = "Idiomatic GO database migration tool"

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.AddCommand(
		&create.Command,
		&down.Command,
		&status.Command,
		&up.Command,
	)

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.BoolP(
		"version", "v", false,
		"Print version information and quit.",
	)
	flags.Bool("help", false, "Show help information.")
	return cmd
}
