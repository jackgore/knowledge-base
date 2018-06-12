package users

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/JonathonGore/knowledge-base/creds"
	"github.com/JonathonGore/knowledge-base/errors"
	"github.com/JonathonGore/knowledge-base/models/user"
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
	w.Write(httputil.JSON(httputil.ErrorResponse{message, code}))
}

/* POST /users
 *
 * Signs up the given user by inserting them into the database.
 *
 * Note: Error messages here are user facing
 */
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	user := user.User{}
	err := httputil.UnmarshalRequestBody(r, &user)
	if err != nil {
		handleError(w, errors.JSONParseError, http.StatusInternalServerError)
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
		handleError(w, errors.InternalServerError, http.StatusInternalServerError)
		return
	}

	user.JoinedOn = time.Now()

	err = h.db.InsertUser(user)
	if err != nil {
		handleError(w, errors.DBInsertError, http.StatusInternalServerError)
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
	err := httputil.UnmarshalRequestBody(r, &attemptedUser)
	if err != nil {
		handleError(w, errors.JSONParseError, http.StatusInternalServerError)
		return
	}

	log.Printf("Received the following user to login: %v", attemptedUser.Username)

	if attemptedUser.Username == "" || attemptedUser.Password == "" {
		handleError(w, errors.EmptyCredentialsError, http.StatusBadRequest)
		return
	}

	if h.sessionManager.HasSession(r) {
		// Already logged in so the request has succeeded
		w.WriteHeader(http.StatusOK)
		return
	}

	actualUser, err := h.db.GetUserByUsername(attemptedUser.Username)
	if err != nil {
		handleError(w, errors.InvalidCredentialsError, http.StatusUnauthorized)
		return
	}

	// NOTE: attemptedUser.Password is plaintext and actualUser.Password is bcrypted hash of password
	valid := creds.CheckPasswordHash(attemptedUser.Password, actualUser.Password)
	if !valid {
		handleError(w, errors.InvalidCredentialsError, http.StatusUnauthorized)
		return
	}

	// Successfully logged in make sure we have a session -- will insert a session id into the ResponseWriters cookies
	s, err := h.sessionManager.SessionStart(w, r, actualUser.Username)
	if err != nil {
		handleError(w, errors.LoginFailedError, http.StatusInternalServerError)
		return
	}

	w.Write(httputil.JSON(httputil.LoginResponse{s.SID}))
}

/* POST /logout
 *
 * Logout the requesting user by deleting the session
 */
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	err := h.sessionManager.SessionDestroy(w, r)
	if err != nil {
		handleError(w, errors.LogoutFailedError, http.StatusInternalServerError)
		return
	}

	w.Write(httputil.JSON(httputil.SuccessResponse{"Success", http.StatusOK}))
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
		handleError(w, errors.DBGetError, http.StatusNotFound)
		return
	}

	user.Organizations, err = h.getUserOrgNames(user.ID)
	if err != nil {
		handleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	// Now that we have the user set password field to ""
	user.Password = ""

	w.Write(httputil.JSON(user))
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
		handleError(w, errors.DBGetError, http.StatusNotFound)
		return
	}

	user.Organizations, err = h.getUserOrgNames(user.ID)
	if err != nil {
		handleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	// Now that we have the user set password field to ""
	user.Password = ""

	w.Write(httputil.JSON(user))
}
