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

// MakeScalarParseRequest creates new protoMessages.ScalarParseRequest from driver.ScalarParseInput
func MakeScalarParseRequest(input driver.ScalarParseInput) (req *protoMessages.ScalarParseRequest, err error) {
	v, err := anyToValue(input.Value)
	if err == nil {
		req = &protoMessages.ScalarParseRequest{
			Function: &protoMessages.Function{
				Name: input.Function.Name,
			},
			Value: v,
		}
	}
	return
}

// MakeScalarParseOutput creates new driver.ScalarParseOutput from protoMessages.ScalarParseResponse
func MakeScalarParseOutput(resp *protoMessages.ScalarParseResponse) driver.ScalarParseOutput {
	var out driver.ScalarParseOutput
	var err error
	var r interface{}
	if respErr := resp.GetError(); respErr != nil {
		err = fmt.Errorf(respErr.GetMsg())
	} else {
		r, err = valueToAny(nil, resp.GetValue())
		if err == nil {
			out.Response = r
		}
	}
	if err != nil {
		out.Error = &driver.Error{Message: err.Error()}
	}
	return out
}

// MakeScalarSerializeRequest creates new protoMessages.ScalarSerializeRequest from driver.ScalarSerializeInput
func MakeScalarSerializeRequest(input driver.ScalarSerializeInput) (req *protoMessages.ScalarSerializeRequest, err error) {
	v, err := anyToValue(input.Value)
	if err == nil {
		req = &protoMessages.ScalarSerializeRequest{
			Function: &protoMessages.Function{
				Name: input.Function.Name,
			},
			Value: v,
		}
	}
	return
}

// MakeScalarSerializeOutput creates new driver.ScalarSerializeOutput from protoMessages.ScalarSerializeResponse
func MakeScalarSerializeOutput(resp *protoMessages.ScalarSerializeResponse) driver.ScalarSerializeOutput {
	var out driver.ScalarSerializeOutput
	var err error
	var r interface{}
	if respErr := resp.GetError(); respErr != nil {
		err = fmt.Errorf(respErr.GetMsg())
	} else {
		r, err = valueToAny(nil, resp.GetValue())
		if err == nil {
			out.Response = r
		}
	}
	if err != nil {
		out.Error = &driver.Error{Message: err.Error()}
	}
	return out
}

// MakeScalarParseInput creates new driver.ScalarParseInput from protoMessages.ScalarParseRequest
func MakeScalarParseInput(req *protoMessages.ScalarParseRequest) (driver.ScalarParseInput, error) {
	var input driver.ScalarParseInput
	val, err := valueToAny(nil, req.GetValue())
	if err == nil {
		input = driver.ScalarParseInput{
			Function: types.Function{
				Name: req.GetFunction().GetName(),
			},
			Value: val,
		}
	}
	return input, err
}

// MakeScalarParseResponse creates new protoMessages.ScalarParseResposne from any value
func MakeScalarParseResponse(value interface{}) *protoMessages.ScalarParseResponse {
	var protoResponse protoMessages.ScalarParseResponse
	v, err := anyToValue(value)
	if err != nil {
		protoResponse.Error = &protoMessages.Error{
			Msg: err.Error(),
		}
	} else {
		protoResponse.Value = v
	}
	return &protoResponse
}

// MakeScalarSerializeInput creates new driver.ScalarSerializeInput from protoMessages.ScalarSerializeRequest
func MakeScalarSerializeInput(req *protoMessages.ScalarSerializeRequest) (driver.ScalarSerializeInput, error) {
	var input driver.ScalarSerializeInput
	val, err := valueToAny(nil, req.GetValue())
	if err == nil {
		input = driver.ScalarSerializeInput{
			Function: types.Function{
				Name: req.GetFunction().GetName(),
			},
			Value: val,
		}
	}
	return input, err
}

// MakeScalarSerializeResponse creates new protoMessages.ScalarSerializeResposne from any value
func MakeScalarSerializeResponse(value interface{}) *protoMessages.ScalarSerializeResponse {
	var protoResponse protoMessages.ScalarSerializeResponse
	v, err := anyToValue(value)
	if err != nil {
		protoResponse.Error = &protoMessages.Error{
			Msg: err.Error(),
		}
	} else {
		protoResponse.Value = v
	}
	return &protoResponse
}

// ReadScalarSerializeInput reads io.Reader until io.EOF and returs driver.ScalarSerializeInput
func ReadScalarSerializeInput(r io.Reader) (driver.ScalarSerializeInput, error) {
	var err error
	var b []byte
	var out driver.ScalarSerializeInput
	protoMsg := new(protoMessages.ScalarSerializeRequest)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out, err = MakeScalarSerializeInput(protoMsg)
		}
	}
	return out, err
}

// WriteScalarSerializeInput writes ScalarSerializeInput into io.Writer
func WriteScalarSerializeInput(w io.Writer, input driver.ScalarSerializeInput) error {
	req, err := MakeScalarSerializeRequest(input)
	if err == nil {
		var b []byte
		b, err = protobuf.Marshal(req)
		if err == nil {
			_, err = w.Write(b)
		}
	}
	return err
}

// ReadScalarSerializeOutput reads io.Reader until io.EOF and returs driver.ScalarSerializeOutput
func ReadScalarSerializeOutput(r io.Reader) (driver.ScalarSerializeOutput, error) {
	var err error
	var b []byte
	var out driver.ScalarSerializeOutput
	protoMsg := new(protoMessages.ScalarSerializeResponse)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out = MakeScalarSerializeOutput(protoMsg)
		}
	}
	return out, err
}

// WriteScalarSerializeOutput writes ScalarSerializeOutput into io.Writer
func WriteScalarSerializeOutput(w io.Writer, r interface{}) error {
	req := MakeScalarSerializeResponse(r)
	b, err := protobuf.Marshal(req)
	if err == nil {
		_, err = w.Write(b)
	}
	return err
}

// ReadScalarParseInput reads io.Reader until io.EOF and returs driver.ScalarParseInput
func ReadScalarParseInput(r io.Reader) (driver.ScalarParseInput, error) {
	var err error
	var b []byte
	var out driver.ScalarParseInput
	protoMsg := new(protoMessages.ScalarParseRequest)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out, err = MakeScalarParseInput(protoMsg)
		}
	}
	return out, err
}

// WriteScalarParseInput writes ScalarParseInput into io.Writer
func WriteScalarParseInput(w io.Writer, input driver.ScalarParseInput) error {
	req, err := MakeScalarParseRequest(input)
	if err == nil {
		var b []byte
		b, err = protobuf.Marshal(req)
		if err == nil {
			_, err = w.Write(b)
		}
	}
	return err
}

// ReadScalarParseOutput reads io.Reader until io.EOF and returs driver.ScalarParseOutput
func ReadScalarParseOutput(r io.Reader) (driver.ScalarParseOutput, error) {
	var err error
	var b []byte
	var out driver.ScalarParseOutput
	protoMsg := new(protoMessages.ScalarParseResponse)
	if b, err = ioutil.ReadAll(r); err == nil {
		if err = protobuf.Unmarshal(b, protoMsg); err == nil {
			out = MakeScalarParseOutput(protoMsg)
		}
	}
	return out, err
}

// WriteScalarParseOutput writes ScalarParseOutput into io.Writer
func WriteScalarParseOutput(w io.Writer, r interface{}) error {
	req := MakeScalarParseResponse(r)
	b, err := protobuf.Marshal(req)
	if err == nil {
		_, err = w.Write(b)
	}
	return err
}
