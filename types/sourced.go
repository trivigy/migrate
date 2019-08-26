package types

// Sourced represents a driver that is able to return its connection string.
type Sourced interface {
	Source() (string, error)
}
