package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct{}

func New() (*Handler, error) {
	return &Handler{}, nil
}

func (h *Handler) HelloWorld(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	w.Write([]byte(fmt.Sprintf("Hello, %v", name)))
}
