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

func TestClientInterfaceResolveType(t *testing.T) {
	prototest.RunInterfaceResolveTypeClientTests(t, func(t *testing.T, tt prototest.InterfaceResolveTypeClientTest) {
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if tt.ProtoError != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(tt.ProtoError.Error()))
				return
			}
			header := req.Header.Get("content-type")
			assert.Equal(t, "application/x-protobuf; message=InterfaceResolveTypeRequest", header)
			body, err := ioutil.ReadAll(req.Body)
			assert.NoError(t, err)
			req.Body.Close()
			var p protoMessages.InterfaceResolveTypeRequest
			assert.NoError(t, proto.Unmarshal(body, &p))
			rw.Header().Add("content-type", "application/x-protobuf; message=InterfaceResolveTypeResponse")
			b, _ := protobuf.Marshal(tt.ProtoResponse)
			rw.Write(b)
		}))
		defer srv.Close()
		client := protohttp.NewClient(protohttp.Config{
			Client: srv.Client(),
			URL:    srv.URL,
		})
		out := client.InterfaceResolveType(tt.Input)
		if tt.Expected.Error != nil {
			assert.Contains(t, out.Error.Message, tt.Expected.Error.Message)
		} else {
			assert.Equal(t, tt.Expected, out)
		}
	})
}

func TestServerInterfaceResolveType(t *testing.T) {
	prototest.RunInterfaceResolveTypeServerTests(t, func(t *testing.T, tt prototest.InterfaceResolveTypeServerTest) {
		var r http.Request
		b, _ := protobuf.Marshal(tt.Input)
		r.Header = make(http.Header)
		r.Body = ioutil.NopCloser(bytes.NewReader(b))
		r.Header.Add("content-type", "application/x-protobuf; message=InterfaceResolveTypeRequest")
		responseRecorder := httptest.NewRecorder()
		mockMuxer := new(mockMuxer)
		handler := &protohttp.Handler{
			Muxer: mockMuxer,
		}
		mockMuxer.On("InterfaceResolveType", tt.HandlerInput).Return(tt.HandlerOutput, tt.HandlerError)
		handler.ServeHTTP(responseRecorder, &r)
		mockMuxer.AssertCalled(t, "InterfaceResolveType", tt.HandlerInput)
		assert.Equal(t, "application/x-protobuf; message=InterfaceResolveTypeResponse", responseRecorder.Header().Get("content-type"))
		var protoResp protoMessages.InterfaceResolveTypeResponse
		assert.NoError(t, protobuf.Unmarshal(responseRecorder.Body.Bytes(), &protoResp))
	})
}
