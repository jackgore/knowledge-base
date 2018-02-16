package models

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	JoinedOn  time.Time `json:"joined-on"`
}
