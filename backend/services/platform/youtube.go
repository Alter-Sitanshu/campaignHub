package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type YTClient struct {
	APIKey     string
	httpClient *http.Client
}

type Stats struct {
	ViewCount string `json:"viewCount"`
	LikeCount string `json:"likeCount"`
}

type YTThumbnails struct {
	Medium struct {
		URL string `json:"url"`
	} `json:"medium"`
	// Optional low quality thumbnail
	// Low struct {
	// 	URL    string `json:"url"`
	// } `json:"default"`
}

type Snippet struct {
	Title      string       `json:"title"`
	Thumbs     YTThumbnails `json:"thumbnails"`
	UploadedAt string       `json:"publishedAt"`
}

type Video struct {
	Details    Snippet `json:"snippet"`
	Statistics Stats   `json:"statistics"`
}

type YoutubeResponse struct {
	Items []Video `json:"items"`
}

type Thumbnail struct {
	Raw         []byte `json:"raw"`
	ContentType string `json:"content-type"`
}

// Return structured metadata
type VideoMetadata struct {
	VideoID    string    `json:"video_id"`
	Platform   string    `json:"platform"`
	Title      string    `json:"title"`
	ViewCount  int       `json:"view_count"`
	LikeCount  int       `json:"like_count"`
	Thumbnails Thumbnail `json:"thumbnails,omitempty"`
	UploadedAt string    `json:"uploaded_at"`
}

// constructor for the YT Client
func NewYTClient(key string) *YTClient {
	return &YTClient{
		APIKey:     key,
		httpClient: &http.Client{},
	}
}

func (yt *YTClient) GetVideoDetails(ctx context.Context, VideoID string) (any, error) {
	if VideoID == "" || VideoID == " " {
		log.Printf("error: video id invalid %q\n", VideoID)
		return nil, fmt.Errorf("invalid video id")
	}
	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&id=%s&key=%s",
		VideoID, yt.APIKey,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("error: %v", err.Error())
		return nil, err
	}

	resp, err := yt.httpClient.Do(req)
	if err != nil {
		log.Printf("error: %v", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("youtube api returned status %d", resp.StatusCode)
	}
	var data YoutubeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("error: %v", err.Error())
		return nil, err
	}

	if len(data.Items) == 0 {
		log.Printf("error: %v", "invalid video id")
		return nil, fmt.Errorf("video not found")
	}

	video := data.Items[0]

	// Convert string counts to integers
	viewCount, _ := strconv.Atoi(video.Statistics.ViewCount)
	likeCount, _ := strconv.Atoi(video.Statistics.LikeCount)
	donwloadThumb, Type, err := DownloadFile(video.Details.Thumbs.Medium.URL)
	if err != nil {
		return nil, err
	}
	thumbs := Thumbnail{
		Raw:         donwloadThumb,
		ContentType: Type,
	}
	//{
	// 	Quality: "low",
	// 	URL:     video.Details.Thumbs.Low.URL,
	// 	Width:   video.Details.Thumbs.Low.Width,
	// 	Height:  video.Details.Thumbs.Low.Height,
	// },
	return &VideoMetadata{
		VideoID:    VideoID,
		Platform:   "youtube",
		Title:      video.Details.Title,
		ViewCount:  viewCount,
		LikeCount:  likeCount,
		Thumbnails: thumbs,
		UploadedAt: video.Details.UploadedAt,
	}, nil
}

func (yt *YTClient) GetVideoDetailsForWorkers(ctx context.Context, VideoID string) (*VideoMetadata, error) {
	if VideoID == "" || VideoID == " " {
		log.Printf("error: video id invalid %q\n", VideoID)
		return nil, fmt.Errorf("invalid video id")
	}
	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&id=%s&key=%s",
		VideoID, yt.APIKey,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("error: %v", err.Error())
		return nil, err
	}

	resp, err := yt.httpClient.Do(req)
	if err != nil {
		log.Printf("error: %v", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("youtube api returned status %d", resp.StatusCode)
	}
	var data YoutubeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("error: %v", err.Error())
		return nil, err
	}

	if len(data.Items) == 0 {
		log.Printf("error: %v", "invalid video id")
		return nil, fmt.Errorf("video not found")
	}

	video := data.Items[0]

	// Convert string counts to integers
	viewCount, _ := strconv.Atoi(video.Statistics.ViewCount)
	likeCount, _ := strconv.Atoi(video.Statistics.LikeCount)

	return &VideoMetadata{
		VideoID:    VideoID,
		Platform:   "youtube",
		Title:      video.Details.Title,
		ViewCount:  viewCount,
		LikeCount:  likeCount,
		UploadedAt: video.Details.UploadedAt,
	}, nil
}
