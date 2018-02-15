package sql

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"

	"github.com/JonathonGore/knowledge-base/models"
)

type driver struct {
	db *sql.DB
}

func (d *driver) InsertQuestion(author models.User, question models.Question) error {
	return nil
}

func New(config Config) (*driver, error) {
	dbinfo := fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=verify-full",
		config.Username, config.Password, config.Host, config.DatabaseName)

	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Printf("unable to connect to database %v", err)
		return nil, err
	}
	defer db.Close()

	log.Printf("Successfully connected to database")
	d := &driver{db}

	return d, nil
}
