package sql

import (
	"database/sql"
	"errors"
	"log"

	"github.com/JonathonGore/knowledge-base/models/question"
)

// DeleteQuestion deletes the question with the given id from the database.
func (d *driver) DeleteQuestion(id int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM question WHERE id = $1;", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM post_of WHERE pid = $1;", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM post WHERE id = $1;", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetQuestion retrieves the question with the given id from the database.
func (d *driver) GetQuestion(id int) (question.Question, error) {
	question := question.Question{}
	err := d.db.QueryRow(
		" SELECT post.id as id, users.username, submitted_on, title, content, author, views, organization.name,"+
			" (SELECT count(*) from answer where post.id=answer.question) as answers"+
			" FROM ((((post NATURAL JOIN question) JOIN users ON (author = users.id))"+
			" JOIN post_of ON (post.id = post_of.pid)) JOIN team ON (team.id = post_of.tid))"+
			" JOIN organization ON (team.org_id = organization.id)"+
			" where post.id=$1",
		id).Scan(&question.ID, &question.Username, &question.SubmittedOn, &question.Title,
		&question.Content, &question.Author, &question.Views, &question.Organization, &question.Answers)
	if err != nil {
		log.Printf("Unable to retrieve question with id %v: %v", id, err)
		return question, err
	}

	return question, nil
}

// ViewQuestion updates the view count by one for the question with the given id
func (d *driver) ViewQuestion(id int) error {
	_, err := d.db.Exec("UPDATE post SET views = views + 1 WHERE id = $1;", id)
	if err != nil {
		log.Printf("Unable to update view count for question with id %v: %v", id, err)
		return err
	}

	return nil
}

// VoteQuestion updates the view count by one for the question with the given id
func (d *driver) VoteQuestion(qid int, uid int, upvote bool) error {
	rows, err := d.db.Query("SELECT upvote FROM vote WHERE uid=$1 AND qid=$2", uid, qid)
	if err != nil {
		return err
	}

	if rows.Next() {
		// we expect a maximum of one result row
		var vote bool
		err := rows.Scan(&vote)
		if err != nil {
			return err
		}

		// Now update the database according to the provided value
		_, err = d.db.Exec("UPDATE vote SET upvote=$1 WHERE uid=$2 AND qid=$3;", upvote, uid, qid)
		if err != nil {
			return err
		}
		return nil
	}

	_, err = d.db.Exec("INSERT INTO vote (qid, uid, upvote) VALUES ($1, $2, $3)", qid, uid, upvote)
	if err != nil {
		return err
	}

	return nil
}

/* Gets a page of questions from the database
 * for the user with the given userid
 */
func (d *driver) GetUserQuestions(uid int) ([]question.Question, error) {
	rows, err := d.db.Query(
		" SELECT post.id as id, submitted_on, title, content, author, views,"+
			" (SELECT count(*) from answer where post.id=answer.question) as answers"+
			" FROM post NATURAL JOIN question"+
			" WHERE post.id NOT IN (SELECT pid FROM post_of) AND author=$1"+
			" order by submitted_on", uid)
	if err != nil {
		log.Printf("Unable to receive questions from the db: %v", err)
		return nil, err
	}

	// TODO The two GetQuestions function have very similar scanning code but differ on 1 column
	//      not sure the best way to abstract this.
	questions := make([]question.Question, 0)
	for rows.Next() {
		question := question.Question{}
		err := rows.Scan(&question.ID, &question.SubmittedOn, &question.Title,
			&question.Content, &question.Author, &question.Views, &question.Answers)
		if err != nil {
			log.Printf("Received error scanning in data from database: %v", err)
			continue
		}
		questions = append(questions, question)
	}

	return questions, err
}

/* Gets a page of questions from the database
 */
func (d *driver) GetQuestions() ([]question.Question, error) {
	rows, err := d.db.Query(
		" SELECT post.id as id, submitted_on, title, content, username, views," +
			" (SELECT count(*) from answer where post.id=answer.question) as answers" +
			" FROM (post NATURAL JOIN question) JOIN users on (users.id = post.author)" +
			" WHERE post.id NOT IN (SELECT pid FROM post_of)" +
			" ORDER BY views DESC")
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

func scanQuestions(rows *sql.Rows) ([]question.Question, error) {
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

	return questions, nil

}

/* Gets a page of questions from the database for the requested team and org
 */
func (d *driver) GetOrgQuestions(org string) ([]question.Question, error) {
	rows, err := d.db.Query(
		" SELECT post.id as id, submitted_on, title, content, username, views,"+
			" (SELECT count(*) from answer where post.id=answer.question) as answers"+
			" FROM (((question NATURAL JOIN post) JOIN users ON (post.author = users.id))"+
			" JOIN post_of ON (post_of.pid = question.id)) JOIN team ON (post_of.tid=team.id)"+
			" WHERE team.org_id=(SELECT distinct id FROM organization WHERE name=$1)", org)
	if err != nil {
		log.Printf("Unable to receive questions from the db: %v", err)
		return nil, err
	}

	return scanQuestions(rows)
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

	return scanQuestions(rows)
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
