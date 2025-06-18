package configs

import (
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
	METRICS_delay           = 300 * time.Second
	LOGGER_filter           = []string{"api:users"}
	LOGGER_enable_timestamp = false
	LOGGER_service_map      = map[string]string{
		"api":     "api",
		"users":   "users",
		"metrics": "metrics",
		"auth":    "auth",
	}
)

var CONFIG = Config{
	Domain: "127.0.0.1",
	Port:   ":8080",
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
