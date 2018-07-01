package wrappers

import (
	"fmt"
	"net/http"

	"github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/storage"
	"github.com/JonathonGore/knowledge-base/util/httputil"
	"github.com/gorilla/mux"
)

type OrgMemberMiddleware struct {
	m  session.Manager
	db storage.Driver
}

func (o *OrgMemberMiddleware) Initialize(m session.Manager, db storage.Driver) {
	o.m = m
	o.db = db
}

func (o *OrgMemberMiddleware) assertMember(w http.ResponseWriter, r *http.Request, f func(http.ResponseWriter, *http.Request), admin bool) {
	// Allow path name to be either org or organization
	org, ok := mux.Vars(r)["org"]
	if !ok {
		org, ok = mux.Vars(r)["organization"]
	}

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

	members, err := o.db.GetOrganizationMembers(org, admin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(httputil.JSON(httputil.ErrorResponse{"internal server error", http.StatusInternalServerError}))
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
func (o *OrgMemberMiddleware) OrgAdmin(f func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		o.assertMember(w, r, f, true)
	}
}

// OrgMember ensures that the incoming request belongs to a user who is a member of the
// org in the path param of the request.
func (o *OrgMemberMiddleware) OrgMember(f func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		o.assertMember(w, r, f, false)
	}
}
