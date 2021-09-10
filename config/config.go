package config

import (
	"errors"
	"fmt"
	"log"
)

// Config struct
type Config struct {
	Env          string       `json:"env"`
	Impl         string       `json:"impl"`
	Server       Server       `json:"server"`
	Database     Database     `json:"database"`
	Cache        Cache        `json:"cache"`
	Integrations Integrations `json:"integrations"`
	Services     Services     `json:"services"`
}

// Server struct config
type Server struct {
	Port           string     `json:"port"`
	Timeout        int64      `json:"timeout"`
	UploadPath     string     `json:"uploadPath"`
	AllowedOrigins []string   `json:"allowedOrigins"`
	AllowedIPs     []string   `json:"allowedIPs"`
	Secure         Secure     `json:"secure"`
	JWT            JWT        `json:"jwt"`
	Workers        int64      `json:"workers"`
	Encryption     Encryption `json:"encryption"`
}

// encryption struct config
type Encryption struct {
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

// Secure struct config
type Secure struct {
	Key      string `json:"key"`
	Cert     string `json:"cert"`
	User     string `json:"user"`
	Password string `json:"password"`
	SslMode  bool   `json:"sslmode"`
}

// JWT struct config
type JWT struct {
	Secret     string `json:"secret"`
	Authorized bool   `json:"authorized"`
	Minutes    int64  `json:"minutes"`
}

// Cache struct config
type Cache struct {
	DefaultExpiration int64 `json:"defaultExpiration"`
	CleanupInterval   int64 `json:"cleanupInterval"`
}

// Database struct config
type Database struct {
	Master DatabaseType `json:"master"`
	Slave  DatabaseType `json:"slave"`
}

// DatabaseType struct config
type DatabaseType struct {
	Host          string `json:"host"`
	Port          string `json:"port"`
	Database      string `json:"database"`
	User          string `json:"user"`
	Password      string `json:"password"`
	Secure        Secure `json:"secure"`
	Timeout       int64  `json:"timeOut"`
	MaxLifetime   int64  `json:"maxLifetime"`
	MaxIdleConns  int64  `json:"maxIdleConns"`
	MaxOpenConns  int64  `json:"maxOpenConns"`
	ExecuteSchema bool   `json:"executeSchema"`
}

// Integrations struct config
type Integrations struct {
	Email   Email   `json:"email"`
	SMS     SMS     `json:"sms"`
	Contact Contact `json:"contact"`
	Payment Payment `json:"payment"`
}

// Email struct config
type Email struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	From     string `json:"from"`
}

// SMS struct config
type SMS struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// Contact struct config
type Contact struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// Payment struct config
type Payment struct {
	Mpesa Mpesa `json:"mpesa"`
}

// Mpesa struct config
type Mpesa struct {
	Env         int    `json:"env"`
	AppKey      string `json:"appKey"`
	AppSecret   string `json:"appSecret"`
	CallBackURL string `json:"callBackUrl"`
	PassKey     string `json:"passKey"`
	ShortCode   string `json:"shortCode"`
}

// Services struct config
type Services struct {
	User Service `json:"user"`
}

// Service struct config
type Service struct {
	Host      string      `json:"host"`
	Port      string      `json:"port"`
	Client    Client      `json:"client"`
	Operation []Operation `json:"operation"`
}

// Client struct config
type Client struct {
	Email    string `json:"email"`
	ClientID string `json:"clientId"`
	Secret   string `json:"secret"`
}

// Operation struct
type Operation struct {
	TemplateID int64  `json:"templateId"`
	ChannelID  int64  `json:"channelId"`
	Operation  string `json:"operation"`
	URI        string `json:"uri"`
}

// Validate checks dependencies in the settings
func (c *Config) validate() {
	validateDefault(c)
}

func validateDefault(c *Config) {
	var err error
	if c.Impl == "" {
		err = errors.New("please configure impl in setting file")
		log.Fatal(fmt.Println(err))
	}
	if c.Env == "" {
		err = errors.New("please configure env in setting file")
		log.Fatal(fmt.Println(err))
	}
	if c.Server.Port == "" {
		err = errors.New("please configure server port in setting file")
		log.Fatal(fmt.Println(err))
	}
	if c.Server.Timeout == 0 {
		err = errors.New("please configure Timeout in setting file")
		log.Fatal(fmt.Println(err))
	}

	if c.Server.JWT.Authorized {
		if c.Server.JWT.Secret == "" {
			err = errors.New("please configure JWT in settings")
			log.Fatal(fmt.Println(err))
		}

		if c.Server.JWT.Minutes == 0 {
			err = errors.New("please configure JWT in settings")
			log.Fatal(fmt.Println(err))
		}
	}

	// validate database
	validateCache(c)

	// validate database
	validateDatabase(c)

	// validate integrations
	validateIntegrations(c)

	// validate service
	validateService(c.Services.User, "User Service")
}

// validateEmail checks database configuration
func validateCache(c *Config) {
	var err error
	if c.Cache.CleanupInterval == 0 {
		err = errors.New("please configure Cache interval")
		log.Fatal(fmt.Println(err))
	}
	if c.Cache.DefaultExpiration == 0 {
		err = errors.New("please configure Cache expiration")
		log.Fatal(fmt.Println(err))
	}
}

// validateIntegrations checks integration configuration
func validateIntegrations(c *Config) {
	// validate email
	validateEmail(c)

	// validate sms
	validateSMS(c)

	// validate contact
	validateContact(c)

	// validate payment
	validatePayment(c)
}

// ValidateDatabase checks database configuration
func validateDatabase(c *Config) {
	var err error
	if c.Database.Master.Host == "" || c.Database.Slave.Host == "" {
		err = errors.New("please configure database host")
		log.Fatal(fmt.Println(err))
	}
	if c.Database.Master.Port == "" || c.Database.Slave.Port == "" {
		err = errors.New("please configure database port")
		log.Fatal(fmt.Println(err))
	}
	if c.Database.Master.Database == "" || c.Database.Slave.Database == "" {
		err = errors.New("please configure database name")
		log.Fatal(fmt.Println(err))
	}
	if c.Database.Master.User == "" || c.Database.Slave.User == "" {
		err = errors.New("please configure database user")
		log.Fatal(fmt.Println(err))
	}
	if c.Database.Master.Password == "" || c.Database.Slave.Password == "" {
		err = errors.New("please configure database user")
		log.Fatal(fmt.Println(err))
	}
	if c.Env == "prod" {
		if !c.Database.Master.Secure.SslMode || !c.Database.Slave.Secure.SslMode {
			err = errors.New("please configure secure ssl mode")
			log.Fatal(fmt.Println(err))
		}
	}

	if c.Database.Master.MaxOpenConns == 0 || c.Database.Slave.MaxOpenConns == 0 {
		err = errors.New("please configure database MaxOpenConns")
		log.Fatal(fmt.Println(err))
	}
	if c.Database.Master.MaxIdleConns == 0 || c.Database.Slave.MaxOpenConns == 0 {
		err = errors.New("please configure database MaxIdleConns")
		log.Fatal(fmt.Println(err))
	}
	if c.Database.Master.MaxLifetime == 0 || c.Database.Slave.MaxOpenConns == 0 {
		err = errors.New("please configure database MaxLifetime")
		log.Fatal(fmt.Println(err))
	}
}

// validateEmail checks database configuration
func validateEmail(c *Config) {
	var err error
	if c.Integrations.Email.Host == "" {
		err = errors.New("please configure Email host")
		log.Fatal(fmt.Println(err))
	}
	if c.Integrations.Email.Port == "" {
		err = errors.New("please configure Email port")
		log.Fatal(fmt.Println(err))
	}
	if c.Integrations.Email.User == "" {
		err = errors.New("please configure Email user")
		log.Fatal(fmt.Println(err))
	}
	if c.Integrations.Email.Password == "" {
		err = errors.New("please configure Email password")
		log.Fatal(fmt.Println(err))
	}
	if c.Integrations.Email.From == "" {
		err = errors.New("please configure Email from")
		log.Fatal(fmt.Println(err))
	}
}

// validateEmail checks database configuration
func validateSMS(c *Config) {
	var err error
	if c.Integrations.Email.Host == "" {
		err = errors.New("please configure SMS host")
		log.Fatal(fmt.Println(err))
	}
	if c.Integrations.Email.Port == "" {
		err = errors.New("please configure SMS port")
		log.Fatal(fmt.Println(err))
	}
	if c.Integrations.Email.User == "" {
		err = errors.New("please configure SMS user")
		log.Fatal(fmt.Println(err))
	}
	if c.Integrations.Email.Password == "" {
		err = errors.New("please configure SMS password")
		log.Fatal(fmt.Println(err))
	}
}

// validateContact checks database configuration
func validateContact(c *Config) {
	var err error
	if c.Integrations.Contact.Email == "" {
		err = errors.New("please configure Contact email")
		log.Fatal(fmt.Println(err))
	}
	if c.Integrations.Contact.Phone == "" {
		err = errors.New("please configure Contact phone")
		log.Fatal(fmt.Println(err))
	}
}

// validatePayment checks payment configuration
func validatePayment(c *Config) {
	var err error
	if c.Integrations.Payment.Mpesa.AppKey == "" {
		err = errors.New("please configure Mpesa Key")
		log.Fatal(fmt.Println(err))
	}

	if c.Integrations.Payment.Mpesa.AppSecret == "" {
		err = errors.New("please configure Mpesa Secret")
		log.Fatal(fmt.Println(err))
	}

	if c.Integrations.Payment.Mpesa.CallBackURL == "" {
		err = errors.New("please configure Mpesa Call back URL")
		log.Fatal(fmt.Println(err))
	}

	if c.Integrations.Payment.Mpesa.PassKey == "" {
		err = errors.New("please configure Mpesa PassKey")
		log.Fatal(fmt.Println(err))
	}

	if c.Integrations.Payment.Mpesa.ShortCode == "" {
		err = errors.New("please configure Mpesa ShortCode")
		log.Fatal(fmt.Println(err))
	}
}

// validateNotify checks database configuration
func validateService(s Service, name string) {
	var err error
	if s.Host == "" {
		err = errors.New("please configure " + name + " host")
		log.Fatal(fmt.Println(err))
	}
	if s.Port == "" {
		err = errors.New("please configure " + name + " port")
		log.Fatal(fmt.Println(err))
	}
	if s.Client.Email == "" {
		err = errors.New("please configure " + name + " Email")
		log.Fatal(fmt.Println(err))
	}
	if s.Client.ClientID == "" {
		err = errors.New("please configure " + name + " Client ID")
		log.Fatal(fmt.Println(err))
	}
	if s.Client.Secret == "" {
		err = errors.New("please configure " + name + " Secret")
		log.Fatal(fmt.Println(err))
	}
	if s.Operation == nil {
		err = errors.New("please configure " + name + " Operation")
		log.Fatal(fmt.Println(err))
	}
}
