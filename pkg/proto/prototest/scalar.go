package prototest

import (
	"fmt"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/types"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// ScalarParseClientTest is basic struct for testing clients implementing proto
type ScalarParseClientTest struct {
	Title         string
	Input         driver.ScalarParseInput
	ProtoRequest  *protoMessages.ScalarParseRequest
	ProtoResponse *protoMessages.ScalarParseResponse
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
			ProtoRequest: &protoMessages.ScalarParseRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Value: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
			},
			ProtoResponse: &protoMessages.ScalarParseResponse{Value: &protoMessages.Value{
				TestValue: &protoMessages.Value_S{
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
			ProtoRequest: &protoMessages.ScalarParseRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Value: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
			},
			ProtoResponse: &protoMessages.ScalarParseResponse{
				Error: &protoMessages.Error{
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
	Input         *protoMessages.ScalarParseRequest
	HandlerInput  driver.ScalarParseInput
	HandlerOutput interface{}
	HandlerError  error
	Expected      *protoMessages.ScalarParseResponse
}

// ScalarParseServerTestData is a data for testing scalar resolution of proto servers
func ScalarParseServerTestData() []ScalarParseServerTest {
	return []ScalarParseServerTest{
		{
			Title: "CallsScalarParseHandler",
			Input: &protoMessages.ScalarParseRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
			},
			HandlerInput: driver.ScalarParseInput{
				Function: types.Function{
					Name: "function",
				},
			},
			HandlerOutput: "scalar",
			Expected: &protoMessages.ScalarParseResponse{Value: &protoMessages.Value{
				TestValue: &protoMessages.Value_S{
					S: "scalar",
				},
			}},
		},
		{
			Title: "ReturnsUserError",
			Input: &protoMessages.ScalarParseRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
			},
			HandlerInput: driver.ScalarParseInput{
				Function: types.Function{
					Name: "function",
				},
			},
			HandlerError: fmt.Errorf("user error"),
			Expected: &protoMessages.ScalarParseResponse{
				Error: &protoMessages.Error{
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
	ProtoRequest  *protoMessages.ScalarSerializeRequest
	ProtoResponse *protoMessages.ScalarSerializeResponse
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
			ProtoRequest: &protoMessages.ScalarSerializeRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Value: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
			},
			ProtoResponse: &protoMessages.ScalarSerializeResponse{Value: &protoMessages.Value{
				TestValue: &protoMessages.Value_S{
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
			ProtoRequest: &protoMessages.ScalarSerializeRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Value: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
			},
			ProtoResponse: &protoMessages.ScalarSerializeResponse{
				Error: &protoMessages.Error{
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
	Input         *protoMessages.ScalarSerializeRequest
	HandlerInput  driver.ScalarSerializeInput
	HandlerOutput interface{}
	HandlerError  error
	Expected      *protoMessages.ScalarSerializeResponse
}

// ScalarSerializeServerTestData is a data for testing scalar resolution of proto servers
func ScalarSerializeServerTestData() []ScalarSerializeServerTest {
	return []ScalarSerializeServerTest{
		{
			Title: "CallsScalarSerializeHandler",
			Input: &protoMessages.ScalarSerializeRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
			},
			HandlerInput: driver.ScalarSerializeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			HandlerOutput: "scalar",
			Expected: &protoMessages.ScalarSerializeResponse{Value: &protoMessages.Value{
				TestValue: &protoMessages.Value_S{
					S: "scalar",
				},
			}},
		},
		{
			Title: "ReturnsUserError",
			Input: &protoMessages.ScalarSerializeRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
			},
			HandlerInput: driver.ScalarSerializeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			HandlerError: fmt.Errorf("user error"),
			Expected: &protoMessages.ScalarSerializeResponse{
				Error: &protoMessages.Error{
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
