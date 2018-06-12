package questions

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/JonathonGore/knowledge-base/errors"
	"github.com/JonathonGore/knowledge-base/models/question"
	"github.com/JonathonGore/knowledge-base/query"
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
		httputil.HandleError(w, errors.DBGetError, http.StatusInternalServerError)
		return
	}

	w.Write(httputil.JSON(questions))
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
		httputil.HandleError(w, fmt.Sprintf("Organization %v does not exist", org), http.StatusBadRequest)
		return
	}

	t, err := h.db.GetTeamByName(org, team)
	if err != nil {
		httputil.HandleError(w, fmt.Sprintf("Team %v does not exist", org), http.StatusBadRequest)
		return
	}

	q.Team = team
	q.Organization = org
	q.SubmittedOn = time.Now()

	id, err := h.db.InsertTeamQuestion(q, t.ID)
	if err != nil {
		httputil.HandleError(w, errors.DBInsertError, http.StatusInternalServerError)
		return
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
		httputil.HandleError(w, errors.JSONParseError, http.StatusInternalServerError)
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

	w.Write(httputil.JSON(httputil.IDResponse{id}))
}

/* GET /question/{id}
 *
 * Retrieves a question from the database with the given id
 */
func (h *Handler) GetQuestion(w http.ResponseWriter, r *http.Request) {
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

/* GET /questions
 *
 * Receives a page of questions
 * TODO: accept query params
 */
func (h *Handler) GetQuestions(w http.ResponseWriter, r *http.Request) {
	var questions []question.Question
	var err error

	qparams := query.ParseParams(r)
	if val, ok := qparams["user"]; ok {
		id, cerr := strconv.Atoi(val)
		if cerr != nil {
			httputil.HandleError(w, errors.InvalidQueryParamError, http.StatusInternalServerError)
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
