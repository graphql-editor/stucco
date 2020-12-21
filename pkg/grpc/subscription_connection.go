package grpc

import (
	"context"
	"fmt"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
)

// SubscriptionConnection marshals a field resolution request through GRPC to a function
// that handles an actual resolution.
func (m *Client) SubscriptionConnection(input driver.SubscriptionConnectionInput) (f driver.SubscriptionConnectionOutput) {
	req, err := protodriver.MakeSubscriptionConnectionRequest(input)
	if err == nil {
		var resp *proto.SubscriptionConnectionResponse
		resp, err = m.Client.SubscriptionConnection(context.Background(), req)
		if err == nil {
			f = protodriver.MakeSubscriptionConnectionOutput(resp)
		}
	}
	if err != nil {
		f.Error = &driver.Error{Message: err.Error()}
	}
	return
}

// SubscriptionConnectionHandler interface implemented by user to handle subscription connection creation
type SubscriptionConnectionHandler interface {
	// Handle takes SubscriptionConnectionInput as a field resolution input and returns
	// arbitrary user response.
	Handle(input driver.SubscriptionConnectionInput) (interface{}, error)
}

// SubscriptionConnectionHandlerFunc is a convienience function wrapper implementing SubscriptionConnectionHandler
type SubscriptionConnectionHandlerFunc func(input driver.SubscriptionConnectionInput) (interface{}, error)

// Handle takes SubscriptionConnectionInput as a field resolution input and returns arbitrary
func (f SubscriptionConnectionHandlerFunc) Handle(input driver.SubscriptionConnectionInput) (interface{}, error) {
	return f(input)
}

// SubscriptionConnection function calls user implemented handler for subscription connection creation
func (m *Server) SubscriptionConnection(ctx context.Context, input *proto.SubscriptionConnectionRequest) (s *proto.SubscriptionConnectionResponse, _ error) {
	defer func() {
		if r := recover(); r != nil {
			s = &proto.SubscriptionConnectionResponse{
				Error: &proto.Error{
					Msg: fmt.Sprintf("%v", r),
				},
			}
		}
	}()
	req, err := protodriver.MakeSubscriptionConnectionInput(input)
	if err == nil {
		s = new(proto.SubscriptionConnectionResponse)
		var resp interface{}
		resp, err = m.SubscriptionConnectionHandler.Handle(req)
		if err == nil {
			*s = protodriver.MakeSubscriptionConnectionResponse(resp)
		}
	}
	if err != nil {
		s.Error = &proto.Error{
			Msg: err.Error(),
		}
	}
	return
}
