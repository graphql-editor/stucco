package grpc_test

import (
	"context"
	"testing"

	"github.com/graphql-editor/stucco/pkg/grpc"
	"github.com/graphql-editor/stucco/pkg/proto/prototest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClientSetSecrets(t *testing.T) {
	prototest.RunSetSecretsClientTests(t, func(t *testing.T, tt prototest.SetSecretsClientTest) {
		driverClientMock := new(driverClientMock)
		driverClientMock.On(
			"SetSecrets",
			mock.Anything,
			tt.ProtoRequest,
		).Return(tt.ProtoResponse, tt.ProtoError)
		client := grpc.Client{
			Client: driverClientMock,
		}
		out := client.SetSecrets(tt.Input)
		assert.Equal(t, tt.Expected, out)
	})
}

func TestServerSetSecrets(t *testing.T) {
	prototest.RunSetSecretsServerTests(t, func(t *testing.T, tt prototest.SetSecretsServerTest) {
		setSecretsMock := new(setSecretsMock)
		setSecretsMock.On("Handle", tt.HandlerInput).Return(tt.HandlerOutput)
		srv := grpc.Server{
			SetSecretsHandler: setSecretsMock,
		}
		out, err := srv.SetSecrets(context.Background(), tt.Input)
		assert.NoError(t, err)
		assert.Equal(t, tt.Expected, out)
	})
	t.Run("RecoversFromPanic", func(t *testing.T) {
		srv := grpc.Server{}
		out, err := srv.SetSecrets(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, out.Error)
		assert.NotEmpty(t, out.Error.Msg)
	})
}
