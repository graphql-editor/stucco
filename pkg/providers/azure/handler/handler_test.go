package handler_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/graphql-editor/azure-functions-golang-worker/api"
	"github.com/graphql-editor/azure-functions-golang-worker/mocks"
	"github.com/graphql-editor/stucco/pkg/providers/azure/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockHTTPHandler struct {
	mock.Mock
}

func (m *mockHTTPHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.Called(rw, req)
	rw.Header().Set("header", "value")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(`data`))
}

type ctxkey string

func TestHandler(t *testing.T) {
	var httpMock mockHTTPHandler
	mockURL, _ := url.Parse("http://mock.url/with/mock/path?key=value")
	reqBody := []byte("sample body")
	reqHeaders := make(http.Header)
	reqHeaders.Set("header", "value")
	ctx := context.WithValue(context.Background(), ctxkey("example"), "value")
	httpMock.On(
		"ServeHTTP",
		mock.MatchedBy(func(v interface{}) bool {
			return assert.Implements(t, (*http.ResponseWriter)(nil), v)
		}),
		mock.MatchedBy(func(v interface{}) bool {
			req, result := v.(*http.Request)
			result = result && assert.Equal(t, mockURL, req.URL)
			b, err := ioutil.ReadAll(req.Body)
			result = result && assert.NoError(t, err)
			req.Body.Close()
			result = result && assert.Contains(t, [][]byte{
				reqBody,
				[]byte(`"sample body"`),
			}, b)
			result = result && assert.Equal(t, reqHeaders, req.Header)
			result = result && assert.Equal(t, ctx, req.Context())
			result = result && assert.Equal(t, "POST", req.Method)
			return result
		}),
	)
	urlWithoutQuery := *mockURL
	urlWithoutQuery.RawQuery = ""
	respHeaders := make(http.Header)
	respHeaders.Set("header", "value")
	respHeaders.Set("content-type", "text/plain; charset=utf-8")
	var logger mocks.Logger
	logger.On("Error", mock.Anything)
	assert.Equal(
		t,
		api.Response{
			Body:       []byte(`data`),
			Headers:    respHeaders,
			StatusCode: http.StatusOK,
		},
		handler.Handler{
			Handler: &httpMock,
		}.ServeHTTP(ctx, &logger, &api.Request{
			Method:  "POST",
			URL:     urlWithoutQuery.String(),
			Query:   mockURL.Query(),
			Headers: reqHeaders,
			Body:    "sample body",
		}),
	)
	assert.Equal(
		t,
		api.Response{
			Body:       []byte(`data`),
			Headers:    respHeaders,
			StatusCode: http.StatusOK,
		},
		handler.Handler{
			Handler: &httpMock,
		}.ServeHTTP(ctx, &logger, &api.Request{
			Method:  "POST",
			URL:     urlWithoutQuery.String(),
			Query:   mockURL.Query(),
			Headers: reqHeaders,
			RawBody: reqBody,
		}),
	)
	assert.Equal(
		t,
		api.Response{
			Body:       []byte(`data`),
			Headers:    respHeaders,
			StatusCode: http.StatusOK,
		},
		handler.Handler{
			Handler: &httpMock,
		}.ServeHTTP(ctx, &logger, &api.Request{
			Method:  "POST",
			URL:     urlWithoutQuery.String(),
			Query:   mockURL.Query(),
			Headers: reqHeaders,
			RawBody: reqBody,
		}),
	)
	respHeaders = make(http.Header)
	respHeaders.Set("content-type", "text/plain; charset=utf-8")
	assert.Equal(
		t,
		api.Response{
			Body:       []byte("invalid url '://bad.url' from function"),
			Headers:    respHeaders,
			StatusCode: http.StatusInternalServerError,
		},
		handler.Handler{
			Handler: &httpMock,
		}.ServeHTTP(context.Background(), &logger, &api.Request{
			URL: "://bad.url",
		}),
	)
	assert.Equal(
		t,
		api.Response{
			Body:       []byte("could not serialize body of type int"),
			Headers:    respHeaders,
			StatusCode: http.StatusInternalServerError,
		},
		handler.Handler{
			Handler: &httpMock,
		}.ServeHTTP(context.Background(), &logger, &api.Request{
			Method:  "POST",
			URL:     urlWithoutQuery.String(),
			Query:   mockURL.Query(),
			Headers: reqHeaders,
			Body:    10,
		}),
	)
	reqHeaders = make(http.Header)
	reqHeaders.Set("content-type", "application/json")
	respHeaders.Set("header", "value")
	assert.Equal(
		t,
		api.Response{
			Body:       []byte(`data`),
			Headers:    respHeaders,
			StatusCode: http.StatusOK,
		},
		handler.Handler{
			Handler: &httpMock,
		}.ServeHTTP(ctx, &logger, &api.Request{
			Method:  "POST",
			URL:     urlWithoutQuery.String(),
			Query:   mockURL.Query(),
			Headers: reqHeaders,
			Body:    "sample body",
		}),
	)
}
