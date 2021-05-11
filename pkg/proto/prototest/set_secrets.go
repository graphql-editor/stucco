package prototest

import (
	"errors"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
)

// SetSecretsClientTest is basic struct for testing clients implementing proto

type SetSecretsClientTest struct {
	Title         string
	Input         driver.SetSecretsInput
	ProtoRequest  *protoMessages.SetSecretsRequest
	ProtoResponse *protoMessages.SetSecretsResponse
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
			ProtoRequest: &protoMessages.SetSecretsRequest{
				Secrets: []*protoMessages.Secret{
					&protoMessages.Secret{
						Key:   "secret",
						Value: "value",
					},
				},
			},
			ProtoResponse: new(protoMessages.SetSecretsResponse),
		},
		{
			Title:         "ReturnsProtoError",
			Input:         driver.SetSecretsInput{},
			ProtoRequest:  new(protoMessages.SetSecretsRequest),
			ProtoError:    errors.New("proto error"),
			ProtoResponse: new(protoMessages.SetSecretsResponse),
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
	Input         *protoMessages.SetSecretsRequest
	HandlerInput  driver.SetSecretsInput
	HandlerOutput error
	Expected      *protoMessages.SetSecretsResponse
}

// SetSecretsServerTestData is a data for testing secrets of proto clients
func SetSecretsServerTestData() []SetSecretsServerTest {
	return []SetSecretsServerTest{
		{
			Title: "CallsHandler",
			Input: &protoMessages.SetSecretsRequest{
				Secrets: []*protoMessages.Secret{
					&protoMessages.Secret{
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
			Expected: new(protoMessages.SetSecretsResponse),
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
