package user

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"time"
)

var (
	emailRegex *regexp.Regexp
)

func init() {
	// Regular expression for email validation.
	// Taken from http://www.golangprograms.com/golang-package-examples/regular-expression-to-validate-email-address.html
	emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:" +
		"[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

}

// User contains information relative to a user of the site.
type User struct {
	ID            int       `json:"id,omitempty"`
	Username      string    `json:"username"`
	Password      string    `json:"password,omitempty"`
	Email         string    `json:"email"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Organizations []string  `json:"organizations"`
	JoinedOn      time.Time `json:"joined-on"`
}

// SafePrint formats the user to a string omitting the password.
// Used if we ever need to log the user object without leaking credentials.
func (user *User) SafePrint() string {
	cuser := *user
	cuser.Password = ""

	b, err := json.Marshal(cuser)
	if err != nil {
		log.Printf("Unable to safe print user object: %v", err)
		return ""
	}

	return string(b)
}

// Validate ensures the given user to make sure all fields meet the required
// specifications.
func Validate(user User) error {
	if err := validateID(user.ID); err != nil {
		return err
	}

	if err := ValidateEmail(user.Email); err != nil {
		return err
	}

	return nil
}

// ValidateEmail consumes the given string and determines if it is a valid
// email address.
func ValidateEmail(email string) error {
	if ok := emailRegex.MatchString(email); !ok {
		return fmt.Errorf("Given email address: %v is not a valid email address", email)
	}

	return nil
}

// validateID ensures the user id is in an acceptable range.
func validateID(id int) error {
	if id < 0 {
		return fmt.Errorf("User id must be a non-negative integer. Received: %v.", id)
	}

	return nil
}
