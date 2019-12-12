package grpc

import (
	"context"
	"fmt"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
)

func (m *Client) ScalarParse(input driver.ScalarParseInput) (s driver.ScalarParseOutput, err error) {
	v, err := anyToValue(input.Value)
	if err != nil {
		s.Error = &driver.Error{
			Message: err.Error(),
		}
		return
	}
	spr := &proto.ScalarParseRequest{
		Function: &proto.Function{
			Name: input.Function.Name,
		},
		Value: v,
	}
	resp, err := m.Client.ScalarParse(context.Background(), spr)
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
	if err != nil {
		s.Error = &driver.Error{Message: err.Error()}
		err = nil
	}
	return
}

func (m *Client) ScalarSerialize(input driver.ScalarSerializeInput) (s driver.ScalarSerializeOutput, err error) {
	ssr := &proto.ScalarSerializeRequest{
		Function: &proto.Function{
			Name: input.Function.Name,
		},
	}
	val, err := anyToValue(input.Value)
	if err != nil {
		s.Error = &driver.Error{Message: err.Error()}
		err = nil
		return
	}
	ssr.Value = val
	resp, err := m.Client.ScalarSerialize(context.Background(), ssr)
	if err != nil {
		s.Error = &driver.Error{Message: err.Error()}
		err = nil
		return
	}
	s.Response, err = valueToAny(nil, resp.GetValue())
	if err != nil {
		s.Error = &driver.Error{Message: err.Error()}
	} else if rerr := resp.GetError(); rerr != nil {
		s.Error = &driver.Error{Message: rerr.GetMsg()}
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

// ScalarParseHandlerFunc is a convienience wrapper for function implementing ScalarParseHandler
func (f ScalarParseHandlerFunc) Handle(input driver.ScalarParseInput) (interface{}, error) {
	return f(input)
}

// ScalarParse  calls user defined function for parsing a scalar.
func (m *Server) ScalarParse(ctx context.Context, input *proto.ScalarParseRequest) (s *proto.ScalarParseResponse, err error) {
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
	if err != nil {
		return
	}
	resp, err := m.ScalarParseHandler.Handle(driver.ScalarParseInput{
		Function: types.Function{
			Name: input.GetFunction().GetName(),
		},
		Value: val,
	})
	if err == nil {
		s.Value, err = anyToValue(resp)
	}
	if err != nil {
		s.Error = &proto.Error{Msg: err.Error()}
		err = nil
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

// ScalarSerializeHandlerFunc is a convienience wrapper for function implementing ScalarSerializeHandler
func (f ScalarSerializeHandlerFunc) Handle(input driver.ScalarSerializeInput) (interface{}, error) {
	return f(input)
}

func (m *Server) ScalarSerialize(ctx context.Context, input *proto.ScalarSerializeRequest) (s *proto.ScalarSerializeResponse, err error) {
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
	if err != nil {
		return
	}
	resp, err := m.ScalarSerializeHandler.Handle(driver.ScalarSerializeInput{
		Function: types.Function{
			Name: input.GetFunction().GetName(),
		},
		Value: val,
	})
	if err == nil {
		s.Value, err = anyToValue(resp)
	}
	if err != nil {
		s.Error = &proto.Error{Msg: err.Error()}
		err = nil
	}
	return
}
