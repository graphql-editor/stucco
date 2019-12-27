package grpc

import (
	"context"
	"fmt"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
)

// InterfaceResolveType handles type resolution for interface through GRPC
func (m *Client) InterfaceResolveType(input driver.InterfaceResolveTypeInput) (i driver.InterfaceResolveTypeOutput) {
	req, err := protodriver.MakeInterfaceResolveTypeRequest(input)
	if err == nil {
		var resp *proto.InterfaceResolveTypeResponse
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
func (m *Server) InterfaceResolveType(ctx context.Context, input *proto.InterfaceResolveTypeRequest) (f *proto.InterfaceResolveTypeResponse, _ error) {
	defer func() {
		if r := recover(); r != nil {
			f = &proto.InterfaceResolveTypeResponse{
				Error: &proto.Error{
					Msg: fmt.Sprintf("%v", r),
				},
			}
		}
	}()
	req, err := protodriver.MakeInterfaceResolveTypeInput(input)
	if err == nil {
		var resp string
		resp, err = m.InterfaceResolveTypeHandler.Handle(req)
		f = new(proto.InterfaceResolveTypeResponse)
		if err == nil {
			*f = protodriver.MakeInterfaceResolveTypeResponse(resp)
		}
	}
	if err != nil {
		f.Error = &proto.Error{Msg: err.Error()}
	}
	return
}
