package aphgrpc

// JSONAPIAllowedParams interface should be implement by all grpc-gateway services
// that supports JSON API specifications.
type JSONAPIAllowedParams interface {
	// Relationships that could be included
	AllowedInclude() []string
	// Attribute fields that are allowed
	AllowedFields() []string
}
