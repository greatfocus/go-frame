package server

import (
	"encoding/json"
	"net/http"

	"github.com/greatfocus/gf-sframe/crypt/pki"
)

// Response data
type Response struct {
	Payload string `json:"data,omitempty"`
}

// Success returns response as json
func Success(w http.ResponseWriter, statusCode int, data interface{}, publicKey string) {
	payload, _ := pki.Encrypt(publicKey, data.(string))
	res := Response{Payload: payload}
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(res)
}

// Error returns error as json
func Error(w http.ResponseWriter, statusCode int, err error, publicKey string) {
	if err != nil {
		Success(w, statusCode, struct {
			Error string `json:"error"`
		}{Error: err.Error()}, publicKey)
		return
	}
	Success(w, http.StatusBadRequest, nil, publicKey)
}
