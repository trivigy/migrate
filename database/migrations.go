package database

// Migrations represents multiple database migrations.
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
	return s[i].Tag.LT(s[j].Tag)
}
