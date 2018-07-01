package users

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/JonathonGore/knowledge-base/creds"
	"github.com/JonathonGore/knowledge-base/errors"
	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/user"
	sess "github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/util/httputil"
	"github.com/gorilla/mux"
)

// storage describes the interface methods required from an storage component
type storage interface {
	GetUser(userID int) (user.User, error)
	GetUserByUsername(username string) (user.User, error)
	GetUserOrganizations(uid int) ([]organization.Organization, error)
	InsertUser(user user.User) error
}

// session describes the interface methods required from an session component
type session interface {
	GetSession(r *http.Request) (sess.Session, error)
	HasSession(r *http.Request) bool
	SessionStart(w http.ResponseWriter, r *http.Request, username string) (sess.Session, error)
	SessionDestroy(w http.ResponseWriter, r *http.Request) error
}

// Handler describes the http handler struct used for managing users data.
type Handler struct {
	db             storage
	sessionManager session
}

// New creates a new users handler with the given storage
// and session component.
func New(d storage, sm session) (*Handler, error) {
	if d == nil || sm == nil {
		return nil, fmt.Errorf("storage driver and session manager must not be nil")
	}

	return &Handler{d, sm}, nil
}

// GetUserOrgNames retrieves a list of organization names that the user with
// the provided id belongs to.
func (h *Handler) getUserOrgNames(id int) ([]string, error) {
	orgs, err := h.db.GetUserOrganizations(id)
	if err != nil {
		return nil, err
	}

	orgNames := make([]string, len(orgs))
	for i, org := range orgs {
		orgNames[i] = org.Name
	}

	return orgNames, nil
}

/* POST /users
 *
 * Signs up the given user by inserting them into the database.
 *
 * Note: Error messages here are user facing
 */
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	u := user.User{}
	err := httputil.UnmarshalRequestBody(r, &u)
	if err != nil {
		httputil.HandleError(w, errors.JSONParseError, http.StatusBadRequest)
		return
	}

	log.Printf("Received the following user to signup: %v", u.SafePrint())

	if err := user.Validate(u); err != nil {
		httputil.HandleError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate signup credentials is done here instead of user.Validate - as
	// this is the only time we can validate the plain text password.
	if err = creds.ValidateSignupCredentials(u.Username, u.Password); err != nil {
		httputil.HandleError(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = h.db.GetUserByUsername(u.Username)
	if err == nil {
		msg := fmt.Sprintf("User with username %v already exists", u.Username)
		httputil.HandleError(w, msg, http.StatusBadRequest)
		return
	}

	u.Password, err = creds.HashPassword(u.Password)
	if err != nil {
		httputil.HandleError(w, errors.InternalServerError, http.StatusInternalServerError)
		return
	}

	u.JoinedOn = time.Now()

	err = h.db.InsertUser(u)
	if err != nil {
		httputil.HandleError(w, errors.DBInsertError, http.StatusInternalServerError)
		return
	}

	httputil.Success(w)
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
		httputil.HandleError(w, errors.JSONParseError, http.StatusBadRequest)
		return
	}

	log.Printf("Received the following user to login: %v", attemptedUser.Username)

	if attemptedUser.Username == "" || attemptedUser.Password == "" {
		httputil.HandleError(w, errors.EmptyCredentialsError, http.StatusBadRequest)
		return
	}

	if h.sessionManager.HasSession(r) {
		// Already logged in so the request has succeeded
		w.WriteHeader(http.StatusOK)
		return
	}

	actualUser, err := h.db.GetUserByUsername(attemptedUser.Username)
	if err != nil {
		httputil.HandleError(w, errors.InvalidCredentialsError, http.StatusUnauthorized)
		return
	}

	// NOTE: attemptedUser.Password is plaintext and actualUser.Password is bcrypted hash of password
	valid := creds.CheckPasswordHash(attemptedUser.Password, actualUser.Password)
	if !valid {
		httputil.HandleError(w, errors.InvalidCredentialsError, http.StatusUnauthorized)
		return
	}

	// Successfully logged in make sure we have a session -- will insert a session id into the ResponseWriters cookies
	s, err := h.sessionManager.SessionStart(w, r, actualUser.Username)
	if err != nil {
		httputil.HandleError(w, errors.LoginFailedError, http.StatusInternalServerError)
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
		httputil.HandleError(w, errors.LogoutFailedError, http.StatusInternalServerError)
		return
	}

	httputil.Success(w)
}

/* GET /profile
 *
 * Retrieves the user from the db, inferring from session cookie
 */
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	sess, err := h.sessionManager.GetSession(r)
	if err != nil {
		msg := "Must be logged in to view profile"
		httputil.HandleError(w, msg, http.StatusUnauthorized)
		return
	}

	user, err := h.db.GetUserByUsername(sess.Username)
	if err != nil {
		httputil.HandleError(w, errors.DBGetError, http.StatusNotFound)
		return
	}

	user.Organizations, err = h.getUserOrgNames(user.ID)
	if err != nil {
		httputil.HandleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	// Now that we have the user set password field to ""
	user.Password = ""

	w.Write(httputil.JSON(user))
}

/* GET /users/{username}
 *
 * Retrieves the user from the database with the given username
 *
 * Note: Right now we do not require any authorization for users.
 */
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	username, ok := mux.Vars(r)["username"]
	if !ok {
		httputil.HandleError(w, errors.InternalServerError, http.StatusInternalServerError)
		return
	}

	log.Printf("Attempting to retrieve user with username: %v", username)

	user, err := h.db.GetUserByUsername(username)
	if err != nil {
		httputil.HandleError(w, errors.DBGetError, http.StatusNotFound)
		return
	}

	user.Organizations, err = h.getUserOrgNames(user.ID)
	if err != nil {
		httputil.HandleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	// Now that we have the user set password field to ""
	user.Password = ""

	w.Write(httputil.JSON(user))
}
