package answer

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AnswerTestSuite struct {
	suite.Suite
}

func (s *AnswerTestSuite) SetupSuite() {
	log.SetOutput(ioutil.Discard)
}

func (s *AnswerTestSuite) TestValidateAnswer() {
	s.NotNil(Validate(Answer{}))
}

func (s *AnswerTestSuite) TestValidateContent() {
	s.NotNil(validateContent(""))
	s.NotNil(validateContent("abc"))
	s.Nil(validateContent("this is a good answer"))
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(AnswerTestSuite))
}
