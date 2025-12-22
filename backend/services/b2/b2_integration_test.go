//go:build integration
// +build integration

package b2

import (
	"context"
	"testing"

	"github.com/Alter-Sitanshu/campaignHub/env"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func setupIntegrationTest(t *testing.T) *B2Storage {
	keyID := env.GetString("B2KeyID", "")
	appKey := env.GetString("B2API_KEY", "")
	endpoint := env.GetString("B2Endpoint", "")
	region := env.GetString("B2Region", "")
	bucket := env.GetString("RootBucket", "")

	if keyID == "" || appKey == "" || endpoint == "" || region == "" || bucket == "" {
		t.Skip("Skipping integration test: B2 credentials not set")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(keyID, appKey, ""),
		),
	)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return NewB2(bucket, client)
}

func TestIntegration_Upload(t *testing.T) {
	b2 := setupIntegrationTest(t)

	testData := []byte("integration test data")
	objKey := "test/integration-test.txt"

	err := b2.UploadFile(objKey, testData, "text/plain")
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// Verify by downloading
	data, err := b2.DownloadFile(objKey)
	if err != nil {
		t.Fatalf("GetObjectBytes failed: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Downloaded data = %s, expected %s", string(data), string(testData))
	}

	// Cleanup
	err = b2.DeleteFile(objKey)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = b2.DownloadFile(objKey)
	if err == nil {
		t.Fatalf("Expected error when downloading deleted object, got nil")
	}
}

func TestIntegration_GetSignedURL(t *testing.T) {
	b2 := setupIntegrationTest(t)

	objKey := "test/integration-test.txt"

	url, err := b2.GetSignedURL(&objKey, GetObj)
	if err != nil {
		t.Fatalf("GetSignedURL failed: %v", err)
	}

	if url == "" {
		t.Error("Expected non-empty URL")
	}

	t.Logf("Generated URL: %s\n", url)
}
