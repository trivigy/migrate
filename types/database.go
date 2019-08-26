package types

// Database defines the interface for a database driver.
type Database interface {
	Driver
	Sourced
}
