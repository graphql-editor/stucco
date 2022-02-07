package protohttp

import (
	"bytes"
	"net/http"

	"github.com/graphql-editor/stucco/pkg/driver"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// UnionResolveType over http
func (c *Client) UnionResolveType(input driver.UnionResolveTypeInput) driver.UnionResolveTypeOutput {
	var out driver.UnionResolveTypeOutput
	var body bytes.Buffer
	err := protodriver.WriteUnionResolveTypeInput(&body, input)
	if err == nil {
		var b []byte
		if b, err = c.do(message{
			contentType:         unionResolveTypeRequestMessage,
			responseContentType: unionResolveTypeResponseMessage,
			b:                   body.Bytes(),
		}); err == nil {
			out, err = protodriver.ReadUnionResolveTypeOutput(bytes.NewReader(b))
		}
	}
	if err != nil {
		out.Error = &driver.Error{
			Message: err.Error(),
		}
	}
	return out
}

func (h *Handler) unionResolveType(req *http.Request, rw http.ResponseWriter) error {
	rw.Header().Add(contentTypeHeader, interfaceResolveTypeResponseMessage.String())
	in, err := protodriver.ReadUnionResolveTypeInput(req.Body)
	if err == nil {
		req.Body.Close()
		if err == nil {
			var driverResp string
			driverResp, err = h.UnionResolveType(in)
			if err == nil {
				err = protodriver.WriteUnionResolveTypeOutput(rw, driverResp)
			}
		}
	}
	if err != nil {
		err = writeProto(rw, &protoMessages.UnionResolveTypeResponse{
			Error: &protoMessages.Error{
				Msg: err.Error(),
			},
		})
	}
	return err
}
