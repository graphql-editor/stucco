package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"

	"github.com/graphql-editor/azure-functions-golang-worker/api"
	"github.com/pkg/errors"
)

type responseWriter struct {
	body       []byte
	statusCode int
	headers    http.Header
}

func (w *responseWriter) Header() http.Header {
	return w.headers
}

func (w *responseWriter) Write(b []byte) (int, error) {
	statusCode := w.statusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	w.WriteHeader(statusCode)
	w.body = append(w.body, b...)
	return len(b), nil
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = http.StatusOK
}

// Handler is a proxy that allows using http.Handler with function
type Handler struct {
	http.Handler
}

func makeReader(b []byte) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader(b))
}

func serializeBody(v interface{}, contentType string) (r io.ReadCloser, ok bool) {
	if v == nil {
		return
	}
	b, ok := v.([]byte)
	if !ok {
		mt, _, _ := mime.ParseMediaType(contentType)
		switch mt {
		case "application/json":
			var err error
			if b, err = json.Marshal(v); err == nil {
				ok = true
			}
		default:
			var s string
			s, ok = v.(string)
			if ok {
				b = []byte(s)
			}
		}
	}
	if ok {
		r = makeReader(b)
	}
	return
}

func errorResponse(err error, logger api.Logger) api.Response {
	logger.Error(err.Error())
	headers := make(http.Header)
	headers.Set("content-type", "text/plain; charset=utf-8")
	return api.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       []byte(err.Error()),
		Headers:    headers,
	}
}

// ServeHTTP using http.Handler
func (h Handler) ServeHTTP(ctx context.Context, logger api.Logger, req *api.Request) api.Response {
	u := req.URL
	q := req.Query.Encode()
	if len(q) > 0 {
		u += "?" + q
	}
	reqURL, err := url.Parse(u)
	if err != nil {
		return errorResponse(errors.Errorf("invalid url '%s' from function", u), logger)
	}
	resp := responseWriter{
		headers: make(http.Header),
	}
	httpReq := &http.Request{
		Method: req.Method,
		URL:    reqURL,
		Header: req.Headers,
	}
	httpReq = httpReq.WithContext(ctx)
	contentType := req.Headers.Get("content-type")
	if r, ok := serializeBody(req.RawBody, contentType); ok {
		httpReq.Body = r
	} else if r, ok = serializeBody(req.Body, contentType); ok {
		httpReq.Body = r
	} else {
		if req.RawBody != nil || req.Body != nil {
			body := req.RawBody
			if body == nil {
				body = req.Body
			}
			return errorResponse(errors.Errorf("could not serialize body of type %T", body), logger)
		}
	}
	h.Handler.ServeHTTP(&resp, httpReq)
	if contentType := resp.headers.Get("content-type"); contentType == "" {
		resp.headers.Set("content-type", http.DetectContentType(resp.body))
	}
	return api.Response{
		Body:       resp.body,
		Headers:    resp.headers,
		StatusCode: resp.statusCode,
	}
}
