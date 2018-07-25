package wrappers

import (
	"fmt"
	"net/http"

	"github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/storage"
	"github.com/JonathonGore/knowledge-base/util/httputil"
	"github.com/gorilla/mux"
)

type TeamMemberMiddleware struct {
	m  session.Manager
	db storage.Driver
}

func (t *TeamMemberMiddleware) Initialize(m session.Manager, db storage.Driver) {
	t.m = m
	t.db = db
}

func (o *TeamMemberMiddleware) assertMember(w http.ResponseWriter, r *http.Request, f func(http.ResponseWriter, *http.Request), admin bool) {
	team, ok := mux.Vars(r)["team"]
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(httputil.JSON(httputil.ErrorResponse{"unauthorized", http.StatusUnauthorized}))
		return
	}

	org, ok := mux.Vars(r)["organization"]
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(httputil.JSON(httputil.ErrorResponse{"unauthorized", http.StatusUnauthorized}))
		return
	}

	// Ensure there is a session cookie attached to the request
	sess, err := o.m.GetSession(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(httputil.JSON(httputil.ErrorResponse{
			"must be logged in to perform this action",
			http.StatusUnauthorized,
		}))
		return
	}

	members, err := o.db.GetTeamMembers(org, team, admin)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(httputil.JSON(httputil.ErrorResponse{"unauthorized", http.StatusUnauthorized}))
		return
	}

	for _, val := range members {
		if sess.Username == val {
			// Only proceed if the user is a member of the organization
			f(w, r) // Proceed down the call chain
			return
		}
	}

	memberText := "member"
	if admin {
		memberText = "admin"
	}

	w.WriteHeader(http.StatusUnauthorized)
	w.Write(httputil.JSON(httputil.ErrorResponse{
		fmt.Sprintf("you must be a %v of the %v organization to perform this action", memberText, org),
		http.StatusUnauthorized,
	}))
}

// OrgAdmin ensures that the incoming request belongs to a user who is an admin of the
// org in the path param of the request.
func (t *TeamMemberMiddleware) TeamAdmin(f func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.assertMember(w, r, f, true)
	}
}

// OrgMember ensures that the incoming request belongs to a user who is a member of the
// org in the path param of the request.
func (t *TeamMemberMiddleware) TeamMember(f func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.assertMember(w, r, f, false)
	}
}
