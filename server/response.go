package server

import (
	"encoding/json"
	"net/http"
)

// response returns payload
func response(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

// Error returns error as json
func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	if data != nil {
		response(w, statusCode, data)
		return
	}
	response(w, http.StatusBadRequest, nil)
}

// Error returns error as json
func Error(w http.ResponseWriter, statusCode int, err error) {
	if err != nil {
		response(w, statusCode, struct {
			Error string `json:"error"`
		}{Error: err.Error()})
		return
	}
	response(w, http.StatusBadRequest, nil)
}
