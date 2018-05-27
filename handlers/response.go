package handlers

import (
	"encoding/json"
	"log"
)

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

func JSONString(e interface{}) string {
	return string(JSON(e))
}

func JSON(e interface{}) []byte {
	contents, err := json.Marshal(e)
	if err != nil {
		log.Printf("Unable to convert erorr response to JSON: %v", err)
		return []byte{}
	}

	return contents
}
