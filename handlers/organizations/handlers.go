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
	"github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/storage"
	"github.com/JonathonGore/knowledge-base/util/httputil"
	"github.com/gorilla/mux"
)

type Handler struct {
	db             storage.Driver
	sessionManager session.Manager
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

	// TODO: Add existance check
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
	return

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
		httputil.HandleError(w, errors.CreateResourceError, http.StatusBadRequest)
		return
	}

	_, err = h.db.GetOrganizationByName(org.Name)
	if err == nil {
		msg := fmt.Sprintf("Attempted to create organization %v but name already exists", org.Name)
		httputil.HandleError(w, msg, http.StatusBadRequest)
		return
	}

	org.CreatedOn = time.Now()

	id, err := h.db.InsertOrganization(org)
	if err != nil {
		log.Printf("Unable to insert organization into database: %v", err)
		httputil.HandleError(w, errors.DBInsertError, http.StatusBadRequest)
		return
	}

	sess, err := h.sessionManager.GetSession(r)
	if err != nil {
		msg := "Must be logged in to create an organization"
		httputil.HandleError(w, msg, http.StatusUnauthorized)
		return
	}

	err = h.db.InsertOrgMember(sess.Username, org.Name, true)
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