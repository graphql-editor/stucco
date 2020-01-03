package protohttp

import (
	"io/ioutil"
	"net/http"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
)

// FieldResolve over http
func (c *Client) FieldResolve(input driver.FieldResolveInput) driver.FieldResolveOutput {
	var out driver.FieldResolveOutput
	req, err := protodriver.MakeFieldResolveRequest(input)
	if err == nil {
		resp := new(proto.FieldResolveResponse)
		if err = c.do(message{
			contentType: fieldResolveRequestMessage,
			proto:       req,
		}, message{
			contentType: fieldResolveResponseMessage,
			proto:       resp,
		}); err == nil {
			out = protodriver.MakeFieldResolveOutput(resp)
		}
	}
	if err != nil {
		out.Error = &driver.Error{
			Message: err.Error(),
		}
	}
	return out
}

func (h *Handler) fieldResolve(req *http.Request) *proto.FieldResolveResponse {
	resp := new(proto.FieldResolveResponse)
	protoReq := new(proto.FieldResolveRequest)
	var err error
	var b []byte
	if b, err = ioutil.ReadAll(req.Body); err == nil {
		defer req.Body.Close()
		if err = protobuf.Unmarshal(b, protoReq); err == nil {
			var in driver.FieldResolveInput
			in, err = protodriver.MakeFieldResolveInput(protoReq)
			if err == nil {
				var driverResp interface{}
				driverResp, err = h.FieldResolve(in)
				if err == nil {
					*resp = protodriver.MakeFieldResolveResponse(driverResp)
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
