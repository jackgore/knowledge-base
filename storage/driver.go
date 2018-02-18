package storage

import (
	"github.com/JonathonGore/knowledge-base/models"
)

type Driver interface {
	InsertQuestion(question models.Question) error
	GetQuestions() ([]models.Question, error)

	InsertUser(user models.User) error
	GetUser(userID int) (models.User, error)
}
