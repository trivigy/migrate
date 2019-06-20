package migrate

// DataSource defines database configuration to use for database connection.
type DataSource struct {
	Driver string `json:"driver" yaml:"driver"`
	Source string `json:"source" yaml:"source"`
}
