package types

// Patches implements a collection of patching elements each of which executes
// a pre-execution hook function.
type Patches []*Patch

// Filter is a method which allows filtering specific patches based on a
// command chaining path.
func (r *Patches) Filter(name string) *Patches {
	patches := &Patches{}
	for _, patch := range *r {
		if contains(patch.Filters, name) {
			*patches = append(*patches, patch)
		}
	}
	return patches
}

func contains(list []string, value string) bool {
	for _, each := range list {
		if value == each {
			return true
		}
	}
	return false
}
