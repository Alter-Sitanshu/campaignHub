package platform

import (
	"context"
	"fmt"
	"regexp"
)

type Client interface {
	GetVideoDetails(ctx context.Context, VideoID string) (any, error)
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
