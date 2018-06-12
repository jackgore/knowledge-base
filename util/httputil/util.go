package httputil

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
)

// HandleError is a utility function to responsd to a response writer with
// the given message and error code as well as log the message
func HandleError(w http.ResponseWriter, message string, code int) {
	_, fn, line, _ := runtime.Caller(1)
	log.Printf("Error at: %v:%v - %v", fn, line, message)
	w.WriteHeader(code)
	w.Write(JSON(ErrorResponse{message, code}))
}

// JSONString consumes an interface and marshals it into a JSON representation
// in string format.
func JSONString(e interface{}) string {
	return string(JSON(e))
}

// JSON consumes an interface and marshals it into a JSON representation
// stored in a byte slice.
func JSON(e interface{}) []byte {
	contents, err := json.Marshal(e)
	if err != nil {
		return []byte{}
	}

	return contents
}

// UnmarshalRequestBody consumes an http request and marshals the body into the
// provided interface.
func UnmarshalRequestBody(r *http.Request, v interface{}) error {
	if r == nil {
		return errors.New("http request must be non-nil")
	}

	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(contents, v)
	if err != nil {
		return err
	}

	return nil
}
