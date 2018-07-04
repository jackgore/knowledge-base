package sql

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/JonathonGore/knowledge-base/config"
	_ "github.com/lib/pq"
)

const (
	MaxRetries = 3
	RetryDelay = 10 // delay between retrying db connection
)

type driver struct {
	db *sql.DB
}

// Connect is a helper function to attempt to establish
// a connection to the database X times.
func connect(db *sql.DB, retries int) {
	err := db.Ping()
	if err != nil && retries > 0 {
		time.Sleep(RetryDelay * time.Second)
		log.Printf("Retrying to connect to the database")
		connect(db, retries-1)
	} else if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
}

// New consumes an sql.Config object and creates a new postgres driver.
func New(conf config.DBConfig) (*driver, error) {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		conf.Host, 5432, conf.User, conf.Password, conf.Name)

	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Printf("error opening connection to database %v", err)
		return nil, err
	}

	log.Printf("Attempting to connect to the database")
	connect(db, MaxRetries)
	log.Printf("Successfully connected to database")

	return &driver{db}, nil
}
