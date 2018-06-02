package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/JonathonGore/knowledge-base/models/answer"
	"github.com/JonathonGore/knowledge-base/models/organization"
	"github.com/JonathonGore/knowledge-base/models/question"
	"github.com/JonathonGore/knowledge-base/models/team"
	"github.com/JonathonGore/knowledge-base/models/user"
	"github.com/JonathonGore/knowledge-base/session"
	_ "github.com/lib/pq"
)

type driver struct {
	db *sql.DB
}

/* Gets the session with the given sid from the database.
 */
func (d *driver) GetSession(sid string) (session.Session, error) {
	s := session.Session{}
	err := d.db.QueryRow("SELECT sid, username, created_on, expires_on FROM session WHERE sid=$1",
		sid).Scan(&s.SID, &s.Username, &s.CreatedOn, &s.ExpiresOn)
	if err != nil {
		log.Printf("Unable to retrieve session with sid %v: %v", sid, err)
		return s, err
	}

	return s, nil
}

/* Inserts the given session into the database
 */
func (d *driver) InsertSession(s session.Session) error {
	_, err := d.db.Exec("INSERT INTO session(sid, username, created_on, expires_on) VALUES($1, $2, $3, $4)",
		s.SID, s.Username, s.CreatedOn, s.ExpiresOn)
	if err != nil {
		log.Printf("Unable to insert session: %v", err)
		return err
	}

	return nil
}

/* Deletes the session with the sid from the database
 */
func (d *driver) DeleteSession(sid string) error {
	_, err := d.db.Exec("DELETE FROM session WHERE sid=$1", sid)
	if err != nil {
		log.Printf("Unable to delete session: %v", err)
		return err
	}

	return nil
}

/* Gets the org with the given ID from the database.
 */
func (d *driver) GetOrganization(orgID int) (organization.Organization, error) {
	org := organization.Organization{}
	err := d.db.QueryRow("SELECT id, name, created_on, is_public FROM team WHERE id=$1",
		orgID).Scan(&org.ID, &org.Name, &org.CreatedOn, &org.IsPublic)
	if err != nil {
		log.Printf("Unable to retrieve org with id %v: %v", orgID, err)
		return org, err
	}

	return org, nil
}

/* Gets the org with the given name.
 */
func (d *driver) GetOrganizationByName(name string) (organization.Organization, error) {
	org := organization.Organization{}
	err := d.db.QueryRow("SELECT id, name, created_on, is_public FROM organization WHERE name=$1",
		name).Scan(&org.ID, &org.Name, &org.CreatedOn, &org.IsPublic)
	if err != nil {
		log.Printf("Unable to retrieve org with name %v: %v", name, err)
		return org, err
	}

	return org, nil
}

/* Gets a page of organizations from the database
 */
func (d *driver) GetOrganizations() ([]organization.Organization, error) {
	rows, err := d.db.Query("SELECT id, name, created_on, is_public" +
		" FROM organization" +
		" order by name")
	if err != nil {
		log.Printf("Unable to receive organizations from the db: %v", err)
		return nil, err
	}

	orgs := make([]organization.Organization, 0)
	for rows.Next() {
		org := organization.Organization{}
		err := rows.Scan(&org.ID, &org.Name, &org.CreatedOn, &org.IsPublic)
		if err != nil {
			log.Printf("Received error scanning in data from database: %v", err)
			continue
		}
		orgs = append(orgs, org)
	}

	return orgs, err
}

/* Inserts the given organization into the database
 */
func (d *driver) InsertOrganization(org organization.Organization) error {
	_, err := d.db.Exec("INSERT INTO organization(name, created_on, is_public) VALUES($1, $2, $3)",
		org.Name, org.CreatedOn, org.IsPublic)
	if err != nil {
		log.Printf("Unable to insert org: %v", err)
		return err
	}

	return nil
}

/* Gets the teams for the given org from the database.
 */
func (d *driver) GetTeams(org string) ([]team.Team, error) {
	rows, err := d.db.Query("SELECT team.id, team.org_id, team.name, team.created_on, team.is_public"+
		" FROM team JOIN organization on (team.org_id = organization.id)"+
		" WHERE organization.name = $1"+
		" order by team.name", org)
	if err != nil {
		log.Printf("Unable to receive teams for org %v from the db: %v", org, err)
		return nil, err
	}

	teams := make([]team.Team, 0)
	for rows.Next() {
		team := team.Team{}
		err := rows.Scan(&team.ID, &team.Organization, &team.Name, &team.CreatedOn, &team.IsPublic)
		if err != nil {
			log.Printf("Received error scanning in data from database: %v", err)
			continue
		}
		teams = append(teams, team)
	}

	return teams, err
}

/* Gets the team with the given ID from the database.
 */
func (d *driver) GetTeam(teamID int) (team.Team, error) {
	t := team.Team{}
	err := d.db.QueryRow("SELECT id, org_id, name, created_on, is_public FROM team WHERE id=$1",
		teamID).Scan(&t.ID, &t.Organization, &t.Name, &t.CreatedOn, &t.IsPublic)
	if err != nil {
		log.Printf("Unable to retrieve team with id %v: %v", teamID, err)
		return t, err
	}

	return t, nil
}

/* Gets the team with the given name from the given org from the database.
 */
func (d *driver) GetTeamByName(org, name string) (team.Team, error) {
	t := team.Team{}
	err := d.db.QueryRow("SELECT team.id, team.org_id, team.name, team.created_on, team.is_public FROM team JOIN organization ON "+
		" (team.org_id = organization.id) WHERE organization.name=$1 and team.name=$2",
		org, name).Scan(&t.ID, &t.Organization, &t.Name, &t.CreatedOn, &t.IsPublic)
	if err != nil {
		log.Printf("Unable to retrieve team with name %v from organization %v: %v", name, org, err)
		return t, err
	}

	return t, nil
}

/* Inserts the given team into the database
 */
func (d *driver) InsertTeam(t team.Team) error {
	_, err := d.db.Exec("INSERT INTO team(org_id, name, created_on, is_public) VALUES($1, $2, $3, $4)",
		t.Organization, t.Name, t.CreatedOn, t.IsPublic)
	if err != nil {
		log.Printf("Unable to insert team: %v", err)
		return err
	}

	return nil
}

/* Inserts the given user into the database.
 * This is an all or nothing insertion.
 *
 * Note: Assumes the password in the user object has already been hashed
 *
 * TODO: This function should return the id of the inserted user
 */
func (d *driver) InsertUser(user user.User) error {
	_, err := d.db.Exec("INSERT INTO users(first_name, last_name, username, password, email, joined_on) VALUES($1, $2, $3, $4, $5, $6)",
		user.FirstName, user.LastName, user.Username, user.Password, user.Email, user.JoinedOn)
	if err != nil {
		log.Printf("Unable to insert user: %v", err)
		return err
	}

	return nil
}

/* Attempts to retrieve the user with the given username from the database.
 * Usually used when attempting to see if a username is attached to a user.
 */
func (d *driver) GetUserByUsername(username string) (user.User, error) {
	user := user.User{}
	err := d.db.QueryRow("SELECT id, first_name, last_name, joined_on, password FROM users WHERE username=$1",
		username).Scan(&user.ID, &user.FirstName, &user.LastName, &user.JoinedOn, &user.Password)
	if err != nil {
		log.Printf("User with username %v not found: %v", username, err)
		return user, err
	}

	user.Username = username
	return user, nil
}

/* Gets the user with the given userID from the database.
 */
func (d *driver) GetUser(userID int) (user.User, error) {
	user := user.User{}
	err := d.db.QueryRow("SELECT id, first_name, last_name, joined_on FROM users WHERE id=$1",
		userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.JoinedOn)
	if err != nil {
		log.Printf("Unable to retrieve user with id %v: %v", userID, err)
		return user, err
	}

	return user, nil
}

/* Gets the question with the given id from the database.
 */
func (d *driver) GetQuestion(id int) (question.Question, error) {
	question := question.Question{}
	err := d.db.QueryRow(
		" SELECT post.id as id, users.username, submitted_on, title, content, author, views,"+
			" (SELECT count(*) from answer where post.id=answer.question) as answers"+
			" FROM (post NATURAL JOIN question) JOIN users ON (author = users.id)"+
			" where post.id=$1",
		id).Scan(&question.ID, &question.Username, &question.SubmittedOn, &question.Title,
		&question.Content, &question.Author, &question.Views, &question.Answers)
	if err != nil {
		log.Printf("Unable to retrieve question with id %v: %v", id, err)
		return question, err
	}

	return question, nil
}

/* Updates the view count by one for the question with the given id
 */
func (d *driver) ViewQuestion(id int) error {
	_, err := d.db.Exec("UPDATE post SET views = views + 1 WHERE id = $1;", id)
	if err != nil {
		log.Printf("Unable to update view count for question with id %v: %v", id, err)
		return err
	}

	return nil
}

/* Gets a page of questions from the database
 */
func (d *driver) GetQuestions() ([]question.Question, error) {
	rows, err := d.db.Query(
		" SELECT post.id as id, submitted_on, title, content, username, views," +
			" (SELECT count(*) from answer where post.id=answer.question) as answers" +
			" FROM (post NATURAL JOIN question) JOIN users on (users.id = post.author)" +
			" order by submitted_on")
	if err != nil {
		log.Printf("Unable to receive questions from the db: %v", err)
		return nil, err
	}

	questions := make([]question.Question, 0)
	for rows.Next() {
		question := question.Question{}
		err := rows.Scan(&question.ID, &question.SubmittedOn, &question.Title,
			&question.Content, &question.Username, &question.Views, &question.Answers)
		if err != nil {
			log.Printf("Received error scanning in data from database: %v", err)
			continue
		}
		questions = append(questions, question)
	}

	return questions, err
}

/* Inserts the given question into the database.
 * This is an all or nothing insertion.
 */
func (d *driver) InsertQuestion(question question.Question) (int, error) {
	postID := -1
	tx, err := d.db.Begin()
	if err != nil {
		log.Printf("Unable to begin transaction: %v", err)
		return postID, err
	}

	err = tx.QueryRow("INSERT INTO post(submitted_on, title, content, author) VALUES($1,$2,$3,$4) returning id;",
		question.SubmittedOn, question.Title, question.Content, question.Author).Scan(&postID)
	if err != nil {
		log.Printf("Unable to insert post: %v", err)
		return postID, tx.Rollback() // Not sure if we want to return this error
	}

	_, err = tx.Exec("INSERT INTO question(id) VALUES($1)", postID)
	if err != nil {
		log.Printf("Unable to insert post: %v", err)
		return postID, tx.Rollback() // Not sure if we want to return this error
	}

	return postID, tx.Commit()
}

/* Gets a page of questions from the database for the requested team and org
 */
func (d *driver) GetTeamQuestions(team, org string) ([]question.Question, error) {
	rows, err := d.db.Query(
		" SELECT post.id as id, submitted_on, title, content, username, views,"+
			" (SELECT count(*) from answer where post.id=answer.question) as answers"+
			" FROM ((question NATURAL JOIN post) JOIN users ON (post.author = users.id))"+
			" JOIN post_of ON (post_of.pid = question.id) WHERE"+
			" post_of.tid = (SELECT distinct team.id FROM team, organization WHERE team.name = $1 AND organization.name = $2)", team, org)
	if err != nil {
		log.Printf("Unable to receive questions from the db: %v", err)
		return nil, err
	}

	questions := make([]question.Question, 0)
	for rows.Next() {
		question := question.Question{}
		err := rows.Scan(&question.ID, &question.SubmittedOn, &question.Title, &question.Content,
			&question.Username, &question.Views, &question.Answers)
		if err != nil {
			log.Printf("Received error scanning in data from database: %v", err)
			return questions, err
		}
		questions = append(questions, question)
	}

	return questions, err
}

/* Inserts the given question into the database for the given team.
 * This is an all or nothing insertion.
 */
func (d *driver) InsertTeamQuestion(q question.Question, teamID int) (int, error) {
	postID := -1

	if q.Team == "" || q.Organization == "" {
		return postID, errors.New("Team and organization both must not be empty")
	}

	tx, err := d.db.Begin()
	if err != nil {
		log.Printf("Unable to begin transaction: %v", err)
		return postID, err
	}

	err = tx.QueryRow("INSERT INTO post(submitted_on, title, content, author) VALUES($1,$2,$3,$4) returning id;",
		q.SubmittedOn, q.Title, q.Content, q.Author).Scan(&postID)
	if err != nil {
		log.Printf("Unable to insert post: %v", err)
		return postID, tx.Rollback() // Not sure if we want to return this error
	}

	_, err = tx.Exec("INSERT INTO question(id) VALUES($1)", postID)
	if err != nil {
		log.Printf("Unable to insert post: %v", err)
		return postID, tx.Rollback() // Not sure if we want to return this error
	}

	_, err = tx.Exec("INSERT INTO post_of(pid, tid) VALUES($1,$2)", postID, teamID)
	if err != nil {
		log.Printf("Unable to insert post: %v", err)
		return postID, tx.Rollback() // Not sure if we want to return this error
	}

	return postID, tx.Commit()
}

/* Gets a page of answers from the database
 */
func (d *driver) GetAnswers(qid int) ([]answer.Answer, error) {
	rows, err := d.db.Query(
		"SELECT answer.id, question, accepted, content, submitted_on, author, username"+
			" FROM (answer NATURAL JOIN followup) JOIN users ON (users.id = author) WHERE question=$1;", qid)
	if err != nil {
		log.Printf("Unable to receive answers from the db: %v", err)
		return nil, err
	}

	answers := make([]answer.Answer, 0)
	for rows.Next() {
		ans := answer.Answer{}
		err := rows.Scan(&ans.ID, &ans.Question, &ans.Accepted, &ans.Content, &ans.SubmittedOn, &ans.Author, &ans.Username)
		if err != nil {
			log.Printf("Received error scanning in data from database: %v", err)
			continue
		}
		answers = append(answers, ans)
	}

	return answers, err
}

/* Inserts the given answer into the database.
 * This is an all or nothing insertion.
 */
func (d *driver) InsertAnswer(answer answer.Answer) error {
	tx, err := d.db.Begin()
	if err != nil {
		log.Printf("Unbale to begin transaction: %", err)
		return err
	}

	var followID int
	err = tx.QueryRow("INSERT INTO followup(submitted_on, content, author) VALUES($1,$2,$3) returning id;",
		answer.SubmittedOn, answer.Content, answer.Author).Scan(&followID)
	if err != nil {
		log.Printf("Unable to insert answer: %v", err)
		tx.Rollback() // Not sure if we want to return this error
		return err
	}

	_, err = tx.Exec("INSERT INTO answer(id, question, accepted) VALUES($1,$2,$3)",
		followID, answer.Question, answer.Accepted)
	if err != nil {
		log.Printf("Unable to insert answer: %v", err)
		tx.Rollback() // Not sure if we want to return this error
		return err
	}

	return tx.Commit()
}

/* Creates a new postgres driver consuming an sql.Config object
 * which specifies connection information for the DB.
 */
func New(config Config) (*driver, error) {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, 5432, config.Username, config.Password, config.DatabaseName)

	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Printf("unable to connect to database %v", err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	log.Printf("Successfully connected to database")
	d := &driver{db}

	return d, nil
}
