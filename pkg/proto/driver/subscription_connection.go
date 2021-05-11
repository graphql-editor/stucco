package protodriver

import (
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/types"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// MakeSubscriptionConnectionRequest creates a new proto SubscriptionConnectionRequest from driver input
func MakeSubscriptionConnectionRequest(input driver.SubscriptionConnectionInput) (r *protoMessages.SubscriptionConnectionRequest, err error) {
	ret := protoMessages.SubscriptionConnectionRequest{
		Function: &protoMessages.Function{
			Name: input.Function.Name,
		},
		Query:         input.Query,
		OperationName: input.OperationName,
	}
	for k, v := range input.VariableValues {
		if ret.VariableValues == nil {
			ret.VariableValues = make(map[string]*protoMessages.Value)
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
func MakeSubscriptionConnectionOutput(resp *protoMessages.SubscriptionConnectionResponse) (out driver.SubscriptionConnectionOutput) {
	var err error
	out.Response, err = valueToAny(nil, resp.GetResponse())
	if err != nil {
		out.Error = &driver.Error{Message: err.Error()}
	} else if rerr := resp.GetError(); rerr != nil {
		out.Error = &driver.Error{Message: rerr.GetMsg()}
	}
	return out
}

// MakeSubscriptionConnectionInput creates driver.SubscriptionConnectionInput from protoMessages.SubscriptionConnectionRequest
func MakeSubscriptionConnectionInput(input *protoMessages.SubscriptionConnectionRequest) (f driver.SubscriptionConnectionInput, err error) {
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

// MakeSubscriptionConnectionResponse creates a protoMessages.SubscriptionConnectionRespone from a value
func MakeSubscriptionConnectionResponse(resp interface{}) protoMessages.SubscriptionConnectionResponse {
	protoResponse := protoMessages.SubscriptionConnectionResponse{}
	v, err := anyToValue(resp)
	if err == nil {
		protoResponse.Response = v
	} else {
		protoResponse.Error = &protoMessages.Error{
			Msg: err.Error(),
		}
	}
	return protoResponse
}
