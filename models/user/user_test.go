package user

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
	suite.Suite
}

func (s *UserTestSuite) SetupSuite() {
	log.SetOutput(ioutil.Discard)
}

func (s *UserTestSuite) TestValidate() {
	s.NotNil(Validate(User{ID: -1}))
}

func (s *UserTestSuite) TestValidateID() {
	s.NotNil(validateID(-1))
	s.NotNil(validateID(-9))
	s.Nil(validateID(1))
	s.Nil(validateID(100))
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
