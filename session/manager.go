package session

import (
	"net/http"
)

type Manager interface {
	GetSession(r *http.Request) (Session, error)
	HasSession(r *http.Request) bool
	SessionStart(w http.ResponseWriter, r *http.Request, username string) (Session, error)
	SessionDestroy(w http.ResponseWriter, r *http.Request) error
}
