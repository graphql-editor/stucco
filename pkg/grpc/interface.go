package grpc

import (
	"context"
	"fmt"

	"github.com/graphql-editor/stucco/pkg/driver"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// InterfaceResolveType handles type resolution for interface through GRPC
func (m *Client) InterfaceResolveType(input driver.InterfaceResolveTypeInput) (i driver.InterfaceResolveTypeOutput) {
	req, err := protodriver.MakeInterfaceResolveTypeRequest(input)
	if err == nil {
		var resp *protoMessages.InterfaceResolveTypeResponse
		resp, err = m.Client.InterfaceResolveType(context.Background(), req)
		if err == nil {
			i = protodriver.MakeInterfaceResolveTypeOutput(resp)
		}
	}
	if err != nil {
		i.Error = &driver.Error{Message: err.Error()}
		err = nil
	}
	return
}

// InterfaceResolveTypeHandler interface implemented by user to handle interface type resolution
type InterfaceResolveTypeHandler interface {
	// Handle takes InterfaceResolveTypeInput as a type resolution input and returns
	// type name.
	Handle(driver.InterfaceResolveTypeInput) (string, error)
}

// InterfaceResolveTypeHandlerFunc is a convienience function wrapper implementing InterfaceResolveTypeHandler
type InterfaceResolveTypeHandlerFunc func(driver.InterfaceResolveTypeInput) (string, error)

// Handle takes InterfaceResolveTypeInput as a type resolution input and returns
// type name.
func (f InterfaceResolveTypeHandlerFunc) Handle(in driver.InterfaceResolveTypeInput) (string, error) {
	return f(in)
}

// InterfaceResolveType handles type resolution request with user defined function
func (m *Server) InterfaceResolveType(ctx context.Context, input *protoMessages.InterfaceResolveTypeRequest) (f *protoMessages.InterfaceResolveTypeResponse, _ error) {
	defer func() {
		if r := recover(); r != nil {
			f = &protoMessages.InterfaceResolveTypeResponse{
				Error: &protoMessages.Error{
					Msg: fmt.Sprintf("%v", r),
				},
			}
		}
	}()
	req, err := protodriver.MakeInterfaceResolveTypeInput(input)
	if err == nil {
		var resp string
		resp, err = m.InterfaceResolveTypeHandler.Handle(req)
		if err == nil {
			f = protodriver.MakeInterfaceResolveTypeResponse(resp)
		}
	}
	if err != nil {
		f = &protoMessages.InterfaceResolveTypeResponse{
			Error: &protoMessages.Error{Msg: err.Error()},
		}
	}
	return
}
