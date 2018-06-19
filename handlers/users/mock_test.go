package users

import (
	"errors"
	"net/http"

	"github.com/JonathonGore/knowledge-base/models/organization"
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

func (m *MockStorage) GetUserOrganizations(uid int) ([]organization.Organization, error) {
	if uid == validUserID {
		orgs := []organization.Organization{
			organization.Organization{Name: "Jack"},
			organization.Organization{Name: "Hello"},
		}

		return orgs, nil
	}

	return nil, errors.New("invalid user id")
}

func (m *MockStorage) InsertUser(user user.User) error {
	return nil
}

func (m *MockStorage) GetUser(userID int) (user.User, error) {
	var u user.User

	return u, nil
}

func (m *MockStorage) GetUserByUsername(username string) (user.User, error) {
	var u user.User

	return u, nil
}
