package sql

import (
	"log"

	"github.com/JonathonGore/knowledge-base/models/answer"
)

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
		log.Printf("Unbale to begin transaction: %v", err)
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
