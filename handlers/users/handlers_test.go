package users

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/JonathonGore/knowledge-base/models/user"
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

var logoutTests = []struct {
	sessionCookie string
	code          int
}{
	{validCookieValue, 200}, // Valid cookie value should be able to logout
	{"", 500},               // No cookie value should fail
	{"cookie", 500},         // Invalid cookie value should fail
}

var getProfileTests = []struct {
	sessionCookie string
	code          int
	u             user.User
}{
	{validCookieValue, 200, validUser}, // Valid cookie value should be able to logout
	{"", 401, user.User{}},             // No cookie value should fail
	{"cookie", 401, user.User{}},       // Invalid cookie value should fail
}

var getUserTests = []struct {
	username string
	code     int
	u        user.User
}{
	{validUsername, 200, validUser},        // Valid cookie value should be able to logout
	{"invalid username", 404, user.User{}}, // No cookie value should fail
}

func init() {
	log.SetOutput(ioutil.Discard)

	handler = Handler{&MockStorage{}, &MockSession{}}

	router = mux.NewRouter()
	router.HandleFunc("/signup", handler.Signup).Methods(http.MethodPost)
	router.HandleFunc("/login", handler.Login).Methods(http.MethodPost)
	router.HandleFunc("/logout", handler.Logout).Methods(http.MethodPost)
	router.HandleFunc("/profile", handler.GetProfile).Methods(http.MethodGet)
	router.HandleFunc("/users/{username}", handler.GetUser).Methods(http.MethodGet)
}

func TestNew(t *testing.T) {
	_, err := New(nil, nil)
	if err == nil {
		t.Errorf("Expected to receive error when passing nil interfaces")
	}
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

func TestLogout(t *testing.T) {
	for _, test := range logoutTests {
		r, err := http.NewRequest(http.MethodPost, "/logout", nil)
		if err != nil {
			t.Errorf("unexepceted error when creating request %v", err)
		}

		r.Header.Set("Cookie", fmt.Sprintf("%v=%v", testCookieName, test.sessionCookie))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		if test.code != w.Code {
			t.Errorf("Received status code: %v Expected: %v for logout", w.Code, test.code)
		}
	}
}

func TestGetProfile(t *testing.T) {
	for _, test := range getProfileTests {
		r, err := http.NewRequest(http.MethodGet, "/profile", nil)
		if err != nil {
			t.Errorf("unexpected error when creating request %v", err)
		}

		r.Header.Set("Cookie", fmt.Sprintf("%v=%v", testCookieName, test.sessionCookie))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		if test.code != w.Code {
			t.Errorf("Received status code: %v Expected: %v for logout", w.Code, test.code)
		}

		contents, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Errorf("unexpected error when creating request %v", err)
		}

		var u user.User
		err = json.Unmarshal(contents, &u)
		if err != nil {
			t.Errorf("unexpected error when parsing user %v", err)
		}
	}
}

func TestGetUser(t *testing.T) {
	for _, test := range getUserTests {
		r, err := http.NewRequest(http.MethodGet, "/users/"+test.username, nil)
		if err != nil {
			t.Errorf("unexpected error when creating request %v", err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		if test.code != w.Code {
			t.Errorf("Received status code: %v Expected: %v for logout", w.Code, test.code)
		}

		contents, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Errorf("unexpected error when creating request %v", err)
		}

		var u user.User
		err = json.Unmarshal(contents, &u)
		if err != nil {
			t.Errorf("unexpected error when parsing user %v", err)
		}
	}
}
