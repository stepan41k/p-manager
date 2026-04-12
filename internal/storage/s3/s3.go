package s3

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	sconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stepan41k/p-manager/internal/config"
)

type Storage struct {
	log    *slog.Logger
	client *s3.Client
	bucket string
}

func New(ctx context.Context, cfg *config.S3Config, log *slog.Logger) (*Storage, error) {
	sConfig, err := sconfig.LoadDefaultConfig(ctx,
		sconfig.WithRegion(cfg.Region),
		sconfig.WithLogger(s3SlogAdapter{log: log}),
		sconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}

	client := s3.NewFromConfig(sConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &Storage{
		log:    log,
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func (s *Storage) Upload(ctx context.Context, key string, body io.Reader) error {
	s.log.Info("Загрузка объекта %s в бакет %s", key, s.bucket)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   body,
	})

	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	return nil
}

func (s *Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}

	return result.Body, nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}
