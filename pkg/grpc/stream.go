package grpc

import (
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
	m.Response, err = valueToAny(nil, g.last.GetResponse())
	if err != nil {
		m.Error = &driver.Error{Message: err.Error()}
	} else if serr := g.last.GetError(); serr != nil {
		m.Error = &driver.Error{Message: serr.GetMsg()}
	}
	return m
}

// Close is no-op
func (g *grpcDriverStreamReader) Close() {}

// Stream TODO: client side stream requests
func (m *Client) Stream(input driver.StreamInput) (s driver.StreamOutput, err error) {
	return
}

// StreamHandler interface must be implemented by user to handle stream requests from subscriptions
type StreamHandler interface {
	// Handle handles subscription streaming requests
	Handle(*proto.StreamRequest, proto.Driver_StreamServer) error
}

// StreamHandlerFunc is a convienience wrapper implementing StreamHandler interface
type StreamHandlerFunc func(*proto.StreamRequest, proto.Driver_StreamServer) error

// Handle handles subscription streaming requests
func (f StreamHandlerFunc) Handle(s *proto.StreamRequest, ss proto.Driver_StreamServer) error {
	return f(s, ss)
}

func (m *Server) Stream(s *proto.StreamRequest, ss proto.Driver_StreamServer) error {
	return m.StreamHandler.Handle(s, ss)
}
