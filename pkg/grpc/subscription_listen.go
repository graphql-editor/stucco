package grpc

import (
	"github.com/graphql-editor/stucco/pkg/driver"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
	protoDriverService "github.com/graphql-editor/stucco_proto/go/driver_service"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// SubscriptionListen returns a subscription event reader.
func (m *Client) SubscriptionListen(input driver.SubscriptionListenInput) (out driver.SubscriptionListenOutput) {
	req, err := protodriver.MakeSubscriptionListenRequest(input)
	if err == nil {
		out.Reader, err = protodriver.NewSubscriptionReader(m.Client, req)
	}
	if err != nil {
		out.Error = &driver.Error{Message: err.Error()}
	}
	return
}

type subscriptionListenEmitter struct {
	srv protoDriverService.Driver_SubscriptionListenServer
}

func (s subscriptionListenEmitter) Emit() error {
	return s.srv.Send(&protoMessages.SubscriptionListenMessage{
		Next: true,
	})
}

func (s subscriptionListenEmitter) Close() error {
	return s.srv.Send(&protoMessages.SubscriptionListenMessage{
		Next: false,
	})
}

// SubscriptionListen implements protoMessages.DriverServer
func (m *Server) SubscriptionListen(req *protoMessages.SubscriptionListenRequest, srv protoDriverService.Driver_SubscriptionListenServer) error {
	input, err := protodriver.MakeSubscriptionListenInput(req)
	if err == nil {
		err = m.SubscriptionListenHandler.Handle(input, subscriptionListenEmitter{
			srv: srv,
		})
	}
	return err
}
