package protodriver

import (
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
	rp, err := makeProtoResponsePath(input.Path)
	if err != nil {
		return
	}
	r = &proto.FieldResolveInfo{
		FieldName:      input.FieldName,
		Path:           rp,
		ReturnType:     makeProtoTypeRef(input.ReturnType),
		ParentType:     makeProtoTypeRef(input.ParentType),
		VariableValues: variableValues,
		Operation:      od,
	}
	return
}

// MakeFieldResolveRequest creates a new proto FieldResolveRequest from driver input
func MakeFieldResolveRequest(input driver.FieldResolveInput) (r *proto.FieldResolveRequest, err error) {
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

// MakeFieldResolveResponse creates new driver.FieldResolveOutput from proto response
func MakeFieldResolveOutput(resp *proto.FieldResolveResponse) (out driver.FieldResolveOutput) {
	var err error
	out.Response, err = valueToAny(nil, resp.GetResponse())
	if err != nil {
		out.Error = &driver.Error{Message: err.Error()}
	} else if rerr := resp.GetError(); rerr != nil {
		out.Error = &driver.Error{Message: rerr.GetMsg()}
	}
	return out
}

func makeDriverFieldResolveInfo(input *proto.FieldResolveInfo) (f driver.FieldResolveInfo, err error) {
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
	f = driver.FieldResolveInfo{
		FieldName:      input.GetFieldName(),
		Path:           rp,
		ReturnType:     makeDriverTypeRef(input.GetReturnType()),
		ParentType:     makeDriverTypeRef(input.GetParentType()),
		VariableValues: variableValues,
		Operation:      od,
	}
	return
}

// MakeFieldResolveInput creates driver.FieldResolveInput from proto.FieldResolverequest
func MakeFieldResolveInput(input *proto.FieldResolveRequest) (f driver.FieldResolveInput, err error) {
	variables := initVariablesWithDefaults(
		input.GetInfo().GetVariableValues(),
		input.GetInfo().GetOperation(),
	)
	source, err := valueToAny(nil, input.GetSource())
	if err != nil {
		return
	}
	protocol, err := valueToAny(nil, input.GetProtocol())
	if err != nil {
		return
	}
	info, err := makeDriverFieldResolveInfo(input.GetInfo())
	if err != nil {
		return
	}
	args, err := mapOfValueToMapOfAny(variables, input.GetArguments())
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

// MakeFieldResolveResponse creates a proto.FieldResolveResponse from a value
func MakeFieldResolveResponse(resp interface{}) proto.FieldResolveResponse {
	protoResponse := proto.FieldResolveResponse{}
	v, err := anyToValue(resp)
	if err == nil {
		protoResponse.Response = v
	} else {
		protoResponse.Error = &proto.Error{
			Msg: err.Error(),
		}
	}
	return protoResponse
}
