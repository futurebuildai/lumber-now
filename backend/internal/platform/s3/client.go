package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/builderwire/lumber-now/backend/internal/platform/circuitbreaker"
)

type Client struct {
	client  *s3.Client
	bucket  string
	breaker *circuitbreaker.Breaker
}

func NewClient(endpoint, bucket, region, accessKey, secretKey string) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		}
	})

	return &Client{
		client:  client,
		bucket:  bucket,
		breaker: circuitbreaker.New(5, 60*time.Second),
	}, nil
}

// BreakerState returns the current circuit breaker state for health reporting.
func (c *Client) BreakerState() string {
	return c.breaker.State().String()
}

func (c *Client) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	return c.breaker.Execute(func() error {
		_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(c.bucket),
			Key:         aws.String(key),
			Body:        body,
			ContentType: aws.String(contentType),
		})
		if err != nil {
			return fmt.Errorf("S3 PutObject: %w", err)
		}
		return nil
	})
}

func (c *Client) Download(ctx context.Context, key string) (io.ReadCloser, string, error) {
	var body io.ReadCloser
	var ct string
	err := c.breaker.Execute(func() error {
		out, err := c.client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(c.bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return fmt.Errorf("S3 GetObject: %w", err)
		}
		ct = "application/octet-stream"
		if out.ContentType != nil {
			ct = *out.ContentType
		}
		body = out.Body
		return nil
	})
	if err != nil {
		return nil, "", err
	}
	return body, ct, nil
}

func (c *Client) PresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	var url string
	err := c.breaker.Execute(func() error {
		presigner := s3.NewPresignClient(c.client)
		req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(c.bucket),
			Key:    aws.String(key),
		}, s3.WithPresignExpires(expiry))
		if err != nil {
			return fmt.Errorf("S3 presign: %w", err)
		}
		url = req.URL
		return nil
	})
	if err != nil {
		return "", err
	}
	return url, nil
}
