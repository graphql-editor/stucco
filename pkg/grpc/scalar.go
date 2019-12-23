package grpc

import (
	"context"
	"fmt"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
)

// ScalarParse executes server side ScalarParse rpc
func (m *Client) ScalarParse(input driver.ScalarParseInput) (s driver.ScalarParseOutput) {
	v, err := anyToValue(input.Value)
	if err == nil {
		var resp *proto.ScalarParseResponse
		resp, err = m.Client.ScalarParse(
			context.Background(),
			&proto.ScalarParseRequest{
				Function: &proto.Function{
					Name: input.Function.Name,
				},
				Value: v,
			},
		)
		if err == nil {
			var r interface{}
			if respErr := resp.GetError(); respErr != nil {
				err = fmt.Errorf(respErr.GetMsg())
			} else {
				r, err = valueToAny(nil, resp.GetValue())
				if err == nil {
					s.Response = r
				}
			}
		}
	}
	if err != nil {
		s.Error = &driver.Error{Message: err.Error()}
		err = nil
	}
	return
}

// ScalarSerialize executes server side ScalarSerialize rpc
func (m *Client) ScalarSerialize(input driver.ScalarSerializeInput) (s driver.ScalarSerializeOutput) {
	v, err := anyToValue(input.Value)
	if err == nil {
		var resp *proto.ScalarSerializeResponse
		resp, err = m.Client.ScalarSerialize(
			context.Background(),
			&proto.ScalarSerializeRequest{
				Function: &proto.Function{
					Name: input.Function.Name,
				},
				Value: v,
			},
		)
		if err == nil {
			var r interface{}
			if respErr := resp.GetError(); respErr != nil {
				err = fmt.Errorf(respErr.GetMsg())
			} else {
				r, err = valueToAny(nil, resp.GetValue())
				if err == nil {
					s.Response = r
				}
			}
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
func (m *Server) ScalarParse(ctx context.Context, input *proto.ScalarParseRequest) (s *proto.ScalarParseResponse, _ error) {
	defer func() {
		if r := recover(); r != nil {
			s = &proto.ScalarParseResponse{
				Error: &proto.Error{
					Msg: fmt.Sprintf("%v", r),
				},
			}
		}
	}()
	s = new(proto.ScalarParseResponse)
	val, err := valueToAny(nil, input.GetValue())
	if err == nil {
		var resp interface{}
		resp, err = m.ScalarParseHandler.Handle(driver.ScalarParseInput{
			Function: types.Function{
				Name: input.GetFunction().GetName(),
			},
			Value: val,
		})
		if err == nil {
			s.Value, err = anyToValue(resp)
		}
	}
	if err != nil {
		s.Error = &proto.Error{Msg: err.Error()}
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
func (m *Server) ScalarSerialize(ctx context.Context, input *proto.ScalarSerializeRequest) (s *proto.ScalarSerializeResponse, _ error) {
	defer func() {
		if r := recover(); r != nil {
			s = &proto.ScalarSerializeResponse{
				Error: &proto.Error{
					Msg: fmt.Sprintf("%v", r),
				},
			}
		}
	}()
	s = new(proto.ScalarSerializeResponse)
	val, err := valueToAny(nil, input.GetValue())
	if err == nil {
		var resp interface{}
		resp, err = m.ScalarSerializeHandler.Handle(driver.ScalarSerializeInput{
			Function: types.Function{
				Name: input.GetFunction().GetName(),
			},
			Value: val,
		})
		if err == nil {
			s.Value, err = anyToValue(resp)
		}
	}
	if err != nil {
		s.Error = &proto.Error{Msg: err.Error()}
		err = nil
	}
	return
}
