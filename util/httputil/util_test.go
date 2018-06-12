package httputil

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UtilsTestSuite struct {
	suite.Suite
}

func (s *UtilsTestSuite) SetupSuite() {
	log.SetOutput(ioutil.Discard)
}

func (s *UtilsTestSuite) TestValidateSignupCredentials() {
	// Too short of username should fail
	s.NotNil(UnmarshalRequestBody(nil, nil))
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}
