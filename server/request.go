package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/greatfocus/gf-sframe/crypt/pki"
)

// Request data
type Request struct {
	Payload string `json:"data,omitempty"`
}

// Success returns response as json
func GetPayload(w http.ResponseWriter, r *http.Request, privateKey string) (bool, []byte) {
	// check if the body is valid
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Error(w, http.StatusBadRequest, err, privateKey)
		return false, nil
	}

	// Get the Payload string and convert to requst struct
	req := Request{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		Error(w, http.StatusBadRequest, err, privateKey)
		return false, nil
	}

	// decrypt the string and return byte
	payload, _ := pki.Decrypt(privateKey, req.Payload)
	return true, []byte(payload)
}
