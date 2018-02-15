package storage

import (
	"github.com/JonathonGore/knowledge-base/models"
)

type Driver interface {
	InsertQuestion(author models.User, question models.Question) error
}
