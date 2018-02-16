package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

/* Consumes an http request and marshals the body into the
 * provided interface.
 *
 * Returns: error should one occur
 */
func UnmarshalRequestBody(r *http.Request, v interface{}) error {
	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Unable to read incoming request body: %v", err)
		return err
	}

	err = json.Unmarshal(contents, v)
	if err != nil {
		log.Printf("Unable to parse incoming request body: %v", err)
		return err

	}

	return nil
}
