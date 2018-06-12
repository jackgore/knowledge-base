package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/JonathonGore/knowledge-base/errors"
	"github.com/JonathonGore/knowledge-base/handlers/questions"
	"github.com/JonathonGore/knowledge-base/handlers/users"
	"github.com/JonathonGore/knowledge-base/models/answer"
	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/team"
	"github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/storage"
	"github.com/JonathonGore/knowledge-base/util/httputil"
	"github.com/gorilla/mux"
)

type Handler struct {
	users.UserRoutes
	questions.QuestionRoutes

	db             storage.Driver
	sessionManager session.Manager
}

func New(d storage.Driver, sm session.Manager) (*Handler, error) {
	userHandler, err := users.New(d, sm)
	if err != nil {
		return nil, err
	}

	questionHandler, err := questions.New(d, sm)
	if err != nil {
		return nil, err
	}

	handler := &Handler{
		UserRoutes:     userHandler,
		QuestionRoutes: questionHandler,
		db:             d,
		sessionManager: sm,
	}

	return handler, nil
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

/* GET /organizations/<organization>/teams/{team}
 *
 * Receives the team within the requested organization
 */
func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	orgName := params["organization"]
	teamName := params["team"]

	_, err := h.db.GetOrganizationByName(orgName)
	if err != nil {
		httputil.HandleError(w, errors.ResourceNotFoundError, http.StatusBadRequest)
		return
	}

	team, err := h.db.GetTeamByName(orgName, teamName)
	if err != nil {
		httputil.HandleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	contents, err := json.Marshal(team)
	if err != nil {
		httputil.HandleError(w, errors.JSONError, http.StatusInternalServerError)
		return
	}

	w.Write(contents)
}

/* GET /organizations/<organization>/teams
 *
 * Receives a page of teams within an organization
 * TODO: accept query params
 */
func (h *Handler) GetTeams(w http.ResponseWriter, r *http.Request) {
	orgName, _ := mux.Vars(r)["organization"]

	_, err := h.db.GetOrganizationByName(orgName)
	if err != nil {
		httputil.HandleError(w, errors.ResourceNotFoundError, http.StatusBadRequest)
		return
	}

	teams, err := h.db.GetTeams(orgName)
	if err != nil {
		httputil.HandleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	contents, err := json.Marshal(teams)
	if err != nil {
		httputil.HandleError(w, errors.JSONError, http.StatusInternalServerError)
		return
	}

	w.Write(contents)
}

/* POST /organizations/<organization>/teams
 *
 * Creates a new team within the organization
 *
 * Note: Error messages here are user facing
 */
func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	orgName, _ := mux.Vars(r)["organization"]

	t := team.Team{}
	err := httputil.UnmarshalRequestBody(r, &t)
	if err != nil {
		httputil.HandleError(w, errors.JSONParseError, http.StatusInternalServerError)
		return
	}

	err = team.Validate(t)
	if err != nil {
		httputil.HandleError(w, errors.CreateResourceError, http.StatusBadRequest)
		return
	}

	org, err := h.db.GetOrganizationByName(orgName)
	if err != nil {
		httputil.HandleError(w, errors.ResourceNotFoundError, http.StatusBadRequest)
		return
	}

	_, err = h.db.GetTeamByName(orgName, t.Name)
	if err == nil {
		msg := fmt.Sprintf("Team with name %v within %v already exists", t.Name, orgName)
		httputil.HandleError(w, msg, http.StatusBadRequest)
		return
	}

	t.Organization = org.ID // Link the team to the org

	err = h.db.InsertTeam(t)
	if err != nil {
		httputil.HandleError(w, errors.DBInsertError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // TODO: Simple response body instead of just code
}

/* POST /questions/{id}/answers
 *
 * Expected: { "content":<string> }
 *
 *	All other values will be inferred from context/path paramater
 *
 * Receives an answer to the question with id
 * and submits it as an answer
 */
func (h *Handler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.HandleError(w, errors.BadIDError, http.StatusBadRequest)
		return
	}

	ans := answer.Answer{}
	err = httputil.UnmarshalRequestBody(r, &ans)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		httputil.HandleError(w, errors.JSONParseError, http.StatusBadRequest)
		return
	}

	ans.Question = id
	ans.SubmittedOn = time.Now()

	err = answer.Validate(ans)
	if err != nil {
		msg := fmt.Sprintf("Invalid answer: %v", err)
		httputil.HandleError(w, msg, http.StatusBadRequest)
		return
	}

	sess, err := h.sessionManager.GetSession(r)
	if err != nil {
		msg := fmt.Sprintf("You must be logged in to answer a question")
		httputil.HandleError(w, msg, http.StatusUnauthorized)
		return
	}

	u, err := h.db.GetUserByUsername(sess.Username)
	if err != nil {
		msg := fmt.Sprintf("Received answer authored by a user that doesn't exist.")
		httputil.HandleError(w, msg, http.StatusBadRequest)
		return
	}

	ans.Author = u.ID

	// Ensure the question with the given id actually exists
	_, err = h.db.GetQuestion(id)
	if err != nil {
		msg := fmt.Sprintf("Received answer to a question that doesn't exist.")
		httputil.HandleError(w, msg, http.StatusBadRequest)
		return
	}

	err = h.db.InsertAnswer(ans)
	if err != nil {
		httputil.HandleError(w, errors.DBInsertError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // TODO Success JSON response body
}

/* GET /questions/{id}/answers
 *
 * Retrieves answers to the question with id
 */
func (h *Handler) GetAnswers(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.HandleError(w, errors.BadIDError, http.StatusBadRequest)
		return
	}

	ans, err := h.db.GetAnswers(id)
	if err != nil {
		httputil.HandleError(w, errors.ResourceNotFoundError, http.StatusNotFound)
		return
	}

	w.Write(JSON(ans))
	return
}
