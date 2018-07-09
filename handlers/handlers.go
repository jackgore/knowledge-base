package handlers

import (
	"github.com/JonathonGore/knowledge-base/handlers/answers"
	"github.com/JonathonGore/knowledge-base/handlers/organizations"
	"github.com/JonathonGore/knowledge-base/handlers/questions"
	"github.com/JonathonGore/knowledge-base/handlers/teams"
	"github.com/JonathonGore/knowledge-base/handlers/users"
	"github.com/JonathonGore/knowledge-base/search"
	"github.com/JonathonGore/knowledge-base/session"
	"github.com/JonathonGore/knowledge-base/storage"
)

type Handler struct {
	users.UserRoutes
	questions.QuestionRoutes
	organizations.OrganizationRoutes
	answers.AnswerRoutes
	teams.TeamRoutes

	db             storage.Driver
	sessionManager session.Manager
}

func New(d storage.Driver, sm session.Manager, search search.Search) (*Handler, error) {
	userHandler, err := users.New(d, sm)
	if err != nil {
		return nil, err
	}

	questionHandler, err := questions.New(d, sm, search)
	if err != nil {
		return nil, err
	}

	orgHandler, err := organizations.New(d, sm)
	if err != nil {
		return nil, err
	}

	teamHandler, err := teams.New(d, sm)
	if err != nil {
		return nil, err
	}

	answerHandler, err := answers.New(d, sm)
	if err != nil {
		return nil, err
	}

	handler := &Handler{
		UserRoutes:         userHandler,
		QuestionRoutes:     questionHandler,
		OrganizationRoutes: orgHandler,
		AnswerRoutes:       answerHandler,
		TeamRoutes:         teamHandler,
		db:                 d,
		sessionManager:     sm,
	}

	return handler, nil
}
