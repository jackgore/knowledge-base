package wrappers

import (
	"fmt"
	"net/http"

	"github.com/JonathonGore/knowledge-base/session"
)

type LoggedInMiddleware struct {
	m session.Manager
}

func (l *LoggedInMiddleware) Initialize(m session.Manager) {
	l.m = m
}

func (l *LoggedInMiddleware) LoggedIn(f func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ensure there is a session cookie attached to the request
		_, err := l.m.GetSession(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)

			// TODO: Currently the JSON function is in handlers package - need to do some refactoring
			// to make that in its own util package

			w.Write([]byte(fmt.Sprintf("{\"message\": \"unauthorized\"}")))
			return
		}

		f(w, r) // Proceed down the call chain
	}
}
