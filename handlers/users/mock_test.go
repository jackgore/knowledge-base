package users

import (
	"net/http"

	"github.com/JonathonGore/knowledge-base/models/user"
	sess "github.com/JonathonGore/knowledge-base/session"
)

type MockSession struct{}

func (m *MockSession) GetSession(r *http.Request) (sess.Session, error) {
	var s sess.Session

	return s, nil
}

func (m *MockSession) HasSession(r *http.Request) bool {
	return true
}

func (m *MockSession) SessionStart(w http.ResponseWriter, r *http.Request, username string) (sess.Session, error) {
	var s sess.Session

	return s, nil
}

func (m *MockSession) SessionDestroy(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type MockStorage struct{}

func (m *MockStorage) GetUser(userID int) (user.User, error) {
	var u user.User

	return u, nil
}

func (m *MockStorage) GetUserByUsername(username string) (user.User, error) {
	var u user.User

	return u, nil
}
