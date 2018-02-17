package models

import (
	"time"
)

type Question struct {
	ID          int       `json:"id"`
	SubmittedOn time.Time `json:"submitted-on"`
	AuthoredBy  int       `json:"authored-by"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
}
