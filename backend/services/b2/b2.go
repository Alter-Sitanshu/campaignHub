package b2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client interface for mocking
type S3Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
	GetObjectAcl(ctx context.Context, params *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// PresignClient interface for mocking
type PresignClient interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
	PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

const (
	GetObj  = "GET"
	PutObj  = "PUT"
	URLExp  = time.Minute * 10
	Timeout = time.Second * 20
)

var (
	ErrInvalidReq       = errors.New("cannot sign url. invalid nil arguments passed")
	ErrInternalCrash    = errors.New("server error. unexpected error")
	ErrUnsupportedTask  = errors.New("cannot sign url. task not supported")
	ErrFileUploadError  = errors.New("cannot upload file to storage. unexpected error")
	ErrDownloadFile     = errors.New("cannot download file. unexpected error")
	ErrInsufficientPerm = errors.New("access denied. permissions insufficient")
)

type B2Storage struct {
	BucketName string
	Client     S3Client
	signClient PresignClient
}

// CheckDeletePermission checks if you have delete permissions
func (b2 *B2Storage) CheckDeletePermission(objKey string) (bool, error) {
	if objKey == "" {
		return false, ErrInvalidReq
	}

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	// Get object ACL
	aclOutput, err := b2.Client.GetObjectAcl(ctx, &s3.GetObjectAclInput{
		Bucket: aws.String(b2.BucketName),
		Key:    aws.String(objKey),
	})

	if err != nil {
		log.Printf("error getting ACL for %s: %v", objKey, err)
		return false, err
	}

	// Check if you're the owner (owners have full control)
	if aclOutput.Owner != nil {
		log.Printf("Object owner ID: %s\n", *aclOutput.Owner.ID)
		return true, nil
	}

	return false, nil
}

// NewB2 creates a new B2Storage instance
func NewB2(bucketName string, client *s3.Client) *B2Storage {
	return &B2Storage{
		BucketName: bucketName,
		Client:     client,
		signClient: s3.NewPresignClient(client),
	}
}

// Handles creation of a new S3 bucket
func (b2 *B2Storage) NewBucket(bucket string) error {
	if bucket == "" {
		return ErrInvalidReq
	}
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	_, err := b2.Client.CreateBucket(ctx,
		&s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		},
	)
	if err != nil {
		log.Printf("error creating new bucket: %s\n", err.Error())
		return err
	}
	log.Printf("Successfully created bucket %s\n", bucket)
	return nil
}

// Handles generation of PreSignedURLs to perform S3 tasks
// Supported Tasks : [GET, PUT]
func (b2 *B2Storage) GetSignedURL(objKey *string, task string) (string, error) {
	if objKey == nil {
		return "", ErrInvalidReq
	}
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	switch task {
	case GetObj:
		signedURL, err := b2.signClient.PresignGetObject(ctx,
			&s3.GetObjectInput{
				Bucket: aws.String(b2.BucketName),
				Key:    objKey,
			},
			s3.WithPresignExpires(URLExp),
		)
		if err != nil {
			log.Printf("unexpected error signing url for %s: %s\n", task, err.Error())
			return "", ErrInternalCrash
		}

		return signedURL.URL, nil
	case PutObj:
		signedURL, err := b2.signClient.PresignPutObject(ctx,
			&s3.PutObjectInput{
				Bucket: aws.String(b2.BucketName),
				Key:    objKey,
			},
			s3.WithPresignExpires(URLExp),
		)
		if err != nil {
			log.Printf("unexpected error signing url for %s: %s\n", task, err.Error())
			return "", ErrInternalCrash
		}

		return signedURL.URL, nil
	default:
		return "", ErrUnsupportedTask
	}
}

// Handles the upload of files through byte streams directly
// inefficient for large files to load onto memory: Use UploadFromReader
func (b2 *B2Storage) UploadFile(objKey string, fileData []byte, contentType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	// detect the content type if not provided
	if contentType == "" {
		contentType = http.DetectContentType(fileData)
	}
	_, err := b2.Client.PutObject(ctx,
		&s3.PutObjectInput{
			Bucket:      aws.String(b2.BucketName),
			Key:         aws.String(objKey),
			Body:        bytes.NewReader(fileData),
			ContentType: aws.String(contentType),
		},
	)
	if err != nil {
		log.Printf("error uploading file: %s", err.Error())
		return ErrFileUploadError
	}
	return nil
}

// Handles the upload of large files by leveraging io.Reader and not copying the entire content onto memory
func (b2 *B2Storage) UploadFromReader(objKey string, reader io.Reader, contentType string) error {
	if objKey == "" {
		return ErrInvalidReq
	}
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	// content type is required
	if contentType == "" {
		return ErrInvalidReq
	}

	_, err := b2.Client.PutObject(ctx,
		&s3.PutObjectInput{
			Bucket:             aws.String(b2.BucketName),
			Key:                aws.String(objKey),
			Body:               reader,
			ContentDisposition: aws.String(contentType),
		},
	)
	if err != nil {
		log.Printf("error uploading file: %s\n", err.Error())
		return ErrFileUploadError
	}

	return nil
}

// Handles the multipart form file uploads
func (b2 *B2Storage) UploadMultiPart(objKey string, file *multipart.FileHeader) error {
	if file == nil {
		return ErrInvalidReq
	}

	// Open the multipart file
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Detect content type from file extension
	contentType := mime.TypeByExtension(filepath.Ext(file.Filename))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// upload from the reader
	return b2.UploadFromReader(objKey, src, contentType)
}

// Handles the retreival of the blob object from the Bucket
func (b2 *B2Storage) DownloadFile(objKey string) ([]byte, error) {
	if objKey == "" {
		return nil, ErrInvalidReq
	}
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	file, err := b2.Client.GetObject(ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(b2.BucketName),
			Key:    aws.String(objKey),
		},
	)
	if err != nil {
		log.Printf("error fetching file: %v\n", err)
		return nil, ErrDownloadFile
	}
	defer file.Body.Close()

	data, err := io.ReadAll(file.Body)
	if err != nil {
		log.Printf("error reading object body %s: %v", objKey, err)
		return nil, ErrDownloadFile
	}

	return data, nil
}

func (b2 *B2Storage) DeleteFile(objKey string) error {
	// Execute the Delete operation
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	_, err := b2.Client.DeleteObject(ctx,
		&s3.DeleteObjectInput{
			Bucket: aws.String(b2.BucketName),
			Key:    aws.String(objKey),
		},
	)

	if err != nil {
		log.Printf("could not delete file. error: %v\n", err.Error())
		return ErrInternalCrash
	}

	return nil
}

func (b2 *B2Storage) DeleteFileWithPerms(objKey string) error {
	if objKey == "" {
		return ErrInvalidReq
	}
	permisson, err := b2.CheckDeletePermission(objKey)
	if err != nil {
		log.Printf("error could not resolve permission: %v\n", err.Error())
		return err
	}
	// Access Denied
	if !permisson {
		return ErrInsufficientPerm
	}

	return b2.DeleteFile(objKey)
}
