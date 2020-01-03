package protohttp_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/graphql-editor/stucco/pkg/driver/protohttp"
	"github.com/graphql-editor/stucco/pkg/proto"
	"github.com/graphql-editor/stucco/pkg/proto/prototest"
	"github.com/stretchr/testify/assert"
)

func TestClientFieldResolve(t *testing.T) {
	prototest.RunFieldResolveClientTests(t, func(t *testing.T, tt prototest.FieldResolveClientTest) {
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			header := req.Header.Get("content-type")
			assert.Equal(t, "application/x-protobuf; message=FieldResolveRequest", header)
			var protoRequest proto.FieldResolveRequest
			body, _ := ioutil.ReadAll(req.Body)
			req.Body.Close()
			protobuf.Unmarshal(body, &protoRequest)
			assert.Equal(t, tt.ProtoRequest, &protoRequest)
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
		var protoResp proto.FieldResolveResponse
		assert.NoError(t, protobuf.Unmarshal(responseRecorder.Body.Bytes(), &protoResp))
		assert.Equal(t, tt.Expected, &protoResp)
	})
}
