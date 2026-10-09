package config

import "os"

// Config contains the runtime values needed by the API bootstrap. Feature
// modules will extend this without reading process environment directly.
type Config struct {
	Addr        string
	Environment string
	AppURL      string
}

func Load() Config {
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	environment := os.Getenv("APP_ENV")
	if environment == "" {
		environment = "development"
	}
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:8180"
	}
	return Config{Addr: addr, Environment: environment, AppURL: appURL}
}
