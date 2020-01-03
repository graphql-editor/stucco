package protohttp

import (
	"io/ioutil"
	"net/http"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
)

// ScalarParse over http
func (c *Client) ScalarParse(input driver.ScalarParseInput) driver.ScalarParseOutput {
	var out driver.ScalarParseOutput
	req, err := protodriver.MakeScalarParseRequest(input)
	if err == nil {
		resp := new(proto.ScalarParseResponse)
		if err = c.do(message{
			contentType: scalarParseRequestMessage,
			proto:       req,
		}, message{
			contentType: scalarParseResponseMessage,
			proto:       resp,
		}); err == nil {
			out = protodriver.MakeScalarParseOutput(resp)
		}
	}
	if err != nil {
		out.Error = &driver.Error{
			Message: err.Error(),
		}
	}
	return out
}

// ScalarSerialize over http
func (c *Client) ScalarSerialize(input driver.ScalarSerializeInput) driver.ScalarSerializeOutput {
	var out driver.ScalarSerializeOutput
	req, err := protodriver.MakeScalarSerializeRequest(input)
	if err == nil {
		resp := new(proto.ScalarSerializeResponse)
		if err = c.do(message{
			contentType: scalarSerializeRequestMessage,
			proto:       req,
		}, message{
			contentType: scalarSerializeResponseMessage,
			proto:       resp,
		}); err == nil {
			out = protodriver.MakeScalarSerializeOutput(resp)
		}
	}
	if err != nil {
		out.Error = &driver.Error{
			Message: err.Error(),
		}
	}
	return out
}

func (h *Handler) scalarParse(req *http.Request) *proto.ScalarParseResponse {
	resp := new(proto.ScalarParseResponse)
	protoReq := new(proto.ScalarParseRequest)
	var err error
	var b []byte
	if b, err = ioutil.ReadAll(req.Body); err == nil {
		defer req.Body.Close()
		if err = protobuf.Unmarshal(b, protoReq); err == nil {
			var in driver.ScalarParseInput
			in, err = protodriver.MakeScalarParseInput(protoReq)
			if err == nil {
				var scalar interface{}
				scalar, err = h.ScalarParse(in)
				if err == nil {
					*resp = protodriver.MakeScalarParseResponse(scalar)
				}
			}
		}
	}
	if err != nil {
		resp.Error = &proto.Error{
			Msg: err.Error(),
		}
	}
	return resp
}

func (h *Handler) scalarSerialize(req *http.Request) *proto.ScalarSerializeResponse {
	resp := new(proto.ScalarSerializeResponse)
	protoReq := new(proto.ScalarSerializeRequest)
	var err error
	var b []byte
	if b, err = ioutil.ReadAll(req.Body); err == nil {
		defer req.Body.Close()
		if err = protobuf.Unmarshal(b, protoReq); err == nil {
			var in driver.ScalarSerializeInput
			in, err = protodriver.MakeScalarSerializeInput(protoReq)
			if err == nil {
				var scalar interface{}
				scalar, err = h.ScalarSerialize(in)
				if err == nil {
					*resp = protodriver.MakeScalarSerializeResponse(scalar)
				}
			}
		}
	}
	if err != nil {
		resp.Error = &proto.Error{
			Msg: err.Error(),
		}
	}
	return resp
}
