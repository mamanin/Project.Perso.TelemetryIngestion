package redis

import (
	"crypto/tls"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Client represents a Redis client.
type Client = redis.Client

// Config holds configuration for connecting to a Redis instance.
type Config struct {
	// TODO
}

// AspireConfig holds configuration for connecting to an Aspire Redis instance.
type AspireConfig struct {
	Host        string
	Port        int
	Password    string
	Connections int
}

// NewRedis creates a new redis.Client based on the provided configuration.
func NewRedis(_ Config) *Client {
	panic("not implemented")
}

// NewRedisForAspire creates a new redis.Client based on the provided Aspire configuration.
func NewRedisForAspire(cfg AspireConfig) *Client {
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

	return client
}
