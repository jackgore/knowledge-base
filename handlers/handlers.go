package handlers

import (
	//	"fmt"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/JonathonGore/knowledge-base/models"
	"github.com/JonathonGore/knowledge-base/storage"
	"github.com/JonathonGore/knowledge-base/utils"
	"github.com/gorilla/mux"
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
 */
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	user := models.User{}
	err := utils.UnmarshalRequestBody(r, &user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		//w.Write([]{"some json error"}) TODO
		return
	}

	err = h.db.InsertUser(user)
	if err != nil {
		log.Printf("Unable to insert user into database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		//w.Write([]{"some json error"})
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* GET /users/{user-id}
 *
 * Signs up the given user by inserting them into the database.
 */
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["user-id"]

	id, err := strconv.Atoi(userID)
	if err != nil {
		log.Printf("Received bad user id in request paramater: %v", userID)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.db.GetUser(id)
	if err != nil {
		log.Printf("Unable to get user from database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		//w.Write([]{"some json error"})
		return
	}

	contents, err := json.Marshal(user)
	if err != nil {
		log.Printf("Unable to convert user to byte array")
		w.WriteHeader(http.StatusInternalServerError)
		//w.Write([]{"some json error"})
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
	question := models.Question{}
	err := utils.UnmarshalRequestBody(r, &question)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		//w.Write([]{"some json error"}) TODO
		return
	}

	// TODO: Validate we have the needed fields in our question object

	err = h.db.InsertQuestion(question)
	if err != nil {
		log.Printf("Unable to insert question into database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		//w.Write([]{"some json error"})
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
		//w.Write([]{"some json error"})
		return
	}

	contents, err := json.Marshal(questions)
	if err != nil {
		log.Printf("Unable to convert questions to byte array")
		w.WriteHeader(http.StatusInternalServerError)
		//w.Write([]{"some json error"})
		return
	}

	w.Write(contents)
	return
}
