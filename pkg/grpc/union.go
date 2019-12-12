package grpc

import (
	"context"
	"fmt"

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
	rp, err := makeProtoResponsePath(input.Path)
	if err != nil {
		return
	}
	r = &proto.UnionResolveTypeInfo{
		FieldName:      input.FieldName,
		Path:           rp,
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
	if input.Function.Name == "" {
		return nil, fmt.Errorf("function name is required")
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

func (m *Client) UnionResolveType(input driver.UnionResolveTypeInput) (f driver.UnionResolveTypeOutput, err error) {
	req, err := makeUnionResolveTypeRequest(input)
	if err != nil {
		f.Error = &driver.Error{Message: err.Error()}
		err = nil
		return
	}
	resp, err := m.Client.UnionResolveType(context.Background(), req)
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
	variables := input.GetVariableValues()
	variableValues, err := mapOfValueToMapOfAny(nil, variables)
	if err != nil {
		return
	}
	variables = initVariablesWithDefaults(variables, input.GetOperation())
	od, err := makeDriverOperationDefinition(variables, input.GetOperation())
	if err != nil {
		return
	}
	rp, err := makeDriverResponsePath(variables, input.GetPath())
	if err != nil {
		return
	}
	u = driver.UnionResolveTypeInfo{
		FieldName:      input.GetFieldName(),
		Path:           rp,
		ReturnType:     makeDriverTypeRef(input.GetReturnType()),
		ParentType:     makeDriverTypeRef(input.GetParentType()),
		VariableValues: variableValues,
		Operation:      od,
	}
	return
}

func makeUnionResolveTypeInput(input *proto.UnionResolveTypeRequest) (u driver.UnionResolveTypeInput, err error) {
	val, err := valueToAny(nil, input.GetValue())
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

// UnionResolveTypeHandler union implemented by user to handle union type resolution
type UnionResolveTypeHandler interface {
	// Handle takes UnionResolveTypeInput as a type resolution input and returns
	// type name.
	Handle(driver.UnionResolveTypeInput) (string, error)
}

// UnionResolveTypeHandlerFunc is a convienience function wrapper implementing UnionResolveTypeHandler
type UnionResolveTypeHandlerFunc func(driver.UnionResolveTypeInput) (string, error)

// Handle takes UnionResolveTypeInput as a type resolution input and returns
// type name.
func (f UnionResolveTypeHandlerFunc) Handle(in driver.UnionResolveTypeInput) (string, error) {
	return f(in)
}

func (m *Server) UnionResolveType(ctx context.Context, input *proto.UnionResolveTypeRequest) (f *proto.UnionResolveTypeResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			f = &proto.UnionResolveTypeResponse{
				Error: &proto.Error{
					Msg: fmt.Sprintf("%v", r),
				},
			}
		}
	}()
	req, err := makeUnionResolveTypeInput(input)
	if err == nil {
		var resp string
		resp, err = m.UnionResolveTypeHandler.Handle(req)
		f = new(proto.UnionResolveTypeResponse)
		if err == nil {
			f.Type = &proto.TypeRef{TestTyperef: &proto.TypeRef_Name{Name: resp}}
		}
	}
	if err != nil {
		f.Error = &proto.Error{Msg: err.Error()}
		err = nil
	}
	return
}
