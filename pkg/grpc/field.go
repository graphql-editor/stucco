package grpc

import (
	"context"
	"fmt"

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

// FieldResolve marshals a field resolution request through GRPC to a function
// that handles an actual resolution.
func (m *Client) FieldResolve(input driver.FieldResolveInput) (f driver.FieldResolveOutput) {
	req, err := makeFieldResolveRequest(input)
	if err == nil {
		var resp *proto.FieldResolveResponse
		resp, err = m.Client.FieldResolve(context.Background(), req)
		if err == nil {
			f.Response, err = valueToAny(nil, resp.GetResponse())
		}
		if resp.GetError() != nil {
			err = fmt.Errorf(resp.GetError().GetMsg())
		}
	}
	if err != nil {
		f.Error = &driver.Error{Message: err.Error()}
		return
	}
	return
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

func makeFieldResolveInput(input *proto.FieldResolveRequest) (f driver.FieldResolveInput, err error) {
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

// FieldResolveHandler interface implemented by user to handle field resolution request.
type FieldResolveHandler interface {
	// Handle takes FieldResolveInput as a field resolution input and returns arbitrary
	// user response.
	Handle(input driver.FieldResolveInput) (interface{}, error)
}

// FieldResolveHandlerFunc is a convienience function wrapper implementing FieldResolveHandler
type FieldResolveHandlerFunc func(input driver.FieldResolveInput) (interface{}, error)

// Handle takes FieldResolveInput as a field resolution input and returns arbitrary
func (f FieldResolveHandlerFunc) Handle(input driver.FieldResolveInput) (interface{}, error) {
	return f(input)
}

// FieldResolve function calls user implemented handler for field resolution
func (m *Server) FieldResolve(ctx context.Context, input *proto.FieldResolveRequest) (f *proto.FieldResolveResponse, _ error) {
	defer func() {
		if r := recover(); r != nil {
			f = &proto.FieldResolveResponse{
				Error: &proto.Error{
					Msg: fmt.Sprintf("%v", r),
				},
			}
		}
	}()
	req, err := makeFieldResolveInput(input)
	if err == nil {
		f = new(proto.FieldResolveResponse)
		var resp interface{}
		resp, err = m.FieldResolveHandler.Handle(req)
		if err == nil {
			f.Response, err = anyToValue(resp)
		}
	}
	if err != nil {
		f.Error = &proto.Error{
			Msg: err.Error(),
		}
	}
	return
}
