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

func TestClientScalarParse(t *testing.T) {
	data := []struct {
		title        string
		input        driver.ScalarParseInput
		grpcRequest  *proto.ScalarParseRequest
		grpcResponse *proto.ScalarParseResponse
		grpcError    error
		expected     driver.ScalarParseOutput
		expectedErr  assert.ErrorAssertionFunc
	}{
		{
			title: "CallsGRPCScalarParse",
			input: driver.ScalarParseInput{
				Function: types.Function{
					Name: "function",
				},
			},
			grpcRequest: &proto.ScalarParseRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			grpcResponse: &proto.ScalarParseResponse{Value: &proto.Value{
				TestValue: &proto.Value_S{
					S: "scalar",
				},
			}},
			expected:    driver.ScalarParseOutput{Response: "scalar"},
			expectedErr: assert.NoError,
		},
		{
			title: "ReturnsUserError",
			input: driver.ScalarParseInput{
				Function: types.Function{
					Name: "function",
				},
			},
			grpcRequest: &proto.ScalarParseRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			grpcResponse: &proto.ScalarParseResponse{
				Error: &proto.Error{
					Msg: "user error",
				},
			},
			expected: driver.ScalarParseOutput{
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
				"ScalarParse",
				mock.Anything,
				tt.grpcRequest,
			).Return(tt.grpcResponse, tt.grpcError)
			client := grpc.Client{
				Client: driverClientMock,
			}
			out, err := client.ScalarParse(tt.input)
			tt.expectedErr(t, err)
			assert.Equal(t, tt.expected, out)
		})
	}
}

func TestClientScalarSerialize(t *testing.T) {
	data := []struct {
		title        string
		input        driver.ScalarSerializeInput
		grpcRequest  *proto.ScalarSerializeRequest
		grpcResponse *proto.ScalarSerializeResponse
		grpcError    error
		expected     driver.ScalarSerializeOutput
		expectedErr  assert.ErrorAssertionFunc
	}{
		{
			title: "CallsGRPCScalarSerialize",
			input: driver.ScalarSerializeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			grpcRequest: &proto.ScalarSerializeRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			grpcResponse: &proto.ScalarSerializeResponse{Value: &proto.Value{
				TestValue: &proto.Value_S{
					S: "scalar",
				},
			}},
			expected:    driver.ScalarSerializeOutput{Response: "scalar"},
			expectedErr: assert.NoError,
		},
		{
			title: "ReturnsUserError",
			input: driver.ScalarSerializeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			grpcRequest: &proto.ScalarSerializeRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			grpcResponse: &proto.ScalarSerializeResponse{
				Error: &proto.Error{
					Msg: "user error",
				},
			},
			expected: driver.ScalarSerializeOutput{
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
				"ScalarSerialize",
				mock.Anything,
				tt.grpcRequest,
			).Return(tt.grpcResponse, tt.grpcError)
			client := grpc.Client{
				Client: driverClientMock,
			}
			out, err := client.ScalarSerialize(tt.input)
			tt.expectedErr(t, err)
			assert.Equal(t, tt.expected, out)
		})
	}
}

func TestServerScalarParse(t *testing.T) {
	data := []struct {
		title         string
		input         *proto.ScalarParseRequest
		handlerInput  driver.ScalarParseInput
		handlerOutput interface{}
		handlerError  error
		expected      *proto.ScalarParseResponse
		expectedErr   assert.ErrorAssertionFunc
	}{
		{
			title: "CallsScalarParseHandler",
			input: &proto.ScalarParseRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			handlerInput: driver.ScalarParseInput{
				Function: types.Function{
					Name: "function",
				},
			},
			handlerOutput: "scalar",
			expected: &proto.ScalarParseResponse{Value: &proto.Value{
				TestValue: &proto.Value_S{
					S: "scalar",
				},
			}},
			expectedErr: assert.NoError,
		},
		{
			title: "ReturnsUserError",
			input: &proto.ScalarParseRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			handlerInput: driver.ScalarParseInput{
				Function: types.Function{
					Name: "function",
				},
			},
			handlerError: fmt.Errorf("user error"),
			expected: &proto.ScalarParseResponse{
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
			scalarParseMock := new(scalarParseMock)
			scalarParseMock.On(
				"Handle",
				tt.handlerInput,
			).Return(tt.handlerOutput, tt.handlerError)
			srv := grpc.Server{
				ScalarParseHandler: scalarParseMock,
			}
			out, err := srv.ScalarParse(context.Background(), tt.input)
			tt.expectedErr(t, err)
			assert.Equal(t, tt.expected, out)
		})
	}
	t.Run("RecoversFromPanic", func(t *testing.T) {
		srv := grpc.Server{}
		out, err := srv.ScalarParse(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, out.Error)
		assert.NotEmpty(t, out.Error.Msg)
	})
}

func TestServerScalarSerialize(t *testing.T) {
	data := []struct {
		title         string
		input         *proto.ScalarSerializeRequest
		handlerInput  driver.ScalarSerializeInput
		handlerOutput interface{}
		handlerError  error
		expected      *proto.ScalarSerializeResponse
		expectedErr   assert.ErrorAssertionFunc
	}{
		{
			title: "CallsScalarSerializeHandler",
			input: &proto.ScalarSerializeRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			handlerInput: driver.ScalarSerializeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			handlerOutput: "scalar",
			expected: &proto.ScalarSerializeResponse{Value: &proto.Value{
				TestValue: &proto.Value_S{
					S: "scalar",
				},
			}},
			expectedErr: assert.NoError,
		},
		{
			title: "ReturnsUserError",
			input: &proto.ScalarSerializeRequest{
				Value: new(proto.Value),
				Function: &proto.Function{
					Name: "function",
				},
			},
			handlerInput: driver.ScalarSerializeInput{
				Function: types.Function{
					Name: "function",
				},
			},
			handlerError: fmt.Errorf("user error"),
			expected: &proto.ScalarSerializeResponse{
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
			scalarSerializeMock := new(scalarSerializeMock)
			scalarSerializeMock.On(
				"Handle",
				tt.handlerInput,
			).Return(tt.handlerOutput, tt.handlerError)
			srv := grpc.Server{
				ScalarSerializeHandler: scalarSerializeMock,
			}
			out, err := srv.ScalarSerialize(context.Background(), tt.input)
			tt.expectedErr(t, err)
			assert.Equal(t, tt.expected, out)
		})
	}
	t.Run("RecoversFromPanic", func(t *testing.T) {
		srv := grpc.Server{}
		out, err := srv.ScalarSerialize(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, out.Error)
		assert.NotEmpty(t, out.Error.Msg)
	})
}
