package s3

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// Client обёртка над aws s3.Client для работы с SeaweedFS S3 API.
type Client struct {
	*s3.Client
	log *slog.Logger
}

// Config настройки подключения к S3-совместимому хранилищу SeaweedFS.
type Config struct {
	Endpoint     string
	AccessKey    string
	SecretKey    string
	UseSSL       bool
	Region       string
	GamesBucket  string
	MediaBucket  string
	BuildsBucket string
}

// NewClient инициализирует клиент AWS SDK v2 для SeaweedFS S3.
func NewClient(ctx context.Context, cfg Config, log *slog.Logger) (*Client, error) {
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
		return nil, fmt.Errorf("load aws sdk config: %w", err)
	}

	s3Client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true // Для SeaweedFS обязателен path-style доступ: http://host:port/bucket/key
	})

	return &Client{
		Client: s3Client,
		log:    log,
	}, nil
}

// EnsureBuckets проверяет существование бакетов и создаёт отсутствующие.
func (c *Client) EnsureBuckets(ctx context.Context, buckets ...string) error {
	for _, bucket := range buckets {
		if bucket == "" {
			continue
		}

		_, err := c.Client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(bucket),
		})
		if err == nil {
			continue
		}

		var apiErr smithy.APIError
		var notFound *types.NotFound
		isNotFound := errors.As(err, &notFound) || (errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "404"))

		if isNotFound || err != nil {
			c.log.Info("creating S3 bucket in SeaweedFS", slog.String("bucket", bucket))
			_, createErr := c.Client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: aws.String(bucket),
			})
			if createErr != nil {
				// Если бакет уже создан другим процессом, игнорируем ошибку
				var bucketExists *types.BucketAlreadyExists
				var bucketOwned *types.BucketAlreadyOwnedByYou
				if errors.As(createErr, &bucketExists) || errors.As(createErr, &bucketOwned) {
					continue
				}
				return fmt.Errorf("create bucket %q: %w", bucket, createErr)
			}
		}
	}
	return nil
}
