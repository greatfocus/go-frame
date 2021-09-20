package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Impl struct
type Impl struct {
	VaultURL  string `json:"-"`
	VaultUser string `json:"vaultUser"`
	VaultPass string `json:"vaultPass"`
	Service   string `json:"service"`
	Env       string `json:"env"`
}

// GetConfig method gets configf from impl
func (i *Impl) GetConfig() Config {
	request := Impl{
		Env:       i.Env,
		Service:   i.Service,
		VaultUser: i.VaultUser,
		VaultPass: i.VaultPass,
	}
	reqBody, err := json.Marshal(request)
	if err != nil {
		log.Fatal(fmt.Println("Failed to get Impl config", err))
	}
	if err != nil {
		log.Fatal(fmt.Println("Failed to get Impl config", err))
	}

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100
	client := http.Client{
		Timeout:   10 * time.Second,
		Transport: t,
	}

	// make API call to impl
	resp, err := client.Post(i.VaultURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		log.Fatal(fmt.Println("Failed to get Impl config"))
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(fmt.Println("Failed to get Impl config"))
	}

	// marshal te response
	var config Config
	err = json.Unmarshal(body, &config)
	if err != nil {
		log.Fatal(fmt.Println("Failed to get Impl config"))
	}

	// verify response
	if config.Impl == "" {
		log.Fatal(fmt.Println("Failed to get Impl config"))
	}

	// validate
	config.validate()

	return config
}
