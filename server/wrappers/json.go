package wrappers

import (
	"net/http"
)

// JSONResponse adds a Content-Type header to wrapped requests.
func JSONResponse(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler.ServeHTTP(w, r)
	})
}
