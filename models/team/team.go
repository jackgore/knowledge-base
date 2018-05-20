package team

import (
	"fmt"
)

const (
	maxNameLength = 100 // Maximum length of the name of the org
	minNameLength = 1   // Minimum length of the name of the org
)

type Team struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Organization string `json:"organization"`
	MemberCount int     `json:"member-count"`
	AdminCount  int     `json:"admin-count"`
}

/* Validates the given org to make sure all fields all
 * meet the required specifications.
 */
func Validate(org Team) error {
	err := validateID(org.ID)
	if err != nil {
		return err
	}

	err = validateName(org.Name)
	if err != nil {
		return err
	}

	return nil
}

func validateName(name string) error {
	if len(name) > maxNameLength {
		return fmt.Errorf("Length of team name must be less than %v. Has length of %v.", maxNameLength, len(name))
	} else if len(name) < minNameLength {
		return fmt.Errorf("Length of team name must be at least %v. Has length of %v.", minNameLength, len(name))
	}

	return nil
}

func validateID(id int) error {
	if id < 0 {
		return fmt.Errorf("ID must be a non-negative integer. Received: %v.", id)
	}

	return nil
}
