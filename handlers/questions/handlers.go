package questions

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/JonathonGore/knowledge-base/errors"
	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/question"
	"github.com/JonathonGore/knowledge-base/models/team"
	"github.com/JonathonGore/knowledge-base/models/user"
	"github.com/JonathonGore/knowledge-base/query"
	"github.com/JonathonGore/knowledge-base/search"
	sess "github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/util/httputil"
	"github.com/gorilla/mux"
)

type Handler struct {
	db             storage
	sessionManager session
	search         search.Search
}

type storage interface {
	GetOrgQuestions(org string) ([]question.Question, error)
	GetOrganizationByName(name string) (organization.Organization, error)
	GetQuestion(id int) (question.Question, error)
	GetQuestions() ([]question.Question, error)
	GetTeamQuestions(team, org string) ([]question.Question, error)
	GetTeamByName(org, team string) (team.Team, error)
	GetUserByUsername(username string) (user.User, error)
	GetUsernameOrganizations(username string) ([]organization.Organization, error)
	GetUserQuestions(id int) ([]question.Question, error)
	InsertQuestion(question question.Question) (int, error)
	InsertTeamQuestion(question question.Question, tid int) (int, error)
	ViewQuestion(id int) error
}

type session interface {
	GetSession(r *http.Request) (sess.Session, error)
	HasSession(r *http.Request) bool
	SessionStart(w http.ResponseWriter, r *http.Request, username string) (sess.Session, error)
	SessionDestroy(w http.ResponseWriter, r *http.Request) error
}

// New creates a new questions handler with the given storage
// driver and session manager.
func New(d storage, sm session, s search.Search) (*Handler, error) {
	if d == nil || sm == nil {
		return nil, fmt.Errorf("storage drive and session manager must not be nil")
	}

	return &Handler{d, sm, s}, nil
}

// PrepareQuestion is a helper function to validate the the question contained
// in the given request is valid and is created by a valid user.
func (h *Handler) prepareQuestion(w http.ResponseWriter, r *http.Request) (question.Question, error) {
	q := question.Question{}
	err := httputil.UnmarshalRequestBody(r, &q)
	if err != nil {
		httputil.HandleError(w, errors.JSONParseError, http.StatusInternalServerError)
		return q, err
	}

	err = question.Validate(q)
	if err != nil {
		httputil.HandleError(w, err.Error(), http.StatusBadRequest)
		return q, err
	}

	sess, err := h.sessionManager.GetSession(r)
	if err != nil {
		msg := "Must be logged in to create a question"
		httputil.HandleError(w, msg, http.StatusUnauthorized)
		return q, err
	}

	u, err := h.db.GetUserByUsername(sess.Username)
	if err != nil {
		msg := "Received question authored by a user that doesn't exist."
		httputil.HandleError(w, msg, http.StatusBadRequest)
		return q, err
	}

	q.Author = u.ID
	q.SubmittedOn = time.Now()

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
		httputil.HandleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	w.Write(httputil.JSON(questions))
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
		httputil.HandleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	w.Write(httputil.JSON(questions))
}

// InsertQuestion inserts the given question for the given team and org. Returns the id
// if no error is produced. Otherwise returns an error and the http status code to return.
func (h *Handler) insertQuestion(q question.Question, team, org string) (int, error) {
	_, err := h.db.GetOrganizationByName(org)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("Organization %v does not exist", org)
	}

	t, err := h.db.GetTeamByName(org, team)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("Default team for org %v does not exist", org)
	}

	id, err := h.db.InsertTeamQuestion(q, t.ID)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("%v", errors.DBInsertError)
	}

	return id, nil
}

/* POST /organizations/{org}/questions
 *
 * Receives a question to insert for the given org, validates it
 * and puts it into the database.
 *
 * Expected: { title: <string>, content: <string> }
 * Author will be inferred from the session attached to the request
 */
func (h *Handler) SubmitOrgQuestion(w http.ResponseWriter, r *http.Request) {
	org := mux.Vars(r)["org"]

	q, err := h.prepareQuestion(w, r)
	if err != nil {
		return // We write to w in prepareQuestion
	}

	q.Team = "default"
	q.Organization = org
	q.SubmittedOn = time.Now()

	id, err := h.insertQuestion(q, "default", org)
	if err != nil {
		httputil.HandleError(w, fmt.Sprintf("%v", err), id)
		return
	}

	q.ID = id // Attach id to request

	if err := h.search.IndexQuestion(q); err != nil {
		log.Printf("Unable to index question in elasticsearch: %v", err)
	}

	w.Write(httputil.JSON(httputil.IDResponse{id}))
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

	q.Team = team
	q.Organization = org
	q.SubmittedOn = time.Now()

	id, err := h.insertQuestion(q, team, org)
	if err != nil {
		httputil.HandleError(w, fmt.Sprintf("%v", err), id)
		return
	}

	q.ID = id

	if err := h.search.IndexQuestion(q); err != nil {
		log.Printf("Unable to index question in elasticsearch: %v", err)
	}

	w.Write(httputil.JSON(httputil.IDResponse{id}))
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
		httputil.HandleError(w, errors.JSONParseError, http.StatusBadRequest)
		return
	}

	err = question.Validate(q)
	if err != nil {
		httputil.HandleError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sess, err := h.sessionManager.GetSession(r)
	if err != nil {
		msg := "Must be logged in to create a question"
		httputil.HandleError(w, msg, http.StatusUnauthorized)
		return
	}

	u, err := h.db.GetUserByUsername(sess.Username)
	if err != nil {
		msg := "Received question authored by a user that doesn't exist."
		httputil.HandleError(w, msg, http.StatusBadRequest)
		return
	}

	q.Author = u.ID
	q.SubmittedOn = time.Now()

	id, err := h.db.InsertQuestion(q)
	if err != nil {
		httputil.HandleError(w, errors.DBInsertError, http.StatusInternalServerError)
		return
	}

	q.ID = id

	if err := h.search.IndexQuestion(q); err != nil {
		// Dont cause the operation to fail if this break
		log.Printf("Unable to index question in elasticsearch: %v", err)
	}

	w.Write(httputil.JSON(httputil.IDResponse{id}))
}

/* GET /question/{id}
 *
 * Retrieves a question from the database with the given id
 */
func (h *Handler) GetQuestion(w http.ResponseWriter, r *http.Request) {
	// TODO: Ensure user is allowed to view the question

	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.HandleError(w, errors.BadIDError, http.StatusBadRequest)
		return
	}

	question, err := h.db.GetQuestion(id)
	if err != nil {
		httputil.HandleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	w.Write(httputil.JSON(question))
}

/* GET /search
 *
 * Search through questions to retrieve questions relavent to the provided query
 * Params:
 *		query: the query string
 *      organization: the org to look for questions in
 */
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	var err error
	orgs := make([]string, 0)

	qparams := query.ParseParams(r)
	query, ok := qparams["query"]
	if !ok {
		httputil.HandleError(w, "search requires you to pass a query string", http.StatusBadRequest)
		return
	}

	org, ok := qparams["org"]
	if !ok {
		// TODO: Eventually need to allow for public orgs to be searched
		s, err := h.sessionManager.GetSession(r)
		if err != nil {
			for _, cookie := range r.Cookies() {
				log.Printf("Cookie: %v", *cookie)
			}
			httputil.HandleError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		orgList, err := h.db.GetUsernameOrganizations(s.Username)
		if err != nil {
			log.Printf("Error getting username organizations: %v", err)
			httputil.HandleError(w, errors.InternalServerError, http.StatusInternalServerError)
			return
		}

		for _, o := range orgList {
			orgs = append(orgs, o.Name)
		}
	} else {
		orgs = append(orgs, org)
	}

	questions, err := h.search.Search(query, orgs)
	if err != nil {
		httputil.HandleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(httputil.JSON(questions))
}

/* GET /questions
 *
 * Receives a page of questions
 * TODO: accept query params
 */
func (h *Handler) GetQuestions(w http.ResponseWriter, r *http.Request) {
	var questions []question.Question
	var err error

	qparams := query.ParseParams(r)
	if userVal, ok := qparams["user"]; ok {
		id, err := strconv.Atoi(userVal)
		if err != nil {
			httputil.HandleError(w, errors.InvalidQueryParamError, http.StatusBadRequest)
			return
		}
		questions, err = h.db.GetUserQuestions(id)
	} else {
		questions, err = h.db.GetQuestions()
	}

	if err != nil {
		httputil.HandleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	w.Write(httputil.JSON(questions))
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
		httputil.HandleError(w, errors.BadIDError, http.StatusBadRequest)
		return
	}

	err = h.db.ViewQuestion(id)
	if err != nil {
		log.Printf("Unable to update view count for question with id: %v. Error: %v", id, err)
		httputil.HandleError(w, errors.DBUpdateError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // TODO: include JSON body
}
