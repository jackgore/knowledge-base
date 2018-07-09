package search

import "github.com/JonathonGore/knowledge-base/models/question"

type Search interface {
	IndexQuestion(question.Question) error
	Search(query string) ([]question.Question, error)
}
