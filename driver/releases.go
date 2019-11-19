package driver

import (
	"github.com/trivigy/migrate/v2/types"
)

type WithReleases interface {
	Releases() *types.Releases
}
