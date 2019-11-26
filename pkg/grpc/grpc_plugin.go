package grpc

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type DriverGRPCPlugin struct {
	plugin.Plugin
	Impl                driver.Driver
	StreamServerHandler func(*proto.StreamRequest, proto.Driver_StreamServer) error
}

func (p *DriverGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterDriverServer(s, &GRPCServer{
		Impl:    p.Impl,
		Handler: p.StreamServerHandler,
	})
	return nil
}

func (p *DriverGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewDriverClient(c)}, nil
}
