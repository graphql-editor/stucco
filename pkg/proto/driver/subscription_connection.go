package protodriver

import (
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
)

// MakeSubscriptionConnectionRequest creates a new proto SubscriptionConnectionRequest from driver input
func MakeSubscriptionConnectionRequest(input driver.SubscriptionConnectionInput) (r *proto.SubscriptionConnectionRequest, err error) {
	ret := proto.SubscriptionConnectionRequest{
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
	proto, err := anyToValue(input.Protocol)
	if err == nil {
		ret.Protocol = proto
		r = &ret
	}
	return
}

// MakeSubscriptionConnectionOutput creates new driver.SubscriptionConnectionOutput from proto response
func MakeSubscriptionConnectionOutput(resp *proto.SubscriptionConnectionResponse) (out driver.SubscriptionConnectionOutput) {
	var err error
	out.Response, err = valueToAny(nil, resp.GetResponse())
	if err != nil {
		out.Error = &driver.Error{Message: err.Error()}
	} else if rerr := resp.GetError(); rerr != nil {
		out.Error = &driver.Error{Message: rerr.GetMsg()}
	}
	return out
}

// MakeSubscriptionConnectionInput creates driver.SubscriptionConnectionInput from proto.SubscriptionConnectionRequest
func MakeSubscriptionConnectionInput(input *proto.SubscriptionConnectionRequest) (f driver.SubscriptionConnectionInput, err error) {
	f = driver.SubscriptionConnectionInput{
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
			f = driver.SubscriptionConnectionInput{}
			return
		}
	}
	if pr := input.GetProtocol(); pr != nil {
		f.Protocol, err = valueToAny(nil, pr)
	}
	return
}

// MakeSubscriptionConnectionResponse creates a proto.SubscriptionConnectionRespone from a value
func MakeSubscriptionConnectionResponse(resp interface{}) proto.SubscriptionConnectionResponse {
	protoResponse := proto.SubscriptionConnectionResponse{}
	v, err := anyToValue(resp)
	if err == nil {
		protoResponse.Response = v
	} else {
		protoResponse.Error = &proto.Error{
			Msg: err.Error(),
		}
	}
	return protoResponse
}
