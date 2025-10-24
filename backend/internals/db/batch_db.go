package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Alter-Sitanshu/campaignHub/internals"
)

type BatchRepository struct {
	db *sql.DB
}

func NewBatchRepository(db *sql.DB) *BatchRepository {
	return &BatchRepository{db: db}
}

// updates multiple submissions in a single transaction
func (r *BatchRepository) BatchUpdateSubmissions(ctx context.Context, updates []*internals.BatchUpdate) error {
	// nothing to update
	if len(updates) == 0 {
		return nil
	}

	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statement for all the updates
	stmt, err := tx.PrepareContext(ctx, `
        UPDATE submissions 
        SET 
            views = views + $1,
            earnings = earnings + $2,
            like_count = $3,
            video_title = COALESCE(NULLIF($4, ''), video_title),
            thumbnail_url = COALESCE(NULLIF($5, ''), thumbnail_url),
            metadata_last_updated = NOW(),
            last_synced_at = NOW()
        WHERE id = $6
            AND views + $1 >= 0
    `)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute updates
	successCount := 0
	for _, update := range updates {
		_, err := stmt.ExecContext(ctx,
			update.ViewsDelta,
			update.EarningsDelta,
			update.LikeCount,
			update.VideoTitle,
			update.ThumbnailURL,
			update.SubmissionID,
		)

		if err != nil {
			log.Printf("Failed to update submission %s: %v", update.SubmissionID, err)
			continue // Skip this update, continue with others
		}

		successCount++
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Batch updated %d/%d submissions", successCount, len(updates))
	return nil
}

// updates multiple creator balances
func (r *BatchRepository) BatchUpdateCreatorBalances(ctx context.Context, updates map[string]float64) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
        UPDATE users 
        SET balance = balance + $1 
        WHERE id = $2
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for creatorID, delta := range updates {
		if _, err := stmt.ExecContext(ctx, delta, creatorID); err != nil {
			log.Printf("Failed to update creator %s balance: %v", creatorID, err)
		}
	}

	return tx.Commit()
}

// BatchUpdateCampaignBudgets updates multiple campaign budgets
func (r *BatchRepository) BatchUpdateCampaignBudgets(ctx context.Context, updates map[string]float64) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
        UPDATE campaigns 
        SET budget = budget - $1 
        WHERE id = $2
            AND budget - $1 >= 0
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for campaignID, amount := range updates {
		if _, err := stmt.ExecContext(ctx, amount, campaignID); err != nil {
			log.Printf("Failed to update campaign %s budget: %v", campaignID, err)
		}
	}

	return tx.Commit()
}
