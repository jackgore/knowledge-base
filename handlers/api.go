package handlers

import (
	"net/http"
)

type API interface {
	HelloWorld(w http.ResponseWriter, r *http.Request)
}
