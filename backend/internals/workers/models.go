package workers

import (
	"context"
	"sync"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals"
	"github.com/Alter-Sitanshu/campaignHub/internals/cache"
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/Alter-Sitanshu/campaignHub/services/platform"
)

type GroupedUpdates struct {
	Submissions     []*internals.BatchUpdate
	CreatorBalances map[string]float64 // creatorID -> delta
	CampaignBudgets map[string]float64 // campaignID -> delta
}
type BatchWorker struct {
	cache     *cache.Service
	repo      *db.Store
	interval  time.Duration
	batchSize int
	stopOnce  sync.Once // Guard against multiple close attempts concurrently
	stopChan  chan struct{}
}

type PollingWorker struct {
	repo           *db.Store
	cache          *cache.Service
	platformClient *platform.Factory
	interval       time.Duration
	stopOnce       sync.Once
	stopChan       chan struct{}
}

type AppWorkers struct {
	Batch  *BatchWorker
	Poll   *PollingWorker
	cancel context.CancelFunc
}

func NewAppWorker(
	cache *cache.Service, repo *db.Store,
	factory *platform.Factory,
	BatchInterval, PollInterval time.Duration,
) *AppWorkers {
	return &AppWorkers{
		Batch: NewBatchWorker(
			cache,
			repo,
			BatchInterval,
		),
		Poll: NewPollingWorker(
			repo,
			cache,
			factory,
			PollInterval,
		),
	}
}

func (aw *AppWorkers) SetCancel(cancel context.CancelFunc) {
	aw.cancel = cancel
}
