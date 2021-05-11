package protohttp

import (
	"io/ioutil"
	"net/http"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// SetSecrets using http
func (c *Client) SetSecrets(input driver.SetSecretsInput) driver.SetSecretsOutput {
	var out driver.SetSecretsOutput
	req := protodriver.MakeSetSecretsRequest(input)
	resp := new(protoMessages.SetSecretsResponse)
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

func (h *Handler) setSecrets(req *http.Request) *protoMessages.SetSecretsResponse {
	resp := new(protoMessages.SetSecretsResponse)
	protoReq := new(protoMessages.SetSecretsRequest)
	var err error
	var b []byte
	if b, err = ioutil.ReadAll(req.Body); err == nil {
		defer req.Body.Close()
		if err = protobuf.Unmarshal(b, protoReq); err == nil {
			err = h.SetSecrets(protodriver.MakeSetSecretsInput(protoReq))
		}
	}
	if err != nil {
		resp.Error = &protoMessages.Error{
			Msg: err.Error(),
		}
	}
	return resp
}
