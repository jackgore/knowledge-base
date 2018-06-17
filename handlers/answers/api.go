package answers

import (
	"net/http"
)

type AnswerRoutes interface {
	GetAnswers(w http.ResponseWriter, r *http.Request)
	SubmitAnswer(w http.ResponseWriter, r *http.Request)
}
