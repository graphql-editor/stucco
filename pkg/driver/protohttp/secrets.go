package protohttp

import (
	"io/ioutil"
	"net/http"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
)

// SetSecrets using http
func (c *Client) SetSecrets(input driver.SetSecretsInput) driver.SetSecretsOutput {
	var out driver.SetSecretsOutput
	req := protodriver.MakeSetSecretsRequest(input)
	resp := new(proto.SetSecretsResponse)
	var err error
	if err = c.do(message{
		contentType: setSecretsRequestMessage,
		proto:       req,
	}, message{
		contentType: setSecretsResponseMessage,
		proto:       resp,
	}); err == nil {
		out = protodriver.MakeSetSecretsOutput(resp)
	}
	if err != nil {
		out.Error = &driver.Error{
			Message: err.Error(),
		}
	}
	return out
}

func (h *Handler) setSecrets(req *http.Request) *proto.SetSecretsResponse {
	resp := new(proto.SetSecretsResponse)
	protoReq := new(proto.SetSecretsRequest)
	var err error
	var b []byte
	if b, err = ioutil.ReadAll(req.Body); err == nil {
		defer req.Body.Close()
		if err = protobuf.Unmarshal(b, protoReq); err == nil {
			err = h.SetSecrets(protodriver.MakeSetSecretsInput(protoReq))
		}
	}
	if err != nil {
		resp.Error = &proto.Error{
			Msg: err.Error(),
		}
	}
	return resp
}
