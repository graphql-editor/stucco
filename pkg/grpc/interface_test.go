package grpc_test

import (
	"context"
	"testing"

	"github.com/graphql-editor/stucco/pkg/grpc"
	"github.com/graphql-editor/stucco/pkg/proto/prototest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClientInterfaceResolveType(t *testing.T) {
	prototest.RunInterfaceResolveTypeClientTests(t, func(t *testing.T, tt prototest.InterfaceResolveTypeClientTest) {
		driverClientMock := new(driverClientMock)
		driverClientMock.On(
			"InterfaceResolveType",
			mock.Anything,
			tt.ProtoRequest,
		).Return(tt.ProtoResponse, tt.ProtoError)
		client := grpc.Client{
			Client: driverClientMock,
		}
		out, err := client.InterfaceResolveType(tt.Input)
		tt.ExpectedErr(t, err)
		assert.Equal(t, tt.Expected, out)
	})
}

func TestServerInterfaceResolveType(t *testing.T) {
	prototest.RunInterfaceResolveTypeServerTests(t, func(t *testing.T, tt prototest.InterfaceResolveTypeServerTest) {
		interfaceResolveTypeMock := new(interfaceResolveTypeMock)
		interfaceResolveTypeMock.On("Handle", tt.HandlerInput).Return(tt.HandlerOutput, tt.HandlerError)
		srv := grpc.Server{
			InterfaceResolveTypeHandler: interfaceResolveTypeMock,
		}
		out, err := srv.InterfaceResolveType(context.Background(), tt.Input)
		tt.ExpectedErr(t, err)
		assert.Equal(t, tt.Expected, out)
	})
	t.Run("RecoversFromPanic", func(t *testing.T) {
		srv := grpc.Server{}
		out, err := srv.InterfaceResolveType(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, out.Error)
		assert.NotEmpty(t, out.Error.Msg)
	})
}
