package grpc

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
)

func makeProtoFieldResolveInfo(input driver.FieldResolveInfo) (r *proto.FieldResolveInfo, err error) {
	variableValues, err := mapOfAnyToMapOfValue(input.VariableValues)
	if err != nil {
		return
	}
	od, err := makeProtoOperationDefinition(input.Operation)
	if err != nil {
		return
	}

	r = &proto.FieldResolveInfo{
		FieldName:      input.FieldName,
		Path:           makeProtoResponsePath(input.Path),
		ReturnType:     makeProtoTypeRef(input.ReturnType),
		ParentType:     makeProtoTypeRef(input.ParentType),
		VariableValues: variableValues,
		Operation:      od,
	}
	return
}

func makeFieldResolveRequest(input driver.FieldResolveInput) (r *proto.FieldResolveRequest, err error) {
	source, err := anyToValue(input.Source)
	if err != nil {
		return
	}
	args, err := mapOfAnyToMapOfValue(input.Arguments)
	if err != nil {
		return
	}
	info, err := makeProtoFieldResolveInfo(input.Info)
	if err != nil {
		return
	}
	protocol, err := anyToValue(input.Protocol)
	if err != nil {
		return
	}
	r = &proto.FieldResolveRequest{
		Function: &proto.Function{
			Name: input.Function.Name,
		},
		Source:    source,
		Info:      info,
		Arguments: args,
		Protocol:  protocol,
	}
	return
}

func (m *GRPCClient) FieldResolve(input driver.FieldResolveInput) (f driver.FieldResolveOutput, err error) {
	req, err := makeFieldResolveRequest(input)
	if err != nil {
		f.Error = &driver.Error{Message: err.Error()}
		err = nil
		return
	}
	resp, err := m.client.FieldResolve(context.Background(), req)
	if err != nil {
		f.Error = &driver.Error{Message: err.Error()}
		err = nil
		return
	}
	f.Response, err = valueToAny(resp.GetResponse())
	if err != nil {
		f.Error = &driver.Error{Message: err.Error()}
	} else if rerr := resp.GetError(); rerr != nil {
		f.Error = &driver.Error{Message: rerr.GetMsg()}
	}
	return
}

func makeDriverFieldResolveInfo(input *proto.FieldResolveInfo) (f driver.FieldResolveInfo, err error) {
	variableValues, err := mapOfValueToMapOfAny(input.GetVariableValues())
	if err != nil {
		return
	}
	od, err := makeDriverOperationDefinition(input.GetOperation())
	if err != nil {
		return
	}
	f = driver.FieldResolveInfo{
		FieldName:      input.GetFieldName(),
		Path:           makeDriverResponsePath(input.GetPath()),
		ReturnType:     makeDriverTypeRef(input.GetReturnType()),
		ParentType:     makeDriverTypeRef(input.GetParentType()),
		VariableValues: variableValues,
		Operation:      od,
	}
	return
}

func makeFieldResolveInput(input *proto.FieldResolveRequest) (f driver.FieldResolveInput, err error) {
	source, err := valueToAny(input.GetSource())
	if err != nil {
		return
	}
	protocol, err := valueToAny(input.GetProtocol())
	if err != nil {
		return
	}
	info, err := makeDriverFieldResolveInfo(input.GetInfo())
	if err != nil {
		return
	}
	args, err := mapOfValueToMapOfAny(input.GetArguments())
	if err != nil {
		return
	}
	f = driver.FieldResolveInput{
		Function: types.Function{
			Name: input.GetFunction().GetName(),
		},
		Source:    source,
		Info:      info,
		Arguments: args,
		Protocol:  protocol,
	}
	return
}

func (m *GRPCServer) FieldResolve(ctx context.Context, input *proto.FieldResolveRequest) (f *proto.FieldResolveResponse, err error) {
	req, err := makeFieldResolveInput(input)
	if err != nil {
		return
	}
	resp, err := m.Impl.FieldResolve(req)
	if err != nil {
		f.Error = &proto.Error{Msg: err.Error()}
		return
	}
	f = new(proto.FieldResolveResponse)
	if err != nil {
		f.Error = &proto.Error{Msg: err.Error()}
	} else {
		v, err := anyToValue(resp.Response)
		if err != nil {
			f.Error = &proto.Error{Msg: err.Error()}
		} else {
			f.Response = v
		}
	}
	return
}
