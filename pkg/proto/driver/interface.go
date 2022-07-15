package protodriver

import (
	"fmt"
	"io"
	"io/ioutil"

	protobuf "google.golang.org/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/types"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

func makeProtoInterfaceResolveTypeInfo(input driver.InterfaceResolveTypeInfo) (r *protoMessages.InterfaceResolveTypeInfo, err error) {
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
	r = &protoMessages.InterfaceResolveTypeInfo{
		FieldName:      input.FieldName,
		Path:           rp,
		ReturnType:     makeProtoTypeRef(input.ReturnType),
		ParentType:     makeProtoTypeRef(input.ParentType),
		VariableValues: variableValues,
		Operation:      od,
	}
	return
}

// MakeInterfaceResolveTypeRequest creates new protoMessages.InterfaceResolveTypeRequest from driver.InterfaceResolveTypeInput
func MakeInterfaceResolveTypeRequest(input driver.InterfaceResolveTypeInput) (r *protoMessages.InterfaceResolveTypeRequest, err error) {
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
	r = &protoMessages.InterfaceResolveTypeRequest{
		Function: &protoMessages.Function{
			Name: input.Function.Name,
		},
		Value: value,
		Info:  info,
	}
	return
}

// MakeInterfaceResolveTypeOutput creates new driver.InterfaceResolveTypeOutput from protoMessages.InterfaceResolveTypeResponse
func MakeInterfaceResolveTypeOutput(resp *protoMessages.InterfaceResolveTypeResponse) driver.InterfaceResolveTypeOutput {
	out := driver.InterfaceResolveTypeOutput{}
	if err := resp.GetError(); err != nil {
		out.Error = &driver.Error{
			Message: err.GetMsg(),
		}
	} else if t := resp.GetType(); t != nil {
		out.Type = *makeDriverTypeRef(t)
	}
	return out
}

func makeDriverInterfaceResolveTypeInfo(input *protoMessages.InterfaceResolveTypeInfo) (i driver.InterfaceResolveTypeInfo, err error) {
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

// MakeInterfaceResolveTypeInput creates new driver.InterfaceResolveTypeInput from protoMessages.InterfaceResolveTypeRequest
func MakeInterfaceResolveTypeInput(input *protoMessages.InterfaceResolveTypeRequest) (i driver.InterfaceResolveTypeInput, err error) {
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

// MakeInterfaceResolveTypeResponse creates new protoMessages.InterfaceResolveTypeResponse from type string
func MakeInterfaceResolveTypeResponse(resp string) *protoMessages.InterfaceResolveTypeResponse {
	return &protoMessages.InterfaceResolveTypeResponse{
		Type: &protoMessages.TypeRef{
			TestTyperef: &protoMessages.TypeRef_Name{
				Name: resp,
			},
		},
	}
}

// ReadInterfaceResolveTypeInput reads io.Reader until io.EOF and returs driver.InterfaceResolveTypeInput
func ReadInterfaceResolveTypeInput(r io.Reader) (driver.InterfaceResolveTypeInput, error) {
	var err error
	var b []byte
	var out driver.InterfaceResolveTypeInput
	protoMsg := new(protoMessages.InterfaceResolveTypeRequest)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out, err = MakeInterfaceResolveTypeInput(protoMsg)
		}
	}
	return out, err
}

// WriteInterfaceResolveTypeInput writes InterfaceResolveTypeInput into io.Writer
func WriteInterfaceResolveTypeInput(w io.Writer, input driver.InterfaceResolveTypeInput) error {
	req, err := MakeInterfaceResolveTypeRequest(input)
	if err == nil {
		var b []byte
		b, err = protobuf.Marshal(req)
		if err == nil {
			_, err = w.Write(b)
		}
	}
	return err
}

// ReadInterfaceResolveTypeOutput reads io.Reader until io.EOF and returs driver.InterfaceResolveTypeOutput
func ReadInterfaceResolveTypeOutput(r io.Reader) (driver.InterfaceResolveTypeOutput, error) {
	var err error
	var b []byte
	var out driver.InterfaceResolveTypeOutput
	protoMsg := new(protoMessages.InterfaceResolveTypeResponse)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out = MakeInterfaceResolveTypeOutput(protoMsg)
		}
	}
	return out, err
}

// WriteInterfaceResolveTypeOutput writes InterfaceResolveTypeOutput into io.Writer
func WriteInterfaceResolveTypeOutput(w io.Writer, r string) error {
	req := MakeInterfaceResolveTypeResponse(r)
	b, err := protobuf.Marshal(req)
	if err == nil {
		_, err = w.Write(b)
	}
	return err
}
