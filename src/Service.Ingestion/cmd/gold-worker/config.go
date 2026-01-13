package main

const (
	Key = "Value"
)

// Config holds the entire configuration for the application.
type Config struct {
}

// LoadConfig loads configuration from environment variables and azure app configuration.
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	return cfg, nil
}
