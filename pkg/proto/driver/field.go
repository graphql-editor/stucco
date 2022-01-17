package protodriver

import (
	"io"
	"io/ioutil"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/types"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

func makeProtoFieldResolveInfo(input driver.FieldResolveInfo) (r *protoMessages.FieldResolveInfo, err error) {
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
	var rt *protoMessages.Value
	if input.RootValue != nil {
		rt, err = anyToValue(input.RootValue)
		if err != nil {
			return
		}
	}
	r = &protoMessages.FieldResolveInfo{
		FieldName:      input.FieldName,
		Path:           rp,
		ReturnType:     makeProtoTypeRef(input.ReturnType),
		ParentType:     makeProtoTypeRef(input.ParentType),
		VariableValues: variableValues,
		Operation:      od,
		RootValue:      rt,
	}
	return
}

// MakeFieldResolveRequest creates a new proto FieldResolveRequest from driver input
func MakeFieldResolveRequest(input driver.FieldResolveInput) (r *protoMessages.FieldResolveRequest, err error) {
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
	subscriptionPayload, err := anyToValue(input.SubscriptionPayload)
	if err != nil {
		return
	}
	r = &protoMessages.FieldResolveRequest{
		Function: &protoMessages.Function{
			Name: input.Function.Name,
		},
		Source:              source,
		Info:                info,
		Arguments:           args,
		Protocol:            protocol,
		SubscriptionPayload: subscriptionPayload,
	}
	return
}

// MakeFieldResolveOutput creates new driver.FieldResolveOutput from proto response
func MakeFieldResolveOutput(resp *protoMessages.FieldResolveResponse) (out driver.FieldResolveOutput) {
	var err error
	out.Response, err = valueToAny(nil, resp.GetResponse())
	if err != nil {
		out.Error = &driver.Error{Message: err.Error()}
	} else if rerr := resp.GetError(); rerr != nil {
		out.Error = &driver.Error{Message: rerr.GetMsg()}
	}
	return out
}

func makeDriverFieldResolveInfo(input *protoMessages.FieldResolveInfo) (f driver.FieldResolveInfo, err error) {
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

// MakeFieldResolveInput creates driver.FieldResolveInput from protoMessages.FieldResolverequest
func MakeFieldResolveInput(input *protoMessages.FieldResolveRequest) (f driver.FieldResolveInput, err error) {
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
	subscriptionPayload, err := valueToAny(nil, input.GetSubscriptionPayload())
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
		Source:              source,
		Info:                info,
		Arguments:           args,
		Protocol:            protocol,
		SubscriptionPayload: subscriptionPayload,
	}
	return
}

// MakeFieldResolveResponse creates a protoMessages.FieldResolveResponse from a value
func MakeFieldResolveResponse(resp interface{}) *protoMessages.FieldResolveResponse {
	protoResponse := protoMessages.FieldResolveResponse{}
	v, err := anyToValue(resp)
	if err == nil {
		protoResponse.Response = v
	} else {
		protoResponse.Error = &protoMessages.Error{
			Msg: err.Error(),
		}
	}
	return &protoResponse
}

// ReadFieldResolveInput reads io.Reader until io.EOF and returs driver.FieldResolveInput
func ReadFieldResolveInput(r io.Reader) (driver.FieldResolveInput, error) {
	var err error
	var b []byte
	var out driver.FieldResolveInput
	protoMsg := new(protoMessages.FieldResolveRequest)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out, err = MakeFieldResolveInput(protoMsg)
		}
	}
	return out, err
}

// WriteFieldResolveInput writes FieldResolveInput into io.Writer
func WriteFieldResolveInput(w io.Writer, input driver.FieldResolveInput) error {
	req, err := MakeFieldResolveRequest(input)
	if err == nil {
		var b []byte
		b, err = protobuf.Marshal(req)
		if err == nil {
			_, err = w.Write(b)
		}
	}
	return err
}

// ReadFieldResolveOutput reads io.Reader until io.EOF and returs driver.FieldResolveOutput
func ReadFieldResolveOutput(r io.Reader) (driver.FieldResolveOutput, error) {
	var err error
	var b []byte
	var out driver.FieldResolveOutput
	protoMsg := new(protoMessages.FieldResolveResponse)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out = MakeFieldResolveOutput(protoMsg)
		}
	}
	return out, err
}

// WriteFieldResolveOutput writes FieldResolveOutput into io.Writer
func WriteFieldResolveOutput(w io.Writer, r interface{}) error {
	req := MakeFieldResolveResponse(r)
	b, err := protobuf.Marshal(req)
	if err == nil {
		_, err = w.Write(b)
	}
	return err
}
