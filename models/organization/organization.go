package organization

import (
	"fmt"
	"strings"
	"time"
)

const (
	// TODO: These should be read in from a file
	maxNameLength = 100 // Maximum length of the name of the org
	minNameLength = 1   // Minimum length of the name of the org
)

type Organization struct {
	ID          int       `json:"id"`
	MemberCount int       `json:"member-count"`
	TeamCount   int       `json:"team-count"`
	AdminCount  int       `json:"admin-count"`
	Name        string    `json:"name"`
	CreatedOn   time.Time `json:"created-on"`
	IsPublic    bool      `json:"is-public"`

	Members []int `json:"members"` // Note: Not sure if it makes sense to have these fields
	Admins  []int `json:"admins"`
}

// Validate ensures the given org meets the required specifications.
func Validate(org Organization) error {
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
		return fmt.Errorf("Length of org name must be less than %v. Has length of %v.", maxNameLength, len(name))
	} else if len(name) < minNameLength {
		return fmt.Errorf("Length of org namemust be at least %v. Has length of %v.", minNameLength, len(name))
	}

	if strings.Contains(name, " ") {
		return fmt.Errorf("Org names cannot contain spaces")
	}

	return nil
}

func validateID(id int) error {
	if id < 0 {
		return fmt.Errorf("ID must be a non-negative integer. Received: %v.", id)
	}

	return nil
}
