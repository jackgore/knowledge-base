package users

import (
	"reflect"
	"testing"
)

const (
	validUserID   = 1
	invalidUserID = 2
)

var handler Handler

func init() {
	handler = Handler{&MockStorage{}, &MockSession{}}
}

func TestGetUserOrgNames(t *testing.T) {
	orgs, err := handler.getUserOrgNames(validUserID)
	if !reflect.DeepEqual([]string{"Jack", "Hello"}, orgs) || err != nil {
		t.Errorf("Received unexpected error")
	}
}
