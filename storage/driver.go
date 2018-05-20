package storage

import (
	"github.com/JonathonGore/knowledge-base/models/answer"
	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/question"
	"github.com/JonathonGore/knowledge-base/models/team"
	"github.com/JonathonGore/knowledge-base/models/user"
)

type Driver interface {
	InsertAnswer(answer answer.Answer) error

	// TODO: GetQuestion should return an additional boolean to indicate existance
	GetQuestion(id int) (question.Question, error)
	GetQuestions() ([]question.Question, error)
	InsertQuestion(question question.Question) error
	ViewQuestion(id int) error

	InsertUser(user user.User) error
	GetUser(userID int) (user.User, error)
	GetUserByUsername(username string) (user.User, error)

	GetTeam(teamID int) (team.Team, error)
	InsertTeam(team.Team) error

	GetOrganization(orgID int) (organization.Organization, error)
	InsertOrganization(organization.Organization) error
}
