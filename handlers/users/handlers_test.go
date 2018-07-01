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
	validSignupUser    = `{"username": "Jacky", "password": "password"}`
	validLoginUser     = `{"username": "jacky", "password": "password"}`
	noUsernameUser     = `{"username": "", "password": "password"}`
	noPasswordUser     = `{"username": "jacky", "password": ""}`
	invalidJSONUser    = `"username": "jacky", "password": "password"}`
	spacesUsernameUser = `{"username": "jacky jacky", "password": "password"}`
	shortPasswordUser  = `{"username": "jacky jacky", "password": "x"}`
)

var (
	handler Handler
	router  *mux.Router
)

var signupTests = []struct {
	body string
	code int
}{
	{validSignupUser, 200},
	{invalidJSONUser, 400},
	{noUsernameUser, 400},
	{noPasswordUser, 400},
	{spacesUsernameUser, 400},
	{shortPasswordUser, 400},
}

var loginTests = []struct {
	body string
	code int
}{
	{validLoginUser, 200},
	{invalidJSONUser, 400},
	{noUsernameUser, 400},
	{noPasswordUser, 400},
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
