package redis

import "github.com/go-redis/redis/v8"

type Store struct {
	raw *redis.Client
}
