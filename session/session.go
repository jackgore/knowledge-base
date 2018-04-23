package session

import "time"

type Session struct {
	SID      string    `json:"sid"`
	Username string    `json:"username"`
	Expiry   time.Time `json:"expiry"`
}
