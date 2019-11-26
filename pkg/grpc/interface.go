package grpc

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
)

func makeProtoInterfaceResolveTypeInfo(input driver.InterfaceResolveTypeInfo) (r *proto.InterfaceResolveTypeInfo, err error) {
	variableValues, err := mapOfAnyToMapOfValue(input.VariableValues)
	if err != nil {
		return
	}
	od, err := makeProtoOperationDefinition(input.Operation)
	if err != nil {
		return
	}
	r = &proto.InterfaceResolveTypeInfo{
		FieldName:      input.FieldName,
		Path:           makeProtoResponsePath(input.Path),
		ReturnType:     makeProtoTypeRef(input.ReturnType),
		ParentType:     makeProtoTypeRef(input.ParentType),
		VariableValues: variableValues,
		Operation:      od,
	}
	return
}

func makeInterfaceResolveTypeRequest(input driver.InterfaceResolveTypeInput) (r *proto.InterfaceResolveTypeRequest, err error) {
	info, err := makeProtoInterfaceResolveTypeInfo(input.Info)
	if err != nil {
		return
	}
	value, err := anyToValue(input.Value)
	if err != nil {
		return
	}
	r = &proto.InterfaceResolveTypeRequest{
		Function: &proto.Function{
			Name: input.Function.Name,
		},
		Value: value,
		Info:  info,
	}
	return
}

func (m *GRPCClient) InterfaceResolveType(input driver.InterfaceResolveTypeInput) (f driver.InterfaceResolveTypeOutput, err error) {
	req, err := makeInterfaceResolveTypeRequest(input)
	if err != nil {
		f.Error = &driver.Error{Message: err.Error()}
		err = nil
		return
	}
	resp, err := m.client.InterfaceResolveType(context.Background(), req)
	if err != nil {
		f.Error = &driver.Error{Message: err.Error()}
		err = nil
		return
	}
	if r := resp.GetType(); r != nil {
		f.Type = *makeDriverTypeRef(r)
	}
	if err := resp.GetError(); err != nil {
		f.Error = &driver.Error{Message: err.GetMsg()}
	}
	return
}

func makeDriverInterfaceResolveTypeInfo(input *proto.InterfaceResolveTypeInfo) (i driver.InterfaceResolveTypeInfo, err error) {
	variableValues, err := mapOfValueToMapOfAny(input.GetVariableValues())
	if err != nil {
		return
	}
	od, err := makeDriverOperationDefinition(input.GetOperation())
	if err != nil {
		return
	}
	i = driver.InterfaceResolveTypeInfo{
		FieldName:      input.GetFieldName(),
		Path:           makeDriverResponsePath(input.GetPath()),
		ReturnType:     makeDriverTypeRef(input.GetReturnType()),
		ParentType:     makeDriverTypeRef(input.GetParentType()),
		VariableValues: variableValues,
		Operation:      od,
	}
	return
}

func makeInterfaceResolveTypeInput(input *proto.InterfaceResolveTypeRequest) (i driver.InterfaceResolveTypeInput, err error) {
	val, err := valueToAny(input.GetValue())
	if err != nil {
		return
	}
	info, err := makeDriverInterfaceResolveTypeInfo(input.GetInfo())
	if err != nil {
		return
	}
	i = driver.InterfaceResolveTypeInput{
		Function: types.Function{
			Name: input.GetFunction().GetName(),
		},
		Value: val,
		Info:  info,
	}
	return
}

func (m *GRPCServer) InterfaceResolveType(ctx context.Context, input *proto.InterfaceResolveTypeRequest) (f *proto.InterfaceResolveTypeResponse, err error) {
	req, err := makeInterfaceResolveTypeInput(input)
	if err != nil {
		f.Error = &proto.Error{Msg: err.Error()}
		err = nil
		return
	}
	resp, err := m.Impl.InterfaceResolveType(req)
	if err != nil {
		f.Error = &proto.Error{Msg: err.Error()}
		err = nil
		return
	}
	f = new(proto.InterfaceResolveTypeResponse)
	if err != nil {
		f.Error = &proto.Error{Msg: err.Error()}
	} else {
		f.Type = makeProtoTypeRef(&resp.Type)
	}
	return
}
