package grpc

import (
	"github.com/graphql-editor/stucco/pkg/driver"
	protoDriverService "github.com/graphql-editor/stucco_proto/go/driver_service"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
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
	Handle(*protoMessages.StreamRequest, protoDriverService.Driver_StreamServer) error
}

// StreamHandlerFunc is a convienience wrapper implementing StreamHandler interface
type StreamHandlerFunc func(*protoMessages.StreamRequest, protoDriverService.Driver_StreamServer) error

// Handle implements StreamHandler.Handle method
func (f StreamHandlerFunc) Handle(s *protoMessages.StreamRequest, ss protoDriverService.Driver_StreamServer) error {
	return f(s, ss)
}

// Stream hands over stream request to user defined handler for stream
func (m *Server) Stream(s *protoMessages.StreamRequest, ss protoDriverService.Driver_StreamServer) error {
	return m.StreamHandler.Handle(s, ss)
}
