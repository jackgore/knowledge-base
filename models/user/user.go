package user

import (
	"fmt"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	JoinedOn  time.Time `json:"joined-on"`
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
