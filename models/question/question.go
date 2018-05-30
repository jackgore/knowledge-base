package question

import (
	"fmt"
	"time"
)

const (
	maxTitleLength = 200 // Maximum length of a title of a question
	minTitleLength = 10  // Minimum length of a title of a question

	maxContentLength = 2000 // Maximum length of a question
	minContentLength = 10   // Minimum length of a question
)

type Question struct {
	ID           int       `json:"id"`
	SubmittedOn  time.Time `json:"submitted-on"`
	Author       int       `json:"author,omitempty"`
	Username     string    `json:"username"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Views        int       `json:"views"`
	Upvotes      int       `json:"upvotes"`
	Team         string    `json:"team,omitempty"`
	Organization string    `json:"organization,omitempty"`
}

/* Validates the given question to make sure all fields all
 * meet the required specifications.
 */
func Validate(question Question) error {
	err := validateID(question.ID)
	if err != nil {
		return err
	}

	err = validateQuestionTitle(question.Title)
	if err != nil {
		return err
	}

	// Note we will ignore the submitted time as we will replace
	// whatever we receive from the client.

	return nil
}

func validateQuestionContent(content string) error {
	if len(content) > maxContentLength {
		return fmt.Errorf("Length of question must be less than %v. Has length of %v.", maxContentLength, len(content))
	} else if len(content) < minContentLength {
		return fmt.Errorf("Length of question must be at least %v. Has length of %v.", minContentLength, len(content))
	}

	return nil
}

func validateQuestionTitle(title string) error {
	if len(title) > maxTitleLength {
		return fmt.Errorf("Length of question title must be less than %v. Has length of %v.", maxTitleLength, len(title))
	} else if len(title) < minTitleLength {
		return fmt.Errorf("Length of question title must be at least %v. Has length of %v.", minTitleLength, len(title))
	}

	return nil
}

func validateID(id int) error {
	if id < 0 {
		return fmt.Errorf("ID must be a non-negative integer. Received: %v.", id)
	}

	return nil
}
