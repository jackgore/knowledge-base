package sql

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/JonathonGore/knowledge-base/models"
	_ "github.com/lib/pq"
)

type driver struct {
	db *sql.DB
}

/* Inserts the given user into the database.
 * This is an all or nothing insertion.
 */
func (d *driver) InsertUser(user models.User) error {
	_, err := d.db.Exec("INSERT INTO author(first_name, last_name, joined_on) VALUES($1, $2, $3)", user.FirstName, user.LastName, user.JoinedOn)
	if err != nil {
		log.Printf("Unable to insert user: %v", err)
		return err
	}

	return nil
}

/* Gets the user with the given userID from the database.
 */
func (d *driver) GetUser(userID int) (models.User, error) {
	user := models.User{}

	err := d.db.QueryRow("SELECT first_name, last_name, joined_on"+
		" FROM author WHERE id=$1", userID).Scan(&user.FirstName, &user.LastName, &user.JoinedOn)
	if err != nil {
		log.Printf("Unable to retrieve user with id %v: %v", userID, err)
		return user, err
	}

	return user, nil
}

/* Gets a page of questions from the database
 */
func (d *driver) GetQuestions() ([]models.Question, error) {
	rows, err := d.db.Query(
		" SELECT post.id as id, submitted_on, title, content, author" +
			" FROM post NATURAL JOIN question" +
			" order by submitted_on")
	if err != nil {
		log.Printf("Unable to receive questions from the db: %v", err)
		return nil, err
	}

	questions := make([]models.Question, 10)
	for rows.Next() {
		question := models.Question{}
		err := rows.Scan(&question.ID, &question.SubmittedOn, &question.Title, &question.Content, &question.AuthoredBy)
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
func (d *driver) InsertQuestion(question models.Question) error {
	tx, err := d.db.Begin()
	if err != nil {
		log.Printf("Unable to begin transaction: %v", err)
		return err
	}

	var postID int
	err = tx.QueryRow("INSERT INTO post(submitted_on, title, content, author) VALUES($1,$2,$3,$4) returning id;",
		question.SubmittedOn, question.Title, question.Content, question.AuthoredBy).Scan(&postID)
	if err != nil {
		log.Printf("Unable to insert post: %v", err)
		return tx.Rollback() // Not sure if we want to return this error
	}

	_, err = tx.Exec("INSERT INTO question(id) VALUES($1)", postID)
	if err != nil {
		log.Printf("Unable to insert post: %v", err)
		return tx.Rollback() // Not sure if we want to return this error
	}

	return tx.Commit()
}

/* Inserts the given answer into the database.
 * This is an all or nothing insertion.
 */
func (d *driver) InsertAnswer(answer models.Answer) error {
	tx, err := d.db.Begin()
	if err != nil {
		log.Printf("Unbale to begin transaction: %", err)
		return err
	}

	var followID int
	err = tx.QueryRow("INSERT INTO followup(submitted_on, content, author) VALUES($1,$2,$3,$4) returning id;",
		answer.SubmittedOn, answer.Content, answer.AuthoredBy).Scan(&followID)
	if err != nil {
		log.Printf("Unable to insert answer: %v", err)
		return tx.Rollback() // Not sure if we want to return this error
	}

	_, err = tx.Exec("INSERT INTO answer(id, question, accepted) VALUES($1,$2,$3)",
		followID, answer.Question, answer.Accepted)
	if err != nil {
		log.Printf("Unable to insert answer: %v", err)
		return tx.Rollback() // Not sure if we want to return this error
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
		log.Fatalf("Unable to connect to database: %v")
	}

	log.Printf("Successfully connected to database")
	d := &driver{db}

	return d, nil
}
