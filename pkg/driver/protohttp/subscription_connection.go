package protohttp

import (
	"bytes"
	"net/http"

	"github.com/graphql-editor/stucco/pkg/driver"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// SubscriptionConnection implements driver.SubscriptionConnection over HTTP
func (c *Client) SubscriptionConnection(input driver.SubscriptionConnectionInput) driver.SubscriptionConnectionOutput {
	var out driver.SubscriptionConnectionOutput
	var body bytes.Buffer
	err := protodriver.WriteSubscriptionConnectionInput(&body, input)
	if err == nil {
		var b []byte
		if b, err = c.do(message{
			contentType: fieldResolveRequestMessage,
			b:           body.Bytes(),
		}); err == nil {
			out, err = protodriver.ReadSubscriptionConnectionOutput(bytes.NewReader(b))
		}
	}
	if err != nil {
		out.Error = &driver.Error{
			Message: err.Error(),
		}
	}
	return out
}

func (h *Handler) subscriptionConnection(req *http.Request, rw http.ResponseWriter) error {
	rw.Header().Add(contentTypeHeader, subscriptionConnectionResponseMessage.String())
	in, err := protodriver.ReadSubscriptionConnectionInput(req.Body)
	if err == nil {
		req.Body.Close()
		if err == nil {
			var driverResp interface{}
			driverResp, err = h.SubscriptionConnection(in)
			if err == nil {
				err = protodriver.WriteSubscriptionConnectionOutput(rw, driverResp)
			}
		}
	}
	if err != nil {
		err = writeProto(rw, &protoMessages.SubscriptionConnectionResponse{
			Error: &protoMessages.Error{
				Msg: err.Error(),
			},
		})
	}
	return err
}
