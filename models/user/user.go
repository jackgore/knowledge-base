package user

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Bio       string    `json:"bio"`
	JoinedOn  time.Time `json:"joined-on"`
}
