package user

import (
	"io/ioutil"
	"log"
	"testing"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

var validateEmailTests = []struct {
	email   string
	success bool
}{
	{"", false},
	{" ", false},
	{"email", false},
	{"@", false},
	{"@com", false},
	{"jack@", false},
	{"jack@test.com", true},
}

var validateIDTests = []struct {
	id      int
	success bool
}{
	{-1, false},
	{-9, false},
	{1, true},
	{100, true},
}

var validateTests = []struct {
	user    User
	success bool
}{
	{User{ID: -1}, false},
}

func TestValidateEmail(t *testing.T) {
	for _, test := range validateEmailTests {
		if (ValidateEmail(test.email) == nil) != test.success {
			text := "succeed"
			if !test.success {
				text = "fail"
			}
			t.Errorf("Expected test to %v for email: %v but it did not", text, test.email)
		}
	}

}

func TestValidate(t *testing.T) {
	for _, test := range validateTests {
		if (Validate(test.user) == nil) != test.success {
			text := "succeed"
			if !test.success {
				text = "fail"
			}
			t.Errorf("Expected test to %v for user: %+v but it did not", text, test.user)
		}
	}

}

func TestValidateID(t *testing.T) {
	for _, test := range validateIDTests {
		if (validateID(test.id) == nil) != test.success {
			text := "succeed"
			if !test.success {
				text = "fail"
			}
			t.Errorf("Expected test to %v for id: %v but it did not", text, test.id)
		}
	}
}
