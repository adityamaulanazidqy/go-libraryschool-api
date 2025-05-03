package helpers

import (
	"encoding/json"
	"net/http"
)

type ApiResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ApiResponseAuthorization struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Token   string      `json:"token"`
}

func SendJson(w http.ResponseWriter, code int, response ApiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func SendJsonAuthorization(w http.ResponseWriter, code int, response ApiResponseAuthorization) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}
