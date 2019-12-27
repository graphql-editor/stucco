package grpc

import (
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
)

// Stream TODO: client side stream requests
func (m *Client) Stream(input driver.StreamInput) (s driver.StreamOutput) {
	return driver.StreamOutput{
		Error: &driver.Error{
			Message: "Streaming not yet implemented",
		},
	}
}

// StreamHandler interface must be implemented by user to handle stream requests from subscriptions
type StreamHandler interface {
	// Handle handles subscription streaming requests
	Handle(*proto.StreamRequest, proto.Driver_StreamServer) error
}

// StreamHandlerFunc is a convienience wrapper implementing StreamHandler interface
type StreamHandlerFunc func(*proto.StreamRequest, proto.Driver_StreamServer) error

// Handle implements StreamHandler.Handle method
func (f StreamHandlerFunc) Handle(s *proto.StreamRequest, ss proto.Driver_StreamServer) error {
	return f(s, ss)
}

// Stream hands over stream request to user defined handler for stream
func (m *Server) Stream(s *proto.StreamRequest, ss proto.Driver_StreamServer) error {
	return m.StreamHandler.Handle(s, ss)
}
