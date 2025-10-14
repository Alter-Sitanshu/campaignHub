package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals"
)

// QueueBatchUpdate adds an update to the batch queue
func (s *Service) QueueBatchUpdate(ctx context.Context, update *internals.BatchUpdate) error {
	// Set timestamp if not set
	if update.Timestamp == "" {
		update.Timestamp = time.Now().String()
	}

	// Serialize to JSON
	data, err := json.Marshal(update)
	if err != nil {
		return err
	}

	// Add to Redis list (FIFO queue)
	return s.client.LPush(ctx, batchQueueKey, data).Err()
}

// GetPendingBatchUpdates retrieves pending updates
func (s *Service) GetPendingBatchUpdates(ctx context.Context, limit int64) ([]*internals.BatchUpdate, error) {
	// Get items from queue (FIFO - oldest first)
	results, err := s.client.LRange(ctx, batchQueueKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return []*internals.BatchUpdate{}, nil
	}

	// Remove retrieved items from queue
	if err := s.client.LTrim(ctx, batchQueueKey, int64(len(results)), -1).Err(); err != nil {
		return nil, err
	}

	// Deserialize
	updates := make([]*internals.BatchUpdate, 0, len(results))
	for _, data := range results {
		var update internals.BatchUpdate
		if err := json.Unmarshal([]byte(data), &update); err != nil {
			continue // Skip malformed updates
		}
		updates = append(updates, &update)
	}

	return updates, nil
}

// GetBatchQueueLength returns number of pending updates
func (s *Service) GetBatchQueueLength(ctx context.Context) (int64, error) {
	return s.client.LLen(ctx, batchQueueKey).Result()
}

// ClearBatchQueue removes all pending updates (emergency use)
func (s *Service) ClearBatchQueue(ctx context.Context) error {
	return s.client.Del(ctx, batchQueueKey).Err()
}
