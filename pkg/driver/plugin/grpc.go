package plugin

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/grpc"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/hashicorp/go-plugin"
	googlegrpc "google.golang.org/grpc"
)

// GRPC implement GRPCPlugin interface fro go-plugin
type GRPC struct {
	plugin.Plugin
	FieldResolveHandler         grpc.FieldResolveHandler
	InterfaceResolveTypeHandler grpc.InterfaceResolveTypeHandler
	ScalarParseHandler          grpc.ScalarParseHandler
	ScalarSerializeHandler      grpc.ScalarSerializeHandler
	UnionResolveTypeHandler     grpc.UnionResolveTypeHandler
	SetSecretsHandler           grpc.SetSecretsHandler
	StreamHandler               grpc.StreamHandler
	StdoutHandler               grpc.StdoutHandler
	StderrHandler               grpc.StderrHandler
}

// RegisterDriverServer registers an concrete implementation of a grpc server for a protocol.
var RegisterDriverServer = proto.RegisterDriverServer

// GRPCServer returns a server implementation for go-plugin
func (g *GRPC) GRPCServer(broker *plugin.GRPCBroker, s *googlegrpc.Server) error {
	RegisterDriverServer(s, &grpc.Server{
		FieldResolveHandler:         g.FieldResolveHandler,
		InterfaceResolveTypeHandler: g.InterfaceResolveTypeHandler,
		ScalarParseHandler:          g.ScalarParseHandler,
		ScalarSerializeHandler:      g.ScalarSerializeHandler,
		UnionResolveTypeHandler:     g.UnionResolveTypeHandler,
		StreamHandler:               g.StreamHandler,
		StdoutHandler:               g.StdoutHandler,
		StderrHandler:               g.StderrHandler,
	})
	return nil
}

// NewDriverClient creates a grpc client for protocol using connection
var NewDriverClient = proto.NewDriverClient

// GRPCClient returns a client implementation for go-plugin
func (g *GRPC) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *googlegrpc.ClientConn) (interface{}, error) {
	return &grpc.Client{Client: NewDriverClient(c)}, nil
}
