package organizations

import (
	"encoding/json"
	errs "errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/JonathonGore/knowledge-base/errors"
	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/team"
	"github.com/JonathonGore/knowledge-base/models/user"
	"github.com/JonathonGore/knowledge-base/query"
	sess "github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/util"
	"github.com/JonathonGore/knowledge-base/util/httputil"
	"github.com/gorilla/mux"
)

// storage is the interface required by the organizations handlers to store and
// and retrieve organization data.
type storage interface {
	GetOrganization(orgID int) (organization.Organization, error)
	GetOrganizationByName(name string) (organization.Organization, error)
	GetOrganizations(public bool) ([]organization.Organization, error)
	GetUserByUsername(username string) (user.User, error)
	GetUserOrganizations(uid int) ([]organization.Organization, error)
	GetUsernameOrganizations(username string) ([]organization.Organization, error)
	GetOrganizationMembers(org string, admins bool) ([]string, error)
	InsertOrganization(organization.Organization) (int, error)
	InsertOrgMember(username, org string, isAdmin bool) error
	InsertTeam(t team.Team) error
}

// session is the interface required by the organizations handler for
// interacting with user sessions.
type session interface {
	GetSession(r *http.Request) (sess.Session, error)
}

// Handler is responsible for interacting with the data store
// and session manager to perform business logic.
type Handler struct {
	db             storage
	sessionManager session
}

// orgAddition is used for adding a user to an organization
type orgAddition struct {
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
}

// New creates a new handler for handling requests concerning organizations.
func New(d storage, sm session) (*Handler, error) {
	if d == nil || sm == nil {
		return nil, errs.New("storage driver and session manager must both not be nil")
	}

	return &Handler{d, sm}, nil
}

// joinOrgs consumes to slices of organizations and merges them into a single one
// while removing any duplicates. Right now this operation will take O(n^2).
// Will want to eventually optimize this operation. joinOrgs requires each of orgs1
// and orgs2 to have no duplicates within themselves.
func joinOrgs(orgs1 []organization.Organization, orgs2 []organization.Organization) []organization.Organization {
	for _, org2 := range orgs2 {
		duplicate := false
		for _, org1 := range orgs1 {
			if org1.Name == org2.Name {
				duplicate = true
				break
			}
		}

		if !duplicate {
			orgs1 = append(orgs1, org2)
		}
	}

	return orgs1
}

/* GET /organizations
 *
 * Receieves a page of organizations that are viewable by the requesting user.
 * This ends up being all public organizations and organizations the user
 * belongs to.
 *
 * Query Params:
 *		username: if present fetches only organizations the user belongs to.
 */
func (h *Handler) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	var err error
	var username string

	publicOrgs := []organization.Organization{}
	userOrgs := []organization.Organization{}
	public := true // Used as a paramter to only fetch public organizations

	params := query.ParseParams(r)

	username, ok := params["username"]
	if !ok {
		// If no username param provided that means we want to retrieve all
		// organizations viewable by the requesting user.
		publicOrgs, err = h.db.GetOrganizations(public)
		if err != nil {
			httputil.HandleError(w, errors.DBGetError, http.StatusInternalServerError)
			return
		}

		session, err := h.sessionManager.GetSession(r)
		if err != nil {
			// No session attached to request so set username to "" to prevent retrieval
			// of user organizations
			username = ""
		} else {
			username = session.Username
		}
	}

	// Only retrieve organizations for a username if there is a non-empty username
	if username != "" {
		// Ensure the user is allowed to get orgs for the requested user
		session, err := h.sessionManager.GetSession(r)
		if err != nil || session.Username != username {
			httputil.HandleError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userOrgs, err = h.db.GetUsernameOrganizations(username)
		if err != nil {
			httputil.HandleError(w, errors.DBGetError, http.StatusInternalServerError)
			return
		}
	}

	// Join userOrgs and publicOrgs removing duplicates.
	orgs := joinOrgs(publicOrgs, userOrgs)

	contents, err := json.Marshal(orgs)
	if err != nil {
		httputil.HandleError(w, errors.JSONError, http.StatusInternalServerError)
		return
	}

	w.Write(contents)
}

/* GET /organization/{organization}
 *
 * Retrieves the organization with the provided name.
 *
 * TODO: Require the requester to be a member of the organization or the org to be public.
 */
func (h *Handler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	orgName, ok := mux.Vars(r)["organization"]
	if !ok {
		httputil.HandleError(w, errors.InternalServerError, http.StatusInternalServerError)
		return
	}

	org, err := h.db.GetOrganizationByName(orgName)
	if err != nil {
		httputil.HandleError(w, errors.DBGetError, http.StatusNotFound)
		return
	}

	// Now we must ensure that the requested org is public or the requesting user belongs to the org
	if !org.IsPublic {
		s, err := h.sessionManager.GetSession(r)
		if err != nil {
			httputil.HandleError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		members, err := h.db.GetOrganizationMembers(orgName, false)
		if err != nil {
			httputil.HandleError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if !util.Contains(members, s.Username) {
			// If the requesting user is not an org member then they are unauthorized
			httputil.HandleError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
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
	orgName, ok := mux.Vars(r)["organization"]
	if !ok {
		httputil.HandleError(w, errors.InternalServerError, http.StatusInternalServerError)
		return
	}

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
