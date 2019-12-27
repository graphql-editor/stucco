package protodriver

import (
	"fmt"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
)

// MakeScalarParseRequest creates new proto.ScalarParseRequest from driver.ScalarParseInput
func MakeScalarParseRequest(input driver.ScalarParseInput) (req *proto.ScalarParseRequest, err error) {
	v, err := anyToValue(input.Value)
	if err == nil {
		req = &proto.ScalarParseRequest{
			Function: &proto.Function{
				Name: input.Function.Name,
			},
			Value: v,
		}
	}
	return
}

// MakeScalarParseOutput creates new driver.ScalarParseOutput from proto.ScalarParseResponse
func MakeScalarParseOutput(resp *proto.ScalarParseResponse) driver.ScalarParseOutput {
	var out driver.ScalarParseOutput
	var err error
	var r interface{}
	if respErr := resp.GetError(); respErr != nil {
		err = fmt.Errorf(respErr.GetMsg())
	} else {
		r, err = valueToAny(nil, resp.GetValue())
		if err == nil {
			out.Response = r
		}
	}
	if err != nil {
	}
	if err != nil {
		out.Error = &driver.Error{Message: err.Error()}
	}
	return out
}

// MakeScalarSerializeRequest creates new proto.ScalarSerializeRequest from driver.ScalarSerializeInput
func MakeScalarSerializeRequest(input driver.ScalarSerializeInput) (req *proto.ScalarSerializeRequest, err error) {
	v, err := anyToValue(input.Value)
	if err == nil {
		req = &proto.ScalarSerializeRequest{
			Function: &proto.Function{
				Name: input.Function.Name,
			},
			Value: v,
		}
	}
	return
}

// MakeScalarSerializeOutput creates new driver.ScalarSerializeOutput from proto.ScalarSerializeResponse
func MakeScalarSerializeOutput(resp *proto.ScalarSerializeResponse) driver.ScalarSerializeOutput {
	var out driver.ScalarSerializeOutput
	var err error
	var r interface{}
	if respErr := resp.GetError(); respErr != nil {
		err = fmt.Errorf(respErr.GetMsg())
	} else {
		r, err = valueToAny(nil, resp.GetValue())
		if err == nil {
			out.Response = r
		}
	}
	if err != nil {
	}
	if err != nil {
		out.Error = &driver.Error{Message: err.Error()}
	}
	return out
}

// MakeScalarParseInput creates new driver.ScalarParseInput from proto.ScalarParseRequest
func MakeScalarParseInput(req *proto.ScalarParseRequest) (driver.ScalarParseInput, error) {
	var input driver.ScalarParseInput
	val, err := valueToAny(nil, req.GetValue())
	if err == nil {
		input = driver.ScalarParseInput{
			Function: types.Function{
				Name: req.GetFunction().GetName(),
			},
			Value: val,
		}
	}
	return input, err
}

// MakeScalarParseResponse creates new proto.ScalarParseResposne from any value
func MakeScalarParseResponse(value interface{}) proto.ScalarParseResponse {
	var protoResponse proto.ScalarParseResponse
	v, err := anyToValue(value)
	if err != nil {
		protoResponse.Error = &proto.Error{
			Msg: err.Error(),
		}
	} else {
		protoResponse.Value = v
	}
	return protoResponse
}

// MakeScalarSerializeInput creates new driver.ScalarSerializeInput from proto.ScalarSerializeRequest
func MakeScalarSerializeInput(req *proto.ScalarSerializeRequest) (driver.ScalarSerializeInput, error) {
	var input driver.ScalarSerializeInput
	val, err := valueToAny(nil, req.GetValue())
	if err == nil {
		input = driver.ScalarSerializeInput{
			Function: types.Function{
				Name: req.GetFunction().GetName(),
			},
			Value: val,
		}
	}
	return input, err
}

// MakeScalarSerializeResponse creates new proto.ScalarSerializeResposne from any value
func MakeScalarSerializeResponse(value interface{}) proto.ScalarSerializeResponse {
	var protoResponse proto.ScalarSerializeResponse
	v, err := anyToValue(value)
	if err != nil {
		protoResponse.Error = &proto.Error{
			Msg: err.Error(),
		}
	} else {
		protoResponse.Value = v
	}
	return protoResponse
}
