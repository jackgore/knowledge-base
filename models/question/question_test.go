package question

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/suite"
)

type QuestionTestSuite struct {
	suite.Suite
}

func (s *QuestionTestSuite) SetupSuite() {
	log.SetOutput(ioutil.Discard)
}

func (s *QuestionTestSuite) TestValidateID() {
	s.NotNil(validateID(-1))
	s.NotNil(validateID(-9))
	s.Nil(validateID(1))
	s.Nil(validateID(100))
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(QuestionTestSuite))
}
