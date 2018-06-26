package organizations

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/JonathonGore/knowledge-base/errors"
	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/team"
	"github.com/JonathonGore/knowledge-base/query"
	"github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/storage"
	"github.com/JonathonGore/knowledge-base/util"
	"github.com/JonathonGore/knowledge-base/util/httputil"
	"github.com/gorilla/mux"
)

type Handler struct {
	db             storage.Driver
	sessionManager session.Manager
}

type orgAddition struct {
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
}

func New(d storage.Driver, sm session.Manager) (*Handler, error) {
	return &Handler{d, sm}, nil
}

/* GET /organizations
 *
 * Receives a page of organizations
 * TODO: accept query params
 */
func (h *Handler) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.db.GetOrganizations()
	if err != nil {
		httputil.HandleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	contents, err := json.Marshal(orgs)
	if err != nil {
		httputil.HandleError(w, errors.JSONError, http.StatusInternalServerError)
		return
	}

	w.Write(contents)
}

/* GET /organization/{organization}
 *
 * Receives a single organization
 * TODO: accept query params
 */
func (h *Handler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	orgName := mux.Vars(r)["organization"]

	org, err := h.db.GetOrganizationByName(orgName)
	if err != nil {
		httputil.HandleError(w, errors.DBGetError, http.StatusNotFound)
		return
	}

	contents, err := json.Marshal(org)
	if err != nil {
		httputil.HandleError(w, errors.JSONError, http.StatusInternalServerError)
		return
	}

	w.Write(contents)
}

/* GET /organization/{organization}/members
 *
 * Receives a single organization
 * TODO: accept query params
 */
func (h *Handler) GetOrganizationMembers(w http.ResponseWriter, r *http.Request) {
	orgName := mux.Vars(r)["organization"]
	params := query.ParseParams(r)

	admins := false
	if val, ok := params["admins"]; ok {
		if val == "true" {
			admins = true
		}
	}

	org, err := h.db.GetOrganizationByName(orgName)
	if err != nil {
		httputil.HandleError(w, errors.DBGetError, http.StatusNotFound)
		return
	}

	members, err := h.db.GetOrganizationMembers(org.Name, admins)
	if err != nil {
		httputil.HandleError(w, errors.JSONError, http.StatusInternalServerError)
		return
	}

	contents, err := json.Marshal(members)
	if err != nil {
		httputil.HandleError(w, errors.JSONError, http.StatusInternalServerError)
		return
	}

	w.Write(contents)
}

/* POST /organizations/{organization}/members
 *
 * Adds the member to the organization
 *
 * Expected body: { "username": "<username>", "admin": false }
 *
 * NOTE: We need to assume that this function is called by and admin of the org
 * which should be handled by our middleware
 */
func (h *Handler) InsertOrganizationMember(w http.ResponseWriter, r *http.Request) {
	org := mux.Vars(r)["organization"]

	_, err := h.db.GetOrganizationByName(org)
	if err != nil {
		msg := fmt.Sprintf("Organization %v does not exist", org)
		httputil.HandleError(w, msg, http.StatusBadRequest)
		return
	}

	member := orgAddition{}
	err = httputil.UnmarshalRequestBody(r, &member)
	if err != nil {
		httputil.HandleError(w, errors.JSONParseError, http.StatusBadRequest)
		return
	}

	user, err := h.db.GetUserByUsername(member.Username)
	if err != nil {
		msg := fmt.Sprintf("User %v does not exist", member.Username)
		httputil.HandleError(w, msg, http.StatusBadRequest)
		return
	}

	// If user is already a member return a 400
	members, err := h.db.GetOrganizationMembers(org, false)
	if err != nil {
		httputil.HandleError(w, "Internal server error", http.StatusBadRequest)
		return
	}

	if util.Contains(members, user.Username) {
		msg := fmt.Sprintf("User: %v is already a member of organization: %v", user.Username, org)
		httputil.HandleError(w, msg, http.StatusBadRequest)
		return
	}

	err = h.db.InsertOrgMember(user.Username, org, true)
	if err != nil {
		log.Printf("unable to insert user as member: %v", err)
		httputil.HandleError(w, errors.DBInsertError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* POST /organizations
 *
 * Creates a new organization
 *
 * Note: Error messages here are user facing
 */
func (h *Handler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	org := organization.Organization{}
	err := httputil.UnmarshalRequestBody(r, &org)
	if err != nil {
		httputil.HandleError(w, errors.JSONParseError, http.StatusBadRequest)
		return
	}

	err = organization.Validate(org)
	if err != nil {
		httputil.HandleError(w, err.Error(), http.StatusBadRequest)
		return
	}

	o, err := h.db.GetOrganizationByName(org.Name)
	if err == nil {
		msg := fmt.Sprintf("Organization %v already exists", o.Name)
		httputil.HandleError(w, msg, http.StatusBadRequest)
		return
	} else {
		log.Printf("Error: %v", err)
	}

	org.CreatedOn = time.Now()

	id, err := h.db.InsertOrganization(org)
	if err != nil {
		log.Printf("Unable to insert organization %v into database: %v", org.Name, err)
		httputil.HandleError(w, errors.DBInsertError, http.StatusInternalServerError)
		return
	}

	sess, err := h.sessionManager.GetSession(r)
	if err != nil {
		msg := "Must be logged in to create an organization"
		httputil.HandleError(w, msg, http.StatusUnauthorized)
		return
	}

	err = h.db.InsertOrgMember(sess.Username, org.Name, true) // Org creator is added as an admin
	if err != nil {
		log.Printf("unable to insert user as member: %v", err)
		httputil.HandleError(w, errors.DBInsertError, http.StatusInternalServerError)
		return
	}

	// We want to have a default team for every organization - call it `default`
	defaultTeam := team.Team{
		Name:         "default",
		Organization: id,
		CreatedOn:    time.Now(),
		IsPublic:     org.IsPublic,
		MemberCount:  1,
		AdminCount:   1,
	}

	err = h.db.InsertTeam(defaultTeam)
	if err != nil {
		httputil.HandleError(w, errors.DBInsertError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
