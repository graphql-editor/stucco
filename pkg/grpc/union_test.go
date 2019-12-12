package grpc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/grpc"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClientUnionResolveType(t *testing.T) {
	data := []struct {
		title        string
		input        driver.UnionResolveTypeInput
		grpcRequest  *proto.UnionResolveTypeRequest
		grpcResposne *proto.UnionResolveTypeResponse
		grpcError    error
		expected     driver.UnionResolveTypeOutput
		expectedErr  assert.ErrorAssertionFunc
	}{
		{
			title: "CallsGRPCUnionResolveTypeInput",
			input: driver.UnionResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			grpcRequest: &proto.UnionResolveTypeRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Value: new(proto.Value),
				Info:  &proto.UnionResolveTypeInfo{},
			},
			grpcResposne: &proto.UnionResolveTypeResponse{
				Type: &proto.TypeRef{
					TestTyperef: &proto.TypeRef_Name{Name: "SomeType"},
				},
			},
			expected: driver.UnionResolveTypeOutput{
				Type: types.TypeRef{
					Name: "SomeType",
				},
			},
			expectedErr: assert.NoError,
		},
		{
			title: "ErrorOnMissingFunction",
			input: driver.UnionResolveTypeInput{},
			expected: driver.UnionResolveTypeOutput{
				Error: &driver.Error{
					Message: "function name is required",
				},
			},
			expectedErr: assert.NoError,
		},
		{
			title: "PassthroughGRPCError",
			input: driver.UnionResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			grpcRequest: &proto.UnionResolveTypeRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Value: new(proto.Value),
				Info:  &proto.UnionResolveTypeInfo{},
			},
			grpcError: fmt.Errorf("grpc error"),
			expected: driver.UnionResolveTypeOutput{
				Error: &driver.Error{
					Message: "grpc error",
				},
			},
			expectedErr: assert.NoError,
		},
		{
			title: "PassthroughUserError",
			input: driver.UnionResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			grpcRequest: &proto.UnionResolveTypeRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Value: new(proto.Value),
				Info:  &proto.UnionResolveTypeInfo{},
			},
			grpcResposne: &proto.UnionResolveTypeResponse{
				Error: &proto.Error{
					Msg: "user error",
				},
			},
			expected: driver.UnionResolveTypeOutput{
				Error: &driver.Error{
					Message: "user error",
				},
			},
			expectedErr: assert.NoError,
		},
	}
	for i := range data {
		tt := data[i]
		t.Run(tt.title, func(t *testing.T) {
			driverClientMock := new(driverClientMock)
			driverClientMock.On(
				"UnionResolveType",
				mock.Anything,
				tt.grpcRequest,
			).Return(tt.grpcResposne, tt.grpcError)
			client := grpc.Client{
				Client: driverClientMock,
			}
			out, err := client.UnionResolveType(tt.input)
			tt.expectedErr(t, err)
			assert.Equal(t, tt.expected, out)
		})
	}
}

func TestServerUnionResolveType(t *testing.T) {
	data := []struct {
		title         string
		input         *proto.UnionResolveTypeRequest
		handlerInput  driver.UnionResolveTypeInput
		handlerOutput string
		handlerError  error
		expected      *proto.UnionResolveTypeResponse
		expectedErr   assert.ErrorAssertionFunc
	}{
		{
			title:         "CallsUserHandler",
			input:         new(proto.UnionResolveTypeRequest),
			handlerInput:  driver.UnionResolveTypeInput{},
			handlerOutput: "SomeType",
			expected: &proto.UnionResolveTypeResponse{
				Type: &proto.TypeRef{
					TestTyperef: &proto.TypeRef_Name{Name: "SomeType"},
				},
			},
			expectedErr: assert.NoError,
		},
		{
			title:        "ReturnsUserError",
			input:        new(proto.UnionResolveTypeRequest),
			handlerInput: driver.UnionResolveTypeInput{},
			handlerError: fmt.Errorf("user error"),
			expected: &proto.UnionResolveTypeResponse{
				Error: &proto.Error{
					Msg: "user error",
				},
			},
			expectedErr: assert.NoError,
		},
	}
	for i := range data {
		tt := data[i]
		t.Run(tt.title, func(t *testing.T) {
			unionResolveTypeMock := new(unionResolveTypeMock)
			unionResolveTypeMock.On("Handle", tt.handlerInput).Return(tt.handlerOutput, tt.handlerError)
			srv := grpc.Server{
				UnionResolveTypeHandler: unionResolveTypeMock,
			}
			out, err := srv.UnionResolveType(context.Background(), tt.input)
			tt.expectedErr(t, err)
			assert.Equal(t, tt.expected, out)
		})
	}
	t.Run("RecoversFromPanic", func(t *testing.T) {
		srv := grpc.Server{}
		out, err := srv.UnionResolveType(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, out.Error)
		assert.NotEmpty(t, out.Error.Msg)
	})
}
