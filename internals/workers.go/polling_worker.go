// internal/workers/polling_worker.go
package workers

import (
	"context"
	"log"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals"
	"github.com/Alter-Sitanshu/campaignHub/internals/cache"
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/Alter-Sitanshu/campaignHub/internals/platform"
)

type PollingWorker struct {
	submissionRepo *db.SubmissionStore
	campaignRepo   *db.CampaignStore
	cache          *cache.Service
	platformClient *platform.Factory
	interval       time.Duration
	stopChan       chan struct{}
}

func NewPollingWorker(
	submissionRepo *db.SubmissionStore,
	campaignRepo *db.CampaignStore,
	cache *cache.Service,
	platformClient *platform.Factory,
	interval time.Duration,
) *PollingWorker {
	return &PollingWorker{
		submissionRepo: submissionRepo,
		campaignRepo:   campaignRepo,
		cache:          cache,
		platformClient: platformClient,
		interval:       interval,
		stopChan:       make(chan struct{}),
	}
}

func (w *PollingWorker) Start(ctx context.Context) {
	log.Println("Polling worker started")

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Run immediately
	w.poll(ctx)

	for {
		select {
		case <-ticker.C:
			w.poll(ctx)
		case <-w.stopChan:
			log.Println("Polling worker stopped")
			return
		case <-ctx.Done():
			log.Println("Polling worker context cancelled")
			return
		}
	}
}

func (w *PollingWorker) poll(ctx context.Context) {
	submissions, err := w.submissionRepo.GetSubmissionsForSync(ctx)
	if err != nil {
		log.Printf("Failed to fetch submissions: %v", err)
		return
	}

	log.Printf(" %d submissions...", len(submissions))

	for _, submission := range submissions {
		if err := w.syncSubmission(ctx, submission); err != nil {
			log.Printf("Error syncing submission %s: %v", submission.Id, err)
		}
	}
}

func (w *PollingWorker) syncSubmission(ctx context.Context, submission db.PollingSubmission) error {
	// Parse video URL
	parsed, err := platform.ParseVideoURL(submission.Url)
	if err != nil {
		return err
	}

	// Fetch metadata from platform
	response, err := w.platformClient.GetVideoDetails(ctx, parsed.Name, parsed.VideoID)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return err
	}
	metadata, ok := response.(*platform.VideoMetadata)
	if !ok {
		log.Printf("invalid response, video meta data packet invalid\n")
		return err
	}

	// Update cache immediately (real-time data)
	cacheMetadata := &cache.VideoMetadata{
		SubmissionID: submission.Id,
		VideoID:      metadata.VideoID,
		Platform:     metadata.Platform,
		Title:        metadata.Title,
		ViewCount:    metadata.ViewCount,
		LikeCount:    metadata.LikeCount,
		Thumbnail:    metadata.Thumbnails,
		UploadedAt:   metadata.UploadedAt,
	}
	w.cache.SetVideoMetadata(ctx, submission.Id, cacheMetadata)

	// Calculate changes
	viewsDelta := metadata.ViewCount - submission.Views

	// Only queue batch update if significant change
	if abs(viewsDelta) >= 10 {
		// Get campaign for CPM calculation
		campaign, err := w.campaignRepo.GetCampaign(ctx, submission.CampaignId)
		if err != nil {
			return err
		}

		// Calculate earnings
		earningsDelta := float64(viewsDelta) * campaign.CPM / 1000.0

		// Create batch update
		batchUpdate := &internals.BatchUpdate{
			SubmissionID:  submission.Id,
			OldViews:      submission.Views,
			NewViews:      metadata.ViewCount,
			ViewsDelta:    viewsDelta,
			VideoTitle:    metadata.Title,
			LikeCount:     metadata.LikeCount,
			ThumbnailURL:  metadata.Thumbnails.URL,
			EarningsDelta: earningsDelta,
			CampaignID:    submission.CampaignId,
			CreatorID:     submission.CreatorId,
			Timestamp:     time.Now().String(),
			Source:        "polling_worker",
		}

		// Queue for batch processing
		if err := w.cache.QueueBatchUpdate(ctx, batchUpdate); err != nil {
			log.Printf("Failed to queue batch update: %v", err)
		}
		// Adjust sync frequency based on video age
		w.adjustSyncFrequency(ctx, submission)

		// Also update cache for immediate read access
		w.cache.IncrementSubmissionEarnings(ctx, submission.Id, earningsDelta)
		w.cache.UpdateUserBalance(ctx, submission.CreatorId, earningsDelta)
		w.cache.DecrementCampaignBudget(ctx, submission.CampaignId, earningsDelta)
	}

	return nil
}

func (w *PollingWorker) Stop() {
	close(w.stopChan)
}

func (w *PollingWorker) adjustSyncFrequency(ctx context.Context, submission db.PollingSubmission) {
	// Calculate video age
	Created_At, _ := time.Parse(time.RFC3339, submission.CreatedAt)
	ageInDays := time.Since(Created_At).Hours() / 24

	var newFrequency int
	switch {
	case ageInDays < 1:
		newFrequency = 5 // Check every 5 minutes (new viral video!)
	case ageInDays < 7:
		newFrequency = 30 // Check every 30 minutes
	default:
		newFrequency = 120 // Check every 2 hours
	}

	// Update in database
	if newFrequency != submission.SyncFrequency {
		w.submissionRepo.UpdateSyncFrequency(ctx, submission.Id, newFrequency)
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
