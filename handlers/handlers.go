package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/JonathonGore/knowledge-base/creds"
	"github.com/JonathonGore/knowledge-base/models/answer"
	"github.com/JonathonGore/knowledge-base/models/question"
	"github.com/JonathonGore/knowledge-base/models/user"
	"github.com/JonathonGore/knowledge-base/storage"
	"github.com/JonathonGore/knowledge-base/utils"
	"github.com/gorilla/mux"
)

const (
	JSONParseError          = "Unable to parse request body as JSON"
	JSONError               = "Unable to convert into JSON"
	DBInsertError           = "Unable to insert into databse"
	DBGetError              = "Unable to retrieve from databse"
	InvalidPathParamError   = "Received bad bath paramater"
	InvalidCredentialsError = "Invalid username or password"
	EmptyCredentialsError   = "Username and password both must be non-empty"
)

type Handler struct {
	db storage.Driver
}

func New(d storage.Driver) (*Handler, error) {
	return &Handler{d}, nil
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
		w.Write((&ErrorResponse{JSONParseError, http.StatusInternalServerError}).toJSON())
		return
	}

	log.Printf("Received the following user to signup: %v", user.SafePrint())

	// verify username and password meet out criteria of valid
	if err = creds.ValidateSignupCredentials(user.Username, user.Password); err != nil {
		log.Printf("Attempted to sign up user %v with invalid credentials - %v", user.Username, err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write((&ErrorResponse{JSONParseError, http.StatusBadRequest}).toJSON())
		return
	}

	_, err = h.db.GetUserByUsername(user.Username)
	if err == nil {
		log.Printf("Attempted to sign up user %v but username already exists", user.Username)
		w.WriteHeader(http.StatusBadRequest)
		w.Write((&ErrorResponse{fmt.Sprintf("Attempted to sign up with username: %v - but username already exists", user.Username),
			http.StatusBadRequest}).toJSON())
		return
	}

	// Hash our password to avoid storing plaintext in database
	user.Password, err = creds.HashPassword(user.Password)
	if err != nil {
		log.Printf("Error hashing user password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write((&ErrorResponse{"Internal server error", http.StatusBadRequest}).toJSON())
		return
	}

	err = h.db.InsertUser(user)
	if err != nil {
		log.Printf("Unable to insert user into database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write((&ErrorResponse{DBInsertError, http.StatusInternalServerError}).toJSON())
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
		http.Error(w, (&ErrorResponse{JSONParseError, http.StatusInternalServerError}).toJSONString(), http.StatusInternalServerError)
		return
	}

	log.Printf("Received the following user to login: %v", attemptedUser.Username)

	if attemptedUser.Username == "" || attemptedUser.Password == "" {
		log.Printf("Received empty credentials when performing a login")
		http.Error(w, (&ErrorResponse{EmptyCredentialsError, http.StatusBadRequest}).toJSONString(), http.StatusBadRequest)
		return
	}

	/*
		// First check to see if the user is already logged in
		if handler.SessionManager.HasSession(r) {
			// Already logged in so the request has succeeded
			w.WriteHeader(http.StatusOK)
			return
		}
	*/

	actualUser, err := h.db.GetUserByUsername(attemptedUser.Username)
	if err != nil {
		log.Printf("Unable to retrieve user %v from db: %v", attemptedUser.Username, err)
		// If the user does not exist return 401 (Unauthorized) for security reasons
		http.Error(w, (&ErrorResponse{InvalidCredentialsError, http.StatusUnauthorized}).toJSONString(), http.StatusUnauthorized)
		return
	}

	// NOTE: attemptedUser.Password is plaintext and actualUser.Password is bcrypted hash of password
	valid := creds.CheckPasswordHash(attemptedUser.Password, actualUser.Password)
	if !valid {
		log.Printf("Bad credentials attempting to authenticate user %v", attemptedUser.Username)
		http.Error(w, (&ErrorResponse{InvalidCredentialsError, http.StatusUnauthorized}).toJSONString(), http.StatusUnauthorized)
		return
	}

	/*
		// Successfully logged in make sure we have a session -- will insert a session id into the ResponseWriters cookies
		session := handler.SessionManager.SessionStart(w, r)
		// Links the session id to our username
		session.Set("username", attemptedUser.Username)
	*/
	w.WriteHeader(http.StatusOK)
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
		w.Write((&ErrorResponse{DBGetError, http.StatusNotFound}).toJSON())
		return
	}

	// Now that we have the user set password field to ""
	user.Password = ""

	contents, err := json.Marshal(user)
	if err != nil {
		log.Printf("Unable to convert user to byte array")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write((&ErrorResponse{JSONError, http.StatusInternalServerError}).toJSON())
		return
	}

	w.Write(contents)
	return
}

/* POST /questions
 *
 * Receives a question to insert, validates it and puts it into the
 * database
 */
func (h *Handler) SubmitQuestion(w http.ResponseWriter, r *http.Request) {
	q := question.Question{}
	err := utils.UnmarshalRequestBody(r, &q)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write((&ErrorResponse{JSONParseError, http.StatusInternalServerError}).toJSON())
		return
	}

	err = question.Validate(q)
	if err != nil {
		log.Printf("Received invalid question: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write((&ErrorResponse{err.Error(), http.StatusBadRequest}).toJSON())
		return
	}

	_, err = h.db.GetUser(q.AuthoredBy)
	if err != nil {
		// The case where we receive a question authroed by an invalid user
		log.Printf("Received question authored by a user that doesn't exist.")
		w.WriteHeader(http.StatusBadRequest)
		w.Write((&ErrorResponse{"invalid author", http.StatusBadRequest}).toJSON())
		return
	}

	err = h.db.InsertQuestion(q)
	if err != nil {
		log.Printf("Unable to insert question into database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write((&ErrorResponse{DBInsertError, http.StatusInternalServerError}).toJSON())
		return
	}

	w.WriteHeader(http.StatusOK)
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
		w.Write((&ErrorResponse{DBGetError, http.StatusInternalServerError}).toJSON())
		return
	}

	contents, err := json.Marshal(questions)
	if err != nil {
		log.Printf("Unable to convert questions to byte array")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write((&ErrorResponse{JSONError, http.StatusInternalServerError}).toJSON())
		return
	}

	w.Write(contents)
	return
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
		w.Write((&ErrorResponse{"invalid question id", http.StatusBadRequest}).toJSON())
		return
	}

	ans := answer.Answer{}
	err = utils.UnmarshalRequestBody(r, &ans)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write((&ErrorResponse{JSONParseError, http.StatusBadRequest}).toJSON())
		return
	}

	err = answer.Validate(ans)
	if err != nil {
		log.Printf("Received invalid answer: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write((&ErrorResponse{err.Error(), http.StatusBadRequest}).toJSON())
		return
	}

	// Ensure the answer is authored by a valid user
	// TODO: This should be its own function
	_, err = h.db.GetUser(ans.AuthoredBy)
	if err != nil {
		log.Printf("Received answer authored by a user that doesn't exist.")
		w.WriteHeader(http.StatusBadRequest)
		w.Write((&ErrorResponse{"invalid author", http.StatusBadRequest}).toJSON())
		return
	}

	// Ensure the question with the given id actually exists
	_, err = h.db.GetQuestion(id)
	if err != nil {
		log.Printf("Received answer to a question that doesn't exist.")
		w.WriteHeader(http.StatusBadRequest)
		w.Write((&ErrorResponse{"invalid question", http.StatusBadRequest}).toJSON())
		return
	}

	err = h.db.InsertAnswer(ans)
	if err != nil {
		log.Printf("Unable to insert answer into database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write((&ErrorResponse{DBInsertError, http.StatusInternalServerError}).toJSON())
		return
	}

	w.WriteHeader(http.StatusOK)
}
