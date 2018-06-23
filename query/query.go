package query

import (
	"net/http"
)

// ParseParams consumes an http request and parses the query params
// into a flat map.
func ParseParams(r *http.Request) map[string]interface{} {
	params := make(map[string]interface{})

	m := r.URL.Query()
	for key, vals := range m {
		for _, val := range vals {
			params[key] = val
		}
	}

	return params
}
