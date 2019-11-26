package grpc

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
)

func makeUnionResolveTypeInfo(input driver.UnionResolveTypeInfo) (r *proto.UnionResolveTypeInfo, err error) {
	variableValues, err := mapOfAnyToMapOfValue(input.VariableValues)
	if err != nil {
		return
	}
	od, err := makeProtoOperationDefinition(input.Operation)
	if err != nil {
		return
	}
	r = &proto.UnionResolveTypeInfo{
		FieldName:      input.FieldName,
		Path:           makeProtoResponsePath(input.Path),
		ReturnType:     makeProtoTypeRef(input.ReturnType),
		ParentType:     makeProtoTypeRef(input.ParentType),
		VariableValues: variableValues,
		Operation:      od,
	}
	return
}

func makeUnionResolveTypeRequest(input driver.UnionResolveTypeInput) (r *proto.UnionResolveTypeRequest, err error) {
	info, err := makeUnionResolveTypeInfo(input.Info)
	if err != nil {
		return
	}
	value, err := anyToValue(input.Value)
	if err != nil {
		return
	}
	r = &proto.UnionResolveTypeRequest{
		Function: &proto.Function{
			Name: input.Function.Name,
		},
		Value: value,
		Info:  info,
	}
	return
}

func (m *GRPCClient) UnionResolveType(input driver.UnionResolveTypeInput) (f driver.UnionResolveTypeOutput, err error) {
	req, err := makeUnionResolveTypeRequest(input)
	if err != nil {
		f.Error = &driver.Error{Message: err.Error()}
		err = nil
		return
	}
	resp, err := m.client.UnionResolveType(context.Background(), req)
	if err != nil {
		f.Error = &driver.Error{Message: err.Error()}
		err = nil
		return
	}
	if r := resp.GetType(); r != nil {
		f.Type = *makeDriverTypeRef(r)
	}
	if err := resp.GetError(); err != nil {
		f.Error = &driver.Error{
			Message: err.Msg,
		}
	}
	return
}

func makeDriverUnionResolveTypeInfo(input *proto.UnionResolveTypeInfo) (u driver.UnionResolveTypeInfo, err error) {
	variableValues, err := mapOfValueToMapOfAny(input.GetVariableValues())
	if err != nil {
		return
	}
	od, err := makeDriverOperationDefinition(input.GetOperation())
	if err != nil {
		return
	}
	u = driver.UnionResolveTypeInfo{
		FieldName:      input.GetFieldName(),
		Path:           makeDriverResponsePath(input.GetPath()),
		ReturnType:     makeDriverTypeRef(input.GetReturnType()),
		ParentType:     makeDriverTypeRef(input.GetParentType()),
		VariableValues: variableValues,
		Operation:      od,
	}
	return
}

func makeUnionResolveTypeInput(input *proto.UnionResolveTypeRequest) (u driver.UnionResolveTypeInput, err error) {
	val, err := valueToAny(input.GetValue())
	if err != nil {
		return
	}
	info, err := makeDriverUnionResolveTypeInfo(input.GetInfo())
	if err != nil {
		return
	}
	u = driver.UnionResolveTypeInput{
		Function: types.Function{
			Name: input.GetFunction().GetName(),
		},
		Value: val,
		Info:  info,
	}
	return
}

func (m *GRPCServer) UnionResolveType(ctx context.Context, input *proto.UnionResolveTypeRequest) (f *proto.UnionResolveTypeResponse, err error) {
	req, err := makeUnionResolveTypeInput(input)
	if err != nil {
		return
	}
	resp, err := m.Impl.UnionResolveType(req)
	f = new(proto.UnionResolveTypeResponse)
	if err != nil {
		f.Error = &proto.Error{Msg: err.Error()}
	} else {
		f.Type = makeProtoTypeRef(&resp.Type)
	}
	return
}
