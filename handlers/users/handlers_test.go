package users

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
)

const (
	validUserID   = 1
	invalidUserID = 2
	emptyUserID   = 3

	validUserSignup = `{"username": "jacky", "password": "password"}`
)

var (
	handler Handler
	router  *mux.Router
)

func init() {
	log.SetOutput(ioutil.Discard)

	handler = Handler{&MockStorage{}, &MockSession{}}
	router = mux.NewRouter()
	router.HandleFunc("/signup", handler.Signup).Methods(http.MethodPost)
}

func TestGetUserOrgNames(t *testing.T) {
	// User id belonging to no orgs should produce no org names
	orgs, err := handler.getUserOrgNames(emptyUserID)
	if !reflect.DeepEqual([]string{}, orgs) || err != nil {
		t.Errorf("Received unexpected error")
	}

	// Valid user id should produce org names
	orgs, err = handler.getUserOrgNames(validUserID)
	if !reflect.DeepEqual([]string{"Jack", "Hello"}, orgs) || err != nil {
		t.Errorf("Received unexpected error")
	}

	// Invalid user id should produce an error
	_, err = handler.getUserOrgNames(invalidUserID)
	if err == nil {
		t.Errorf("Expected to receive error")
	}
}

func TestSignup(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(validUserSignup))
	if err != nil {
		t.Errorf("unexepceted error when creating request %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if http.StatusOK != w.Code {
		t.Errorf("Received unexpected status code when signing up user with valid credentials: %v", w.Code)
	}
}
