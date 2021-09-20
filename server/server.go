package server

import (
	"log"
	"net/http"
	"time"

	gfcron "github.com/greatfocus/gf-cron"
	"github.com/greatfocus/gf-sframe/config"
	"github.com/greatfocus/gf-sframe/database"
)

// HandlerFunc custom server handler
type HandlerFunc func(http.ResponseWriter, *http.Request)

// Meta struct
type Meta struct {
	Env    string
	Mux    *http.ServeMux
	Config *config.Config
	DB     *database.Conn
	Cron   *gfcron.Cron
	JWT    *JWT
}

// Start the server
func (m *Meta) Start() {
	// setUploadPath creates an upload path
	m.setUploadPath()

	// serve creates server instance
	m.serve()
}

// setUploadPath creates an upload path
func (m *Meta) setUploadPath() {
	if m.Config.Server.UploadPath != "" {
		fs := http.FileServer(http.Dir(m.Config.Server.UploadPath + "/"))
		m.Mux.Handle("/file/", http.StripPrefix("/file/", fs))
	}
}

// serve creates server instance
func (m *Meta) serve() {
	addr := ":" + m.Config.Server.Port
	srv := &http.Server{
		Addr:           addr,
		ReadTimeout:    time.Duration(m.Config.Server.Timeout) * time.Second,
		WriteTimeout:   time.Duration(m.Config.Server.Timeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        m.Mux,
	}

	// create server connection
	log.Println("Listening to port HTTP", addr)
	log.Fatal(srv.ListenAndServe())
}
