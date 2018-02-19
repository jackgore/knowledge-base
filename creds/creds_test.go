package creds

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	validUsername = "username"
	validPassword = "password"
)

type CredsTestSuite struct {
	suite.Suite
}

func (s *CredsTestSuite) SetupSuite() {
	log.SetOutput(ioutil.Discard)
}

func (s *CredsTestSuite) TestValidateSignupCredentials() {
	// Too short of username should fail
	s.NotNil(ValidateSignupCredentials("hi", validPassword))

	// Too short of passowrd should fail
	s.NotNil(ValidateSignupCredentials(validUsername, "hi"))

	// Usernames that are not url safe should fail
	s.NotNil(ValidateSignupCredentials("hi there jack", validPassword))

	// Passwords that are not url safe should fail
	s.NotNil(ValidateSignupCredentials(validUsername, "hey there jacob"))

	// Valid username pass should succeed
	s.Nil(ValidateSignupCredentials(validUsername, validPassword))
}

func (s *CredsTestSuite) TestIsURLSafe() {
	s.True(isURLSafe('a'))
	s.True(isURLSafe('A'))
	s.True(isURLSafe('u'))
	s.True(isURLSafe('-'))
	s.True(isURLSafe('.'))
	s.True(isURLSafe('_'))
	s.True(isURLSafe('~'))
	s.True(isURLSafe('1'))
	s.False(isURLSafe('*'))
	s.False(isURLSafe(')'))
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(CredsTestSuite))
}
