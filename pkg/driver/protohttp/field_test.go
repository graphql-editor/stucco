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

func TestClientFieldResolve(t *testing.T) {
	prototest.RunFieldResolveClientTests(t, func(t *testing.T, tt prototest.FieldResolveClientTest) {
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			header := req.Header.Get("content-type")
			assert.Equal(t, "application/x-protobuf; message=FieldResolveRequest", header)
			body, err := ioutil.ReadAll(req.Body)
			assert.NoError(t, err)
			req.Body.Close()
			var p protoMessages.FieldResolveRequest
			assert.NoError(t, proto.Unmarshal(body, &p))
			rw.Header().Add("content-type", "application/x-protobuf; message=FieldResolveResponse")
			b, _ := protobuf.Marshal(tt.ProtoResponse)
			rw.Write(b)
		}))
		defer srv.Close()
		client := protohttp.NewClient(protohttp.Config{
			Client: srv.Client(),
			URL:    srv.URL,
		})
		out := client.FieldResolve(tt.Input)
		assert.Equal(t, tt.Expected, out)
	})
}

func TestServerFieldResolve(t *testing.T) {
	prototest.RunFieldResolveServerTests(t, func(t *testing.T, tt prototest.FieldResolveServerTest) {
		var r http.Request
		b, _ := protobuf.Marshal(tt.Input)
		r.Header = make(http.Header)
		r.Body = ioutil.NopCloser(bytes.NewReader(b))
		r.Header.Add("content-type", "application/x-protobuf; message=FieldResolveRequest")
		responseRecorder := httptest.NewRecorder()
		mockMuxer := new(mockMuxer)
		handler := &protohttp.Handler{
			Muxer: mockMuxer,
		}
		mockMuxer.On("FieldResolve", tt.HandlerInput).Return(tt.HandlerResponse, tt.HandlerError)
		handler.ServeHTTP(responseRecorder, &r)
		mockMuxer.AssertCalled(t, "FieldResolve", tt.HandlerInput)
		assert.Equal(t, "application/x-protobuf; message=FieldResolveResponse", responseRecorder.Header().Get("content-type"))
		var protoResp protoMessages.FieldResolveResponse
		assert.NoError(t, protobuf.Unmarshal(responseRecorder.Body.Bytes(), &protoResp))
	})
}
