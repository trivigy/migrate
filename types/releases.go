package types

// Releases represents a collection of kubernetes releases.
type Releases []*Release

// Len returns length of releases collection
func (r Releases) Len() int {
	return len(r)
}

// Swap swaps two releases inside the collection by its indices
func (r Releases) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// Less checks if release at index i is less than release at index j
func (r Releases) Less(i, j int) bool {
	return r[i].Name < r[j].Name || (r[i].Name == r[j].Name && r[i].Version.LT(r[j].Version))
}
