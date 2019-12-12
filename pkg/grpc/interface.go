package grpc

import (
	"context"
	"fmt"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
)

func makeProtoInterfaceResolveTypeInfo(input driver.InterfaceResolveTypeInfo) (r *proto.InterfaceResolveTypeInfo, err error) {
	variables := input.VariableValues
	variableValues, err := mapOfAnyToMapOfValue(variables)
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
	r = &proto.InterfaceResolveTypeInfo{
		FieldName:      input.FieldName,
		Path:           rp,
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
	if input.Function.Name == "" {
		return nil, fmt.Errorf("function name is required")
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

// InterfaceResolveType handles type resolution for interface through GRPC
func (m *Client) InterfaceResolveType(input driver.InterfaceResolveTypeInput) (f driver.InterfaceResolveTypeOutput, err error) {
	req, err := makeInterfaceResolveTypeRequest(input)
	if err != nil {
		f.Error = &driver.Error{Message: err.Error()}
		err = nil
		return
	}
	resp, err := m.Client.InterfaceResolveType(context.Background(), req)
	if err == nil {
		if t := resp.GetType(); t != nil {
			f.Type = *makeDriverTypeRef(t)
		}
		if respErr := resp.GetError(); respErr != nil {
			err = fmt.Errorf(respErr.GetMsg())
		}
	}
	if err != nil {
		f.Error = &driver.Error{Message: err.Error()}
		err = nil
	}
	return
}

func makeDriverInterfaceResolveTypeInfo(input *proto.InterfaceResolveTypeInfo) (i driver.InterfaceResolveTypeInfo, err error) {
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
	i = driver.InterfaceResolveTypeInfo{
		FieldName:      input.GetFieldName(),
		Path:           rp,
		ReturnType:     makeDriverTypeRef(input.GetReturnType()),
		ParentType:     makeDriverTypeRef(input.GetParentType()),
		VariableValues: variableValues,
		Operation:      od,
	}
	return
}

func makeInterfaceResolveTypeInput(input *proto.InterfaceResolveTypeRequest) (i driver.InterfaceResolveTypeInput, err error) {
	val, err := valueToAny(nil, input.GetValue())
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

func (m *Server) InterfaceResolveType(ctx context.Context, input *proto.InterfaceResolveTypeRequest) (f *proto.InterfaceResolveTypeResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			f = &proto.InterfaceResolveTypeResponse{
				Error: &proto.Error{
					Msg: fmt.Sprintf("%v", r),
				},
			}
		}
	}()
	req, err := makeInterfaceResolveTypeInput(input)
	if err == nil {
		var resp string
		resp, err = m.InterfaceResolveTypeHandler.Handle(req)
		f = new(proto.InterfaceResolveTypeResponse)
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
