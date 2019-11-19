// Package global implements global definitions for migrate package.
package global

const (
	// Unknown defines a value used for enum equivalency checks.
	Unknown = iota

	// UnknownStr defines a value used for returning default enum string.
	UnknownStr = "unknown"

	RefRoot = "root"
)

const (
	// DefaultEnvironment defines the name of a default environment.
	DefaultEnvironment = "development"

	// DefaultUsageTemplate defines the default Usage template used across
	// builtin commands.
	DefaultUsageTemplate = `Usage:
  {{.UseLine}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
`
)
