package user

type LoginAttempt struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
