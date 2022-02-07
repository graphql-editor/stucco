package protohttp

import (
	"bytes"
	"net/http"

	"github.com/graphql-editor/stucco/pkg/driver"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// Authorize over http
func (c *Client) Authorize(input driver.AuthorizeInput) driver.AuthorizeOutput {
	var out driver.AuthorizeOutput
	var body bytes.Buffer
	err := protodriver.WriteAuthorizeInput(&body, input)
	if err == nil {
		var b []byte
		if b, err = c.do(message{
			contentType:         authorizeRequestMessage,
			responseContentType: authorizeResponseMessage,
			b:                   body.Bytes(),
		}); err == nil {
			out, err = protodriver.ReadAuthorizeOutput(bytes.NewReader(b))
		}
	}
	if err != nil {
		out.Error = &driver.Error{
			Message: err.Error(),
		}
	}
	return out
}

func (h *Handler) authorize(req *http.Request, rw http.ResponseWriter) error {
	rw.Header().Add(contentTypeHeader, authorizeResponseMessage.String())
	in, err := protodriver.ReadAuthorizeInput(req.Body)
	if err == nil {
		req.Body.Close()
		if err == nil {
			var driverResp bool
			driverResp, err = h.Authorize(in)
			if err == nil {
				err = protodriver.WriteAuthorizeOutput(rw, driverResp)
			}
		}
	}
	if err != nil {
		err = writeProto(rw, &protoMessages.AuthorizeResponse{
			Error: &protoMessages.Error{
				Msg: err.Error(),
			},
		})
	}
	return err
}
