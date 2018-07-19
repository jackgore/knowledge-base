package questions

import (
	"net/http"

	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/question"
	"github.com/JonathonGore/knowledge-base/models/team"
	"github.com/JonathonGore/knowledge-base/models/user"
	sess "github.com/JonathonGore/knowledge-base/session"
)

type MockSearch struct{}

func (m *MockSearch) Search(q string, orgs []string) ([]question.Question, error) {
	return nil, nil
}

func (m *MockSearch) IndexQuestion(q question.Question) error {
	return nil
}

type MockSession struct{}

func (m *MockSession) GetSession(r *http.Request) (sess.Session, error) {
	var s sess.Session

	return s, nil
}

func (m *MockSession) HasSession(r *http.Request) bool {
	return true
}

func (m *MockSession) SessionStart(w http.ResponseWriter, r *http.Request, username string) (sess.Session, error) {
	var s sess.Session

	return s, nil
}

func (m *MockSession) SessionDestroy(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type MockStorage struct{}

func (m *MockStorage) GetOrgQuestions(org string) ([]question.Question, error) {
	return nil, nil
}

func (m *MockStorage) GetUsernameOrganizations(username string) ([]organization.Organization, error) {
	return nil, nil
}

func (m *MockStorage) GetOrganizationByName(name string) (organization.Organization, error) {
	return organization.Organization{}, nil
}

func (m *MockStorage) DeleteQuestion(id int) error {
	return nil
}

func (m *MockStorage) VoteQuestion(id, uid int, upvote bool) error {
	return nil
}

func (m *MockStorage) GetOrganizationMembers(org string, admin bool) ([]string, error) {
	return nil, nil
}

func (m *MockStorage) GetQuestion(id int) (question.Question, error) {
	return question.Question{}, nil
}

func (m *MockStorage) GetQuestions() ([]question.Question, error) {
	return nil, nil
}

func (m *MockStorage) GetTeamQuestions(team, org string) ([]question.Question, error) {
	return nil, nil
}

func (m *MockStorage) GetTeamByName(org, t string) (team.Team, error) {
	return team.Team{}, nil
}

func (m *MockStorage) GetUserByUsername(username string) (user.User, error) {
	return user.User{}, nil
}

func (m *MockStorage) GetUserQuestions(id int) ([]question.Question, error) {
	return nil, nil
}

func (m *MockStorage) InsertQuestion(question question.Question) (int, error) {
	return 0, nil
}

func (m *MockStorage) InsertTeamQuestion(question question.Question, tid int) (int, error) {
	return 0, nil
}

func (m *MockStorage) ViewQuestion(id int) error {
	return nil
}
