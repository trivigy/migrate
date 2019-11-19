package types

import (
	"context"

	"github.com/spf13/cobra"
)

// Resource represents a command object that is able to generate commands.
type Resource interface {
	NewCommand(ctx context.Context, fqn string) *cobra.Command
}
