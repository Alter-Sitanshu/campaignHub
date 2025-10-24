package cache

import (
	"context"
	"time"
)

// ==================================
// Submission Earnings Operations
// ==================================

func (s *Service) SetSubmissionEarnings(ctx context.Context, submissionID string, amount float64) error {
	key := SubmissionEarningsKey(submissionID)
	return s.Set(ctx, key, amount, TTLEarnings)
}

func (s *Service) GetSubmissionEarnings(ctx context.Context, submissionID string) (float64, error) {
	key := SubmissionEarningsKey(submissionID)
	return s.GetFloat(ctx, key)
}

func (s *Service) IncrementSubmissionEarnings(ctx context.Context, submissionID string, delta float64) error {
	key := SubmissionEarningsKey(submissionID)
	return s.IncrByFloat(ctx, key, delta)
}

func (s *Service) InvalidateSubmissionEarnings(ctx context.Context, submissionID string) error {
	key := SubmissionEarningsKey(submissionID)
	return s.Delete(ctx, key)
}

// ==================================
// Submission Status Operations
// ==================================

func (s *Service) SetSubmissionStatus(ctx context.Context, submissionID string, status int) error {
	key := SubmissionStatusKey(submissionID)
	return s.Set(ctx, key, status, 30*time.Minute)
}

func (s *Service) GetSubmissionStatus(ctx context.Context, submissionID string) (int, error) {
	key := SubmissionStatusKey(submissionID)
	return s.GetInt(ctx, key)
}

func (s *Service) InvalidateSubmissionStatus(ctx context.Context, submissionID string) error {
	key := SubmissionStatusKey(submissionID)
	return s.Delete(ctx, key)
}

// ==================================
// Creator Submissions List
// ==================================

func (s *Service) SetCreatorSubmissions(ctx context.Context, creatorID string, submissionIDs []string) error {
	key := CreatorSubmissionsKey(creatorID)

	// Clear existing set
	s.Delete(ctx, key)

	if len(submissionIDs) == 0 {
		return nil
	}

	// Add all submissions
	members := make([]interface{}, len(submissionIDs))
	for i, id := range submissionIDs {
		members[i] = id
	}

	if err := s.SAdd(ctx, key, members...); err != nil {
		return err
	}

	// Set expiration
	return s.client.Expire(ctx, key, 15*time.Minute).Err()
}

func (s *Service) GetCreatorSubmissions(ctx context.Context, creatorID string) ([]string, error) {
	key := CreatorSubmissionsKey(creatorID)
	return s.SMembers(ctx, key)
}

func (s *Service) InvalidateCreatorSubmissions(ctx context.Context, creatorID string) error {
	key := CreatorSubmissionsKey(creatorID)
	return s.Delete(ctx, key)
}

func (s *Service) InvalidateOneCreatorSubmissions(ctx context.Context, creatorID, sub_id string) error {
	key := CreatorSubmissionsKey(creatorID)
	return s.client.SRem(ctx, key, []string{sub_id}).Err()
}
