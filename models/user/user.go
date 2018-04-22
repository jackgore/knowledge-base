package user

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type User struct {
	ID        int       `json:"id,omitempty"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	JoinedOn  time.Time `json:"joined-on"`
}

// Formats the users to a string omitting the password
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

/* Validates the given user to make sure all fields all
 * meet the required specifications.
 */
func Validate(user User) error {
	err := validateID(user.ID)
	if err != nil {
		return err
	}

	return nil
}

func validateID(id int) error {
	if id < 0 {
		return fmt.Errorf("ID must be a non-negative integer. Received: %v.", id)
	}

	return nil
}
