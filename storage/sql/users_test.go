package sql

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/suite"
)

const (
	initialDBUserID = 1
	dbFile          = "test.db"
)

type UsersTestSuite struct {
	suite.Suite
	d *driver
}

func (s *UsersTestSuite) TestGetUser() {
	_, err := s.d.GetUser(initialDBUserID)
	s.NotNil(err)
}

func (s *UsersTestSuite) SetupSuite() {
	os.Remove(dbFile)

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("Unable to create sqlite database")
	}

	s.d = &driver{db}
}

func (s *UsersTestSuite) TearDownSuite() {
	s.d.db.Close()
	os.Remove(dbFile)
}

func TestExampleTestSuite(t *testing.T) {
	tests := &UsersTestSuite{}
	suite.Run(t, tests)
}
