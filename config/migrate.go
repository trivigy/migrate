package config

// Migrate represents a configurations object of a migrate command.
type Migrate struct {
	Cluster  Cluster  `json:"cluster" yaml:"cluster"`
	Database Database `json:"database" yaml:"database"`
}
