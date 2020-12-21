package protohttp

import (
	"io/ioutil"
	"net/http"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
)

// SubscriptionConnection implements driver.SubscriptionConnection over HTTP
func (c *Client) SubscriptionConnection(input driver.SubscriptionConnectionInput) driver.SubscriptionConnectionOutput {
	var out driver.SubscriptionConnectionOutput
	req, err := protodriver.MakeSubscriptionConnectionRequest(input)
	if err == nil {
		resp := new(proto.SubscriptionConnectionResponse)
		if err = c.do(message{
			contentType: subscriptionConnectionRequestMessage,
			proto:       req,
		}, message{
			contentType: subscriptionConnectionResponseMessage,
			proto:       resp,
		}); err == nil {
			out = protodriver.MakeSubscriptionConnectionOutput(resp)
		}
	}
	if err != nil {
		out.Error = &driver.Error{
			Message: err.Error(),
		}
	}
	return out
}

func (h *Handler) subscriptionConnection(req *http.Request) *proto.SubscriptionConnectionResponse {
	resp := new(proto.SubscriptionConnectionResponse)
	protoReq := new(proto.SubscriptionConnectionRequest)
	var err error
	var b []byte
	if b, err = ioutil.ReadAll(req.Body); err == nil {
		defer req.Body.Close()
		if err = protobuf.Unmarshal(b, protoReq); err == nil {
			var in driver.SubscriptionConnectionInput
			in, err = protodriver.MakeSubscriptionConnectionInput(protoReq)
			if err == nil {
				var driverResp interface{}
				driverResp, err = h.SubscriptionConnection(in)
				if err == nil {
					*resp = protodriver.MakeSubscriptionConnectionResponse(driverResp)
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
