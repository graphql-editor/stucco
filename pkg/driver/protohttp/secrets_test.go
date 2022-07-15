package protohttp_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	protobuf "google.golang.org/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver/protohttp"
	"github.com/graphql-editor/stucco/pkg/proto/prototest"
	protoMessages "github.com/graphql-editor/stucco_proto/go/messages"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestClientSetSecrets(t *testing.T) {
	prototest.RunSetSecretsClientTests(t, func(t *testing.T, tt prototest.SetSecretsClientTest) {
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if tt.ProtoError != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(tt.ProtoError.Error()))
				return
			}
			header := req.Header.Get("content-type")
			assert.Equal(t, "application/x-protobuf; message=SetSecretsRequest", header)
			body, err := ioutil.ReadAll(req.Body)
			assert.NoError(t, err)
			req.Body.Close()
			var p protoMessages.SetSecretsRequest
			assert.NoError(t, proto.Unmarshal(body, &p))
			rw.Header().Add("content-type", "application/x-protobuf; message=SetSecretsResponse")
			b, _ := protobuf.Marshal(tt.ProtoResponse)
			rw.Write(b)
		}))
		defer srv.Close()
		client := protohttp.NewClient(protohttp.Config{
			Client: srv.Client(),
			URL:    srv.URL,
		})
		out := client.SetSecrets(tt.Input)
		if tt.Expected.Error != nil {
			assert.Contains(t, out.Error.Message, tt.Expected.Error.Message)
		} else {
			assert.Equal(t, tt.Expected, out)
		}
	})
}

func TestServerSetSecrets(t *testing.T) {
	prototest.RunSetSecretsServerTests(t, func(t *testing.T, tt prototest.SetSecretsServerTest) {
		var r http.Request
		b, _ := protobuf.Marshal(tt.Input)
		r.Header = make(http.Header)
		r.Body = ioutil.NopCloser(bytes.NewReader(b))
		r.Header.Add("content-type", "application/x-protobuf; message=SetSecretsRequest")
		responseRecorder := httptest.NewRecorder()
		mockMuxer := new(mockMuxer)
		handler := &protohttp.Handler{
			Muxer: mockMuxer,
		}
		mockMuxer.On("SetSecrets", tt.HandlerInput).Return(tt.HandlerOutput)
		handler.ServeHTTP(responseRecorder, &r)
		mockMuxer.AssertCalled(t, "SetSecrets", tt.HandlerInput)
		assert.Equal(t, "application/x-protobuf; message=SetSecretsResponse", responseRecorder.Header().Get("content-type"))
		var protoResp protoMessages.SetSecretsResponse
		assert.NoError(t, protobuf.Unmarshal(responseRecorder.Body.Bytes(), &protoResp))
	})
}
