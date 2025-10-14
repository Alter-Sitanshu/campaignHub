package workers

import (
	"context"
	"log"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals"
	"github.com/Alter-Sitanshu/campaignHub/internals/cache"
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
)

type GroupedUpdates struct {
	Submissions     []*internals.BatchUpdate
	CreatorBalances map[string]float64 // creatorID -> delta
	CampaignBudgets map[string]float64 // campaignID -> delta
}
type BatchWorker struct {
	cache     *cache.Service
	batchRepo *db.BatchRepository
	interval  time.Duration
	batchSize int
	stopChan  chan struct{}
}

func NewBatchWorker(
	cache *cache.Service,
	batchRepo *db.BatchRepository,
	interval time.Duration,
) *BatchWorker {
	return &BatchWorker{
		cache:     cache,
		batchRepo: batchRepo,
		interval:  interval,
		batchSize: 100, // Process 100 updates at a time
		stopChan:  make(chan struct{}),
	}
}

func (w *BatchWorker) Start(ctx context.Context) {
	log.Println("Batch worker started")

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Run immediately on start
	w.processBatch(ctx)

	for {
		select {
		case <-ticker.C:
			w.processBatch(ctx)
		case <-w.stopChan:
			log.Println("Batch worker stopped")
			return
		case <-ctx.Done():
			log.Println("Batch worker context cancelled")
			return
		}
	}
}

func (w *BatchWorker) processBatch(ctx context.Context) {
	// Get queue length
	queueLength, err := w.cache.GetBatchQueueLength(ctx)
	if err != nil {
		log.Printf("Failed to get queue length: %v", err)
		return
	}

	if queueLength == 0 {
		log.Println("No pending batch updates")
		return
	}

	log.Printf("Processing %d pending updates...", queueLength)

	// Retrieve pending updates
	updates, err := w.cache.GetPendingBatchUpdates(ctx, int64(w.batchSize))
	if err != nil {
		log.Printf("Failed to fetch batch updates: %v", err)
		return
	}

	if len(updates) == 0 {
		return
	}

	// Group updates by type
	groupedUpdates := w.groupAndMergeUpdates(updates)

	// Update submissions
	if err := w.batchRepo.BatchUpdateSubmissions(ctx, groupedUpdates.Submissions); err != nil {
		log.Printf("Batch submission update failed: %v", err)
	}

	// Update creator balances
	if err := w.batchRepo.BatchUpdateCreatorBalances(ctx, groupedUpdates.CreatorBalances); err != nil {
		log.Printf("Batch creator balance update failed: %v", err)
	}

	// Update campaign budgets
	if err := w.batchRepo.BatchUpdateCampaignBudgets(ctx, groupedUpdates.CampaignBudgets); err != nil {
		log.Printf("Batch campaign budget update failed: %v", err)
	}

	log.Printf("Batch processing complete")
}

// merges multiple updates for the same submission
// the motive of this function is to reduce the number of changes/writes
// that we need to make to the db
func (w *BatchWorker) groupAndMergeUpdates(updates []*internals.BatchUpdate) *GroupedUpdates {
	grouped := &GroupedUpdates{
		CreatorBalances: make(map[string]float64),
		CampaignBudgets: make(map[string]float64),
	}

	// map to merge duplicate submission updates
	submissionMap := make(map[string]*internals.BatchUpdate)

	for _, update := range updates {
		// Merge submission updates
		if existing, exists := submissionMap[update.SubmissionID]; exists {
			// Merge deltas
			existing.ViewsDelta += update.ViewsDelta
			existing.EarningsDelta += update.EarningsDelta
			existing.NewViews = update.NewViews   // Use latest
			existing.LikeCount = update.LikeCount // Use latest

			// Update metadata if provided
			if update.VideoTitle != "" {
				existing.VideoTitle = update.VideoTitle
			}
			if update.ThumbnailURL != "" {
				existing.ThumbnailURL = update.ThumbnailURL
			}
		} else {
			submissionMap[update.SubmissionID] = update
		}

		// Aggregate creator balances
		if update.CreatorID != "" {
			grouped.CreatorBalances[update.CreatorID] += update.EarningsDelta
		}

		// Aggregate campaign budgets
		if update.CampaignID != "" {
			grouped.CampaignBudgets[update.CampaignID] += update.EarningsDelta
		}
	}

	// Convert map to slice
	for _, update := range submissionMap {
		grouped.Submissions = append(grouped.Submissions, update)
	}

	return grouped
}

func (w *BatchWorker) Stop() {
	close(w.stopChan)
}
