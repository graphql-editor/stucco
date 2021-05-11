package grpc

import (
	"context"
	"fmt"

	"github.com/graphql-editor/stucco/pkg/driver"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// ScalarParse executes server side ScalarParse rpc
func (m *Client) ScalarParse(input driver.ScalarParseInput) (s driver.ScalarParseOutput) {
	req, err := protodriver.MakeScalarParseRequest(input)
	if err == nil {
		var resp *protoMessages.ScalarParseResponse
		resp, err = m.Client.ScalarParse(context.Background(), req)
		if err == nil {
			s = protodriver.MakeScalarParseOutput(resp)
		}
	}
	if err != nil {
		s.Error = &driver.Error{Message: err.Error()}
	}
	return
}

// ScalarSerialize executes server side ScalarSerialize rpc
func (m *Client) ScalarSerialize(input driver.ScalarSerializeInput) (s driver.ScalarSerializeOutput) {
	req, err := protodriver.MakeScalarSerializeRequest(input)
	if err == nil {
		var resp *protoMessages.ScalarSerializeResponse
		resp, err = m.Client.ScalarSerialize(context.Background(), req)
		if err == nil {
			s = protodriver.MakeScalarSerializeOutput(resp)
		}
	}
	if err != nil {
		s.Error = &driver.Error{Message: err.Error()}
		err = nil
	}
	return
}

// ScalarParseHandler interface that must be implemented by user to handle scalar parse
// requests
type ScalarParseHandler interface {
	// Handle takes ScalarParseInput as input returning arbitrary parsed value
	Handle(driver.ScalarParseInput) (interface{}, error)
}

// ScalarParseHandlerFunc is a convienience wrapper for function implementing ScalarParseHandler
type ScalarParseHandlerFunc func(driver.ScalarParseInput) (interface{}, error)

// Handle implements ScalarParseHandler.Handle
func (f ScalarParseHandlerFunc) Handle(input driver.ScalarParseInput) (interface{}, error) {
	return f(input)
}

// ScalarParse  calls user defined function for parsing a scalar.
func (m *Server) ScalarParse(ctx context.Context, input *protoMessages.ScalarParseRequest) (s *protoMessages.ScalarParseResponse, _ error) {
	defer func() {
		if r := recover(); r != nil {
			s = &protoMessages.ScalarParseResponse{
				Error: &protoMessages.Error{
					Msg: fmt.Sprintf("%v", r),
				},
			}
		}
	}()
	s = new(protoMessages.ScalarParseResponse)
	v, err := protodriver.MakeScalarParseInput(input)
	if err == nil {
		var resp interface{}
		resp, err = m.ScalarParseHandler.Handle(v)
		if err == nil {
			*s = protodriver.MakeScalarParseResponse(resp)
		}
	}
	if err != nil {
		s.Error = &protoMessages.Error{Msg: err.Error()}
	}
	return
}

// ScalarSerializeHandler interface that must be implemented by user to handle scalar serialize
// requests
type ScalarSerializeHandler interface {
	// Handle takes ScalarSerializeInput as input returning arbitrary serialized value
	Handle(driver.ScalarSerializeInput) (interface{}, error)
}

// ScalarSerializeHandlerFunc is a convienience wrapper for function implementing ScalarSerializeHandler
type ScalarSerializeHandlerFunc func(driver.ScalarSerializeInput) (interface{}, error)

// Handle implements ScalarSerializeHandler.Handle
func (f ScalarSerializeHandlerFunc) Handle(input driver.ScalarSerializeInput) (interface{}, error) {
	return f(input)
}

// ScalarSerialize executes user handler for scalar serialization
func (m *Server) ScalarSerialize(ctx context.Context, input *protoMessages.ScalarSerializeRequest) (s *protoMessages.ScalarSerializeResponse, _ error) {
	defer func() {
		if r := recover(); r != nil {
			s = &protoMessages.ScalarSerializeResponse{
				Error: &protoMessages.Error{
					Msg: fmt.Sprintf("%v", r),
				},
			}
		}
	}()
	s = new(protoMessages.ScalarSerializeResponse)
	val, err := protodriver.MakeScalarSerializeInput(input)
	if err == nil {
		var resp interface{}
		resp, err = m.ScalarSerializeHandler.Handle(val)
		if err == nil {
			*s = protodriver.MakeScalarSerializeResponse(resp)
		}
	}
	if err != nil {
		s.Error = &protoMessages.Error{Msg: err.Error()}
	}
	return
}
