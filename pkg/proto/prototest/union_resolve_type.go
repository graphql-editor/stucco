package prototest

import (
	"fmt"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/types"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// UnionResolveTypeClientTest is basic struct for testing clients implementing proto
type UnionResolveTypeClientTest struct {
	Title         string
	Input         driver.UnionResolveTypeInput
	ProtoRequest  *protoMessages.UnionResolveTypeRequest
	ProtoResponse *protoMessages.UnionResolveTypeResponse
	ProtoError    error
	Expected      driver.UnionResolveTypeOutput
}

// UnionResolveTypeClientTestData is a data for testing union resolution of proto clients
func UnionResolveTypeClientTestData() []UnionResolveTypeClientTest {
	return []UnionResolveTypeClientTest{
		{
			Title: "CallsProtoUnionResolveTypeInput",
			Input: driver.UnionResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			ProtoRequest: &protoMessages.UnionResolveTypeRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Value: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
				Info: &protoMessages.UnionResolveTypeInfo{},
			},
			ProtoResponse: &protoMessages.UnionResolveTypeResponse{
				Type: &protoMessages.TypeRef{
					TestTyperef: &protoMessages.TypeRef_Name{Name: "SomeType"},
				},
			},
			Expected: driver.UnionResolveTypeOutput{
				Type: types.TypeRef{
					Name: "SomeType",
				},
			},
		},
		{
			Title: "ErrorOnMissingFunction",
			Input: driver.UnionResolveTypeInput{},
			Expected: driver.UnionResolveTypeOutput{
				Error: &driver.Error{
					Message: "function name is required",
				},
			},
		},
		{
			Title: "PassthroughError",
			Input: driver.UnionResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			ProtoRequest: &protoMessages.UnionResolveTypeRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Value: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
				Info: &protoMessages.UnionResolveTypeInfo{},
			},
			ProtoError: fmt.Errorf("proto error"),
			Expected: driver.UnionResolveTypeOutput{
				Error: &driver.Error{
					Message: "proto error",
				},
			},
		},
		{
			Title: "PassthroughUserError",
			Input: driver.UnionResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			ProtoRequest: &protoMessages.UnionResolveTypeRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Value: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
				Info: &protoMessages.UnionResolveTypeInfo{},
			},
			ProtoResponse: &protoMessages.UnionResolveTypeResponse{
				Error: &protoMessages.Error{
					Msg: "user error",
				},
			},
			Expected: driver.UnionResolveTypeOutput{
				Error: &driver.Error{
					Message: "user error",
				},
			},
		},
	}
}

// RunUnionResolveTypeClientTests runs all client tests on a function
func RunUnionResolveTypeClientTests(t *testing.T, f func(t *testing.T, tt UnionResolveTypeClientTest)) {
	for _, tt := range UnionResolveTypeClientTestData() {
		t.Run(tt.Title, func(t *testing.T) {
			f(t, tt)
		})
	}
}

// UnionResolveTypeServerTest is basic struct for testing servers implementing proto
type UnionResolveTypeServerTest struct {
	Title         string
	Input         *protoMessages.UnionResolveTypeRequest
	HandlerInput  driver.UnionResolveTypeInput
	HandlerOutput string
	HandlerError  error
	Expected      *protoMessages.UnionResolveTypeResponse
}

// UnionResolveTypeServerTestData is a data for testing union resolution of proto servers
func UnionResolveTypeServerTestData() []UnionResolveTypeServerTest {
	return []UnionResolveTypeServerTest{
		{
			Title:         "CallsUserHandler",
			Input:         new(protoMessages.UnionResolveTypeRequest),
			HandlerInput:  driver.UnionResolveTypeInput{},
			HandlerOutput: "SomeType",
			Expected: &protoMessages.UnionResolveTypeResponse{
				Type: &protoMessages.TypeRef{
					TestTyperef: &protoMessages.TypeRef_Name{Name: "SomeType"},
				},
			},
		},
		{
			Title:        "ReturnsUserError",
			Input:        new(protoMessages.UnionResolveTypeRequest),
			HandlerInput: driver.UnionResolveTypeInput{},
			HandlerError: fmt.Errorf("user error"),
			Expected: &protoMessages.UnionResolveTypeResponse{
				Error: &protoMessages.Error{
					Msg: "user error",
				},
			},
		},
	}
}

// RunUnionResolveTypeServerTests runs all client tests on a function
func RunUnionResolveTypeServerTests(t *testing.T, f func(t *testing.T, tt UnionResolveTypeServerTest)) {
	for _, tt := range UnionResolveTypeServerTestData() {
		t.Run(tt.Title, func(t *testing.T) {
			f(t, tt)
		})
	}
}
