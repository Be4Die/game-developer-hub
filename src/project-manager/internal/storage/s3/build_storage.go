package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"


	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// BuildStorage реализует domain.BuildStorage поверх SeaweedFS S3.
type BuildStorage struct {
	client *Client
	bucket string
}

// NewBuildStorage создаёт экземпляр S3 хранилища сборок.
func NewBuildStorage(client *Client, bucket string) *BuildStorage {
	if bucket == "" {
		bucket = "builds"
	}
	return &BuildStorage{
		client: client,
		bucket: bucket,
	}
}

func (s *BuildStorage) objectKey(projectID int64, version string) string {
	return fmt.Sprintf("projects/%d/%s.zip", projectID, version)
}

// SaveArchiveStream сохраняет архив сборки из потока io.Reader в S3.
func (s *BuildStorage) SaveArchiveStream(ctx context.Context, projectID int64, version string, src io.Reader) (string, int64, error) {
	key := s.objectKey(projectID, version)

	// Используем CountingReader для подсчета байт (или просто Uploader)
	// manager.Uploader умеет разбивать поток на куски и грузить в память
	uploader := manager.NewUploader(s.client.Client)
	
	// Чтобы узнать размер, обернем src (так как Uploader не возвращает размер)
	counter := &countingReader{r: src}

	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        counter,
		ContentType: aws.String("application/zip"),
	})
	if err != nil {
		return "", 0, fmt.Errorf("s3 upload %s: %w", key, err)
	}

	return key, counter.n, nil
}

type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

// SaveArchive сохраняет архив сборки из массива байтов в S3.
func (s *BuildStorage) SaveArchive(ctx context.Context, projectID int64, version string, data []byte) (string, int64, error) {
	return s.SaveArchiveStream(ctx, projectID, version, bytes.NewReader(data))
}

// GetArchivePath возвращает S3-ключ архива версии.
func (s *BuildStorage) GetArchivePath(projectID int64, version string) string {
	return s.objectKey(projectID, version)
}

// DeleteBuild удаляет архив конкретной версии сборки из S3.
func (s *BuildStorage) DeleteBuild(projectID int64, version string) error {
	key := s.objectKey(projectID, version)
	_, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete object %s: %w", key, err)
	}
	return nil
}

// DeleteProject удаляет все архивы проекта из S3.
func (s *BuildStorage) DeleteProject(projectID int64) error {
	ctx := context.Background()
	prefix := fmt.Sprintf("projects/%d/", projectID)

	paginator := s3.NewListObjectsV2Paginator(s.client.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list objects for delete: %w", err)
		}
		if len(page.Contents) == 0 {
			continue
		}

		deleteObjects := make([]types.ObjectIdentifier, len(page.Contents))
		for i, obj := range page.Contents {
			deleteObjects[i] = types.ObjectIdentifier{Key: obj.Key}
		}

		_, err = s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(s.bucket),
			Delete: &types.Delete{
				Objects: deleteObjects,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return fmt.Errorf("delete objects in batch: %w", err)
		}
	}

	return nil
}

// GetObject возвращает io.ReadCloser для чтения объекта архива из S3.
func (s *BuildStorage) GetObject(ctx context.Context, projectID int64, version string) (io.ReadCloser, error) {
	key := s.objectKey(projectID, version)
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object %s: %w", key, err)
	}
	return out.Body, nil
}
