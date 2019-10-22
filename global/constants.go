// Package global implements global definitions for migrate package.
package global

const (
	// Unknown defines a value used for enum equivalency checks.
	Unknown = iota

	// UnknownStr defines a value used for returning default enum string.
	UnknownStr = "unknown"

	// UnknownJSONStr similar to `UnknownStr` but assumes that json encoding
	// occurred.
	UnknownJSONStr = `"unknown"`
)

const (
	// DefaultEnvironment defines the name of a default environment.
	DefaultEnvironment = "default"
)
