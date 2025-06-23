package configs

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// todo: tidy this up into an actual config struct
type Config struct {
	Domain string  // domain to bind the server to
	Port   string  // port to listen on
	DB     Storage // database configuration
	Auth   Auth    // authentication configuration
}
type Storage struct {
	T        string // type for future expansion, default: "mongo"
	MongoURI string // mongo connection uri
	Name     string // database name
}
type Auth struct {
	JWTSecret string // jwt secret for authentication
}

var (
	METRICS_delay           = 12 * time.Second
	DATAGEN_delay           = 36 * time.Second
	LOGGER_filter           = []string{"api:users"}
	LOGGER_enable_timestamp = false
	LOGGER_service_map      = map[string]string{
		"api":     "api",
		"users":   "users",
		"metrics": "metrics",
		"auth":    "auth",
	}
)

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Domain: os.Getenv("DOMAIN"),
		Port:   os.Getenv("PORT"),
		DB: Storage{
			// needs to be updated alongside the storage/ api
			T:        "mongo",
			Name:     "dps_http",
			MongoURI: os.Getenv("MONGO_URI"),
		},
		Auth: Auth{
			JWTSecret: os.Getenv("JWT_SECRET"),
		},
	}
	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	err := cfg.Validate()

	return cfg, err
}

func (cfg *Config) String() string {
	return fmt.Sprintf(`
	Config:
		Domain: %s, 
		Port: %s, 
		DB: %s, 
		Auth: %s`,
		cfg.Domain, cfg.Port, cfg.DB, cfg.Auth)
}

func (cfg *Config) Validate() error {
	// @TODO -- needs to issue a help like command
	if cfg.Domain == "" {
		return errors.New("domain is required")
	}
	if cfg.Port == "" {
		return errors.New("port is required")
	}
	if cfg.DB.T == "" {
		return errors.New("database type is required")
	}
	if cfg.DB.MongoURI == "" {
		return errors.New("mongo uri is required")
	}
	if cfg.DB.Name == "" {
		return errors.New("database name is required")
	}
	if cfg.Auth.JWTSecret == "" {
		return errors.New("jwt secret is required")
	}
	return nil
}
