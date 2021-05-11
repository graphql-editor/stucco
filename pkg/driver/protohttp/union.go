package protohttp

import (
	"io/ioutil"
	"net/http"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// UnionResolveType over http
func (c *Client) UnionResolveType(input driver.UnionResolveTypeInput) driver.UnionResolveTypeOutput {
	var out driver.UnionResolveTypeOutput
	req, err := protodriver.MakeUnionResolveTypeRequest(input)
	if err == nil {
		resp := new(protoMessages.UnionResolveTypeResponse)
		if err = c.do(message{
			contentType: unionResolveTypeRequestMessage,
			proto:       req,
		}, message{
			contentType: unionResolveTypeResponseMessage,
			proto:       resp,
		}); err == nil {
			out = protodriver.MakeUnionResolveTypeOutput(resp)
		}
	}
	if err != nil {
		out.Error = &driver.Error{
			Message: err.Error(),
		}
	}
	return out
}

func (h *Handler) unionResolveType(req *http.Request) *protoMessages.UnionResolveTypeResponse {
	resp := new(protoMessages.UnionResolveTypeResponse)
	protoReq := new(protoMessages.UnionResolveTypeRequest)
	var err error
	var b []byte
	if b, err = ioutil.ReadAll(req.Body); err == nil {
		defer req.Body.Close()
		if err = protobuf.Unmarshal(b, protoReq); err == nil {
			var in driver.UnionResolveTypeInput
			in, err = protodriver.MakeUnionResolveTypeInput(protoReq)
			if err == nil {
				var unionType string
				unionType, err = h.UnionResolveType(in)
				if err == nil {
					*resp = protodriver.MakeUnionResolveTypeResponse(unionType)
				}
			}
		}
	}
	if err != nil {
		resp.Error = &protoMessages.Error{
			Msg: err.Error(),
		}
	}
	return resp
}
