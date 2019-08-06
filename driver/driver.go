package driver

import (
	"io"
)

// Driver represents an interface to an abstruct automation driver.
type Driver interface {
	Setup(out io.Writer) error
	TearDown(out io.Writer) error
}

// Cluster defines the interface for a cluster driver.
type Cluster interface {
	Driver
	KubeConfig() (string, error)
}

// Database defines the interface for a database driver.
type Database interface {
	Driver
	Name() string
	Source() (string, error)
}
