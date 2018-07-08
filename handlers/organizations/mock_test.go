package organizations

import (
	"errors"
	"net/http"

	"github.com/JonathonGore/knowledge-base/creds"
	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/team"
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

	testCookieName   = "kb-test-cookie"
	validCookieValue = "valid cookie"

	publicOrgName  = "publicOrg"
	privateOrgName = "privateOrg"
)

var (
	validUser = user.User{
		ID:       validUserID,
		Username: validUsername,
		Password: "",
	}

	publicOrg = organization.Organization{
		Name:     publicOrgName,
		IsPublic: true,
	}

	privateUserOrg = organization.Organization{
		Name:     privateOrgName,
		IsPublic: false,
	}
)

// MockSession is a mock implementation of the mock session component used by the users handler.
type MockSession struct{}

// GetSession retrieves a session based on the attached cookie.
func (m *MockSession) GetSession(r *http.Request) (sess.Session, error) {
	var s sess.Session

	c, err := r.Cookie(testCookieName)
	if err != nil {
		return s, errors.New("No cookie attached")
	}

	if c.Value != validCookieValue {
		return s, errors.New("Invalid cookie value")
	}

	s.Username = validUsername

	return s, nil
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

func (m *MockStorage) GetOrganization(orgID int) (organization.Organization, error) {
	return organization.Organization{}, nil
}

func (m *MockStorage) GetOrganizationByName(name string) (organization.Organization, error) {
	if name == publicOrgName {
		return publicOrg, nil
	} else if name == privateOrgName {
		return privateUserOrg, nil
	}

	return organization.Organization{}, nil
}

func (m *MockStorage) GetOrganizations(public bool) ([]organization.Organization, error) {
	if public {
		return []organization.Organization{publicOrg}, nil
	}

	return []organization.Organization{}, nil
}

func (m *MockStorage) GetUsernameOrganizations(username string) ([]organization.Organization, error) {
	if username == validUsername {
		return []organization.Organization{privateUserOrg}, nil
	}

	return []organization.Organization{}, nil
}

func (m *MockStorage) GetOrganizationMembers(org string, admins bool) ([]string, error) {
	if org == privateOrgName {
		return []string{validUsername}, nil
	}

	return []string{}, nil
}

func (m *MockStorage) InsertOrganization(organization.Organization) (int, error) {
	return 1, nil
}

func (m *MockStorage) InsertOrgMember(username, org string, isAdmin bool) error {
	return nil
}

func (m *MockStorage) InsertTeam(t team.Team) error {
	return nil
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
