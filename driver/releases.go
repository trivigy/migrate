package driver

import (
	"github.com/trivigy/migrate/v2/types"
)

// WithReleases represents the method interface for extracting releases from a
// driver. This is likely to be used by a kubernetes like driver for obtaining
// releases which contain kubernetes manifests.
type WithReleases interface {
	Releases() *types.Releases
}
