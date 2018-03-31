package server

import (
	"net/http"

	"github.com/JonathonGore/knowledge-base/handlers"
	"github.com/JonathonGore/knowledge-base/handlers/wrappers"
	"github.com/gorilla/mux"
)

type Server struct {
	Router *mux.Router
}

func New(api handlers.API) (*Server, error) {
	s := &Server{Router: mux.NewRouter()}

	s.Router.HandleFunc("/questions", api.SubmitQuestion).Methods(http.MethodPost)
	s.Router.HandleFunc("/questions/{id}", api.SubmitAnswer).Methods(http.MethodPost)
	s.Router.HandleFunc("/questions", api.GetQuestions).Methods(http.MethodGet)
	s.Router.HandleFunc("/users", api.Signup).Methods(http.MethodPost)
	s.Router.HandleFunc("/users/{username}", api.GetUser).Methods(http.MethodGet)

	// Attach middleware to mux router
	s.Router.Use(wrappers.Log)

	return s, nil
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin) // Restrict this to proper origins
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
