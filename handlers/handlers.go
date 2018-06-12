package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/JonathonGore/knowledge-base/errors"
	"github.com/JonathonGore/knowledge-base/handlers/users"
	"github.com/JonathonGore/knowledge-base/models/answer"
	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/question"
	"github.com/JonathonGore/knowledge-base/models/team"
	"github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/storage"
	"github.com/JonathonGore/knowledge-base/util/httputil"
	"github.com/gorilla/mux"
)

type Handler struct {
	users.UserRoutes

	db             storage.Driver
	sessionManager session.Manager
}

func New(d storage.Driver, sm session.Manager) (*Handler, error) {
	userHandler, err := users.New(d, sm)
	if err != nil {
		return nil, err
	}

	return &Handler{UserRoutes: userHandler, db: d, sessionManager: sm}, nil
}

/* Consumes an http request and flattens the query
 * paramaters
 */
func parseQueryParams(r *http.Request) map[string]string {
	params := make(map[string]string)

	m := r.URL.Query()
	for key, vals := range m {
		for _, val := range vals {
			params[key] = val
		}
	}

	return params
}

func (h *Handler) getUserOrgNames(id int) ([]string, error) {
	orgs, err := h.db.GetUserOrganizations(id)
	if err != nil {
		return nil, err
	}

	var orgNames = make([]string, len(orgs))
	for i, org := range orgs {
		orgNames[i] = org.Name
	}

	return orgNames, nil
}

func handleError(w http.ResponseWriter, message string, code int) {
	_, fn, line, _ := runtime.Caller(1)
	log.Printf("Error at: %v:%v - %v", fn, line, message)
	w.WriteHeader(code)
	w.Write(JSON(ErrorResponse{message, code}))
}

func (h *Handler) prepareQuestion(w http.ResponseWriter, r *http.Request) (question.Question, error) {
	q := question.Question{}
	err := httputil.UnmarshalRequestBody(r, &q)
	if err != nil {
		handleError(w, errors.JSONParseError, http.StatusInternalServerError)
		return q, err
	}

	err = question.Validate(q)
	if err != nil {
		handleError(w, err.Error(), http.StatusBadRequest)
		return q, err
	}

	sess, err := h.sessionManager.GetSession(r)
	if err != nil {
		msg := "Must be logged in to create a question"
		handleError(w, msg, http.StatusUnauthorized)
		return q, err
	}

	u, err := h.db.GetUserByUsername(sess.Username)
	if err != nil {
		msg := "Received question authored by a user that doesn't exist."
		handleError(w, msg, http.StatusBadRequest)
		return q, err
	}

	q.Author = u.ID

	return q, nil
}

/* GET /organizations/{org}/questions
 *
 * Receives a page of questions for the provided team
 * TODO: accept query params
 */
func (h *Handler) GetOrgQuestions(w http.ResponseWriter, r *http.Request) {
	org := mux.Vars(r)["org"]

	questions, err := h.db.GetOrgQuestions(org)
	if err != nil {
		handleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	w.Write(JSON(questions))
	return
}

/* GET /organizations/{org}/teams/{team}/questions
 *
 * Receives a page of questions for the provided team
 * TODO: accept query params
 */
func (h *Handler) GetTeamQuestions(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	team := params["team"]
	org := params["org"]

	questions, err := h.db.GetTeamQuestions(team, org)
	if err != nil {
		handleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	w.Write(JSON(questions))
	return
}

/* POST /organizations/{org}/teams/{team}/questions
 *
 * Receives a question to insert for the given team, validates it
 * and puts it into the database.
 *
 * Expected: { title: <string>, content: <string> }
 * Author will be inferred from the session attached to the request
 */
func (h *Handler) SubmitTeamQuestion(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	org := params["org"]
	team := params["team"]

	q, err := h.prepareQuestion(w, r)
	if err != nil {
		return // We write to w in prepareQuestion
	}

	_, err = h.db.GetOrganizationByName(org)
	if err != nil {
		handleError(w, fmt.Sprintf("Organization %v does not exist", org), http.StatusBadRequest)
		return
	}

	t, err := h.db.GetTeamByName(org, team)
	if err != nil {
		handleError(w, fmt.Sprintf("Team %v does not exist", org), http.StatusBadRequest)
		return
	}

	q.Team = team
	q.Organization = org

	id, err := h.db.InsertTeamQuestion(q, t.ID)
	if err != nil {
		handleError(w, errors.DBInsertError, http.StatusInternalServerError)
		return
	}

	w.Write(JSON(IDResponse{id}))
}

/* GET /organizations
 *
 * Receives a page of organizations
 * TODO: accept query params
 */
func (h *Handler) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.db.GetOrganizations()
	if err != nil {
		handleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	contents, err := json.Marshal(orgs)
	if err != nil {
		handleError(w, errors.JSONError, http.StatusInternalServerError)
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
		handleError(w, errors.DBGetError, http.StatusNotFound)
		return
	}

	contents, err := json.Marshal(org)
	if err != nil {
		handleError(w, errors.JSONError, http.StatusInternalServerError)
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
		handleError(w, errors.JSONParseError, http.StatusBadRequest)
		return
	}

	err = organization.Validate(org)
	if err != nil {
		handleError(w, errors.CreateResourceError, http.StatusBadRequest)
		return
	}

	_, err = h.db.GetOrganizationByName(org.Name)
	if err == nil {
		msg := fmt.Sprintf("Attempted to create organization %v but name already exists", org.Name)
		handleError(w, msg, http.StatusBadRequest)
		return
	}

	org.CreatedOn = time.Now()

	id, err := h.db.InsertOrganization(org)
	if err != nil {
		log.Printf("Unable to insert organization into database: %v", err)
		handleError(w, errors.DBInsertError, http.StatusBadRequest)
		return
	}

	sess, err := h.sessionManager.GetSession(r)
	if err != nil {
		msg := "Must be logged in to create an organization"
		handleError(w, msg, http.StatusUnauthorized)
		return
	}

	err = h.db.InsertOrgMember(sess.Username, org.Name, true)
	if err != nil {
		log.Printf("unable to insert user as member: %v", err)
		handleError(w, errors.DBInsertError, http.StatusInternalServerError)
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
		handleError(w, errors.DBInsertError, http.StatusInternalServerError)
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
		handleError(w, errors.ResourceNotFoundError, http.StatusBadRequest)
		return
	}

	team, err := h.db.GetTeamByName(orgName, teamName)
	if err != nil {
		handleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	contents, err := json.Marshal(team)
	if err != nil {
		handleError(w, errors.JSONError, http.StatusInternalServerError)
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
		handleError(w, errors.ResourceNotFoundError, http.StatusBadRequest)
		return
	}

	teams, err := h.db.GetTeams(orgName)
	if err != nil {
		handleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	contents, err := json.Marshal(teams)
	if err != nil {
		handleError(w, errors.JSONError, http.StatusInternalServerError)
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
		handleError(w, errors.JSONParseError, http.StatusInternalServerError)
		return
	}

	err = team.Validate(t)
	if err != nil {
		handleError(w, errors.CreateResourceError, http.StatusBadRequest)
		return
	}

	org, err := h.db.GetOrganizationByName(orgName)
	if err != nil {
		handleError(w, errors.ResourceNotFoundError, http.StatusBadRequest)
		return
	}

	_, err = h.db.GetTeamByName(orgName, t.Name)
	if err == nil {
		msg := fmt.Sprintf("Team with name %v within %v already exists", t.Name, orgName)
		handleError(w, msg, http.StatusBadRequest)
		return
	}

	t.Organization = org.ID // Link the team to the org

	err = h.db.InsertTeam(t)
	if err != nil {
		handleError(w, errors.DBInsertError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // TODO: Simple response body instead of just code
}

/* POST /questions
 *
 * Receives a question to insert, validates it and puts it into the
 * database
 *
 * Expected: { title: <string>, content: <string> }
 * Author will be inferred from the session attached to the request
 */
func (h *Handler) SubmitQuestion(w http.ResponseWriter, r *http.Request) {
	q := question.Question{}
	err := httputil.UnmarshalRequestBody(r, &q)
	if err != nil {
		handleError(w, errors.JSONParseError, http.StatusInternalServerError)
		return
	}

	err = question.Validate(q)
	if err != nil {
		handleError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sess, err := h.sessionManager.GetSession(r)
	if err != nil {
		msg := "Must be logged in to create a question"
		handleError(w, msg, http.StatusUnauthorized)
		return
	}

	u, err := h.db.GetUserByUsername(sess.Username)
	if err != nil {
		msg := "Received question authored by a user that doesn't exist."
		handleError(w, msg, http.StatusBadRequest)
		return
	}

	q.Author = u.ID

	id, err := h.db.InsertQuestion(q)
	if err != nil {
		handleError(w, errors.DBInsertError, http.StatusInternalServerError)
		return
	}

	w.Write(JSON(IDResponse{id}))
}

/* GET /question/{id}
 *
 * Retrieves a question from the database with the given id
 */
func (h *Handler) GetQuestion(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		handleError(w, errors.BadIDError, http.StatusBadRequest)
		return
	}

	question, err := h.db.GetQuestion(id)
	if err != nil {
		handleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	w.Write(JSON(question))
}

/* GET /questions
 *
 * Receives a page of questions
 * TODO: accept query params
 */
func (h *Handler) GetQuestions(w http.ResponseWriter, r *http.Request) {
	var questions []question.Question
	var err error

	qparams := parseQueryParams(r)
	if val, ok := qparams["user"]; ok {
		id, cerr := strconv.Atoi(val)
		if cerr != nil {
			handleError(w, errors.InvalidQueryParamError, http.StatusInternalServerError)
			return
		}

		questions, err = h.db.GetUserQuestions(id)
	} else {
		questions, err = h.db.GetQuestions()
	}

	if err != nil {
		handleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	w.Write(JSON(questions))
}

/* POST /questions/{id}/view
 *
 * Upon receiving this request it will add a view to the requested
 * question in the database
 */
func (h *Handler) ViewQuestion(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		handleError(w, errors.BadIDError, http.StatusBadRequest)
		return
	}

	err = h.db.ViewQuestion(id)
	if err != nil {
		log.Printf("Unable to update view count for question with id: %v. Error: %v", id, err)
		handleError(w, errors.DBUpdateError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // TODO: include JSON body
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
		handleError(w, errors.BadIDError, http.StatusBadRequest)
		return
	}

	ans := answer.Answer{}
	err = httputil.UnmarshalRequestBody(r, &ans)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handleError(w, errors.JSONParseError, http.StatusBadRequest)
		return
	}

	ans.Question = id
	ans.SubmittedOn = time.Now()

	err = answer.Validate(ans)
	if err != nil {
		msg := fmt.Sprintf("Invalid answer: %v", err)
		handleError(w, msg, http.StatusBadRequest)
		return
	}

	sess, err := h.sessionManager.GetSession(r)
	if err != nil {
		msg := fmt.Sprintf("You must be logged in to answer a question")
		handleError(w, msg, http.StatusUnauthorized)
		return
	}

	u, err := h.db.GetUserByUsername(sess.Username)
	if err != nil {
		msg := fmt.Sprintf("Received answer authored by a user that doesn't exist.")
		handleError(w, msg, http.StatusBadRequest)
		return
	}

	ans.Author = u.ID

	// Ensure the question with the given id actually exists
	_, err = h.db.GetQuestion(id)
	if err != nil {
		msg := fmt.Sprintf("Received answer to a question that doesn't exist.")
		handleError(w, msg, http.StatusBadRequest)
		return
	}

	err = h.db.InsertAnswer(ans)
	if err != nil {
		handleError(w, errors.DBInsertError, http.StatusInternalServerError)
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
		handleError(w, errors.BadIDError, http.StatusBadRequest)
		return
	}

	ans, err := h.db.GetAnswers(id)
	if err != nil {
		handleError(w, errors.ResourceNotFoundError, http.StatusNotFound)
		return
	}

	w.Write(JSON(ans))
	return
}
