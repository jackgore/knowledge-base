package organizations

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	org "github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/gorilla/mux"
)

var (
	handler Handler
	router  *mux.Router
)

var getOrganizationsTests = []struct {
	cookie   string
	username string
	code     int
	orgs     []org.Organization
}{
	{"", "", 200, []org.Organization{publicOrg}}, // Getting public organizations should succeed
	{"", "differentUser", 401, nil},              // Trying to request organizations for another user should fail w/ unauthorized
	{validCookieValue, "", 200,
		[]org.Organization{publicOrg, privateUserOrg}}, // While logged in private orgs you belong to are also produced
	{validCookieValue, validUsername, 200,
		[]org.Organization{privateUserOrg}}, // Requesting orgs only for yourself should succeed w/o orgs you don't belong to
}

var joinOrgsTests = []struct {
	orgs1  []org.Organization
	orgs2  []org.Organization
	result []org.Organization
}{
	{[]org.Organization{}, []org.Organization{}, []org.Organization{}},
	{
		[]org.Organization{{Name: "test"}},
		[]org.Organization{},
		[]org.Organization{{Name: "test"}},
	},
	{
		[]org.Organization{{Name: "test"}},
		[]org.Organization{{Name: "jack"}},
		[]org.Organization{{Name: "test"}, {Name: "jack"}},
	},
	{
		[]org.Organization{{Name: "test"}},
		[]org.Organization{{Name: "test"}},
		[]org.Organization{{Name: "test"}},
	},
	{
		[]org.Organization{{Name: "test"}, {Name: "jack"}},
		[]org.Organization{{Name: "jack"}},
		[]org.Organization{{Name: "test"}, {Name: "jack"}},
	},
	{
		[]org.Organization{},
		[]org.Organization{{Name: "test"}, {Name: "jack"}},
		[]org.Organization{{Name: "test"}, {Name: "jack"}},
	},
}

func init() {
	log.SetOutput(ioutil.Discard)

	handler = Handler{&MockStorage{}, &MockSession{}}

	router = mux.NewRouter()
	router.HandleFunc("/organizations", handler.GetOrganizations).Methods(http.MethodGet)
}

func TestNew(t *testing.T) {
	_, err := New(nil, nil)
	if err == nil {
		t.Errorf("Expected to receive error when passing nil interfaces")
	}
}

func TestJoinOrgs(t *testing.T) {
	for _, test := range joinOrgsTests {
		result := joinOrgs(test.orgs1, test.orgs2)

		if !reflect.DeepEqual(result, test.result) {
			t.Errorf("Received result: %v Expected: %v", result, test.result)
		}
	}
}

func TestGetOrganizations(t *testing.T) {
	for _, test := range getOrganizationsTests {
		r, err := http.NewRequest(http.MethodGet, "/organizations", nil)
		if err != nil {
			t.Errorf("unexepceted error when creating request %v", err)
		}

		if test.cookie != "" {
			r.Header.Set("Cookie", fmt.Sprintf("%v=%v", testCookieName, test.cookie))
		}

		if test.username != "" {
			q := r.URL.Query()
			q.Add("username", test.username)
			r.URL.RawQuery = q.Encode()
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		contents, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Errorf("received unexpected error when testing: %v", err)
		}

		orgs := []org.Organization{}
		if json.Unmarshal(contents, &orgs); err != nil {
			t.Errorf("received unexpected error when testing: %v", err)
		}

		if test.code != w.Code {
			t.Errorf("Received status code: %v Expected: %v", w.Code, test.code)
		}

		if test.orgs != nil && !reflect.DeepEqual(orgs, test.orgs) {
			t.Errorf("did not received exepected orgs from /organizations")
		}
	}
}
