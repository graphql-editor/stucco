package protodriver

import (
	"io"
	"io/ioutil"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/types"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// MakeAuthorizeRequest creates a new proto AuthorizeRequest from driver input
func MakeAuthorizeRequest(input driver.AuthorizeInput) (r *protoMessages.AuthorizeRequest, err error) {
	variableValues, err := mapOfAnyToMapOfValue(input.VariableValues)
	if err != nil {
		return
	}
	protocol, err := anyToValue(input.Protocol)
	if err != nil {
		return
	}
	r = &protoMessages.AuthorizeRequest{
		Function: &protoMessages.Function{
			Name: input.Function.Name,
		},
		Query:          input.Query,
		OperationName:  input.OperationName,
		VariableValues: variableValues,
		Protocol:       protocol,
	}
	return
}

// MakeAuthorizeOutput creates new driver.AuthorizeOutput from proto response
func MakeAuthorizeOutput(resp *protoMessages.AuthorizeResponse) (out driver.AuthorizeOutput) {
	out.Response = resp.GetResponse()
	if rerr := resp.GetError(); rerr != nil {
		out.Error = &driver.Error{Message: rerr.GetMsg()}
	}
	return out
}

// MakeAuthorizeInput creates driver.AuthorizeInput from protoMessages.Authorizerequest
func MakeAuthorizeInput(input *protoMessages.AuthorizeRequest) (f driver.AuthorizeInput, err error) {
	variables := input.GetVariableValues()
	variableValues, err := mapOfValueToMapOfAny(nil, variables)
	if err != nil {
		return
	}
	protocol, err := valueToAny(nil, input.GetProtocol())
	if err != nil {
		return
	}
	f = driver.AuthorizeInput{
		Function: types.Function{
			Name: input.GetFunction().GetName(),
		},
		Query:          input.GetQuery(),
		OperationName:  input.GetOperationName(),
		VariableValues: variableValues,
		Protocol:       protocol,
	}
	return
}

// MakeAuthorizeResponse creates a protoMessages.AuthorizeResponse from a value
func MakeAuthorizeResponse(resp bool) *protoMessages.AuthorizeResponse {
	return &protoMessages.AuthorizeResponse{
		Response: resp,
	}
}

// ReadAuthorizeInput reads io.Reader until io.EOF and returs driver.AuthorizeInput
func ReadAuthorizeInput(r io.Reader) (driver.AuthorizeInput, error) {
	var err error
	var b []byte
	var out driver.AuthorizeInput
	protoMsg := new(protoMessages.AuthorizeRequest)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out, err = MakeAuthorizeInput(protoMsg)
		}
	}
	return out, err
}

// WriteAuthorizeInput writes AuthorizeInput into io.Writer
func WriteAuthorizeInput(w io.Writer, input driver.AuthorizeInput) error {
	req, err := MakeAuthorizeRequest(input)
	if err == nil {
		var b []byte
		b, err = protobuf.Marshal(req)
		if err == nil {
			_, err = w.Write(b)
		}
	}
	return err
}

// ReadAuthorizeOutput reads io.Reader until io.EOF and returs driver.AuthorizeOutput
func ReadAuthorizeOutput(r io.Reader) (driver.AuthorizeOutput, error) {
	var err error
	var b []byte
	var out driver.AuthorizeOutput
	protoMsg := new(protoMessages.AuthorizeResponse)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out = MakeAuthorizeOutput(protoMsg)
		}
	}
	return out, err
}

// WriteAuthorizeOutput writes AuthorizeOutput into io.Writer
func WriteAuthorizeOutput(w io.Writer, r bool) error {
	req := MakeAuthorizeResponse(r)
	b, err := protobuf.Marshal(req)
	if err == nil {
		_, err = w.Write(b)
	}
	return err
}
