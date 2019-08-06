package model

// Release defines a collection of kubernetes manifests which can be released
// together as a logical unit.
type Release struct {
	Name    string `db:"name"`
	Version string `db:"version"`
	Status  string `db:"status"`
	// Values    map[string]interface{} `db:"values"`
	// Manifests []interface{}          `db:"manifests"`
}
