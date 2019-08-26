package types

import (
	"k8s.io/client-go/rest"
)

// Driver represents an interface to an abstruct automation driver.
type Driver interface {
	Creator
	Destroyer
}

// Cluster defines the interface for a cluster driver.
type Cluster interface {
	Driver
	KubeConfig() (*rest.Config, error)
}
