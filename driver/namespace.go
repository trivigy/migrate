package driver

// WithNamespace represents a method interface for extracting the namespace on
// which the driver operates. This is likely to be used by a kubernetes driver.
type WithNamespace interface {
	Namespace() *string
}
