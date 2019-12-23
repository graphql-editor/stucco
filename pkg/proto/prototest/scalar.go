package prototest

import (
	"fmt"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
)

// ScalarParseClientTest is basic struct for testing clients implementing proto
type ScalarParseClientTest struct {
	Title         string
	Input         driver.ScalarParseInput
	ProtoRequest  *proto.ScalarParseRequest
	ProtoResponse *proto.ScalarParseResponse
	ProtoError    error
	Expected      driver.ScalarParseOutput
}

// ScalarParseClientTestData is a data for testing scalar resolution of proto clients
func ScalarParseClientTestData() []ScalarParseClientTest {
	return []ScalarParseClientTest{
		{
			Title: "CallsScalarParse",
			Input: driver.ScalarParseInput{
				Function: types.Function{
					Name: "function",
				},
			},
			ProtoRequest: &proto.ScalarParseRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			ProtoResponse: &proto.ScalarParseResponse{Value: &proto.Value{
				TestValue: &proto.Value_S{
					S: "scalar",
				},
			}},
			Expected: driver.ScalarParseOutput{Response: "scalar"},
		},
		{
			Title: "ReturnsUserError",
			Input: driver.ScalarParseInput{
				Function: types.Function{
					Name: "function",
				},
			},
			ProtoRequest: &proto.ScalarParseRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			ProtoResponse: &proto.ScalarParseResponse{
				Error: &proto.Error{
					Msg: "user error",
				},
			},
			Expected: driver.ScalarParseOutput{
				Error: &driver.Error{
					Message: "user error",
				},
			},
		},
	}
}

// RunScalarParseClientTests runs all client tests on a function
func RunScalarParseClientTests(t *testing.T, f func(t *testing.T, tt ScalarParseClientTest)) {
	for _, tt := range ScalarParseClientTestData() {
		t.Run(tt.Title, func(t *testing.T) {
			f(t, tt)
		})
	}
}

// ScalarParseServerTest is basic struct for testing servers implementing proto
type ScalarParseServerTest struct {
	Title         string
	Input         *proto.ScalarParseRequest
	HandlerInput  driver.ScalarParseInput
	HandlerOutput interface{}
	HandlerError  error
	Expected      *proto.ScalarParseResponse
}

// ScalarParseServerTestData is a data for testing scalar resolution of proto servers
func ScalarParseServerTestData() []ScalarParseServerTest {
	return []ScalarParseServerTest{
		{
			Title: "CallsScalarParseHandler",
			Input: &proto.ScalarParseRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			HandlerInput: driver.ScalarParseInput{
				Function: types.Function{
					Name: "function",
				},
			},
			HandlerOutput: "scalar",
			Expected: &proto.ScalarParseResponse{Value: &proto.Value{
				TestValue: &proto.Value_S{
					S: "scalar",
				},
			}},
		},
		{
			Title: "ReturnsUserError",
			Input: &proto.ScalarParseRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			HandlerInput: driver.ScalarParseInput{
				Function: types.Function{
					Name: "function",
				},
			},
			HandlerError: fmt.Errorf("user error"),
			Expected: &proto.ScalarParseResponse{
				Error: &proto.Error{
					Msg: "user error",
				},
			},
		},
	}
}

// RunScalarParseServerTests runs all client tests on a function
func RunScalarParseServerTests(t *testing.T, f func(t *testing.T, tt ScalarParseServerTest)) {
	for _, tt := range ScalarParseServerTestData() {
		t.Run(tt.Title, func(t *testing.T) {
			f(t, tt)
		})
	}
}

// ScalarSerializeClientTest is basic struct for testing clients implementing proto
type ScalarSerializeClientTest struct {
	Title         string
	Input         driver.ScalarSerializeInput
	ProtoRequest  *proto.ScalarSerializeRequest
	ProtoResponse *proto.ScalarSerializeResponse
	ProtoError    error
	Expected      driver.ScalarSerializeOutput
}

// ScalarSerializeClientTestData is a data for testing scalar resolution of proto clients
func ScalarSerializeClientTestData() []ScalarSerializeClientTest {
	return []ScalarSerializeClientTest{
		{
			Title: "CallsScalarSerialize",
			Input: driver.ScalarSerializeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			ProtoRequest: &proto.ScalarSerializeRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			ProtoResponse: &proto.ScalarSerializeResponse{Value: &proto.Value{
				TestValue: &proto.Value_S{
					S: "scalar",
				},
			}},
			Expected: driver.ScalarSerializeOutput{Response: "scalar"},
		},
		{
			Title: "ReturnsUserError",
			Input: driver.ScalarSerializeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			ProtoRequest: &proto.ScalarSerializeRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			ProtoResponse: &proto.ScalarSerializeResponse{
				Error: &proto.Error{
					Msg: "user error",
				},
			},
			Expected: driver.ScalarSerializeOutput{
				Error: &driver.Error{
					Message: "user error",
				},
			},
		},
	}
}

// RunScalarSerializeClientTests runs all client tests on a function
func RunScalarSerializeClientTests(t *testing.T, f func(t *testing.T, tt ScalarSerializeClientTest)) {
	for _, tt := range ScalarSerializeClientTestData() {
		t.Run(tt.Title, func(t *testing.T) {
			f(t, tt)
		})
	}
}

// ScalarSerializeServerTest is basic struct for testing servers implementing proto
type ScalarSerializeServerTest struct {
	Title         string
	Input         *proto.ScalarSerializeRequest
	HandlerInput  driver.ScalarSerializeInput
	HandlerOutput interface{}
	HandlerError  error
	Expected      *proto.ScalarSerializeResponse
}

// ScalarSerializeServerTestData is a data for testing scalar resolution of proto servers
func ScalarSerializeServerTestData() []ScalarSerializeServerTest {
	return []ScalarSerializeServerTest{
		{
			Title: "CallsScalarSerializeHandler",
			Input: &proto.ScalarSerializeRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			HandlerInput: driver.ScalarSerializeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			HandlerOutput: "scalar",
			Expected: &proto.ScalarSerializeResponse{Value: &proto.Value{
				TestValue: &proto.Value_S{
					S: "scalar",
				},
			}},
		},
		{
			Title: "ReturnsUserError",
			Input: &proto.ScalarSerializeRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			HandlerInput: driver.ScalarSerializeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			HandlerError: fmt.Errorf("user error"),
			Expected: &proto.ScalarSerializeResponse{
				Error: &proto.Error{
					Msg: "user error",
				},
			},
		},
	}
}

// RunScalarSerializeServerTests runs all client tests on a function
func RunScalarSerializeServerTests(t *testing.T, f func(t *testing.T, tt ScalarSerializeServerTest)) {
	for _, tt := range ScalarSerializeServerTestData() {
		t.Run(tt.Title, func(t *testing.T) {
			f(t, tt)
		})
	}
}
