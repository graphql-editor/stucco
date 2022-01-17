package protodriver

import (
	"fmt"
	"io"
	"io/ioutil"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/types"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

func makeProtoUnionResolveTypeInfo(input driver.UnionResolveTypeInfo) (r *protoMessages.UnionResolveTypeInfo, err error) {
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
	r = &protoMessages.UnionResolveTypeInfo{
		FieldName:      input.FieldName,
		Path:           rp,
		ReturnType:     makeProtoTypeRef(input.ReturnType),
		ParentType:     makeProtoTypeRef(input.ParentType),
		VariableValues: variableValues,
		Operation:      od,
	}
	return
}

// MakeUnionResolveTypeRequest creates new protoMessages.UnionResolveTypeRequest from driver.UnionResolveTypeInput
func MakeUnionResolveTypeRequest(input driver.UnionResolveTypeInput) (r *protoMessages.UnionResolveTypeRequest, err error) {
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
	r = &protoMessages.UnionResolveTypeRequest{
		Function: &protoMessages.Function{
			Name: input.Function.Name,
		},
		Value: value,
		Info:  info,
	}
	return
}

// MakeUnionResolveTypeOutput creates new driver.UnionResolveTypeOutput from protoMessages.UnionResolveTypeResponse
func MakeUnionResolveTypeOutput(resp *protoMessages.UnionResolveTypeResponse) driver.UnionResolveTypeOutput {
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

func makeDriverUnionResolveTypeInfo(input *protoMessages.UnionResolveTypeInfo) (i driver.UnionResolveTypeInfo, err error) {
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

// MakeUnionResolveTypeInput creates new driver.UnionResolveTypeInput from protoMessages.UnionResolveTypeRequest
func MakeUnionResolveTypeInput(input *protoMessages.UnionResolveTypeRequest) (i driver.UnionResolveTypeInput, err error) {
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

// MakeUnionResolveTypeResponse creates new protoMessages.UnionResolveTypeResponse from type string
func MakeUnionResolveTypeResponse(resp string) *protoMessages.UnionResolveTypeResponse {
	return &protoMessages.UnionResolveTypeResponse{
		Type: &protoMessages.TypeRef{
			TestTyperef: &protoMessages.TypeRef_Name{
				Name: resp,
			},
		},
	}
}

// ReadUnionResolveTypeInput reads io.Reader until io.EOF and returs driver.UnionResolveTypeInput
func ReadUnionResolveTypeInput(r io.Reader) (driver.UnionResolveTypeInput, error) {
	var err error
	var b []byte
	var out driver.UnionResolveTypeInput
	protoMsg := new(protoMessages.UnionResolveTypeRequest)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out, err = MakeUnionResolveTypeInput(protoMsg)
		}
	}
	return out, err
}

// WriteUnionResolveTypeInput writes UnionResolveTypeInput into io.Writer
func WriteUnionResolveTypeInput(w io.Writer, input driver.UnionResolveTypeInput) error {
	req, err := MakeUnionResolveTypeRequest(input)
	if err == nil {
		var b []byte
		b, err = protobuf.Marshal(req)
		if err == nil {
			_, err = w.Write(b)
		}
	}
	return err
}

// ReadUnionResolveTypeOutput reads io.Reader until io.EOF and returs driver.UnionResolveTypeOutput
func ReadUnionResolveTypeOutput(r io.Reader) (driver.UnionResolveTypeOutput, error) {
	var err error
	var b []byte
	var out driver.UnionResolveTypeOutput
	protoMsg := new(protoMessages.UnionResolveTypeResponse)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out = MakeUnionResolveTypeOutput(protoMsg)
		}
	}
	return out, err
}

// WriteUnionResolveTypeOutput writes UnionResolveTypeOutput into io.Writer
func WriteUnionResolveTypeOutput(w io.Writer, r string) error {
	req := MakeUnionResolveTypeResponse(r)
	b, err := protobuf.Marshal(req)
	if err == nil {
		_, err = w.Write(b)
	}
	return err
}
