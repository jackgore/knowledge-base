package wrappers

import (
	"net/http"

	"github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/util/httputil"
	"github.com/gorilla/mux"
)

type IsUserMiddleware struct {
	m session.Manager
}

func (u *IsUserMiddleware) Initialize(m session.Manager) {
	u.m = m
}

func (u *IsUserMiddleware) IsUser(f func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, ok := mux.Vars(r)["username"]
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(httputil.JSON(httputil.ErrorResponse{"unauthorized", http.StatusUnauthorized}))
			return
		}

		sess, err := u.m.GetSession(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(httputil.JSON(httputil.ErrorResponse{
				"must be logged in to perform this action",
				http.StatusUnauthorized,
			}))
			return
		}

		if sess.Username != username {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(httputil.JSON(httputil.ErrorResponse{
				"unauthorized",
				http.StatusUnauthorized,
			}))
		} else {
			f(w, r)
		}
	}
}
