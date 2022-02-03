package plugin

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/grpc"
	protoDriverService "github.com/graphql-editor/stucco_proto/go/driver_service"
	"github.com/hashicorp/go-plugin"
	googlegrpc "google.golang.org/grpc"
)

// GRPC implement GRPCPlugin interface fro go-plugin
type GRPC struct {
	plugin.Plugin
	Authorize                     grpc.AuthorizeHandler
	FieldResolveHandler           grpc.FieldResolveHandler
	InterfaceResolveTypeHandler   grpc.InterfaceResolveTypeHandler
	ScalarParseHandler            grpc.ScalarParseHandler
	ScalarSerializeHandler        grpc.ScalarSerializeHandler
	UnionResolveTypeHandler       grpc.UnionResolveTypeHandler
	SetSecretsHandler             grpc.SetSecretsHandler
	StreamHandler                 grpc.StreamHandler
	StdoutHandler                 grpc.StdoutHandler
	StderrHandler                 grpc.StderrHandler
	SubscriptionConnectionHandler grpc.SubscriptionConnectionHandler
	SubscriptionListenHandler     grpc.SubscriptionListenHandler
}

// RegisterDriverServer registers an concrete implementation of a grpc server for a protocol.
var RegisterDriverServer = protoDriverService.RegisterDriverServer

// GRPCServer returns a server implementation for go-plugin
func (g *GRPC) GRPCServer(broker *plugin.GRPCBroker, s *googlegrpc.Server) error {
	RegisterDriverServer(s, &grpc.Server{
		AuthorizeHandler:              g.Authorize,
		FieldResolveHandler:           g.FieldResolveHandler,
		InterfaceResolveTypeHandler:   g.InterfaceResolveTypeHandler,
		ScalarParseHandler:            g.ScalarParseHandler,
		ScalarSerializeHandler:        g.ScalarSerializeHandler,
		UnionResolveTypeHandler:       g.UnionResolveTypeHandler,
		StreamHandler:                 g.StreamHandler,
		StdoutHandler:                 g.StdoutHandler,
		StderrHandler:                 g.StderrHandler,
		SubscriptionConnectionHandler: g.SubscriptionConnectionHandler,
		SubscriptionListenHandler:     g.SubscriptionListenHandler,
	})
	return nil
}

// NewDriverClient creates a grpc client for protocol using connection
var NewDriverClient = protoDriverService.NewDriverClient

// GRPCClient returns a client implementation for go-plugin
func (g *GRPC) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *googlegrpc.ClientConn) (interface{}, error) {
	return &grpc.Client{Client: NewDriverClient(c)}, nil
}
