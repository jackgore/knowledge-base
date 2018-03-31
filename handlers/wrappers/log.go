package wrappers

import (
	"log"
	"net/http"
	"time"
)

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler.ServeHTTP(w, r)
		log.Printf("%v %v took %v seconds", r.Method, r.URL.Path, time.Since(start).Seconds())
	})
}
