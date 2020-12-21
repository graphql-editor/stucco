package grpc

import (
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	protodriver "github.com/graphql-editor/stucco/pkg/proto/driver"
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
	srv proto.Driver_SubscriptionListenServer
}

func (s subscriptionListenEmitter) Emit() error {
	return s.srv.Send(&proto.SubscriptionListenMessage{
		Next: true,
	})
}

func (s subscriptionListenEmitter) Close() error {
	return s.srv.Send(&proto.SubscriptionListenMessage{
		Next: false,
	})
}

// SubscriptionListen implements proto.DriverServer
func (m *Server) SubscriptionListen(req *proto.SubscriptionListenRequest, srv proto.Driver_SubscriptionListenServer) error {
	input, err := protodriver.MakeSubscriptionListenInput(req)
	if err == nil {
		err = m.SubscriptionListenHandler.Handle(input, subscriptionListenEmitter{
			srv: srv,
		})
	}
	return err
}
