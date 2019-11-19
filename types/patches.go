package types

type Patches []*Patch

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
