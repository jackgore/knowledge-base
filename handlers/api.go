package handlers

import (
	"net/http"
)

type API interface {
	GetAnswers(w http.ResponseWriter, r *http.Request)
	SubmitAnswer(w http.ResponseWriter, r *http.Request)

	SubmitQuestion(w http.ResponseWriter, r *http.Request)
	ViewQuestion(w http.ResponseWriter, r *http.Request)
	GetQuestions(w http.ResponseWriter, r *http.Request)
	GetQuestion(w http.ResponseWriter, r *http.Request)
	GetOrgQuestions(w http.ResponseWriter, r *http.Request)
	GetTeamQuestions(w http.ResponseWriter, r *http.Request)
	Search(w http.ResponseWriter, r *http.Request)
	SubmitTeamQuestion(w http.ResponseWriter, r *http.Request)
	SubmitOrgQuestion(w http.ResponseWriter, r *http.Request)

	DeleteUser(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
	GetProfile(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Signup(w http.ResponseWriter, r *http.Request)

	CreateOrganization(w http.ResponseWriter, r *http.Request)
	GetOrganizations(w http.ResponseWriter, r *http.Request)
	GetOrganization(w http.ResponseWriter, r *http.Request)
	GetOrganizationMembers(w http.ResponseWriter, r *http.Request)
	InsertOrganizationMember(w http.ResponseWriter, r *http.Request)

	GetTeams(w http.ResponseWriter, r *http.Request)
	GetTeam(w http.ResponseWriter, r *http.Request)
	CreateTeam(w http.ResponseWriter, r *http.Request)
}
