package protodriver

import (
	"context"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
)

// MakeSubscriptionListenRequest creates a new proto SubscriptionListenRequest from driver input
func MakeSubscriptionListenRequest(input driver.SubscriptionListenInput) (r *proto.SubscriptionListenRequest, err error) {
	ret := proto.SubscriptionListenRequest{
		Function: &proto.Function{
			Name: input.Function.Name,
		},
		Query:         input.Query,
		OperationName: input.OperationName,
	}
	for k, v := range input.VariableValues {
		if ret.VariableValues == nil {
			ret.VariableValues = make(map[string]*proto.Value)
		}
		ret.VariableValues[k], err = anyToValue(v)
		if err != nil {
			return
		}
	}
	ret.Operation, err = makeProtoOperationDefinition(input.Operation)
	if err == nil {
		ret.Protocol, err = anyToValue(input.Protocol)
	}
	if err == nil {
		r = &ret
	}
	return
}

type sigType struct {
	ok  bool
	v   interface{}
	err error
}

type subscriptionReader struct {
	ctx    context.Context
	cancel context.CancelFunc
	sigCh  chan sigType
	v      interface{}
	err    error
}

// NewSubscriptionReader creates new subscription reader for SubscriptionListen
func NewSubscriptionReader(client proto.DriverClient, req *proto.SubscriptionListenRequest) (driver.SubscriptionListenReader, error) {
	var r subscriptionReader
	r.ctx, r.cancel = context.WithCancel(context.Background())
	subClient, err := client.SubscriptionListen(r.ctx, req)
	if err != nil {
		return nil, err
	}
	r.sigCh = make(chan sigType, 10)
	go func() {
		for {
			m, err := subClient.Recv()
			sig := sigType{
				err: err,
			}
			if m != nil {
				sig.ok = m.Next
				if m.Payload != nil {
					var v interface{}
					v, verr := valueToAny(nil, m.Payload)
					sig.v = v
					if err == nil && verr != nil {
						sig.err = verr
					}
				}
			}
			r.sigCh <- sig
			if !sig.ok || sig.err != nil {
				close(r.sigCh)
				return
			}
		}
	}()
	return &r, nil
}

func (r *subscriptionReader) Error() error {
	if r.err == context.Canceled {
		r.err = nil
	}
	return r.err
}

func (r *subscriptionReader) Next() bool {
	sig := <-r.sigCh
	r.err = sig.err
	r.v = sig.v
	return sig.ok
}

func (r *subscriptionReader) Read() (interface{}, error) {
	return r.v, nil
}

func (r *subscriptionReader) Close() error {
	r.cancel()
	return nil
}

// MakeSubscriptionListenInput creates driver.SubscriptionListenInput from proto.SubscriptionListenRequest
func MakeSubscriptionListenInput(input *proto.SubscriptionListenRequest) (f driver.SubscriptionListenInput, err error) {
	f = driver.SubscriptionListenInput{
		Function: types.Function{
			Name: input.GetFunction().GetName(),
		},
		Query:         input.GetQuery(),
		OperationName: input.GetOperationName(),
	}
	for k, v := range input.GetVariableValues() {
		if f.VariableValues == nil {
			f.VariableValues = make(map[string]interface{})
		}
		f.VariableValues[k], err = valueToAny(nil, v)
		if err != nil {
			return
		}
	}
	if pr := input.GetProtocol(); pr != nil {
		f.Protocol, err = valueToAny(nil, pr)
	}
	return
}
