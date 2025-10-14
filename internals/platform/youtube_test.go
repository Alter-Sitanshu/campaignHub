package platform

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/env"
)

func TestYoutubeAPI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	client, err := NewFactory(env.GetString("YTAPIKEY", ""), "")
	if err != nil {
		t.Fail()
	}

	t.Run("OK", func(t *testing.T) {

		// Random video URL from the Internet
		// 11 Characters video ID
		url := "https://www.youtube.com/watch?v=1234567FrDE"
		ExampleVideo, err := ParseVideoURL(url)
		if err != nil {
			// we expect no error (For now cause the video IS public)
			log.Printf("error fetching video details, may be change the URL: %v\n", err.Error())
			t.Fail()
		}
		if ExampleVideo.VideoID != "1234567FrDE" ||
			ExampleVideo.Name != "youtube" {
			log.Printf("video_id: %q, name: %s", ExampleVideo.VideoID, ExampleVideo.Name)
			t.Fail()
		}
	})
	t.Run("Invalid URL", func(t *testing.T) {
		// Random video URL from the Internet
		_, err := ParseVideoURL("https://www.youtubv.com/watch?v=7FrDEqNnERs")
		if err == nil {
			// we expect an error to occur
			t.Fail()
		}
	})
	t.Run("invalid video ID", func(t *testing.T) {
		_, err := client.GetVideoDetails(ctx, "youtube", "")
		if err == nil {
			t.Fail()
		}
	})
}
