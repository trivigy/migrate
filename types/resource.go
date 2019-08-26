package types

import (
	"github.com/spf13/cobra"
)

// Resource represents a command object that is able to generate commands.
type Resource interface {
	NewCommand(name string) *cobra.Command
}
