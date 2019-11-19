package types

// Migrations represents multiple database migrations.
type Migrations []*Migration

// Len returns length of migrations collection
func (r Migrations) Len() int {
	return len(r)
}

// Swap swaps two migrations inside the collection by its indices
func (r Migrations) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// Less checks if migration at index i is less than migration at index j
func (r Migrations) Less(i, j int) bool {
	return r[i].Tag.LT(r[j].Tag)
}
