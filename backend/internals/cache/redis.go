package cache

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	RedisTimeout = time.Second * 5
)

func NewClient(url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("Redis Cache Logged in")
	return rdb, nil
}

func (s *Service) Close() error {
	return s.client.Close()
}
