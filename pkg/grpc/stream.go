package grpc

import (
	"errors"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
)

type grpcDriverStreamReader struct {
	d    proto.Driver_StreamClient
	err  error
	last *proto.StreamMessage
}

func (g *grpcDriverStreamReader) Error() error {
	return g.err
}

func (g *grpcDriverStreamReader) Next() bool {
	m, err := g.d.Recv()
	if err != nil {
		g.err = err
		return false
	}
	g.last = m
	return true
}

func (g *grpcDriverStreamReader) Read() driver.StreamMessage {
	var m driver.StreamMessage
	var err error
	m.Response, err = valueToAny(g.last.GetResponse())
	if err != nil {
		m.Error = &driver.Error{Message: err.Error()}
	} else if serr := g.last.GetError(); serr != nil {
		m.Error = &driver.Error{Message: serr.GetMsg()}
	}
	return m
}

// Close is no-op
func (g *grpcDriverStreamReader) Close() {}

func (m *GRPCClient) Stream(input driver.StreamInput) (s driver.StreamOutput, err error) {
	return
}

func (m *GRPCServer) Stream(s *proto.StreamRequest, ss proto.Driver_StreamServer) error {
	if m.Handler != nil {
		return m.Handler(s, ss)
	}
	return errors.New("GRPC plugin does not support stream requests")
}
