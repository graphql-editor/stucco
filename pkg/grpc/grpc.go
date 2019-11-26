package grpc

import (
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
)

type GRPCClient struct {
	client proto.DriverClient
}

type GRPCServer struct {
	Impl          driver.Driver
	Handler       func(*proto.StreamRequest, proto.Driver_StreamServer) error
	StdoutHandler StdoutHandlerFunc
	StderrHandler StderrHandlerFunc
}
