package handlers

import (
	"context"
	"net/http"

	"github.com/graphql-editor/stucco/pkg/router"
)

func protocolFromRequest(r *http.Request) map[string]interface{} {
	headers := r.Header.Clone()
	if headers.Get("x-forwarded-proto") == "" {
		if r.TLS != nil {
			headers.Set("x-forwarded-proto", "https")
		}
	}
	return map[string]interface{}{
		"method":        r.Method,
		"headers":       r.Header.Clone(),
		"host":          r.Host,
		"remoteAddress": r.RemoteAddr,
		"proto":         r.Proto,
		"url": map[string]interface{}{
			"query":  r.URL.RawQuery,
			"path":   r.URL.Path,
			"host":   r.URL.Host,
			"scheme": r.URL.Scheme,
		},
	}
}

// WithProtocolInContext appends request headers to context object
func WithProtocolInContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rawSub := r.URL.Query().Get("raw_subscription")
		next.ServeHTTP(
			rw,
			r.WithContext(
				context.WithValue(
					context.WithValue(r.Context(), router.ProtocolKey, protocolFromRequest(r)),
					router.RawSubscriptionKey,
					rawSub == "true",
				),
			),
		)
	})
}
