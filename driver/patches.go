package driver

import (
	"github.com/trivigy/migrate/v2/types"
)

type WithPatches interface {
	Patches(name string) *types.Patches
}
