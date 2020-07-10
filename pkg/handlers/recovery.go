package handlers

import "net/http"

// ErrorLog used by recovery handler
type ErrorLog interface {
	Errorf(string, ...interface{})
}

// RecoveryHandler recovers from panics to return Internal Server Error http response
func RecoveryHandler(next http.Handler, logger ErrorLog) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				if logger != nil {
					logger.Errorf("%v\n", err)
				}
				rw.Header().Set("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusInternalServerError)
				if _, err := rw.Write(
					[]byte("There was an internal server error"),
				); err != nil && logger != nil {
					logger.Errorf("%v\n", err)
				}
			}
		}()
		next.ServeHTTP(rw, r)
	})
}
