package b2

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Mock S3 Client
type MockS3Client struct {
	GetObjectFunc    func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObjectFunc    func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	CreateBucketFunc func(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
	GetObjectAclFunc func(ctx context.Context, params *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error)
	DeleteObjectFunc func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return m.GetObjectFunc(ctx, params, optFns...)
}

func (m *MockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m.PutObjectFunc(ctx, params, optFns...)
}

func (m *MockS3Client) CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	return m.CreateBucketFunc(ctx, params, optFns...)
}

func (m *MockS3Client) GetObjectAcl(ctx context.Context, params *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
	return m.GetObjectAclFunc(ctx, params, optFns...)
}

func (m *MockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return m.DeleteObjectFunc(ctx, params, optFns...)
}

// Mock Presign Client
type MockPresignClient struct {
	PresignGetObjectFunc func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
	PresignPutObjectFunc func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

func (m *MockPresignClient) PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	return m.PresignGetObjectFunc(ctx, params, optFns...)
}

func (m *MockPresignClient) PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	return m.PresignPutObjectFunc(ctx, params, optFns...)
}

// Test New Bucket
func TestB2Storage_CreateBucket(t *testing.T) {
	t.Run("success. new bucket ok", func(t *testing.T) {
		mockClient := &MockS3Client{
			CreateBucketFunc: func(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
				return &s3.CreateBucketOutput{}, nil
			},
		}
		b2 := &B2Storage{
			BucketName: "dummy-bucket",
			Client:     mockClient,
		}

		err := b2.NewBucket("new-test-bucket")
		if err != nil {
			t.Errorf("NewBucket() error: %v, wanted: %v", err, false)
		}
	})
	t.Run("invaid bucket name", func(t *testing.T) {
		mockClient := &MockS3Client{
			CreateBucketFunc: func(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
				return &s3.CreateBucketOutput{}, nil
			},
		}
		b2 := &B2Storage{
			BucketName: "dummy-bucket",
			Client:     mockClient,
		}

		err := b2.NewBucket("")
		if err == nil {
			t.Errorf("NewBucket() error: nil, wanted: %v", ErrInvalidReq)
		} else if !errors.Is(err, ErrInvalidReq) {
			t.Errorf("NewBucket() error: %v, wanted: %v", err, ErrInvalidReq)
		}
	})

}

// Test Upload
func TestB2Storage_Upload(t *testing.T) {
	tests := []struct {
		name        string
		objKey      string
		fileData    []byte
		contentType string
		mockFunc    func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "successful upload",
			objKey:      "test.jpg",
			fileData:    []byte("test image data"),
			contentType: "image/jpeg",
			mockFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				return &s3.PutObjectOutput{}, nil
			},
			wantErr: false,
		},
		{
			name:        "upload fails",
			objKey:      "test.jpg",
			fileData:    []byte("test data"),
			contentType: "image/jpeg",
			mockFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				return nil, errors.New("S3 error")
			},
			wantErr:     true,
			expectedErr: ErrFileUploadError,
		},
		{
			name:        "auto detect content type",
			objKey:      "test.jpg",
			fileData:    []byte("\xff\xd8\xff"), // JPEG magic bytes
			contentType: "",
			mockFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				// Verify content type was detected
				if params.ContentType == nil || *params.ContentType == "" {
					t.Error("ContentType should be auto-detected")
				}
				return &s3.PutObjectOutput{}, nil
			},
			wantErr: false,
		},
		{
			name:        "upload from reader",
			objKey:      "test.jpg",
			fileData:    []byte("test data"),
			contentType: "image/jpeg",
			mockFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				return &s3.PutObjectOutput{}, nil
			},
			wantErr: false,
		},
		{
			name:        "upload from reader",
			objKey:      "test.jpg",
			fileData:    []byte("test data"),
			contentType: "",
			mockFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				return &s3.PutObjectOutput{}, nil
			},
			wantErr:     true,
			expectedErr: ErrInvalidReq,
		},
	}
	var err error
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockS3Client{
				PutObjectFunc: tt.mockFunc,
			}

			b2 := &B2Storage{
				BucketName: "test-bucket",
				Client:     mockClient,
			}
			if tt.name == "upload from reader" {
				err = b2.UploadFromReader(
					tt.objKey, bytes.NewReader(tt.fileData), tt.contentType,
				)
			} else {
				err = b2.UploadFile(tt.objKey, tt.fileData, tt.contentType)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Upload() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && !errors.Is(err, tt.expectedErr) {
				t.Errorf("Upload() error = %v, expectedErr %v", err, tt.expectedErr)
			}
		})
	}
}

// Test GetObjectBytes
func TestB2Storage_GetObjectBytes(t *testing.T) {
	tests := []struct {
		name         string
		objKey       string
		mockFunc     func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
		expectedData []byte
		wantErr      bool
		expectedErr  error
	}{
		{
			name:   "successful get",
			objKey: "test.jpg",
			mockFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return &s3.GetObjectOutput{
					Body: io.NopCloser(strings.NewReader("test data")),
				}, nil
			},
			expectedData: []byte("test data"),
			wantErr:      false,
		},
		{
			name:   "empty key",
			objKey: "",
			mockFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return nil, nil
			},
			wantErr:     true,
			expectedErr: ErrInvalidReq,
		},
		{
			name:   "S3 error",
			objKey: "test.jpg",
			mockFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return nil, errors.New("S3 error")
			},
			wantErr:     true,
			expectedErr: ErrDownloadFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockS3Client{
				GetObjectFunc: tt.mockFunc,
			}

			b2 := &B2Storage{
				BucketName: "test-bucket",
				Client:     mockClient,
			}

			data, err := b2.DownloadFile(tt.objKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && !errors.Is(err, tt.expectedErr) {
				t.Errorf("DownloadFile() error = %v, expectedErr %v", err, tt.expectedErr)
			}

			if !tt.wantErr && string(data) != string(tt.expectedData) {
				t.Errorf("DownloadFile() data = %v, expected %v", string(data), string(tt.expectedData))
			}
		})
	}
}

// Test GetSignedURL
func TestB2Storage_GetSignedURL(t *testing.T) {
	objKey := "test.jpg"

	tests := []struct {
		name        string
		objKey      *string
		task        string
		mockFunc    func() PresignClient
		expectedURL string
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "successful GET presign",
			objKey: &objKey,
			task:   GetObj,
			mockFunc: func() PresignClient {
				return &MockPresignClient{
					PresignGetObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
						return &v4.PresignedHTTPRequest{
							URL: "https://example.com/presigned-get-url",
						}, nil
					},
				}
			},
			expectedURL: "https://example.com/presigned-get-url",
			wantErr:     false,
		},
		{
			name:   "successful PUT presign",
			objKey: &objKey,
			task:   PutObj,
			mockFunc: func() PresignClient {
				return &MockPresignClient{
					PresignPutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
						return &v4.PresignedHTTPRequest{
							URL: "https://example.com/presigned-put-url",
						}, nil
					},
				}
			},
			expectedURL: "https://example.com/presigned-put-url",
			wantErr:     false,
		},
		{
			name:   "nil objKey",
			objKey: nil,
			task:   GetObj,
			mockFunc: func() PresignClient {
				return &MockPresignClient{}
			},
			wantErr:     true,
			expectedErr: ErrInvalidReq,
		},
		{
			name:   "invalid task",
			objKey: &objKey,
			task:   "INVALID",
			mockFunc: func() PresignClient {
				return &MockPresignClient{}
			},
			wantErr:     true,
			expectedErr: ErrUnsupportedTask,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b2 := &B2Storage{
				BucketName: "test-bucket",
				signClient: tt.mockFunc(),
			}

			url, err := b2.GetSignedURL(tt.objKey, tt.task)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSignedURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && !errors.Is(err, tt.expectedErr) {
				t.Errorf("GetSignedURL() error = %v, expectedErr %v", err, tt.expectedErr)
			}

			if !tt.wantErr && url != tt.expectedURL {
				t.Errorf("GetSignedURL() url = %v, expected %v", url, tt.expectedURL)
			}
		})
	}
}
