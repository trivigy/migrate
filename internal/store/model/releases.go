package model

import (
	"github.com/blang/semver"
)

// Releases represents a collection of kubernetes releases.
type Releases []Release

// Len returns length of releases collection
func (s Releases) Len() int {
	return len(s)
}

// Swap swaps two releases inside the collection by its indices
func (s Releases) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less checks if release at index i is less than release at index j
func (s Releases) Less(i, j int) bool {
	iVersion, err := semver.Make(s[i].Version)
	if err != nil {
		panic(err)
	}

	jVersion, err := semver.Make(s[j].Version)
	if err != nil {
		panic(err)
	}

	return s[i].Name < s[j].Name && iVersion.LT(jVersion)
}
