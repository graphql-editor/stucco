package handlers

import (
	"context"
	"net/http"

	"github.com/graphql-editor/stucco/pkg/router"
)

// WithProtocolInContext appends request headers to context object
func WithProtocolInContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(
			rw,
			r.WithContext(
				context.WithValue(
					r.Context(),
					router.ProtocolKey, map[string]interface{}{
						"headers": r.Header,
					},
				),
			),
		)
	})
}
