package types

import (
	"github.com/blang/semver"
)

// Migration defines a set of operations to run on the database.
type Migration struct {
	Name string         `json:"name,omitempty" yaml:"name,omitempty"`
	Tag  semver.Version `json:"tag,omitempty" yaml:"tag,omitempty"`
	Up   []Operation    `json:"up,omitempty" yaml:"up,omitempty"`
	Down []Operation    `json:"down,omitempty" yaml:"down,omitempty"`
}
