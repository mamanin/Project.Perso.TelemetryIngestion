package cache

import (
	"crypto/tls"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Redis = redis.Client

// Config holds configuration for Redis cache.
type Config struct {
	Host        string
	Port        int
	Password    string
	Connections int
}

// NewRedis creates a new Redis client based on the provided configuration.
func NewRedis(cfg *Config) *Redis {
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
