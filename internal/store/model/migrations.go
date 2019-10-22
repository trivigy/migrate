// Package model implements database access objects for the store operations.
package model

import (
	"github.com/blang/semver"
)

// Migrations represents multiple dto migrations.
type Migrations []Migration

// Len returns length of migrations collection
func (s Migrations) Len() int {
	return len(s)
}

// Swap swaps two migrations inside the collection by its indices
func (s Migrations) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less checks if migration at index i is less than migration at index j
func (s Migrations) Less(i, j int) bool {
	iTag, err := semver.Make(s[i].Tag)
	if err != nil {
		panic(err)
	}

	jTag, err := semver.Make(s[j].Tag)
	if err != nil {
		panic(err)
	}
	return iTag.LT(jTag)
}
