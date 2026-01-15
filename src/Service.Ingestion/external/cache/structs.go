package cache

import "github.com/redis/go-redis/v9"

// ItemProcess holds information for processing a single item of any type T.
type ItemProcess[T any] struct {
	Key    string
	Item   *T
	GetCmd *redis.StringCmd
}
