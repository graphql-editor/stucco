package protodriver

import (
	"fmt"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
)

func makeProtoUnionResolveTypeInfo(input driver.UnionResolveTypeInfo) (r *proto.UnionResolveTypeInfo, err error) {
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

// MakeUnionResolveTypeRequest creates new proto.UnionResolveTypeRequest from driver.UnionResolveTypeInput
func MakeUnionResolveTypeRequest(input driver.UnionResolveTypeInput) (r *proto.UnionResolveTypeRequest, err error) {
	info, err := makeProtoUnionResolveTypeInfo(input.Info)
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

// MakeUnionResolveTypeOutput creates new driver.UnionResolveTypeOutput from proto.UnionResolveTypeResponse
func MakeUnionResolveTypeOutput(resp *proto.UnionResolveTypeResponse) driver.UnionResolveTypeOutput {
	out := driver.UnionResolveTypeOutput{}
	if err := resp.GetError(); err != nil {
		out.Error = &driver.Error{
			Message: err.GetMsg(),
		}
	} else if t := resp.GetType(); t != nil {
		out.Type = *makeDriverTypeRef(t)
	}
	return out
}

func makeDriverUnionResolveTypeInfo(input *proto.UnionResolveTypeInfo) (i driver.UnionResolveTypeInfo, err error) {
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
	i = driver.UnionResolveTypeInfo{
		FieldName:      input.GetFieldName(),
		Path:           rp,
		ReturnType:     makeDriverTypeRef(input.GetReturnType()),
		ParentType:     makeDriverTypeRef(input.GetParentType()),
		VariableValues: variableValues,
		Operation:      od,
	}
	return
}

// MakeUnionResolveTypeInput creates new driver.UnionResolveTypeInput from proto.UnionResolveTypeRequest
func MakeUnionResolveTypeInput(input *proto.UnionResolveTypeRequest) (i driver.UnionResolveTypeInput, err error) {
	val, err := valueToAny(nil, input.GetValue())
	if err != nil {
		return
	}
	info, err := makeDriverUnionResolveTypeInfo(input.GetInfo())
	if err != nil {
		return
	}
	i = driver.UnionResolveTypeInput{
		Function: types.Function{
			Name: input.GetFunction().GetName(),
		},
		Value: val,
		Info:  info,
	}
	return
}

// MakeUnionResolveTypeResponse creates new proto.UnionResolveTypeResponse from type string
func MakeUnionResolveTypeResponse(resp string) proto.UnionResolveTypeResponse {
	return proto.UnionResolveTypeResponse{
		Type: &proto.TypeRef{
			TestTyperef: &proto.TypeRef_Name{
				Name: resp,
			},
		},
	}
}
