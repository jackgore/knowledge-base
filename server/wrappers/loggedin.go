package wrappers

import (
	"log"
	"net/http"

	"github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/util/httputil"
)

type LoggedInMiddleware struct {
	m session.Manager
}

// Initialize the provided logged in middleware with a session manager.
func (l *LoggedInMiddleware) Initialize(m session.Manager) {
	l.m = m
}

// LogginIn ensures that the requesting user is logged in to perform the requested operation.
func (l *LoggedInMiddleware) LoggedIn(f func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := l.m.GetSession(r); err != nil {
			log.Printf("Received unauthenticated request: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(httputil.JSON(httputil.ErrorResponse{"unauthorized", http.StatusUnauthorized}))
		} else {
			f(w, r) // Proceed down the call chain
		}
	}
}
