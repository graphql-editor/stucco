package prototest

import (
	"fmt"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/types"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// InterfaceResolveTypeClientTest is basic struct for testing clients implementing proto
type InterfaceResolveTypeClientTest struct {
	Title         string
	Input         driver.InterfaceResolveTypeInput
	ProtoRequest  *protoMessages.InterfaceResolveTypeRequest
	ProtoResponse *protoMessages.InterfaceResolveTypeResponse
	ProtoError    error
	Expected      driver.InterfaceResolveTypeOutput
}

// InterfaceResolveTypeClientTestData is a data for testing interface resolution of proto clients
func InterfaceResolveTypeClientTestData() []InterfaceResolveTypeClientTest {
	return []InterfaceResolveTypeClientTest{
		{
			Title: "CallsProtoInterfaceResolveTypeInput",
			Input: driver.InterfaceResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			ProtoRequest: &protoMessages.InterfaceResolveTypeRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Value: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
				Info: &protoMessages.InterfaceResolveTypeInfo{},
			},
			ProtoResponse: &protoMessages.InterfaceResolveTypeResponse{
				Type: &protoMessages.TypeRef{
					TestTyperef: &protoMessages.TypeRef_Name{Name: "SomeType"},
				},
			},
			Expected: driver.InterfaceResolveTypeOutput{
				Type: types.TypeRef{
					Name: "SomeType",
				},
			},
		},
		{
			Title: "ErrorOnMissingFunction",
			Input: driver.InterfaceResolveTypeInput{},
			Expected: driver.InterfaceResolveTypeOutput{
				Error: &driver.Error{
					Message: "function name is required",
				},
			},
		},
		{
			Title: "PassthroughError",
			Input: driver.InterfaceResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			ProtoRequest: &protoMessages.InterfaceResolveTypeRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Value: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
				Info: &protoMessages.InterfaceResolveTypeInfo{},
			},
			ProtoError: fmt.Errorf("proto error"),
			Expected: driver.InterfaceResolveTypeOutput{
				Error: &driver.Error{
					Message: "proto error",
				},
			},
		},
		{
			Title: "PassthroughUserError",
			Input: driver.InterfaceResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			ProtoRequest: &protoMessages.InterfaceResolveTypeRequest{
				Function: &protoMessages.Function{
					Name: "function",
				},
				Value: &protoMessages.Value{
					TestValue: &protoMessages.Value_Nil{
						Nil: true,
					},
				},
				Info: &protoMessages.InterfaceResolveTypeInfo{},
			},
			ProtoResponse: &protoMessages.InterfaceResolveTypeResponse{
				Error: &protoMessages.Error{
					Msg: "user error",
				},
			},
			Expected: driver.InterfaceResolveTypeOutput{
				Error: &driver.Error{
					Message: "user error",
				},
			},
		},
	}
}

// RunInterfaceResolveTypeClientTests runs all client tests on a function
func RunInterfaceResolveTypeClientTests(t *testing.T, f func(t *testing.T, tt InterfaceResolveTypeClientTest)) {
	for _, tt := range InterfaceResolveTypeClientTestData() {
		t.Run(tt.Title, func(t *testing.T) {
			f(t, tt)
		})
	}
}

// InterfaceResolveTypeServerTest is basic struct for testing servers implementing proto
type InterfaceResolveTypeServerTest struct {
	Title         string
	Input         *protoMessages.InterfaceResolveTypeRequest
	HandlerInput  driver.InterfaceResolveTypeInput
	HandlerOutput string
	HandlerError  error
	Expected      *protoMessages.InterfaceResolveTypeResponse
}

// InterfaceResolveTypeServerTestData is a data for testing interface resolution of proto servers
func InterfaceResolveTypeServerTestData() []InterfaceResolveTypeServerTest {
	return []InterfaceResolveTypeServerTest{
		{
			Title:         "CallsUserHandler",
			Input:         new(protoMessages.InterfaceResolveTypeRequest),
			HandlerInput:  driver.InterfaceResolveTypeInput{},
			HandlerOutput: "SomeType",
			Expected: &protoMessages.InterfaceResolveTypeResponse{
				Type: &protoMessages.TypeRef{
					TestTyperef: &protoMessages.TypeRef_Name{Name: "SomeType"},
				},
			},
		},
		{
			Title:        "ReturnsUserError",
			Input:        new(protoMessages.InterfaceResolveTypeRequest),
			HandlerInput: driver.InterfaceResolveTypeInput{},
			HandlerError: fmt.Errorf("user error"),
			Expected: &protoMessages.InterfaceResolveTypeResponse{
				Error: &protoMessages.Error{
					Msg: "user error",
				},
			},
		},
	}
}

// RunInterfaceResolveTypeServerTests runs all client tests on a function
func RunInterfaceResolveTypeServerTests(t *testing.T, f func(t *testing.T, tt InterfaceResolveTypeServerTest)) {
	for _, tt := range InterfaceResolveTypeServerTestData() {
		t.Run(tt.Title, func(t *testing.T) {
			f(t, tt)
		})
	}
}
