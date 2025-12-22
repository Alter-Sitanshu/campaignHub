package internals

import (
	"crypto/rand"
	"log"
	"math/big"
)

type BatchUpdate struct {
	SubmissionID string `json:"submission_id"`

	// View tracking
	OldViews   int `json:"old_views"`
	NewViews   int `json:"new_views"`
	ViewsDelta int `json:"views_delta"`

	// Metadata updates
	VideoTitle string `json:"video_title,omitempty"`
	LikeCount  int    `json:"like_count"`

	// Earnings
	EarningsDelta float64 `json:"earnings_delta"`

	// Related updates
	CampaignID string `json:"campaign_id"`
	CreatorID  string `json:"creator_id"`

	// Metadata
	Timestamp string `json:"timestamp"`
	Source    string `json:"source"` // "polling_worker", "manual_sync"
}

func RandString(size int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var output string

	for range size {
		// pick a random index
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			log.Printf("error generating random int: %v", err)
			return ""
		}
		output += string(letters[n.Int64()])
	}
	return output
}
