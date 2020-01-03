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

func TestClientScalarParse(t *testing.T) {
	prototest.RunScalarParseClientTests(t, func(t *testing.T, tt prototest.ScalarParseClientTest) {
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if tt.ProtoError != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(tt.ProtoError.Error()))
				return
			}
			header := req.Header.Get("content-type")
			assert.Equal(t, "application/x-protobuf; message=ScalarParseRequest", header)
			var protoRequest proto.ScalarParseRequest
			body, _ := ioutil.ReadAll(req.Body)
			req.Body.Close()
			protobuf.Unmarshal(body, &protoRequest)
			assert.Equal(t, tt.ProtoRequest, &protoRequest)
			rw.Header().Add("content-type", "application/x-protobuf; message=ScalarParseResponse")
			b, _ := protobuf.Marshal(tt.ProtoResponse)
			rw.Write(b)
		}))
		defer srv.Close()
		client := protohttp.NewClient(protohttp.Config{
			Client: srv.Client(),
			URL:    srv.URL,
		})
		out := client.ScalarParse(tt.Input)
		assert.Equal(t, tt.Expected, out)
	})
}

func TestClientScalarSerialize(t *testing.T) {
	prototest.RunScalarSerializeClientTests(t, func(t *testing.T, tt prototest.ScalarSerializeClientTest) {
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if tt.ProtoError != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(tt.ProtoError.Error()))
				return
			}
			header := req.Header.Get("content-type")
			assert.Equal(t, "application/x-protobuf; message=ScalarSerializeRequest", header)
			var protoRequest proto.ScalarSerializeRequest
			body, _ := ioutil.ReadAll(req.Body)
			req.Body.Close()
			protobuf.Unmarshal(body, &protoRequest)
			assert.Equal(t, tt.ProtoRequest, &protoRequest)
			rw.Header().Add("content-type", "application/x-protobuf; message=ScalarSerializeResponse")
			b, _ := protobuf.Marshal(tt.ProtoResponse)
			rw.Write(b)
		}))
		defer srv.Close()
		client := protohttp.NewClient(protohttp.Config{
			Client: srv.Client(),
			URL:    srv.URL,
		})
		out := client.ScalarSerialize(tt.Input)
		assert.Equal(t, tt.Expected, out)
	})
}

func TestServerScalarParse(t *testing.T) {
	prototest.RunScalarParseServerTests(t, func(t *testing.T, tt prototest.ScalarParseServerTest) {
		var r http.Request
		b, _ := protobuf.Marshal(tt.Input)
		r.Header = make(http.Header)
		r.Body = ioutil.NopCloser(bytes.NewReader(b))
		r.Header.Add("content-type", "application/x-protobuf; message=ScalarParseRequest")
		responseRecorder := httptest.NewRecorder()
		mockMuxer := new(mockMuxer)
		handler := &protohttp.Handler{
			Muxer: mockMuxer,
		}
		mockMuxer.On("ScalarParse", tt.HandlerInput).Return(tt.HandlerOutput, tt.HandlerError)
		handler.ServeHTTP(responseRecorder, &r)
		mockMuxer.AssertCalled(t, "ScalarParse", tt.HandlerInput)
		assert.Equal(t, "application/x-protobuf; message=ScalarParseResponse", responseRecorder.Header().Get("content-type"))
		var protoResp proto.ScalarParseResponse
		assert.NoError(t, protobuf.Unmarshal(responseRecorder.Body.Bytes(), &protoResp))
		assert.Equal(t, tt.Expected, &protoResp)
	})
}

func TestServerScalarSerialize(t *testing.T) {
	prototest.RunScalarSerializeServerTests(t, func(t *testing.T, tt prototest.ScalarSerializeServerTest) {
		var r http.Request
		b, _ := protobuf.Marshal(tt.Input)
		r.Header = make(http.Header)
		r.Body = ioutil.NopCloser(bytes.NewReader(b))
		r.Header.Add("content-type", "application/x-protobuf; message=ScalarSerializeRequest")
		responseRecorder := httptest.NewRecorder()
		mockMuxer := new(mockMuxer)
		handler := &protohttp.Handler{
			Muxer: mockMuxer,
		}
		mockMuxer.On("ScalarSerialize", tt.HandlerInput).Return(tt.HandlerOutput, tt.HandlerError)
		handler.ServeHTTP(responseRecorder, &r)
		mockMuxer.AssertCalled(t, "ScalarSerialize", tt.HandlerInput)
		assert.Equal(t, "application/x-protobuf; message=ScalarSerializeResponse", responseRecorder.Header().Get("content-type"))
		var protoResp proto.ScalarSerializeResponse
		assert.NoError(t, protobuf.Unmarshal(responseRecorder.Body.Bytes(), &protoResp))
		assert.Equal(t, tt.Expected, &protoResp)
	})
}
