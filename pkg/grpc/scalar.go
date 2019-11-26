package grpc

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
)

func (m *GRPCClient) ScalarParse(input driver.ScalarParseInput) (s driver.ScalarParseOutput, err error) {
	spr := &proto.ScalarParseRequest{
		Function: &proto.Function{
			Name: input.Function.Name,
		},
	}
	v, err := anyToValue(input.Value)
	if err != nil {
		s.Error = &driver.Error{Message: err.Error()}
		err = nil
		return
	}
	spr.Value = v
	resp, err := m.client.ScalarParse(context.Background(), spr)
	if err != nil {
		s.Error = &driver.Error{Message: err.Error()}
		err = nil
		return
	}
	s.Response, err = valueToAny(resp.GetValue())
	if err != nil {
		s.Error = &driver.Error{Message: err.Error()}
	} else if rerr := resp.GetError(); rerr != nil {
		s.Error = &driver.Error{Message: rerr.GetMsg()}
	}
	return
}

func (m *GRPCClient) ScalarSerialize(input driver.ScalarSerializeInput) (s driver.ScalarSerializeOutput, err error) {
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
	resp, err := m.client.ScalarSerialize(context.Background(), ssr)
	if err != nil {
		s.Error = &driver.Error{Message: err.Error()}
		err = nil
		return
	}
	s.Response, err = valueToAny(resp.GetValue())
	if err != nil {
		s.Error = &driver.Error{Message: err.Error()}
	} else if rerr := resp.GetError(); rerr != nil {
		s.Error = &driver.Error{Message: rerr.GetMsg()}
	}
	return
}

func (m *GRPCServer) ScalarParse(ctx context.Context, input *proto.ScalarParseRequest) (s *proto.ScalarParseResponse, err error) {
	s = new(proto.ScalarParseResponse)
	val, err := valueToAny(input.GetValue())
	if err != nil {
		return
	}
	resp, err := m.Impl.ScalarParse(driver.ScalarParseInput{
		Function: types.Function{
			Name: input.GetFunction().GetName(),
		},
		Value: val,
	})
	if err == nil {
		s.Value, err = anyToValue(resp.Response)
	}
	if err != nil {
		s.Error = &proto.Error{Msg: err.Error()}
		err = nil
	}
	return
}

func (m *GRPCServer) ScalarSerialize(ctx context.Context, input *proto.ScalarSerializeRequest) (s *proto.ScalarSerializeResponse, err error) {
	val, err := valueToAny(input.GetValue())
	if err != nil {
		return
	}
	resp, err := m.Impl.ScalarSerialize(driver.ScalarSerializeInput{
		Function: types.Function{
			Name: input.GetFunction().GetName(),
		},
		Value: val,
	})
	s = new(proto.ScalarSerializeResponse)
	if err == nil {
		s.Value, err = anyToValue(resp.Response)
	}
	if err != nil {
		s.Error = &proto.Error{Msg: err.Error()}
		err = nil
	}
	return
}
