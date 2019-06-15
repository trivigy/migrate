package migrate

// DatabaseConfig defines database configuration to use for database connection.
type DatabaseConfig struct {
	Driver string `json:"driver" yaml:"driver"`
	Source string `json:"source" yaml:"source"`
}
