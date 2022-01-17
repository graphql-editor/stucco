package protohttp

import (
	"bytes"
	"net/http"

	"github.com/graphql-editor/stucco/pkg/driver"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// FieldResolve over http
func (c *Client) FieldResolve(input driver.FieldResolveInput) driver.FieldResolveOutput {
	var out driver.FieldResolveOutput
	var body bytes.Buffer
	err := protodriver.WriteFieldResolveInput(&body, input)
	if err == nil {
		var b []byte
		if b, err = c.do(message{
			contentType: fieldResolveRequestMessage,
			b:           body.Bytes(),
		}); err == nil {
			out, err = protodriver.ReadFieldResolveOutput(bytes.NewReader(b))
		}
	}
	if err != nil {
		out.Error = &driver.Error{
			Message: err.Error(),
		}
	}
	return out
}

func (h *Handler) fieldResolve(req *http.Request, rw http.ResponseWriter) error {
	rw.Header().Add(contentTypeHeader, fieldResolveResponseMessage.String())
	in, err := protodriver.ReadFieldResolveInput(req.Body)
	if err == nil {
		req.Body.Close()
		if err == nil {
			var driverResp interface{}
			driverResp, err = h.FieldResolve(in)
			if err == nil {
				err = protodriver.WriteFieldResolveOutput(rw, driverResp)
			}
		}
	}
	if err != nil {
		err = writeProto(rw, &protoMessages.FieldResolveResponse{
			Error: &protoMessages.Error{
				Msg: err.Error(),
			},
		})
	}
	return err
}
