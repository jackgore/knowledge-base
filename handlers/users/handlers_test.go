package users

import (
	"errors"
	"reflect"
	"testing"

	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/user"
)

const (
	validUserID   = 1
	invalidUserID = 2
)

var handler Handler

func init() {
	handler = Handler{&MockStorage{}, &MockSession{}}
}
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

func TestGetUserOrgNames(t *testing.T) {
	orgs, err := handler.getUserOrgNames(validUserID)
	if !reflect.DeepEqual([]string{"Jack", "Hello"}, orgs) || err != nil {
		t.Errorf("Received unexpected error")
	}
}
