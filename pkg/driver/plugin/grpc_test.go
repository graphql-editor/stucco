package plugin_test

import (
	"context"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver/plugin"
	"github.com/graphql-editor/stucco/pkg/grpc"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	googlegrpc "google.golang.org/grpc"
)

type mockProto struct {
	mock.Mock
}

func (m *mockProto) RegisterDriverServer(s *googlegrpc.Server, d proto.DriverServer) {
	m.Called(s, d)
}

func (m *mockProto) NewDriverClient(c *googlegrpc.ClientConn) proto.DriverClient {
	return m.Called(c).Get(0).(proto.DriverClient)
}

func TestGRPC(t *testing.T) {
	t.Run("GRPCServerCallsRegisterDriverServer", func(t *testing.T) {
		mockProto := new(mockProto)
		gs := googlegrpc.NewServer()
		plugin.RegisterDriverServer = mockProto.RegisterDriverServer
		defer func() {
			plugin.RegisterDriverServer = proto.RegisterDriverServer
		}()
		mockProto.On(
			"RegisterDriverServer",
			gs,
			mock.AnythingOfType("*grpc.Server"),
		)
		s := plugin.GRPC{}
		assert.NoError(t, s.GRPCServer(nil, gs))
		mockProto.AssertCalled(
			t,
			"RegisterDriverServer",
			gs,
			mock.AnythingOfType("*grpc.Server"),
		)
	})
	t.Run("GRPCClientCallsNewDriverClient", func(t *testing.T) {
		mockProto := new(mockProto)
		conn := &googlegrpc.ClientConn{}
		driverClient := proto.NewDriverClient(conn)
		plugin.NewDriverClient = mockProto.NewDriverClient
		defer func() {
			plugin.NewDriverClient = proto.NewDriverClient
		}()
		mockProto.On(
			"NewDriverClient",
			conn,
		).Return(driverClient)
		s := plugin.GRPC{}
		grpcClient, err := s.GRPCClient(context.Background(), nil, conn)
		assert.NoError(t, err)
		assert.IsType(t, &grpc.Client{}, grpcClient)
		mockProto.AssertCalled(
			t,
			"NewDriverClient",
			conn,
		)
		assert.Equal(t, driverClient, grpcClient.(*grpc.Client).Client)
	})
}
