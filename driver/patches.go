package driver

import (
	"github.com/trivigy/migrate/v2/types"
)

// WithPatches represents an interface for extracting patches from a driver.
// Patches are pre-execution hooks for dynamic driver configuration.
type WithPatches interface {
	Patches(name string) *types.Patches
}
