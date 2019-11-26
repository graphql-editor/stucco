package driver

// Driver is an interface that must be defined by an implementation
// for of specific runner.
type Driver interface {
	// SetSecrets defined by user, it's runner's responsibility to pass them
	// to runtime.
	SetSecrets(SetSecretsInput) (SetSecretsOutput, error)
	// FieldResolve requests an execution of defined resolver for a field
	FieldResolve(FieldResolveInput) (FieldResolveOutput, error)
	// InterfaceResolveType requests an execution of defined interface function for a type
	InterfaceResolveType(InterfaceResolveTypeInput) (InterfaceResolveTypeOutput, error)
	// ScalarParse requests an execution of defined parse function for a scalar
	ScalarParse(ScalarParseInput) (ScalarParseOutput, error)
	// ScalarSerialize requests an execution of defined serialize function for a scalar
	ScalarSerialize(ScalarSerializeInput) (ScalarSerializeOutput, error)
	// UnionResolveType requests an execution of defined union function for a type
	UnionResolveType(UnionResolveTypeInput) (UnionResolveTypeOutput, error)
	// Stream begins streaming data between router and runner.
	Stream(StreamInput) (StreamOutput, error)
}
