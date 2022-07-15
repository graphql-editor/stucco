package protodriver

import (
	"io"
	"io/ioutil"

	protobuf "google.golang.org/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/types"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// MakeSubscriptionConnectionRequest creates a new proto SubscriptionConnectionRequest from driver input
func MakeSubscriptionConnectionRequest(input driver.SubscriptionConnectionInput) (r *protoMessages.SubscriptionConnectionRequest, err error) {
	ret := protoMessages.SubscriptionConnectionRequest{
		Function: &protoMessages.Function{
			Name: input.Function.Name,
		},
		Query:         input.Query,
		OperationName: input.OperationName,
	}
	for k, v := range input.VariableValues {
		if ret.VariableValues == nil {
			ret.VariableValues = make(map[string]*protoMessages.Value)
		}
		ret.VariableValues[k], err = anyToValue(v)
		if err != nil {
			return
		}
	}
	proto, err := anyToValue(input.Protocol)
	if err == nil {
		ret.Protocol = proto
		r = &ret
	}
	return
}

// MakeSubscriptionConnectionOutput creates new driver.SubscriptionConnectionOutput from proto response
func MakeSubscriptionConnectionOutput(resp *protoMessages.SubscriptionConnectionResponse) (out driver.SubscriptionConnectionOutput) {
	var err error
	out.Response, err = valueToAny(nil, resp.GetResponse())
	if err != nil {
		out.Error = &driver.Error{Message: err.Error()}
	} else if rerr := resp.GetError(); rerr != nil {
		out.Error = &driver.Error{Message: rerr.GetMsg()}
	}
	return out
}

// MakeSubscriptionConnectionInput creates driver.SubscriptionConnectionInput from protoMessages.SubscriptionConnectionRequest
func MakeSubscriptionConnectionInput(input *protoMessages.SubscriptionConnectionRequest) (f driver.SubscriptionConnectionInput, err error) {
	f = driver.SubscriptionConnectionInput{
		Function: types.Function{
			Name: input.GetFunction().GetName(),
		},
		Query:         input.GetQuery(),
		OperationName: input.GetOperationName(),
	}
	for k, v := range input.GetVariableValues() {
		if f.VariableValues == nil {
			f.VariableValues = make(map[string]interface{})
		}
		f.VariableValues[k], err = valueToAny(nil, v)
		if err != nil {
			f = driver.SubscriptionConnectionInput{}
			return
		}
	}
	if pr := input.GetProtocol(); pr != nil {
		f.Protocol, err = valueToAny(nil, pr)
	}
	return
}

// MakeSubscriptionConnectionResponse creates a protoMessages.SubscriptionConnectionRespone from a value
func MakeSubscriptionConnectionResponse(resp interface{}) *protoMessages.SubscriptionConnectionResponse {
	protoResponse := protoMessages.SubscriptionConnectionResponse{}
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

// ReadSubscriptionConnectionInput reads io.Reader until io.EOF and returs driver.SubscriptionConnectionInput
func ReadSubscriptionConnectionInput(r io.Reader) (driver.SubscriptionConnectionInput, error) {
	var err error
	var b []byte
	var out driver.SubscriptionConnectionInput
	protoMsg := new(protoMessages.SubscriptionConnectionRequest)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out, err = MakeSubscriptionConnectionInput(protoMsg)
		}
	}
	return out, err
}

// WriteSubscriptionConnectionInput writes SubscriptionConnectionInput into io.Writer
func WriteSubscriptionConnectionInput(w io.Writer, input driver.SubscriptionConnectionInput) error {
	req, err := MakeSubscriptionConnectionRequest(input)
	if err == nil {
		var b []byte
		b, err = protobuf.Marshal(req)
		if err == nil {
			_, err = w.Write(b)
		}
	}
	return err
}

// ReadSubscriptionConnectionOutput reads io.Reader until io.EOF and returs driver.SubscriptionConnectionOutput
func ReadSubscriptionConnectionOutput(r io.Reader) (driver.SubscriptionConnectionOutput, error) {
	var err error
	var b []byte
	var out driver.SubscriptionConnectionOutput
	protoMsg := new(protoMessages.SubscriptionConnectionResponse)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out = MakeSubscriptionConnectionOutput(protoMsg)
		}
	}
	return out, err
}

// WriteSubscriptionConnectionOutput writes SubscriptionConnectionOutput into io.Writer
func WriteSubscriptionConnectionOutput(w io.Writer, r interface{}) error {
	req := MakeSubscriptionConnectionResponse(r)
	b, err := protobuf.Marshal(req)
	if err == nil {
		_, err = w.Write(b)
	}
	return err
}
