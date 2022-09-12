package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-go/graphql"
	"k8s.io/klog"
)

func NewWebhookHandler(c Config, gqlHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		protocolData := map[string]interface{}{
			"method": r.Method,
			"url": map[string]string{
				"path":  r.URL.Path,
				"host":  r.URL.Host,
				"query": r.URL.RawQuery,
			},
			"host":          r.Host,
			"remoteAddress": r.RemoteAddr,
			"proto":         r.Proto,
			"headers":       r.Header.Clone(),
		}
		if r.Body != nil {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(rw, "could not read request body", http.StatusInternalServerError)
				return
			}
			r.Body.Close()
			if len(body) > 0 {
				protocolData["body"] = base64.StdEncoding.EncodeToString(body)
			}
		}
		pathParts := strings.Split(strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"), "/"), "/")
		op := strings.ToLower(pathParts[1])
		if len(pathParts) != 3 || (op != "query" && op != "mutation") || len(pathParts) == 0 {
			http.Error(rw, "Invalid webhook path, must be /webhook/<query|mutation>/<field>", http.StatusNotFound)
			return
		}
		r.Body = io.NopCloser(strings.NewReader(`{ "query": "` + op + `{` + pathParts[2] + `}" }`))
		r.Method = "POST"
		r.Header.Set("content-type", "application/json")
		r = r.WithContext(context.WithValue(r.Context(), router.ProtocolKey, protocolData))
		respProxy := webhookResponseWrapper{
			ResponseWriter: rw,
		}
		gqlHandler.ServeHTTP(&respProxy, r)
		status := respProxy.status
		if respProxy.status >= 200 && respProxy.status < 300 {
			var resp graphql.Result
			if err := json.Unmarshal(respProxy.Buffer.Bytes(), &resp); err != nil {
				status = http.StatusInternalServerError
			} else if resp.HasErrors() {
				status = http.StatusBadRequest
			}
		}
		rw.WriteHeader(status)
		if _, err := io.Copy(rw, &respProxy.Buffer); err != nil {
			klog.Error(err)
		}
	})
}
