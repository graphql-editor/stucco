package grpc

import (
	"context"
	"fmt"

	"github.com/graphql-editor/stucco/pkg/driver"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// Authorize marshals a field resolution request through GRPC to a function
// that handles an actual resolution.
func (m *Client) Authorize(input driver.AuthorizeInput) (f driver.AuthorizeOutput) {
	req, err := protodriver.MakeAuthorizeRequest(input)
	if err == nil {
		var resp *protoMessages.AuthorizeResponse
		resp, err = m.Client.Authorize(context.Background(), req)
		if err == nil {
			f = protodriver.MakeAuthorizeOutput(resp)
		}
	}
	if err != nil {
		f.Error = &driver.Error{Message: err.Error()}
	}
	return
}

// AuthorizeHandler interface implemented by user to handle field resolution request.
type AuthorizeHandler interface {
	// Handle takes AuthorizeInput as a field resolution input and returns arbitrary
	// user response.
	Handle(input driver.AuthorizeInput) (bool, error)
}

// AuthorizeHandlerFunc is a convienience function wrapper implementing AuthorizeHandler
type AuthorizeHandlerFunc func(input driver.AuthorizeInput) (bool, error)

// Handle takes AuthorizeInput as a field resolution input and returns arbitrary
func (f AuthorizeHandlerFunc) Handle(input driver.AuthorizeInput) (bool, error) {
	return f(input)
}

// Authorize function calls user implemented handler for field resolution
func (m *Server) Authorize(ctx context.Context, input *protoMessages.AuthorizeRequest) (f *protoMessages.AuthorizeResponse, _ error) {
	defer func() {
		if r := recover(); r != nil {
			f = &protoMessages.AuthorizeResponse{
				Error: &protoMessages.Error{
					Msg: fmt.Sprintf("%v", r),
				},
			}
		}
	}()
	req, err := protodriver.MakeAuthorizeInput(input)
	if err == nil {
		var resp bool
		resp, err = m.AuthorizeHandler.Handle(req)
		if err == nil {
			f = protodriver.MakeAuthorizeResponse(resp)
		}
	}
	if err != nil {
		f = &protoMessages.AuthorizeResponse{
			Error: &protoMessages.Error{
				Msg: err.Error(),
			},
		}
	}
	return
}
