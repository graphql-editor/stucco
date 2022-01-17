package protohttp

import (
	"bytes"
	"net/http"

	"github.com/graphql-editor/stucco/pkg/driver"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// InterfaceResolveType over http
func (c *Client) InterfaceResolveType(input driver.InterfaceResolveTypeInput) driver.InterfaceResolveTypeOutput {
	var out driver.InterfaceResolveTypeOutput
	var body bytes.Buffer
	err := protodriver.WriteInterfaceResolveTypeInput(&body, input)
	if err == nil {
		var b []byte
		if b, err = c.do(message{
			contentType: fieldResolveRequestMessage,
			b:           body.Bytes(),
		}); err == nil {
			out, err = protodriver.ReadInterfaceResolveTypeOutput(bytes.NewReader(b))
		}
	}
	if err != nil {
		out.Error = &driver.Error{
			Message: err.Error(),
		}
	}
	return out
}

func (h *Handler) interfaceResolveType(req *http.Request, rw http.ResponseWriter) error {
	rw.Header().Add(contentTypeHeader, interfaceResolveTypeResponseMessage.String())
	in, err := protodriver.ReadInterfaceResolveTypeInput(req.Body)
	if err == nil {
		req.Body.Close()
		if err == nil {
			var driverResp string
			driverResp, err = h.InterfaceResolveType(in)
			if err == nil {
				err = protodriver.WriteInterfaceResolveTypeOutput(rw, driverResp)
			}
		}
	}
	if err != nil {
		err = writeProto(rw, &protoMessages.InterfaceResolveTypeResponse{
			Error: &protoMessages.Error{
				Msg: err.Error(),
			},
		})
	}
	return err
}
