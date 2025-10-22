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

func NewClient(addr, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

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
