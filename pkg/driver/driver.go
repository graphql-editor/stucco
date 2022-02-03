/*Package driver is an interface that must be implemented by concrete driver implementations
of runners.
*/
package driver

// Driver is an interface that must be defined by an implementation
// for of specific runner.
type Driver interface {
	// Authorize runs a custom auth code on function
	Authorize(AuthorizeInput) AuthorizeOutput
	// SetSecrets defined by user, it's runner's responsibility to pass them
	// to runtime.
	SetSecrets(SetSecretsInput) SetSecretsOutput
	// FieldResolve requests an execution of defined resolver for a field
	FieldResolve(FieldResolveInput) FieldResolveOutput
	// InterfaceResolveType requests an execution of defined interface function for a type
	InterfaceResolveType(InterfaceResolveTypeInput) InterfaceResolveTypeOutput
	// ScalarParse requests an execution of defined parse function for a scalar
	ScalarParse(ScalarParseInput) ScalarParseOutput
	// ScalarSerialize requests an execution of defined serialize function for a scalar
	ScalarSerialize(ScalarSerializeInput) ScalarSerializeOutput
	// UnionResolveType requests an execution of defined union function for a type
	UnionResolveType(UnionResolveTypeInput) UnionResolveTypeOutput
	// Stream begins streaming data between router and runner.
	Stream(StreamInput) StreamOutput
	// SubscriptionConnection creates connection payload for subscription
	SubscriptionConnection(SubscriptionConnectionInput) SubscriptionConnectionOutput
	// SubscriptionListen creates connection payload for subscription
	SubscriptionListen(SubscriptionListenInput) SubscriptionListenOutput
}
