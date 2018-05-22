package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/JonathonGore/knowledge-base/creds"
	"github.com/JonathonGore/knowledge-base/models/answer"
	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/question"
	"github.com/JonathonGore/knowledge-base/models/team"
	"github.com/JonathonGore/knowledge-base/models/user"
	"github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/session/managers"
	"github.com/JonathonGore/knowledge-base/storage"
	"github.com/JonathonGore/knowledge-base/utils"
	"github.com/gorilla/mux"
)

const (
	JSONParseError          = "Unable to parse request body as JSON"
	JSONError               = "Unable to convert into JSON"
	DBInsertError           = "Unable to insert into databse"
	DBUpdateError           = "Unable to update databse"
	DBGetError              = "Unable to retrieve from databse"
	InvalidPathParamError   = "Received bad bath paramater"
	InvalidCredentialsError = "Invalid username or password"
	LoginFailedError        = "Login failed"
	LogoutFailedError       = "Logout failed"
	EmptyCredentialsError   = "Username and password both must be non-empty"
	BadIDError              = "The requested ID does not exist in our system"
)

type Handler struct {
	db             storage.Driver
	sessionManager session.Manager
}

func New(d storage.Driver) (*Handler, error) {
	sm, err := managers.NewSMManager("knowledge_base", 3600*24*365)
	if err != nil {
		return nil, err
	}

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
		log.Printf("Unable to get questions from database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{DBGetError, http.StatusInternalServerError}))
		return
	}

	contents, err := json.Marshal(orgs)
	if err != nil {
		log.Printf("Unable to convert questions to byte array")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{JSONError, http.StatusInternalServerError}))
		return
	}

	w.Write(contents)
	return

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
		log.Printf("Unable to get organization: %v", err)
		w.WriteHeader(http.StatusNotFound)
		w.Write(JSON(ErrorResponse{DBGetError, http.StatusNotFound}))
		return
	}

	contents, err := json.Marshal(org)
	if err != nil {
		log.Printf("Unable to convert organization to byte array")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{JSONError, http.StatusInternalServerError}))
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
		log.Printf("Unable to parse body as JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{JSONParseError, http.StatusInternalServerError}))
		return
	}

	log.Printf("Recieved the following organization to create: %+v", org)

	err = organization.Validate(org)
	if err != nil {
		log.Printf("Unable to create organization: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{fmt.Sprintf("Unable to create organization: %v", err),
			http.StatusBadRequest}))
		return
	}

	_, err = h.db.GetOrganizationByName(org.Name)
	if err == nil {
		log.Printf("Attempted to create organization %v but name already exists", org.Name)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{fmt.Sprintf("Attempted to create organization: %v - but name already exists", org.Name),
			http.StatusBadRequest}))
		return
	}

	err = h.db.InsertOrganization(org)
	if err != nil {
		log.Printf("Unable to insert organization into database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{DBInsertError, http.StatusInternalServerError}))
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* GET /organizations/<org>/teams
 *
 * Receives a page of teams within an organization
 * TODO: accept query params
 */
func (h *Handler) GetTeams(w http.ResponseWriter, r *http.Request) {
	orgName, _ := mux.Vars(r)["organization"]

	_, err := h.db.GetOrganizationByName(orgName)
	if err == nil {
		log.Printf("Unable to get teams, organization %v does not exist", orgName)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{fmt.Sprintf("Unable to get teams, organization %v does not exist", orgName),
			http.StatusBadRequest}))
		return
	}

	orgs, err := h.db.GetOrganizations()
	if err != nil {
		log.Printf("Unable to get questions from database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{DBGetError, http.StatusInternalServerError}))
		return
	}

	contents, err := json.Marshal(orgs)
	if err != nil {
		log.Printf("Unable to convert questions to byte array")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{JSONError, http.StatusInternalServerError}))
		return
	}

	w.Write(contents)
	return

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
		log.Printf("Unable to parse body as JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{JSONParseError, http.StatusInternalServerError}))
		return
	}

	err = team.Validate(t)
	if err != nil {
		log.Printf("Unable to create team: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{fmt.Sprintf("Unable to create team: %v", err),
			http.StatusBadRequest}))
		return
	}

	org, err := h.db.GetOrganizationByName(orgName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{fmt.Sprintf("Organization: %v does not exist", orgName),
			http.StatusBadRequest}))
		return
	}

	_, err = h.db.GetTeamByName(orgName, t.Name)
	if err == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{fmt.Sprintf("Team with name %v within %v already exists", t.Name, orgName),
			http.StatusBadRequest}))
		return

	}

	t.Organization = org.ID // Link the team to the org

	err = h.db.InsertTeam(t)
	if err != nil {
		log.Printf("Unable to insert team into database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{DBInsertError, http.StatusInternalServerError}))
		return
	}

	w.WriteHeader(http.StatusOK)
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
		log.Printf("Unable to parse body as JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{JSONParseError, http.StatusInternalServerError}))
		return
	}

	log.Printf("Received the following user to signup: %v", user.SafePrint())

	// verify username and password meet out criteria of valid
	if err = creds.ValidateSignupCredentials(user.Username, user.Password); err != nil {
		log.Printf("Attempted to sign up user %v with invalid credentials - %v", user.Username, err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{JSONParseError, http.StatusBadRequest}))
		return
	}

	_, err = h.db.GetUserByUsername(user.Username)
	if err == nil {
		log.Printf("Attempted to sign up user %v but username already exists", user.Username)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{fmt.Sprintf("Attempted to sign up with username: %v - but username already exists", user.Username),
			http.StatusBadRequest}))
		return
	}

	// Hash our password to avoid storing plaintext in database
	user.Password, err = creds.HashPassword(user.Password)
	if err != nil {
		log.Printf("Error hashing user password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{"Internal server error", http.StatusBadRequest}))
		return
	}

	err = h.db.InsertUser(user)
	if err != nil {
		log.Printf("Unable to insert user into database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{DBInsertError, http.StatusInternalServerError}))
		return
	}

	w.WriteHeader(http.StatusOK)
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
		log.Printf("Unable to parse body as JSON: %v", err)
		http.Error(w, JSONString(ErrorResponse{JSONParseError, http.StatusInternalServerError}), http.StatusInternalServerError)
		return
	}

	log.Printf("Received the following user to login: %v", attemptedUser.Username)

	if attemptedUser.Username == "" || attemptedUser.Password == "" {
		log.Printf("Received empty credentials when performing a login")
		http.Error(w, JSONString(ErrorResponse{EmptyCredentialsError, http.StatusBadRequest}), http.StatusBadRequest)
		return
	}

	if h.sessionManager.HasSession(r) {
		// Already logged in so the request has succeeded
		w.WriteHeader(http.StatusOK)
		return
	}

	actualUser, err := h.db.GetUserByUsername(attemptedUser.Username)
	if err != nil {
		log.Printf("Unable to retrieve user %v from db: %v", attemptedUser.Username, err)
		// If the user does not exist return 401 (Unauthorized) for security reasons
		http.Error(w, JSONString(ErrorResponse{InvalidCredentialsError, http.StatusUnauthorized}), http.StatusUnauthorized)
		return
	}

	// NOTE: attemptedUser.Password is plaintext and actualUser.Password is bcrypted hash of password
	valid := creds.CheckPasswordHash(attemptedUser.Password, actualUser.Password)
	if !valid {
		log.Printf("Bad credentials attempting to authenticate user %v", attemptedUser.Username)
		http.Error(w, JSONString(ErrorResponse{InvalidCredentialsError, http.StatusUnauthorized}), http.StatusUnauthorized)
		return
	}

	// Successfully logged in make sure we have a session -- will insert a session id into the ResponseWriters cookies
	s, err := h.sessionManager.SessionStart(w, r, actualUser.Username)
	if err != nil {
		log.Printf("Unable to start session for user, login failed")
		http.Error(w, JSONString(ErrorResponse{LoginFailedError, http.StatusInternalServerError}), http.StatusInternalServerError)
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
		log.Printf("Unable to logout user: %v", err)
		http.Error(w, JSONString(ErrorResponse{LogoutFailedError, http.StatusInternalServerError}), http.StatusInternalServerError)
		return
	}

	w.Write(JSON(SuccessResponse{"Success", http.StatusOK}))
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
		log.Printf("Unable to get user from database: %v", err)
		w.WriteHeader(http.StatusNotFound)
		w.Write(JSON(ErrorResponse{DBGetError, http.StatusNotFound}))
		return
	}

	// Now that we have the user set password field to ""
	user.Password = ""

	contents, err := json.Marshal(user)
	if err != nil {
		log.Printf("Unable to convert user to byte array")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{JSONError, http.StatusInternalServerError}))
		return
	}

	w.Write(contents)
	return
}

/* POST /questions
 *
 * Receives a question to insert, validates it and puts it into the
 * database
 *
 * Expected: { author: <int>, title: <string>, content: <string> }
 */
func (h *Handler) SubmitQuestion(w http.ResponseWriter, r *http.Request) {
	q := question.Question{}
	err := utils.UnmarshalRequestBody(r, &q)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{JSONParseError, http.StatusInternalServerError}))
		return
	}

	err = question.Validate(q)
	if err != nil {
		log.Printf("Received invalid question: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{err.Error(), http.StatusBadRequest}))
		return
	}

	_, err = h.db.GetUser(q.Author)
	if err != nil {
		// The case where we receive a question authroed by an invalid user
		log.Printf("Received question authored by a user that doesn't exist.")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{"invalid author", http.StatusBadRequest}))
		return
	}

	err = h.db.InsertQuestion(q)
	if err != nil {
		log.Printf("Unable to insert question into database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{DBInsertError, http.StatusInternalServerError}))
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* GET /question/{id}
 *
 * Retrieves a question from the database with the given id
 */
func (h *Handler) GetQuestion(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Received bad question id when trying to retrieve question", idStr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{BadIDError, http.StatusBadRequest}))
		return
	}

	questions, err := h.db.GetQuestion(id)
	if err != nil {
		log.Printf("Unable to get question from database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{DBGetError, http.StatusInternalServerError}))
		return
	}

	contents, err := json.Marshal(questions)
	if err != nil {
		log.Printf("Unable to convert question to byte array")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{JSONError, http.StatusInternalServerError}))
		return
	}

	w.Write(contents)
}

/* GET /questions
 *
 * Receives a page of questions
 * TODO: accept query params
 */
func (h *Handler) GetQuestions(w http.ResponseWriter, r *http.Request) {
	questions, err := h.db.GetQuestions()
	if err != nil {
		log.Printf("Unable to get questions from database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{DBGetError, http.StatusInternalServerError}))
		return
	}

	contents, err := json.Marshal(questions)
	if err != nil {
		log.Printf("Unable to convert questions to byte array")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{JSONError, http.StatusInternalServerError}))
		return
	}

	w.Write(contents)
	return
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
		log.Printf("Received bad question id when trying to increase view count", idStr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{BadIDError, http.StatusBadRequest}))
		return
	}

	err = h.db.ViewQuestion(id)
	if err != nil {
		log.Printf("Unable to update view count for question with id: %v. Error: %v", id, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{DBUpdateError, http.StatusInternalServerError}))
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* POST /questions/{id}
 *
 * Receives an answer to the question with id
 * and submits it as an answer
 */
func (h *Handler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{"invalid question id", http.StatusBadRequest}))
		return
	}

	ans := answer.Answer{}
	err = utils.UnmarshalRequestBody(r, &ans)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{JSONParseError, http.StatusBadRequest}))
		return
	}

	err = answer.Validate(ans)
	if err != nil {
		log.Printf("Received invalid answer: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{err.Error(), http.StatusBadRequest}))
		return
	}

	// Ensure the answer is authored by a valid user
	// TODO: This should be its own function
	_, err = h.db.GetUser(ans.AuthoredBy)
	if err != nil {
		log.Printf("Received answer authored by a user that doesn't exist.")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{"invalid author", http.StatusBadRequest}))
		return
	}

	// Ensure the question with the given id actually exists
	_, err = h.db.GetQuestion(id)
	if err != nil {
		log.Printf("Received answer to a question that doesn't exist.")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(JSON(ErrorResponse{"invalid question", http.StatusBadRequest}))
		return
	}

	err = h.db.InsertAnswer(ans)
	if err != nil {
		log.Printf("Unable to insert answer into database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(JSON(ErrorResponse{DBInsertError, http.StatusInternalServerError}))
		return
	}

	w.WriteHeader(http.StatusOK)
}
