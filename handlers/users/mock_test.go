package users

import (
	"errors"
	"net/http"

	"github.com/JonathonGore/knowledge-base/creds"
	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/user"
	sess "github.com/JonathonGore/knowledge-base/session"
)

// These constants determine which values are returned by mock functions.
const (
	validUserID   = 1
	invalidUserID = 2
	emptyUserID   = 3

	validUsername = "jacky"
	validPassword = "password"
)

var (
	validUser = user.User{
		Username: validUsername,
		Password: "",
	}
)

// MockSession is a mock implementation of the mock session component used by the users handler.
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

// MockStorage is a mock implementation of the mock storage component used by the users handler.
type MockStorage struct{}

func (m *MockStorage) GetUserOrganizations(uid int) ([]organization.Organization, error) {
	if uid == validUserID {
		orgs := []organization.Organization{
			organization.Organization{Name: "Jack"},
			organization.Organization{Name: "Hello"},
		}

		return orgs, nil
	} else if uid == emptyUserID {
		return []organization.Organization{}, nil
	}

	return nil, errors.New("invalid user id")
}

func (m *MockStorage) InsertUser(user user.User) error {
	return nil
}

func (m *MockStorage) GetUser(userID int) (user.User, error) {
	var u user.User

	if userID == validUserID {
		return validUser, nil
	}

	return u, errors.New("invalid userid")
}

// This function must mimick the behaviour of stored passwords which are hashed with bcrypt.
func (m *MockStorage) GetUserByUsername(username string) (user.User, error) {
	var u user.User

	if username == validUsername {
		u = validUser
		hash, err := creds.HashPassword(u.Password)
		if err != nil {
			return u, nil
		}

		u.Password = hash
		return u, nil
	}

	return u, errors.New("invalid username")
}
