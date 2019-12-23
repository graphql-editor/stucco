package prototest

import (
	"fmt"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
)

// UnionResolveTypeClientTest is basic struct for testing clients implementing proto
type UnionResolveTypeClientTest struct {
	Title         string
	Input         driver.UnionResolveTypeInput
	ProtoRequest  *proto.UnionResolveTypeRequest
	ProtoResponse *proto.UnionResolveTypeResponse
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
			ProtoRequest: &proto.UnionResolveTypeRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Value: new(proto.Value),
				Info:  &proto.UnionResolveTypeInfo{},
			},
			ProtoResponse: &proto.UnionResolveTypeResponse{
				Type: &proto.TypeRef{
					TestTyperef: &proto.TypeRef_Name{Name: "SomeType"},
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
			ProtoRequest: &proto.UnionResolveTypeRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Value: new(proto.Value),
				Info:  &proto.UnionResolveTypeInfo{},
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
			ProtoRequest: &proto.UnionResolveTypeRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Value: new(proto.Value),
				Info:  &proto.UnionResolveTypeInfo{},
			},
			ProtoResponse: &proto.UnionResolveTypeResponse{
				Error: &proto.Error{
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
	Input         *proto.UnionResolveTypeRequest
	HandlerInput  driver.UnionResolveTypeInput
	HandlerOutput string
	HandlerError  error
	Expected      *proto.UnionResolveTypeResponse
}

// UnionResolveTypeServerTestData is a data for testing union resolution of proto servers
func UnionResolveTypeServerTestData() []UnionResolveTypeServerTest {
	return []UnionResolveTypeServerTest{
		{
			Title:         "CallsUserHandler",
			Input:         new(proto.UnionResolveTypeRequest),
			HandlerInput:  driver.UnionResolveTypeInput{},
			HandlerOutput: "SomeType",
			Expected: &proto.UnionResolveTypeResponse{
				Type: &proto.TypeRef{
					TestTyperef: &proto.TypeRef_Name{Name: "SomeType"},
				},
			},
		},
		{
			Title:        "ReturnsUserError",
			Input:        new(proto.UnionResolveTypeRequest),
			HandlerInput: driver.UnionResolveTypeInput{},
			HandlerError: fmt.Errorf("user error"),
			Expected: &proto.UnionResolveTypeResponse{
				Error: &proto.Error{
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
