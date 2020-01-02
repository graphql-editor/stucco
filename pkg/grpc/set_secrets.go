package grpc

import (
	"context"
	"fmt"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
)

// SetSecrets sets a marshals secrets through GRPC
func (m *Client) SetSecrets(input driver.SetSecretsInput) driver.SetSecretsOutput {
	var out driver.SetSecretsOutput
	req := protodriver.MakeSetSecretsRequest(input)
	var resp *proto.SetSecretsResponse
	resp, err := m.Client.SetSecrets(context.Background(), req)
	if err == nil {
		out = protodriver.MakeSetSecretsOutput(resp)
	}
	if err != nil {
		out.Error = &driver.Error{Message: err.Error()}
	}
	return out
}

// SetSecretsHandler interface implemented by user to handle secrets input from client.
type SetSecretsHandler interface {
	// Handle takes SetSecretsHandler as an input and should set a secrets on a server state. It should return nil if there was no error.
	Handle(input driver.SetSecretsInput) error
}

// SetSecretsHandlerFunc is a convienience wrapper around function to implement SetSecretsHandler
type SetSecretsHandlerFunc func(input driver.SetSecretsInput) error

// Handle takes SetSecretsInput as an input, sets a secrets state, and returns a nil if there was no error
func (s SetSecretsHandlerFunc) Handle(input driver.SetSecretsInput) error {
	return s(input)
}

// SetSecrets calls user SetSecrets handler
func (m *Server) SetSecrets(ctx context.Context, input *proto.SetSecretsRequest) (o *proto.SetSecretsResponse, _ error) {
	defer func() {
		if r := recover(); r != nil {
			o = &proto.SetSecretsResponse{
				Error: &proto.Error{
					Msg: fmt.Sprintf("%v", r),
				},
			}
		}
	}()
	return protodriver.MakeSetSecretsResponse(
		m.SetSecretsHandler.Handle(
			protodriver.MakeSetSecretsInput(input),
		),
	), nil
}
