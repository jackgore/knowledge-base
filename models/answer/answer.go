package answer

import (
	"time"
)

type Answer struct {
	SubmittedOn time.Time `json:"submitted-on"`
	AuthoredBy  int       `json:"authored-by"`
	Content     string    `json:"content"`
	Accepted    bool      `json:"accepted"`
	Question    int       `json:"question"`
}
