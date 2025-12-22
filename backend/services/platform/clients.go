package platform

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

type Client interface {
	GetVideoDetails(ctx context.Context, VideoID string) (any, error)
	GetVideoDetailsForWorkers(ctx context.Context, VideoID string) (any, error)
}

type Factory struct {
	youtube   Client
	instagram Client
}

type Platform struct {
	Name    string
	VideoID string
}

func NewFactory(youtubeAPIKey, instagramToken string) (*Factory, error) {
	return &Factory{
		youtube:   NewYTClient(youtubeAPIKey),
		instagram: NewInstagramClient(instagramToken),
	}, nil
}

func (f *Factory) GetClient(platform string) (Client, error) {
	switch platform {
	case "youtube":
		return f.youtube, nil
	case "instagram":
		return f.instagram, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

func (f *Factory) GetVideoDetails(ctx context.Context, platform, VideoID string) (any, error) {
	client, err := f.GetClient(platform)
	if err != nil {
		return 0, err
	}
	return client.GetVideoDetails(ctx, VideoID)
}

func (f *Factory) GetVideoDetailsForWorkers(ctx context.Context, platform, VideoID string) (any, error) {
	client, err := f.GetClient(platform)
	if err != nil {
		return 0, err
	}
	return client.GetVideoDetailsForWorkers(ctx, VideoID)
}

// ParseVideoURL extracts platform and video ID from URL
func ParseVideoURL(url string) (*Platform, error) {
	// YouTube (Primary)
	youtubeRegex := regexp.MustCompile(`(?:youtube\.com/(?:watch\?v=|shorts/)|youtu\.be/)([a-zA-Z0-9_-]{11})`)
	if match := youtubeRegex.FindStringSubmatch(url); match != nil {
		return &Platform{Name: "youtube", VideoID: match[1]}, nil
	}

	// Instagram
	instagramRegex := regexp.MustCompile(`instagram\.com\/(?:p|reel)\/([a-zA-Z0-9_-]+)`)
	if match := instagramRegex.FindStringSubmatch(url); match != nil {
		return &Platform{Name: "instagram", VideoID: match[1]}, nil
	}

	// TikTok
	tiktokRegex := regexp.MustCompile(`tiktok\.com\/@[\w.-]+\/video\/(\d+)`)
	if match := tiktokRegex.FindStringSubmatch(url); match != nil {
		return &Platform{Name: "tiktok", VideoID: match[1]}, nil
	}

	return nil, fmt.Errorf("unsupported URL format: %s", url)
}

func DownloadFile(url string) ([]byte, string, error) {
	client := http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("failed to download thumbnail: %d", resp.StatusCode)
	}

	// Read body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// Get content-type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	return data, contentType, nil
}
