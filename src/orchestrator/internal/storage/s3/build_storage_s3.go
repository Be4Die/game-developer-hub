package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// Config настройки подключения к SeaweedFS S3 для Orchestrator.
type Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Region    string
}

// BuildStorageS3 реализует domain.BuildStorageFS поверх SeaweedFS S3.
type BuildStorageS3 struct {
	client *s3.Client
	bucket string
	log    *slog.Logger
}

// NewBuildStorageS3 создаёт хранилище серверных билдов в SeaweedFS S3.
func NewBuildStorageS3(ctx context.Context, cfg Config, log *slog.Logger) (*BuildStorageS3, error) {
	if log == nil {
		log = slog.Default()
	}

	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	endpoint := cfg.Endpoint
	if endpoint != "" && !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		if cfg.UseSSL {
			endpoint = "https://" + endpoint
		} else {
			endpoint = "http://" + endpoint
		}
	}

	accessKey := cfg.AccessKey
	secretKey := cfg.SecretKey
	if accessKey == "" && secretKey == "" {
		if strings.Contains(endpoint, "8333") || strings.Contains(endpoint, "seaweedfs") || strings.Contains(endpoint, "localhost") {
			accessKey = "seaweedfs"
			secretKey = "seaweedfs"
		}
	}

	optFns := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}
	if accessKey != "" || secretKey != "" {
		optFns = append(optFns, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
	}

	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("load aws sdk config for orchestrator: %w", err)
	}

	s3Client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true
	})

	bucket := cfg.Bucket
	if bucket == "" {
		bucket = "server-builds"
	}

	// Убедимся, что бакет существует
	_, headErr := s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if headErr != nil {
		var apiErr smithy.APIError
		var notFound *types.NotFound
		if errors.As(headErr, &notFound) || (errors.As(headErr, &apiErr) && (apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "404")) {
			log.Info("creating orchestrator server-builds bucket in SeaweedFS", slog.String("bucket", bucket))
			_, _ = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: aws.String(bucket),
			})
		}
	}

	return &BuildStorageS3{
		client: s3Client,
		bucket: bucket,
		log:    log,
	}, nil
}

func (s *BuildStorageS3) objectKey(gameID int64, version string) string {
	return fmt.Sprintf("%d/%s.tar", gameID, version)
}

// Save сохраняет tar-архив билда в SeaweedFS S3.
func (s *BuildStorageS3) Save(gameID int64, version string, reader io.Reader, size int64) (string, error) {
	key := s.objectKey(gameID, version)
	ctx := context.Background()

	uploader := manager.NewUploader(s.client)
	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String("application/x-tar"),
	}
	if size > 0 {
		putInput.ContentLength = aws.Int64(size)
	}

	_, err := uploader.Upload(ctx, putInput)
	if err != nil {
		return "", fmt.Errorf("s3 upload server build %s: %w", key, err)
	}

	return fmt.Sprintf("s3://%s/%s", s.bucket, key), nil
}

// Get возвращает io.ReadCloser для чтения файла билда из S3.
func (s *BuildStorageS3) Get(gameID int64, version string) (io.ReadCloser, error) {
	key := s.objectKey(gameID, version)
	out, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return out.Body, nil
}

// Delete удаляет файл билда из S3.
func (s *BuildStorageS3) Delete(gameID int64, version string) error {
	key := s.objectKey(gameID, version)
	_, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete server build %s: %w", key, err)
	}
	return nil
}

// Exists проверяет существование билда в SeaweedFS S3.
func (s *BuildStorageS3) Exists(gameID int64, version string) bool {
	key := s.objectKey(gameID, version)
	_, err := s.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err == nil
}
