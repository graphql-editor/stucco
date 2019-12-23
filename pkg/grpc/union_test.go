package grpc_test

import (
	"context"
	"testing"

	"github.com/graphql-editor/stucco/pkg/grpc"
	"github.com/graphql-editor/stucco/pkg/proto/prototest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClientUnionResolveType(t *testing.T) {
	prototest.RunUnionResolveTypeClientTests(t, func(t *testing.T, tt prototest.UnionResolveTypeClientTest) {
		driverClientMock := new(driverClientMock)
		driverClientMock.On(
			"UnionResolveType",
			mock.Anything,
			tt.ProtoRequest,
		).Return(tt.ProtoResponse, tt.ProtoError)
		client := grpc.Client{
			Client: driverClientMock,
		}
		out := client.UnionResolveType(tt.Input)
		assert.Equal(t, tt.Expected, out)
	})
}

func TestServerUnionResolveType(t *testing.T) {
	prototest.RunUnionResolveTypeServerTests(t, func(t *testing.T, tt prototest.UnionResolveTypeServerTest) {
		unionResolveTypeMock := new(unionResolveTypeMock)
		unionResolveTypeMock.On("Handle", tt.HandlerInput).Return(tt.HandlerOutput, tt.HandlerError)
		srv := grpc.Server{
			UnionResolveTypeHandler: unionResolveTypeMock,
		}
		out, err := srv.UnionResolveType(context.Background(), tt.Input)
		assert.NoError(t, err)
		assert.Equal(t, tt.Expected, out)
	})
	t.Run("RecoversFromPanic", func(t *testing.T) {
		srv := grpc.Server{}
		out, err := srv.UnionResolveType(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, out.Error)
		assert.NotEmpty(t, out.Error.Msg)
	})
}
