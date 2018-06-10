package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/JonathonGore/knowledge-base/creds"
	"github.com/JonathonGore/knowledge-base/models/answer"
	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/question"
	"github.com/JonathonGore/knowledge-base/models/team"
	"github.com/JonathonGore/knowledge-base/models/user"
	"github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/storage"
	"github.com/JonathonGore/knowledge-base/utils"
	"github.com/gorilla/mux"
)

const (
	JSONParseError          = "Unable to parse request body as JSON"
	JSONError               = "Unable to convert into JSON"
	CreateResourceError     = "Unable to create resource"
	ResourceNotFoundError   = "Unable to find resource"
	DBInsertError           = "Unable to insert into databse"
	DBUpdateError           = "Unable to update databse"
	DBGetError              = "Unable to retrieve data from database"
	InvalidPathParamError   = "Received bad bath paramater"
	InvalidCredentialsError = "Invalid username or password"
	InvalidQueryParamError  = "Invalid query paramater"
	InternalServerError     = "Internal server error"
	LoginFailedError        = "Login failed"
	LogoutFailedError       = "Logout failed"
	EmptyCredentialsError   = "Username and password both must be non-empty"
	BadIDError              = "The requested ID does not exist in our system"
)

type Handler struct {
	db             storage.Driver
	sessionManager session.Manager
}

func New(d storage.Driver, sm session.Manager) (*Handler, error) {
	return &Handler{d, sm}, nil
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
	err := utils.UnmarshalRequestBody(r, &q)
	if err != nil {
		handleError(w, JSONParseError, http.StatusInternalServerError)
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
		handleError(w, DBGetError, http.StatusInternalServerError)
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
		handleError(w, DBGetError, http.StatusInternalServerError)
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
		handleError(w, DBInsertError, http.StatusInternalServerError)
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
		handleError(w, DBGetError, http.StatusInternalServerError)
		return
	}

	contents, err := json.Marshal(orgs)
	if err != nil {
		handleError(w, JSONError, http.StatusInternalServerError)
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
		handleError(w, DBGetError, http.StatusNotFound)
		return
	}

	contents, err := json.Marshal(org)
	if err != nil {
		handleError(w, JSONError, http.StatusInternalServerError)
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
	err := utils.UnmarshalRequestBody(r, &org)
	if err != nil {
		handleError(w, JSONParseError, http.StatusBadRequest)
		return
	}

	err = organization.Validate(org)
	if err != nil {
		handleError(w, CreateResourceError, http.StatusBadRequest)
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
		handleError(w, DBInsertError, http.StatusBadRequest)
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
		handleError(w, DBInsertError, http.StatusInternalServerError)
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
		handleError(w, DBInsertError, http.StatusInternalServerError)
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
		handleError(w, ResourceNotFoundError, http.StatusBadRequest)
		return
	}

	team, err := h.db.GetTeamByName(orgName, teamName)
	if err != nil {
		handleError(w, DBGetError, http.StatusInternalServerError)
		return
	}

	contents, err := json.Marshal(team)
	if err != nil {
		handleError(w, JSONError, http.StatusInternalServerError)
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
		handleError(w, ResourceNotFoundError, http.StatusBadRequest)
		return
	}

	teams, err := h.db.GetTeams(orgName)
	if err != nil {
		handleError(w, DBGetError, http.StatusInternalServerError)
		return
	}

	contents, err := json.Marshal(teams)
	if err != nil {
		handleError(w, JSONError, http.StatusInternalServerError)
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
	err := utils.UnmarshalRequestBody(r, &t)
	if err != nil {
		handleError(w, JSONParseError, http.StatusInternalServerError)
		return
	}

	err = team.Validate(t)
	if err != nil {
		handleError(w, CreateResourceError, http.StatusBadRequest)
		return
	}

	org, err := h.db.GetOrganizationByName(orgName)
	if err != nil {
		handleError(w, ResourceNotFoundError, http.StatusBadRequest)
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
		handleError(w, DBInsertError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // TODO: Simple response body instead of just code
}

/* POST /users
 *
 * Signs up the given user by inserting them into the database.
 *
 * Note: Error messages here are user facing
 */
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	user := user.User{}
	err := utils.UnmarshalRequestBody(r, &user)
	if err != nil {
		handleError(w, JSONParseError, http.StatusInternalServerError)
		return
	}

	log.Printf("Received the following user to signup: %v", user.SafePrint())

	if err = creds.ValidateSignupCredentials(user.Username, user.Password); err != nil {
		handleError(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = h.db.GetUserByUsername(user.Username)
	if err == nil {
		msg := fmt.Sprintf("User with username %v already exists", user.Username)
		handleError(w, msg, http.StatusBadRequest)
		return
	}

	// Hash our password to avoid storing plaintext in database
	user.Password, err = creds.HashPassword(user.Password)
	if err != nil {
		log.Printf("Error hashing user password: %v", err)
		handleError(w, InternalServerError, http.StatusInternalServerError)
		return
	}

	user.JoinedOn = time.Now()

	err = h.db.InsertUser(user)
	if err != nil {
		handleError(w, DBInsertError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // TODO this should be a JSON response
}

/* POST /login
 *
 * Logs the given the user in and creates a new session if needed.
 *
 * Expected body:
 *   { "username": "%v", "password": "%v" }
 *
 * Note: Error messages here are user facing
 */
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	attemptedUser := user.LoginAttempt{}
	err := utils.UnmarshalRequestBody(r, &attemptedUser)
	if err != nil {
		handleError(w, JSONParseError, http.StatusInternalServerError)
		return
	}

	log.Printf("Received the following user to login: %v", attemptedUser.Username)

	if attemptedUser.Username == "" || attemptedUser.Password == "" {
		handleError(w, EmptyCredentialsError, http.StatusBadRequest)
		return
	}

	if h.sessionManager.HasSession(r) {
		// Already logged in so the request has succeeded
		w.WriteHeader(http.StatusOK)
		return
	}

	actualUser, err := h.db.GetUserByUsername(attemptedUser.Username)
	if err != nil {
		handleError(w, InvalidCredentialsError, http.StatusUnauthorized)
		return
	}

	// NOTE: attemptedUser.Password is plaintext and actualUser.Password is bcrypted hash of password
	valid := creds.CheckPasswordHash(attemptedUser.Password, actualUser.Password)
	if !valid {
		handleError(w, InvalidCredentialsError, http.StatusUnauthorized)
		return
	}

	// Successfully logged in make sure we have a session -- will insert a session id into the ResponseWriters cookies
	s, err := h.sessionManager.SessionStart(w, r, actualUser.Username)
	if err != nil {
		handleError(w, LoginFailedError, http.StatusInternalServerError)
		return
	}

	w.Write(JSON(LoginResponse{s.SID}))
}

/* POST /logout
 *
 * Logout the requesting user by deleting the session
 */
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	err := h.sessionManager.SessionDestroy(w, r)
	if err != nil {
		handleError(w, LogoutFailedError, http.StatusInternalServerError)
		return
	}

	w.Write(JSON(SuccessResponse{"Success", http.StatusOK}))
}

/* GET /profile
 *
 * Retrieves the user from the db, inferring from session cookie
 */
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	sess, err := h.sessionManager.GetSession(r)
	if err != nil {
		msg := "Must be logged in to view profile"
		handleError(w, msg, http.StatusUnauthorized)
		return
	}

	user, err := h.db.GetUserByUsername(sess.Username)
	if err != nil {
		handleError(w, DBGetError, http.StatusNotFound)
		return
	}

	user.Organizations, err = h.getUserOrgNames(user.ID)
	if err != nil {
		handleError(w, DBGetError, http.StatusInternalServerError)
		return
	}

	// Now that we have the user set password field to ""
	user.Password = ""

	w.Write(JSON(user))
}

/* GET /users/{username}
 *
 * Retrieves the user from the database with the given username
 */
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]

	log.Printf("Attempting to retrieve user with username: %v", username)

	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		handleError(w, DBGetError, http.StatusNotFound)
		return
	}

	user.Organizations, err = h.getUserOrgNames(user.ID)
	if err != nil {
		handleError(w, DBGetError, http.StatusInternalServerError)
		return
	}

	// Now that we have the user set password field to ""
	user.Password = ""

	w.Write(JSON(user))
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
	err := utils.UnmarshalRequestBody(r, &q)
	if err != nil {
		handleError(w, JSONParseError, http.StatusInternalServerError)
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
		handleError(w, DBInsertError, http.StatusInternalServerError)
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
		handleError(w, BadIDError, http.StatusBadRequest)
		return
	}

	question, err := h.db.GetQuestion(id)
	if err != nil {
		handleError(w, DBGetError, http.StatusInternalServerError)
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
			handleError(w, InvalidQueryParamError, http.StatusInternalServerError)
			return
		}

		questions, err = h.db.GetUserQuestions(id)
	} else {
		questions, err = h.db.GetQuestions()
	}

	if err != nil {
		handleError(w, DBGetError, http.StatusInternalServerError)
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
		handleError(w, BadIDError, http.StatusBadRequest)
		return
	}

	err = h.db.ViewQuestion(id)
	if err != nil {
		log.Printf("Unable to update view count for question with id: %v. Error: %v", id, err)
		handleError(w, DBUpdateError, http.StatusInternalServerError)
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
		handleError(w, BadIDError, http.StatusBadRequest)
		return
	}

	ans := answer.Answer{}
	err = utils.UnmarshalRequestBody(r, &ans)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handleError(w, JSONParseError, http.StatusBadRequest)
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
		handleError(w, DBInsertError, http.StatusInternalServerError)
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
		handleError(w, BadIDError, http.StatusBadRequest)
		return
	}

	ans, err := h.db.GetAnswers(id)
	if err != nil {
		handleError(w, ResourceNotFoundError, http.StatusNotFound)
		return
	}

	w.Write(JSON(ans))
	return
}
