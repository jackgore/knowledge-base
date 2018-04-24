package handlers

import (
	"net/http"
)

type API interface {
	SubmitAnswer(w http.ResponseWriter, r *http.Request)
	SubmitQuestion(w http.ResponseWriter, r *http.Request)
	GetQuestions(w http.ResponseWriter, r *http.Request)

	Signup(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}
