package migrate

// Migration defines a set of operations to run on the database.
type Migration struct {
	Tag  string      `json:"tag" yaml:"tag"`
	Up   []Operation `json:"up" yaml:"up"`
	Down []Operation `json:"down" yaml:"down"`
}
