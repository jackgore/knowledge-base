package questions

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

const (
	validUserID   = 1
	invalidUserID = 2
	emptyUserID   = 3

	validQuestion        = `{"title": "Where is the wifi password", "content": "Not sure where to look"}`
	noTitleQuestion      = `{"title": "", "content": "content"}`
	noContentQuestion    = `{"title": "jacky", "content": ""}`
	shortContentQuestion = `{"title": "jacky", "content": "x"}`
	shortTitleQuestion   = `{"title": "a", "content": "Not sure where to look"}`
	invalidJSONQuestion  = `"title": "jacky", "content": "content"}`
)

var (
	handler Handler
	router  *mux.Router
)

var submitTests = []struct {
	body string
	code int
}{
	{validQuestion, 200},
	{invalidJSONQuestion, 400},
	{noContentQuestion, 400},
	{shortContentQuestion, 400},
	{shortTitleQuestion, 400},
}

func init() {
	log.SetOutput(ioutil.Discard)

	handler = Handler{&MockStorage{}, &MockSession{}, &MockSearch{}}
	router = mux.NewRouter()
	router.HandleFunc("/questions", handler.SubmitQuestion).Methods(http.MethodPost)
}

func TestSubmitQuestion(t *testing.T) {
	for _, test := range submitTests {
		r, err := http.NewRequest(http.MethodPost, "/questions", bytes.NewBufferString(test.body))
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
	_, err := New(nil, nil, nil)
	if err == nil {
		t.Errorf("Expected to receive error when passing nil interfaces")
	}
}
