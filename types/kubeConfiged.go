package types

import (
	"k8s.io/client-go/rest"
)

// KubeConfiged represents a driver that configures a kubernetes cluster and is
// able to return a kubeconfig.
type KubeConfiged interface {
	KubeConfig() (*rest.Config, error)
}
