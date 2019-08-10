package types

import (
	"fmt"

	"github.com/blang/semver"
)

// Release defines a collection of kubernetes manifests which can be released
// together as a logical unit.
type Release struct {
	Name      string         `json:"name" yaml:"name"`
	Version   semver.Version `json:"tag" yaml:"tag"`
	Manifests []interface{}  `json:"manifests" yaml:"manifests"`
}

func (r Release) String() string {
	return fmt.Sprintf("%+v", []string{r.Name, r.Version.String()})
}
