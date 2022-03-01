package server

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	gfbus "github.com/greatfocus/gf-bus"
	gfcron "github.com/greatfocus/gf-cron"
	"github.com/greatfocus/gf-sframe/broker"
	"github.com/greatfocus/gf-sframe/cache"
	"github.com/greatfocus/gf-sframe/database"
	"github.com/greatfocus/gf-sframe/logger"
)

// HandlerFunc custom server handler
type HandlerFunc func(http.ResponseWriter, *http.Request)

// Meta struct
type Meta struct {
	Env        string
	Mux        *http.ServeMux
	DB         *database.Conn
	Cache      *cache.Cache
	Cron       *gfcron.Cron
	JWT        *JWT
	Bus        *gfbus.Bus
	Broker     *broker.Conn
	Logger     *logger.Logger
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	timeout    uint64
}

// Response data
type Response struct {
	Result interface{} `json:"data,omitempty"`
}

// Start the server
func (m *Meta) Start() {
	// Generate encryption keys
	publicKey, privatekey := generatePKI()
	m.publicKey = publicKey
	m.privateKey = privatekey

	// setUploadPath creates an upload path
	m.setUploadPath()

	// serve creates server instance
	m.serve()
}

// setUploadPath creates an upload path
func (m *Meta) setUploadPath() {
	uploadPath := os.Getenv("UPLOAD_PATH")
	if uploadPath != "" {
		fs := http.FileServer(http.Dir(uploadPath + "/"))
		m.Mux.Handle("/file/", http.StripPrefix("/file/", fs))
	}
}

// serve creates server instance
func (m *Meta) serve() {
	timeout, err := strconv.ParseUint(os.Getenv("SERVER_TIMEOUT"), 0, 64)
	m.timeout = timeout
	if err != nil {
		log.Fatal(fmt.Println(err))
	}

	addr := ":" + os.Getenv("SERVER_PORT")
	srv := &http.Server{
		Addr:           addr,
		ReadTimeout:    time.Duration(timeout) * time.Second,
		WriteTimeout:   time.Duration(timeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        m.Mux,
	}

	// create server connection
	crt := os.Getenv("APP_PATH") + "server.crt"
	key := os.Getenv("ENV") + "server.key"

	// generate self-sing key
	err = GenerateSelfSignedCert(os.Getenv("SERVER_HOST"), crt, key)
	if err != nil {
		log.Fatal(fmt.Println(err))
	}

	m.Logger.InfoLogger.Println("Listening to port HTTP", addr)
	log.Fatal(srv.ListenAndServeTLS(crt, key))
}

// Success returns object as json
func (m *Meta) Success(w http.ResponseWriter, r *http.Request, data interface{}) {
	if data != nil {
		m.response(w, r, data, "success", *m.publicKey)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	m.response(w, r, nil, "success", *m.publicKey)
}

// Error returns error as json
func (m *Meta) Error(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		m.response(w, r, struct {
			Error string `json:"error"`
		}{Error: err.Error()}, "error", *m.publicKey)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	m.response(w, r, nil, "error", *m.publicKey)
}

// response returns payload
func (m *Meta) response(w http.ResponseWriter, r *http.Request, data interface{}, message string, publicKey rsa.PublicKey) {
	out, _ := json.Marshal(data)
	res := Response{
		Result: encrypt(string(out), publicKey),
	}
	_ = json.NewEncoder(w).Encode(res)
}

func (m *Meta) Decrypt(cipherText string) string {
	ct, _ := base64.StdEncoding.DecodeString(cipherText)
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rng, m.privateKey, ct, label)
	checkError(err)
	fmt.Println("Plaintext:", string(plaintext))
	return string(plaintext)
}

// generatePKI provides encryption keys
func generatePKI() (*rsa.PublicKey, *rsa.PrivateKey) {
	// generate key
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Cannot generate RSA key\n")
		os.Exit(1)
	}

	return &privatekey.PublicKey, privatekey
}

// encrypt payload
func encrypt(secretMessage string, key rsa.PublicKey) string {
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, &key, []byte(secretMessage), label)
	checkError(err)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

// checkError validate error
func checkError(e error) {
	if e != nil {
		log.Printf("Decryption failed : %s", e)
	}
}

// GenerateSelfSignedCert creates a self-signed certificate and key for the given host.
// Host may be an IP or a DNS name
// The certificate will be created with file mode 0644. The key will be created with file mode 0600.
// If the certificate or key files already exist, they will be overwritten.
// Any parent directories of the certPath or keyPath will be created as needed with file mode 0755.
func GenerateSelfSignedCert(host, certPath, keyPath string) error {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("%s@%d", host, time.Now().Unix()),
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, host)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	// Generate cert
	certBuffer := bytes.Buffer{}
	if err := pem.Encode(&certBuffer, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}

	// Generate key
	keyBuffer := bytes.Buffer{}
	if err := pem.Encode(&keyBuffer, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return err
	}

	// Write cert
	if err := os.MkdirAll(filepath.Dir(certPath), os.FileMode(0755)); err != nil {
		return err
	}
	if err := ioutil.WriteFile(certPath, certBuffer.Bytes(), os.FileMode(0644)); err != nil {
		return err
	}

	// Write key
	if err := os.MkdirAll(filepath.Dir(keyPath), os.FileMode(0755)); err != nil {
		return err
	}
	if err := ioutil.WriteFile(keyPath, keyBuffer.Bytes(), os.FileMode(0600)); err != nil {
		return err
	}

	return nil
}
