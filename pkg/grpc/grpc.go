package grpc

import (
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
)

// Client for github.com/graphql-editor/stucco/pkg/proto
type Client struct {
	Client proto.DriverClient
}

// StdoutHandler interface that must be implemented by user for handling
// stdout bytestream requests by server.
type StdoutHandler interface {
	Handle(*proto.ByteStreamRequest, proto.Driver_StdoutServer) error
}

// StderrHandler interface that must be implemented by user for handling
// stderr bytestream requests by server.
type StderrHandler interface {
	Handle(*proto.ByteStreamRequest, proto.Driver_StderrServer) error
}

// SubscriptionListenEmitter is returned to user to be called each time new subscription should be triggered.
type SubscriptionListenEmitter interface {
	// Emit new subscription event
	Emit() error
	// Close emitter
	Close() error
}

// SubscriptionListenHandler interface that must be implemented by user for handling
// subscription listen handler.
type SubscriptionListenHandler interface {
	Handle(driver.SubscriptionListenInput, SubscriptionListenEmitter) error
}

// Server for github.com/graphql-editor/stucco/pkg/proto
type Server struct {
	FieldResolveHandler           FieldResolveHandler
	InterfaceResolveTypeHandler   InterfaceResolveTypeHandler
	ScalarParseHandler            ScalarParseHandler
	ScalarSerializeHandler        ScalarSerializeHandler
	UnionResolveTypeHandler       UnionResolveTypeHandler
	SetSecretsHandler             SetSecretsHandler
	StreamHandler                 StreamHandler
	StdoutHandler                 StdoutHandler
	StderrHandler                 StderrHandler
	SubscriptionConnectionHandler SubscriptionConnectionHandler
	SubscriptionListenHandler     SubscriptionListenHandler
}
