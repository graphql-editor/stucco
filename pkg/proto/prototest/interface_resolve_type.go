package prototest

import (
	"fmt"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
	"github.com/stretchr/testify/assert"
)

// InterfaceResolveTypeClientTest is basic struct for testing clients implementing proto
type InterfaceResolveTypeClientTest struct {
	Title         string
	Input         driver.InterfaceResolveTypeInput
	ProtoRequest  *proto.InterfaceResolveTypeRequest
	ProtoResponse *proto.InterfaceResolveTypeResponse
	ProtoError    error
	Expected      driver.InterfaceResolveTypeOutput
	ExpectedErr   assert.ErrorAssertionFunc
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
			ProtoRequest: &proto.InterfaceResolveTypeRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Value: new(proto.Value),
				Info:  &proto.InterfaceResolveTypeInfo{},
			},
			ProtoResponse: &proto.InterfaceResolveTypeResponse{
				Type: &proto.TypeRef{
					TestTyperef: &proto.TypeRef_Name{Name: "SomeType"},
				},
			},
			Expected: driver.InterfaceResolveTypeOutput{
				Type: types.TypeRef{
					Name: "SomeType",
				},
			},
			ExpectedErr: assert.NoError,
		},
		{
			Title: "ErrorOnMissingFunction",
			Input: driver.InterfaceResolveTypeInput{},
			Expected: driver.InterfaceResolveTypeOutput{
				Error: &driver.Error{
					Message: "function name is required",
				},
			},
			ExpectedErr: assert.NoError,
		},
		{
			Title: "PassthroughError",
			Input: driver.InterfaceResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			ProtoRequest: &proto.InterfaceResolveTypeRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Value: new(proto.Value),
				Info:  &proto.InterfaceResolveTypeInfo{},
			},
			ProtoError: fmt.Errorf("proto error"),
			Expected: driver.InterfaceResolveTypeOutput{
				Error: &driver.Error{
					Message: "proto error",
				},
			},
			ExpectedErr: assert.NoError,
		},
		{
			Title: "PassthroughUserError",
			Input: driver.InterfaceResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			ProtoRequest: &proto.InterfaceResolveTypeRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Value: new(proto.Value),
				Info:  &proto.InterfaceResolveTypeInfo{},
			},
			ProtoResponse: &proto.InterfaceResolveTypeResponse{
				Error: &proto.Error{
					Msg: "user error",
				},
			},
			Expected: driver.InterfaceResolveTypeOutput{
				Error: &driver.Error{
					Message: "user error",
				},
			},
			ExpectedErr: assert.NoError,
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
	Input         *proto.InterfaceResolveTypeRequest
	HandlerInput  driver.InterfaceResolveTypeInput
	HandlerOutput string
	HandlerError  error
	Expected      *proto.InterfaceResolveTypeResponse
	ExpectedErr   assert.ErrorAssertionFunc
}

// InterfaceResolveTypeServerTestData is a data for testing interface resolution of proto servers
func InterfaceResolveTypeServerTestData() []InterfaceResolveTypeServerTest {
	return []InterfaceResolveTypeServerTest{
		{
			Title:         "CallsUserHandler",
			Input:         new(proto.InterfaceResolveTypeRequest),
			HandlerInput:  driver.InterfaceResolveTypeInput{},
			HandlerOutput: "SomeType",
			Expected: &proto.InterfaceResolveTypeResponse{
				Type: &proto.TypeRef{
					TestTyperef: &proto.TypeRef_Name{Name: "SomeType"},
				},
			},
			ExpectedErr: assert.NoError,
		},
		{
			Title:        "ReturnsUserError",
			Input:        new(proto.InterfaceResolveTypeRequest),
			HandlerInput: driver.InterfaceResolveTypeInput{},
			HandlerError: fmt.Errorf("user error"),
			Expected: &proto.InterfaceResolveTypeResponse{
				Error: &proto.Error{
					Msg: "user error",
				},
			},
			ExpectedErr: assert.NoError,
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
