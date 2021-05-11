package grpc

import (
	"context"
	"fmt"

	"github.com/graphql-editor/stucco/pkg/driver"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// FieldResolve marshals a field resolution request through GRPC to a function
// that handles an actual resolution.
func (m *Client) FieldResolve(input driver.FieldResolveInput) (f driver.FieldResolveOutput) {
	req, err := protodriver.MakeFieldResolveRequest(input)
	if err == nil {
		var resp *protoMessages.FieldResolveResponse
		resp, err = m.Client.FieldResolve(context.Background(), req)
		if err == nil {
			f = protodriver.MakeFieldResolveOutput(resp)
		}
	}
	if err != nil {
		f.Error = &driver.Error{Message: err.Error()}
	}
	return
}

// FieldResolveHandler interface implemented by user to handle field resolution request.
type FieldResolveHandler interface {
	// Handle takes FieldResolveInput as a field resolution input and returns arbitrary
	// user response.
	Handle(input driver.FieldResolveInput) (interface{}, error)
}

// FieldResolveHandlerFunc is a convienience function wrapper implementing FieldResolveHandler
type FieldResolveHandlerFunc func(input driver.FieldResolveInput) (interface{}, error)

// Handle takes FieldResolveInput as a field resolution input and returns arbitrary
func (f FieldResolveHandlerFunc) Handle(input driver.FieldResolveInput) (interface{}, error) {
	return f(input)
}

// FieldResolve function calls user implemented handler for field resolution
func (m *Server) FieldResolve(ctx context.Context, input *protoMessages.FieldResolveRequest) (f *protoMessages.FieldResolveResponse, _ error) {
	defer func() {
		if r := recover(); r != nil {
			f = &protoMessages.FieldResolveResponse{
				Error: &protoMessages.Error{
					Msg: fmt.Sprintf("%v", r),
				},
			}
		}
	}()
	req, err := protodriver.MakeFieldResolveInput(input)
	if err == nil {
		f = new(protoMessages.FieldResolveResponse)
		var resp interface{}
		resp, err = m.FieldResolveHandler.Handle(req)
		if err == nil {
			*f = protodriver.MakeFieldResolveResponse(resp)
		}
	}
	if err != nil {
		f.Error = &protoMessages.Error{
			Msg: err.Error(),
		}
	}
	return
}
