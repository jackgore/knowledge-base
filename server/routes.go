package server

import (
	"fmt"
	"net/http"

	"github.com/JonathonGore/knowledge-base/handlers"
	"github.com/JonathonGore/knowledge-base/server/wrappers"
	"github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/storage"
	"github.com/gorilla/mux"
)

var (
	l = wrappers.LoggedInMiddleware{}
	u = wrappers.IsUserMiddleware{}
	o = wrappers.OrgMemberMiddleware{}
)

type Server struct {
	Router *mux.Router
}

// isPublicHandler consumes the allowPublic configuration variable and determines
// if we should expose routes for manipulating public questions.
func isPublicHandler(allowPublic bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(`{"allow-public": %v}`, allowPublic)))
	}
}

// New creates a new server with routes from the provided api.
func New(api handlers.API, sm session.Manager, db storage.Driver, allowPublic bool) (*Server, error) {
	s := &Server{Router: mux.NewRouter()}

	l.Initialize(sm)
	u.Initialize(sm)
	o.Initialize(sm, db)

	s.Router.HandleFunc("/public", isPublicHandler(allowPublic))

	if allowPublic {
		// Only allow public questions if configured to do so
		s.Router.HandleFunc("/questions", api.SubmitQuestion).Methods(http.MethodPost)
		s.Router.HandleFunc("/questions", api.GetQuestions).Methods(http.MethodGet)
		s.Router.HandleFunc("/questions", api.SubmitQuestion).Methods(http.MethodPost)
	}

	s.Router.HandleFunc("/search", api.Search).Methods(http.MethodGet)
	s.Router.HandleFunc("/questions/{id}/answers", api.SubmitAnswer).Methods(http.MethodPost)
	s.Router.HandleFunc("/questions/{id}/answers", api.GetAnswers).Methods(http.MethodGet)
	s.Router.HandleFunc("/questions/{id}/view", api.ViewQuestion).Methods(http.MethodPost)
	s.Router.HandleFunc("/questions/{id}/upvote", api.UpvoteQuestion).Methods(http.MethodPost)
	s.Router.HandleFunc("/questions/{id}/downvote", api.DownvoteQuestion).Methods(http.MethodPost)
	s.Router.HandleFunc("/questions/{id}", api.GetQuestion).Methods(http.MethodGet)
	s.Router.HandleFunc("/questions/{id}", api.DeleteQuestion).Methods(http.MethodDelete)
	s.Router.HandleFunc("/organizations/{org}/questions", o.OrgMember(api.GetOrgQuestions)).Methods(http.MethodGet)
	s.Router.HandleFunc("/organizations/{org}/questions", o.OrgMember(api.SubmitOrgQuestion)).Methods(http.MethodPost)
	s.Router.HandleFunc("/organizations/{org}/teams/{team}/questions", api.GetTeamQuestions).Methods(http.MethodGet)
	s.Router.HandleFunc("/organizations/{org}/teams/{team}/questions", api.SubmitTeamQuestion).Methods(http.MethodPost)

	s.Router.HandleFunc("/users", api.Signup).Methods(http.MethodPost)
	s.Router.HandleFunc("/users/{username}", api.GetUser).Methods(http.MethodGet)
	s.Router.HandleFunc("/users/{username}", u.IsUser(api.DeleteUser)).Methods(http.MethodDelete)
	s.Router.HandleFunc("/profile", api.GetProfile).Methods(http.MethodGet)
	s.Router.HandleFunc("/login", api.Login).Methods(http.MethodPost)
	s.Router.HandleFunc("/logout", api.Logout).Methods(http.MethodPost)

	s.Router.HandleFunc("/organizations", api.GetOrganizations).Methods(http.MethodGet)
	s.Router.HandleFunc("/organizations/{organization}", api.GetOrganization).Methods(http.MethodGet)
	s.Router.HandleFunc("/organizations/{organization}/members", api.GetOrganizationMembers).Methods(http.MethodGet)
	s.Router.HandleFunc("/organizations/{organization}/members", o.OrgAdmin(api.InsertOrganizationMember)).Methods(http.MethodPost)
	s.Router.HandleFunc("/organizations", l.LoggedIn(api.CreateOrganization)).Methods(http.MethodPost)

	s.Router.HandleFunc("/organizations/{organization}/teams/{team}", api.GetTeam).Methods(http.MethodGet)
	s.Router.HandleFunc("/organizations/{organization}/teams", o.OrgMember(api.CreateTeam)).Methods(http.MethodPost)
	s.Router.HandleFunc("/organizations/{organization}/teams", api.GetTeams).Methods(http.MethodGet)

	s.Router.Use(wrappers.Log)
	s.Router.Use(wrappers.JSONResponse) // All of our routes should return JSON

	return s, nil
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin) // TODO: Restrict this to proper origins
		rw.Header().Set("Access-Control-Allow-Credentials", "true")
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}

	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		return
	}

	// Lets Gorilla work
	s.Router.ServeHTTP(rw, req)
}
