package cluster

// Operation defines a single query with config to be run during a migration.
type Operation struct {
	Query     string
	DisableTx bool
}
