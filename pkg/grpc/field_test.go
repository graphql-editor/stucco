package grpc_test

import (
	"context"
	"testing"

	"github.com/graphql-editor/stucco/pkg/grpc"
	"github.com/graphql-editor/stucco/pkg/proto/prototest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClientFieldResolve(t *testing.T) {
	prototest.RunFieldResolveClientTests(t, func(t *testing.T, tt prototest.FieldResolveClientTest) {
		driverClientMock := new(driverClientMock)
		driverClientMock.On(
			"FieldResolve",
			mock.Anything,
			tt.ProtoRequest,
		).Return(tt.ProtoResponse, nil)
		client := grpc.Client{
			Client: driverClientMock,
		}
		out := client.FieldResolve(tt.Input)
		assert.Equal(t, tt.Expected, out)
		driverClientMock.AssertCalled(t, "FieldResolve", mock.Anything, tt.ProtoRequest)
	})
}

func TestServerFieldResolve(t *testing.T) {
	prototest.RunFieldResolveServerTests(t, func(t *testing.T, tt prototest.FieldResolveServerTest) {
		fieldResolveMock := new(fieldResolveMock)
		fieldResolveMock.On("Handle", tt.HandlerInput).Return(tt.HandlerResponse, tt.HandlerError)
		srv := grpc.Server{
			FieldResolveHandler: fieldResolveMock,
		}
		resp, err := srv.FieldResolve(context.Background(), tt.Input)
		assert.NoError(t, err)
		assert.Equal(t, tt.Expected, resp)
		fieldResolveMock.AssertCalled(t, "Handle", tt.HandlerInput)
	})
	t.Run("RecoversFromPanic", func(t *testing.T) {
		srv := grpc.Server{}
		resp, err := srv.FieldResolve(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.NotEmpty(t, resp.Error.Msg)
	})
}
