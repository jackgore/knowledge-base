package storage

import (
	"github.com/JonathonGore/knowledge-base/models/answer"
	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/question"
	"github.com/JonathonGore/knowledge-base/models/team"
	"github.com/JonathonGore/knowledge-base/models/user"
	"github.com/JonathonGore/knowledge-base/session"
)

type Driver interface {
	InsertAnswer(answer answer.Answer) error
	GetAnswers(qid int) ([]answer.Answer, error)

	// TODO: GetQuestion should return an additional boolean to indicate existance
	GetQuestion(id int) (question.Question, error)
	GetQuestions() ([]question.Question, error)
	GetUserQuestions(id int) ([]question.Question, error)
	GetTeamQuestions(team, org string) ([]question.Question, error)
	GetOrgQuestions(org string) ([]question.Question, error)
	InsertQuestion(question question.Question) (int, error)
	InsertTeamQuestion(question question.Question, tid int) (int, error)
	ViewQuestion(id int) error

	InsertUser(user user.User) error
	GetUser(userID int) (user.User, error)
	GetUserByUsername(username string) (user.User, error)

	InsertSession(s session.Session) error
	GetSession(sid string) (session.Session, error)
	DeleteSession(sid string) error

	GetTeam(teamID int) (team.Team, error)
	GetTeamByName(org, team string) (team.Team, error)
	GetTeams(org string) ([]team.Team, error)
	InsertTeam(team.Team) error
	InsertTeamMember(username, org, team string, isAdmin bool) error

	GetUserOrganizations(uid int) ([]organization.Organization, error)
	GetOrganization(orgID int) (organization.Organization, error)
	GetOrganizationByName(name string) (organization.Organization, error)
	GetOrganizations() ([]organization.Organization, error)
	GetOrganizationMembers(org string, admins bool) ([]string, error)
	InsertOrganization(organization.Organization) (int, error)
	InsertOrgMember(username, org string, isAdmin bool) error
}
