package httputil

type IDResponse struct {
	ID int `json:"id"`
}

type LoginResponse struct {
	SID string `json:"session-id"`
}

type SuccessResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
