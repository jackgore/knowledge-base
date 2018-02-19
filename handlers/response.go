package handlers

import (
	"encoding/json"
	"log"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *ErrorResponse) toJSON() []byte {
	contents, err := json.Marshal(e)
	if err != nil {
		log.Printf("Unable to convert erorr response to JSON: %v", err)
		return []byte{}
	}

	return contents
}
