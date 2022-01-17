package protohttp

import (
	"bytes"
	"net/http"

	"github.com/graphql-editor/stucco/pkg/driver"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// SetSecrets using http
func (c *Client) SetSecrets(input driver.SetSecretsInput) driver.SetSecretsOutput {
	var out driver.SetSecretsOutput
	var body bytes.Buffer
	err := protodriver.WriteSetSecretsInput(&body, input)
	if err == nil {
		var b []byte
		if b, err = c.do(message{
			contentType: fieldResolveRequestMessage,
			b:           body.Bytes(),
		}); err == nil {
			out, err = protodriver.ReadSetSecretsOutput(bytes.NewReader(b))
		}
	}
	if err != nil {
		out.Error = &driver.Error{
			Message: err.Error(),
		}
	}
	return out
}

func (h *Handler) setSecrets(req *http.Request, rw http.ResponseWriter) error {
	rw.Header().Add(contentTypeHeader, setSecretsResponseMessage.String())
	in, err := protodriver.ReadSetSecretsInput(req.Body)
	if err == nil {
		req.Body.Close()
		err = protodriver.WriteFieldResolveOutput(rw, h.SetSecrets(in))
	}
	if err != nil {
		err = writeProto(rw, &protoMessages.SetSecretsResponse{
			Error: &protoMessages.Error{
				Msg: err.Error(),
			},
		})
	}
	return err
}
