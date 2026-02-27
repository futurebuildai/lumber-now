package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

type S3Client interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, string, error)
	PresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

type MediaService struct {
	s3     S3Client
	bucket string
}

func NewMediaService(s3 S3Client, bucket string) *MediaService {
	return &MediaService{s3: s3, bucket: bucket}
}

func (s *MediaService) Upload(ctx context.Context, dealerID uuid.UUID, filename string, body io.Reader, contentType string) (string, error) {
	key := fmt.Sprintf("%s/%s/%s", dealerID.String(), time.Now().Format("2006/01/02"), filename)

	if err := s.s3.Upload(ctx, key, body, contentType); err != nil {
		return "", fmt.Errorf("upload to S3: %w", err)
	}

	return key, nil
}

func (s *MediaService) Download(ctx context.Context, key string) (io.ReadCloser, string, error) {
	return s.s3.Download(ctx, key)
}

func (s *MediaService) GetPresignedURL(ctx context.Context, key string) (string, error) {
	return s.s3.PresignedURL(ctx, key, 15*time.Minute)
}
