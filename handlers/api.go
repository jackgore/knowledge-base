package handlers

import (
	"net/http"
)

type API interface {
	SubmitQuestion(w http.ResponseWriter, r *http.Request)
	Signup(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
}
