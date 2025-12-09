package platform

import (
	"context"
	"net/http"
)

type Instagram struct {
	token      string
	httpClient *http.Client
}

// consttructor for Instagram Client
func NewInstagramClient(token string) *Instagram {
	return &Instagram{
		token:      token,
		httpClient: &http.Client{},
	}
}

// It was too much of hedaache using the Meta API
// Implement it later IF YOU WANT
// Returns the View Count of an Instagram Reel
func (i *Instagram) GetVideoDetails(ctx context.Context, VideoID string) (any, error) {
	return 0, nil
}

func (i *Instagram) GetVideoDetailsForWorkers(ctx context.Context, VideoID string) (any, error) {
	return 0, nil
}
