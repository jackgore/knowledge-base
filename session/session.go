package session

import "time"

type Session struct {
	SID       string    `json:"sid"`
	Username  string    `json:"username"`
	CreatedOn time.Time `json:"created-on"`
	ExpiresOn time.Time `json:"expires-on"`
}
