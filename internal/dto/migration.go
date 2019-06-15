package dto

import (
	"github.com/blang/semver"
)

// Migration defines a set of operations to run on the database.
type Migration struct {
	Tag  semver.Version `json:"tag" yaml:"tag"`
	Up   []Operation    `json:"up" yaml:"up"`
	Down []Operation    `json:"down" yaml:"down"`
}