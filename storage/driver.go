package storage

import (
	"github.com/JonathonGore/knowledge-base/models/answer"
	"github.com/JonathonGore/knowledge-base/models/question"
	"github.com/JonathonGore/knowledge-base/models/user"
)

type Driver interface {
	InsertAnswer(answer answer.Answer) error

	InsertQuestion(question question.Question) error
	ViewQuestion(id int) error
	GetQuestion(id int) (question.Question, error)
	GetQuestions() ([]question.Question, error)

	InsertUser(user user.User) error
	GetUser(userID int) (user.User, error)
	GetUserByUsername(username string) (user.User, error)
}
