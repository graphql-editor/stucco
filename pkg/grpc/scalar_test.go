package grpc_test

import (
	"context"
	"testing"

	"github.com/graphql-editor/stucco/pkg/grpc"
	"github.com/graphql-editor/stucco/pkg/proto/prototest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClientScalarParse(t *testing.T) {
	prototest.RunScalarParseClientTests(t, func(t *testing.T, tt prototest.ScalarParseClientTest) {
		driverClientMock := new(driverClientMock)
		driverClientMock.On(
			"ScalarParse",
			mock.Anything,
			tt.ProtoRequest,
		).Return(tt.ProtoResponse, tt.ProtoError)
		client := grpc.Client{
			Client: driverClientMock,
		}
		out := client.ScalarParse(tt.Input)
		assert.Equal(t, tt.Expected, out)
	})
}

func TestClientScalarSerialize(t *testing.T) {
	prototest.RunScalarSerializeClientTests(t, func(t *testing.T, tt prototest.ScalarSerializeClientTest) {
		driverClientMock := new(driverClientMock)
		driverClientMock.On(
			"ScalarSerialize",
			mock.Anything,
			tt.ProtoRequest,
		).Return(tt.ProtoResponse, tt.ProtoError)
		client := grpc.Client{
			Client: driverClientMock,
		}
		out := client.ScalarSerialize(tt.Input)
		assert.Equal(t, tt.Expected, out)
	})
}

func TestServerScalarParse(t *testing.T) {
	prototest.RunScalarParseServerTests(t, func(t *testing.T, tt prototest.ScalarParseServerTest) {
		scalarParseMock := new(scalarParseMock)
		scalarParseMock.On(
			"Handle",
			tt.HandlerInput,
		).Return(tt.HandlerOutput, tt.HandlerError)
		srv := grpc.Server{
			ScalarParseHandler: scalarParseMock,
		}
		out, err := srv.ScalarParse(context.Background(), tt.Input)
		assert.NoError(t, err)
		assert.Equal(t, tt.Expected, out)
	})
	t.Run("RecoversFromPanic", func(t *testing.T) {
		srv := grpc.Server{}
		out, err := srv.ScalarParse(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, out.Error)
		assert.NotEmpty(t, out.Error.Msg)
	})
}

func TestServerScalarSerialize(t *testing.T) {
	prototest.RunScalarSerializeServerTests(t, func(t *testing.T, tt prototest.ScalarSerializeServerTest) {
		scalarSerializeMock := new(scalarSerializeMock)
		scalarSerializeMock.On(
			"Handle",
			tt.HandlerInput,
		).Return(tt.HandlerOutput, tt.HandlerError)
		srv := grpc.Server{
			ScalarSerializeHandler: scalarSerializeMock,
		}
		out, err := srv.ScalarSerialize(context.Background(), tt.Input)
		assert.NoError(t, err)
		assert.Equal(t, tt.Expected, out)
	})
	t.Run("RecoversFromPanic", func(t *testing.T) {
		srv := grpc.Server{}
		out, err := srv.ScalarSerialize(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, out.Error)
		assert.NotEmpty(t, out.Error.Msg)
	})
}
