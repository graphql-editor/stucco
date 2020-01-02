package prototest

import (
	"errors"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/proto"
)

// SetSecretsClientTest is basic struct for testing clients implementing proto

type SetSecretsClientTest struct {
	Title         string
	Input         driver.SetSecretsInput
	ProtoRequest  *proto.SetSecretsRequest
	ProtoResponse *proto.SetSecretsResponse
	ProtoError    error
	Expected      driver.SetSecretsOutput
}

// SetSecretsClientTestData is a data for testing secrets of proto clients
func SetSecretsClientTestData() []SetSecretsClientTest {
	return []SetSecretsClientTest{
		{
			Title: "Sets secrets",
			Input: driver.SetSecretsInput{
				Secrets: driver.Secrets{
					"secret": "value",
				},
			},
			ProtoRequest: &proto.SetSecretsRequest{
				Secrets: []*proto.Secret{
					&proto.Secret{
						Key:   "secret",
						Value: "value",
					},
				},
			},
			ProtoResponse: new(proto.SetSecretsResponse),
		},
		{
			Title:         "ReturnsProtoError",
			Input:         driver.SetSecretsInput{},
			ProtoRequest:  new(proto.SetSecretsRequest),
			ProtoError:    errors.New("proto error"),
			ProtoResponse: new(proto.SetSecretsResponse),
			Expected: driver.SetSecretsOutput{
				Error: &driver.Error{
					Message: "proto error",
				},
			},
		},
	}
}

// RunSetSecretsClientTests runs all client tests on a function
func RunSetSecretsClientTests(t *testing.T, f func(t *testing.T, tt SetSecretsClientTest)) {
	for _, tt := range SetSecretsClientTestData() {
		t.Run(tt.Title, func(t *testing.T) {
			f(t, tt)
		})
	}
}

// SetSecretsServerTest is basic struct for testing clients implementing proto
type SetSecretsServerTest struct {
	Title         string
	Input         *proto.SetSecretsRequest
	HandlerInput  driver.SetSecretsInput
	HandlerOutput error
	Expected      *proto.SetSecretsResponse
}

// SetSecretsServerTestData is a data for testing secrets of proto clients
func SetSecretsServerTestData() []SetSecretsServerTest {
	return []SetSecretsServerTest{
		{
			Title: "CallsHandler",
			Input: &proto.SetSecretsRequest{
				Secrets: []*proto.Secret{
					&proto.Secret{
						Key:   "secret",
						Value: "value",
					},
				},
			},
			HandlerInput: driver.SetSecretsInput{
				Secrets: driver.Secrets{
					"secret": "value",
				},
			},
			Expected: new(proto.SetSecretsResponse),
		},
	}
}

// RunSetSecretsServerTests runs all client tests on a function
func RunSetSecretsServerTests(t *testing.T, f func(t *testing.T, tt SetSecretsServerTest)) {
	for _, tt := range SetSecretsServerTestData() {
		t.Run(tt.Title, func(t *testing.T) {
			f(t, tt)
		})
	}
}
