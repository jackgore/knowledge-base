package users

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
)

const (
	validSignup          = `{"username": "Jacky", "password": "password", "email": "test@test.com"}`
	spacesUsername       = `{"username": "jacky jacky", "password": "password", "email": "test@test.com"}`
	invalidJSONSignup    = `"username": "jacky", "password": "password", "email": "test@test.com"}`
	noUsernameSignup     = `{"username": "", "password": "password", "email": "test@test.com"}`
	noPasswordSignup     = `{"username": "jacky", "password": "", "email": "test@test.com"}`
	shortPasswordSignup  = `{"username": "jacky", "password": "x", "email": "test@test.com"}`
	noEmailSignup        = `{"username": "Jacky", "password": "password"}`
	emptyEmailSignup     = `{"username": "Jacky", "password": "password", "email": ""}`
	malformedEmailSignup = `{"username": "Jacky", "password": "password", "email": "bad email"}`

	validLogin       = `{"username": "jacky", "password": "password"}`
	invalidJSONLogin = `"username": "jacky", "password": "password"}`
	noUsernameLogin  = `{"username": "", "password": "password"}`
	noPasswordLogin  = `{"username": "jacky", "password": ""}`
)

var (
	handler Handler
	router  *mux.Router
)

var signupTests = []struct {
	body string
	code int
}{
	{validSignup, 200},          // Signing up a user with valid credentials should succeed
	{spacesUsername, 400},       // Usernames cannot contian spaces
	{invalidJSONSignup, 400},    // Non JSON should cause a 400, bad request
	{noUsernameSignup, 400},     // Signup credentials without a username provided should fail
	{noPasswordSignup, 400},     // Signup credentials without a password should fail
	{shortPasswordSignup, 400},  // Password cannot be too short
	{noEmailSignup, 400},        // No email in signup should fail
	{emptyEmailSignup, 400},     // Empty email in signup should fail
	{malformedEmailSignup, 400}, // Malformed email in signup should fail
}

var loginTests = []struct {
	body string
	code int
}{
	{validLogin, 200},       // Should be able to login a user with valid credentials
	{invalidJSONLogin, 400}, // Invalid JSON should cause a bad request
	{noUsernameLogin, 400},  // No username should cause a bad request
	{noPasswordLogin, 400},  // No password should cause a bad request
}

func init() {
	log.SetOutput(ioutil.Discard)

	handler = Handler{&MockStorage{}, &MockSession{}}

	router = mux.NewRouter()
	router.HandleFunc("/signup", handler.Signup).Methods(http.MethodPost)
	router.HandleFunc("/login", handler.Login).Methods(http.MethodPost)
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
	for _, test := range signupTests {
		r, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(test.body))
		if err != nil {
			t.Errorf("unexepceted error when creating request %v", err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		content, _ := ioutil.ReadAll(w.Body)

		if test.code != w.Code {
			t.Errorf("Received status code: %v Expected: %v for body: %v", w.Code, test.code, test.body)
			fmt.Printf("Body: %v\n", string(content))
		}
	}
}

func TestLogin(t *testing.T) {
	for _, test := range loginTests {
		r, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(test.body))
		if err != nil {
			t.Errorf("unexepceted error when creating request %v", err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		if test.code != w.Code {
			t.Errorf("Received status code: %v Expected: %v for body: %v", w.Code, test.code, test.body)
		}
	}
}

func TestNew(t *testing.T) {
	_, err := New(nil, nil)
	if err == nil {
		t.Errorf("Expected to receive error when passing nil interfaces")
	}
}
