package response

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, status int, message string, details interface{}) {
	resp := ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
		Details: details,
	}
	JSON(w, status, resp)
}
