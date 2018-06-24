package wrappers

import (
	"log"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := NewLoggingResponseWriter(w)

		start := time.Now()
		handler.ServeHTTP(lrw, r)
		statusCode := lrw.statusCode

		query := ""
		if r.URL.RawQuery != "" {
			query = "?" + r.URL.RawQuery
		}

		log.Printf("%v %v %v%v took %v seconds", statusCode, r.Method, r.URL.Path, query, time.Since(start).Seconds())
	})
}
