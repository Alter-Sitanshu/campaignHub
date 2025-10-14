package cache

import (
	"context"
	"encoding/json"

	"github.com/Alter-Sitanshu/campaignHub/internals/platform"
	"github.com/redis/go-redis/v9"
)

type VideoMetadata struct {
	SubmissionID string `json:"submission_id"`
	VideoID      string `json:"video_id"`
	Platform     string `json:"platform"`

	// From YouTube API
	Title      string             `json:"title"`
	ViewCount  int                `json:"view_count"`
	LikeCount  int                `json:"like_count"`
	Thumbnail  platform.Thumbnail `json:"thumbnail"`
	UploadedAt string             `json:"uploaded_at"`
}

// ==================================
// Video Metadata Operations
// ==================================

func (s *Service) SetVideoMetadata(ctx context.Context, submissionID string, metadata *VideoMetadata) error {
	key := VideoMetadataKey(submissionID)
	return s.SetJSON(ctx, key, metadata, TTLVideoMetadata)
}

func (s *Service) GetVideoMetadata(ctx context.Context, submissionID string) (*VideoMetadata, error) {
	key := VideoMetadataKey(submissionID)
	var metadata VideoMetadata
	err := s.GetJSON(ctx, key, &metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (s *Service) InvalidateVideoMetadata(ctx context.Context, submissionID string) error {
	key := VideoMetadataKey(submissionID)
	return s.Delete(ctx, key)
}

// Batch get multiple video metadata
func (s *Service) GetMultipleVideoMetadata(ctx context.Context, submissionIDs []string) (map[string]*VideoMetadata, error) {
	if len(submissionIDs) == 0 {
		return make(map[string]*VideoMetadata), nil
	}

	pipe := s.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, id := range submissionIDs {
		key := VideoMetadataKey(id)
		cmds[id] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	results := make(map[string]*VideoMetadata)
	for id, cmd := range cmds {
		if data, err := cmd.Result(); err == nil {
			var metadata VideoMetadata
			if json.Unmarshal([]byte(data), &metadata) == nil {
				results[id] = &metadata
			}
		}
	}

	return results, nil
}

// Backward compatibility: Get just view count
func (s *Service) GetViewCountFromMetadata(ctx context.Context, submissionID string) (int, error) {
	metadata, err := s.GetVideoMetadata(ctx, submissionID)
	if err != nil {
		return 0, err
	}
	return metadata.ViewCount, nil
}
