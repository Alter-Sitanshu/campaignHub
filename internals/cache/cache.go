package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// Default TTLs
const (
	TTLEarnings      = 10 * time.Minute
	TTLCampaign      = 30 * time.Minute
	TTLUserProfile   = 1 * time.Hour
	TTLBalance       = 5 * time.Minute
	TTLActiveCamps   = 5 * time.Minute
	TTLVideoMetadata = 10 * time.Minute
)

type Service struct {
	client *redis.Client
}

// Constructor for the App Cache service
func NewService(client *redis.Client) *Service {
	return &Service{
		client: client,
	}
}

// ==================================
// Generic Operations
// ==================================

func (s *Service) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

func (s *Service) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

func (s *Service) GetInt(ctx context.Context, key string) (int, error) {
	return s.client.Get(ctx, key).Int()
}

func (s *Service) GetFloat(ctx context.Context, key string) (float64, error) {
	return s.client.Get(ctx, key).Float64()
}

func (s *Service) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return s.client.Del(ctx, keys...).Err()
}

func (s *Service) Exists(ctx context.Context, key string) (bool, error) {
	result, err := s.client.Exists(ctx, key).Result()
	return result > 0, err
}

// Increment operations
func (s *Service) Incr(ctx context.Context, key string) error {
	return s.client.Incr(ctx, key).Err()
}

func (s *Service) IncrBy(ctx context.Context, key string, value int64) error {
	return s.client.IncrBy(ctx, key, value).Err()
}

func (s *Service) IncrByFloat(ctx context.Context, key string, value float64) error {
	return s.client.IncrByFloat(ctx, key, value).Err()
}

func (s *Service) Decr(ctx context.Context, key string) error {
	return s.client.Decr(ctx, key).Err()
}

func (s *Service) DecrBy(ctx context.Context, key string, value int64) error {
	return s.client.DecrBy(ctx, key, value).Err()
}

// Set operations (for lists like active campaigns)
func (s *Service) SAdd(ctx context.Context, key string, members ...any) error {
	return s.client.SAdd(ctx, key, members...).Err()
}

func (s *Service) SRem(ctx context.Context, key string, members ...any) error {
	return s.client.SRem(ctx, key, members...).Err()
}

func (s *Service) SMembers(ctx context.Context, key string) ([]string, error) {
	return s.client.SMembers(ctx, key).Result()
}

// JSON operations
func (s *Service) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, ttl).Err()
}

func (s *Service) GetJSON(ctx context.Context, key string, dest any) error {
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// Invalidate is just an alias for Delete
func (s *Service) Invalidate(ctx context.Context, keys ...string) error {
	return s.Delete(ctx, keys...)
}
