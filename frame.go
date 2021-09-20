package frame

import (
	"log"
	"net/http"
	"os"
	"time"

	gfcache "github.com/greatfocus/gf-cache/cache"
	gfcron "github.com/greatfocus/gf-cron"
	"github.com/greatfocus/gf-sframe/config"
	"github.com/greatfocus/gf-sframe/database"
	"github.com/greatfocus/gf-sframe/server"
	"github.com/joho/godotenv"
)

// Frame struct
type Frame struct {
	env    string
	Server *server.Meta
}

// NewFrame get new instance of frame
func NewFrame() *Frame {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}
	env := os.Getenv("ENV")
	service := os.Getenv("SERVICE")
	valtURL := os.Getenv("VALT_URL")
	valtUser := os.Getenv("VALT_USER")
	valtPass := os.Getenv("VALT_PASS")

	// prepare impl config
	impl := config.Impl{
		VaultURL:  valtURL,
		VaultUser: valtUser,
		VaultPass: valtPass,
		Service:   service,
		Env:       env,
	}
	var f = &Frame{env: impl.Env}
	f.Server = f.init(&impl)
	return f
}

// Init provides a way to initialize the frame
func (f *Frame) init(impl *config.Impl) *server.Meta {

	// read the config file and prepare object
	config := f.initConfig(impl)

	// initCron creates instance of cron
	cron := f.initCron()

	// initDB create database connection
	db := f.initDB(config, impl)

	// initCache creates instance of cache
	cache := f.initCache(config.Cache.DefaultExpiration, config.Cache.CleanupInterval)

	// initCron creates instance of cron
	jwt := f.initJWT(config)

	return &server.Meta{
		Env:    impl.Env,
		Config: config,
		Cron:   cron,
		Cache:  cache,
		DB:     db,
		JWT:    jwt,
	}
}

// Start spins up the service
func (f *Frame) Start(mux *http.ServeMux) {
	f.Server.Mux = mux
	f.Server.Start()
}

// initConfig read the configuration file
func (f *Frame) initConfig(impl *config.Impl) *config.Config {
	var config = impl.GetConfig()
	return &config
}

// initCron creates instance of cron
func (f *Frame) initCron() *gfcron.Cron {
	return gfcron.New()
}

func (f *Frame) initCache(defaultExpiration, cleanupInterval int64) *gfcache.Cache {
	// Create a cache with a default expiration time of 5 minutes, and which
	// purges expired items every 10 minutes
	return gfcache.New(time.Duration(defaultExpiration), time.Duration(cleanupInterval))
}

// initDB read the configuration file
func (f *Frame) initDB(config *config.Config, impl *config.Impl) *database.Conn {
	// create database connection
	var db = database.Conn{}
	db.Init(config, impl)
	return &db
}

// initJWT creates instance of auth
func (f *Frame) initJWT(config *config.Config) *server.JWT {
	var jwt = server.JWT{}
	jwt.Init(config)
	return &jwt
}
