package teams

import (
	"net/http"
)

type TeamRoutes interface {
	GetTeams(w http.ResponseWriter, r *http.Request)
	GetTeam(w http.ResponseWriter, r *http.Request)
	CreateTeam(w http.ResponseWriter, r *http.Request)
}
