package redis

import (
	"crypto/tls"
	"fmt"
	"time"

	entraid "github.com/redis/go-redis-entraid"
	"github.com/redis/go-redis/v9"
)

// Client represents a Redis client.
type Client = redis.Client

// Config holds configuration for connecting to a Redis instance.
type Config struct {
	Host string `env:"Host,required"`
	Port int    `env:"Port,required"`
}

// AspireConfig holds configuration for connecting to an Aspire Redis instance.
type AspireConfig struct {
	Host        string
	Port        int
	Password    string
	Connections int
}

// NewRedis creates a new redis.Client based on the provided configuration.
func NewRedis(cfg Config) (*Client, error) {
	provider, err := entraid.NewDefaultAzureCredentialsProvider(entraid.DefaultAzureCredentialsProviderOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to create Entra ID credentials provider: %w", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:                         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		TLSConfig:                    &tls.Config{MinVersion: tls.VersionTLS13},
		ReadTimeout:                  1 * time.Second,
		WriteTimeout:                 1 * time.Second,
		StreamingCredentialsProvider: provider,
	})

	return client, nil
}

// NewRedisForAspire creates a new redis.Client based on the provided Aspire configuration.
func NewRedisForAspire(cfg AspireConfig) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           0,
		MaxIdleConns: cfg.Connections,
		MinIdleConns: cfg.Connections,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	})

	return client, nil
}
