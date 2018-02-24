package storage

import (
	"github.com/JonathonGore/knowledge-base/models/question"
	"github.com/JonathonGore/knowledge-base/models/user"
)

type Driver interface {
	InsertQuestion(question question.Question) error
	GetQuestions() ([]question.Question, error)

	InsertUser(user user.User) error
	GetUser(userID int) (user.User, error)
}
