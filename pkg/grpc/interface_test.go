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

func TestClientInterfaceResolveType(t *testing.T) {
	data := []struct {
		title        string
		input        driver.InterfaceResolveTypeInput
		grpcRequest  *proto.InterfaceResolveTypeRequest
		grpcResposne *proto.InterfaceResolveTypeResponse
		grpcError    error
		expected     driver.InterfaceResolveTypeOutput
		expectedErr  assert.ErrorAssertionFunc
	}{
		{
			title: "CallsGRPCInterfaceResolveTypeInput",
			input: driver.InterfaceResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			grpcRequest: &proto.InterfaceResolveTypeRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Value: new(proto.Value),
				Info:  &proto.InterfaceResolveTypeInfo{},
			},
			grpcResposne: &proto.InterfaceResolveTypeResponse{
				Type: &proto.TypeRef{
					TestTyperef: &proto.TypeRef_Name{Name: "SomeType"},
				},
			},
			expected: driver.InterfaceResolveTypeOutput{
				Type: types.TypeRef{
					Name: "SomeType",
				},
			},
			expectedErr: assert.NoError,
		},
		{
			title: "ErrorOnMissingFunction",
			input: driver.InterfaceResolveTypeInput{},
			expected: driver.InterfaceResolveTypeOutput{
				Error: &driver.Error{
					Message: "function name is required",
				},
			},
			expectedErr: assert.NoError,
		},
		{
			title: "PassthroughGRPCError",
			input: driver.InterfaceResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			grpcRequest: &proto.InterfaceResolveTypeRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Value: new(proto.Value),
				Info:  &proto.InterfaceResolveTypeInfo{},
			},
			grpcError: fmt.Errorf("grpc error"),
			expected: driver.InterfaceResolveTypeOutput{
				Error: &driver.Error{
					Message: "grpc error",
				},
			},
			expectedErr: assert.NoError,
		},
		{
			title: "PassthroughUserError",
			input: driver.InterfaceResolveTypeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			grpcRequest: &proto.InterfaceResolveTypeRequest{
				Function: &proto.Function{
					Name: "function",
				},
				Value: new(proto.Value),
				Info:  &proto.InterfaceResolveTypeInfo{},
			},
			grpcResposne: &proto.InterfaceResolveTypeResponse{
				Error: &proto.Error{
					Msg: "user error",
				},
			},
			expected: driver.InterfaceResolveTypeOutput{
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
				"InterfaceResolveType",
				mock.Anything,
				tt.grpcRequest,
			).Return(tt.grpcResposne, tt.grpcError)
			client := grpc.Client{
				Client: driverClientMock,
			}
			out, err := client.InterfaceResolveType(tt.input)
			tt.expectedErr(t, err)
			assert.Equal(t, tt.expected, out)
		})
	}
}

func TestServerInterfaceResolveType(t *testing.T) {
	data := []struct {
		title         string
		input         *proto.InterfaceResolveTypeRequest
		handlerInput  driver.InterfaceResolveTypeInput
		handlerOutput string
		handlerError  error
		expected      *proto.InterfaceResolveTypeResponse
		expectedErr   assert.ErrorAssertionFunc
	}{
		{
			title:         "CallsUserHandler",
			input:         new(proto.InterfaceResolveTypeRequest),
			handlerInput:  driver.InterfaceResolveTypeInput{},
			handlerOutput: "SomeType",
			expected: &proto.InterfaceResolveTypeResponse{
				Type: &proto.TypeRef{
					TestTyperef: &proto.TypeRef_Name{Name: "SomeType"},
				},
			},
			expectedErr: assert.NoError,
		},
		{
			title:        "ReturnsUserError",
			input:        new(proto.InterfaceResolveTypeRequest),
			handlerInput: driver.InterfaceResolveTypeInput{},
			handlerError: fmt.Errorf("user error"),
			expected: &proto.InterfaceResolveTypeResponse{
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
			interfaceResolveTypeMock := new(interfaceResolveTypeMock)
			interfaceResolveTypeMock.On("Handle", tt.handlerInput).Return(tt.handlerOutput, tt.handlerError)
			srv := grpc.Server{
				InterfaceResolveTypeHandler: interfaceResolveTypeMock,
			}
			out, err := srv.InterfaceResolveType(context.Background(), tt.input)
			tt.expectedErr(t, err)
			assert.Equal(t, tt.expected, out)
		})
	}
	t.Run("RecoversFromPanic", func(t *testing.T) {
		srv := grpc.Server{}
		out, err := srv.InterfaceResolveType(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, out.Error)
		assert.NotEmpty(t, out.Error.Msg)
	})
}
