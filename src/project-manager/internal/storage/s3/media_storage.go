package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"strings"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// MediaStorage реализует domain.MediaStorage поверх SeaweedFS S3.
type MediaStorage struct {
	client *Client
	bucket string
}

// NewMediaStorage создаёт экземпляр S3 хранилища медиафайлов.
func NewMediaStorage(client *Client, bucket string) *MediaStorage {
	if bucket == "" {
		bucket = "media"
	}
	return &MediaStorage{
		client: client,
		bucket: bucket,
	}
}

func (s *MediaStorage) getFileName(mediaType string) (string, error) {
	switch mediaType {
	case "icon":
		return "icon.png", nil
	case "cover":
		return "cover.png", nil
	case "video":
		return "video.mp4", nil
	default:
		return "", domain.ErrInvalidInput
	}
}

func validateMediaMime(r io.Reader, mediaType string) (string, []byte, error) {
	buf := make([]byte, 512)
	n, err := io.ReadFull(r, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", nil, err
	}
	if n == 0 {
		return "", nil, domain.ErrInvalidInput
	}

	contentType := http.DetectContentType(buf[:n])
	switch mediaType {
	case "icon", "cover":
		if !strings.HasPrefix(contentType, "image/") {
			return "", nil, fmt.Errorf("%w: expected image, detected %s", domain.ErrInvalidInput, contentType)
		}
	case "video":
		if !strings.HasPrefix(contentType, "video/") && contentType != "application/octet-stream" {
			return "", nil, fmt.Errorf("%w: expected video, detected %s", domain.ErrInvalidInput, contentType)
		}
	}
	return contentType, buf[:n], nil
}

// SaveMediaStream сохраняет промо-файл из потока io.Reader с валидацией MIME-типа в S3.
func (s *MediaStorage) SaveMediaStream(ctx context.Context, projectID int64, mediaType string, src io.Reader) (string, error) {
	fileName, err := s.getFileName(mediaType)
	if err != nil {
		return "", err
	}

	contentType, headBuf, err := validateMediaMime(src, mediaType)
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("%d/%s", projectID, fileName)
	
	// Объединяем прочитанный буфер и остаток потока
	multi := io.MultiReader(bytes.NewReader(headBuf), src)

	uploader := manager.NewUploader(s.client.Client)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        multi,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("s3 upload %s: %w", key, err)
	}

	return fileName, nil
}

// SaveMedia сохраняет промо-файл из среза байтов.
func (s *MediaStorage) SaveMedia(ctx context.Context, projectID int64, mediaType string, data []byte) (string, error) {
	return s.SaveMediaStream(ctx, projectID, mediaType, bytes.NewReader(data))
}

// DeleteMedia удаляет промо-файл указанного типа из S3.
func (s *MediaStorage) DeleteMedia(projectID int64, mediaType string) error {
	fileName, err := s.getFileName(mediaType)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%d/%s", projectID, fileName)
	_, err = s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// SnapshotMediaForRelease создает неизменяемую копию медиафайла для конкретного релиза через S3 CopyObject.
func (s *MediaStorage) SnapshotMediaForRelease(ctx context.Context, projectID int64, version string, srcPath string, mediaType string) (string, error) {
	if srcPath == "" {
		return "", nil
	}
	fileName, err := s.getFileName(mediaType)
	if err != nil {
		return "", err
	}

	srcKey := srcPath
	srcKey = strings.TrimPrefix(srcKey, "media/")
	if !strings.Contains(srcKey, "/") {
		srcKey = fmt.Sprintf("%d/%s", projectID, fileName)
	}

	destKey := fmt.Sprintf("%d/releases/%s/%s", projectID, version, fileName)
	copySource := url.PathEscape(fmt.Sprintf("%s/%s", s.bucket, srcKey))

	_, err = s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(destKey),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		var notFound *types.NotFound
		var apiErr smithy.APIError
		if errors.As(err, &noSuchKey) || errors.As(err, &notFound) || (errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NoSuchKey" || apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "404")) {
			// Если исходного объекта нет в S3, возвращаем исходный путь
			return srcPath, nil
		}
		return "", fmt.Errorf("s3 snapshot media copy %s -> %s: %w", copySource, destKey, err)
	}

	return destKey, nil
}
