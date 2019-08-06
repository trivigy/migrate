package types

import (
	"github.com/blang/semver"
)

// Release defines a collection of kubernetes manifests which can be released
// together as a logical unit.
type Release struct {
	Name      string         `json:"name" yaml:"name"`
	Version   semver.Version `json:"tag" yaml:"tag"`
	Values    interface{}    `json:"values" yaml:"values"`
	Manifests []interface{}  `json:"manifests" yaml:"manifests"`
}
