package s3

import (
	"context"
	"encoding/json"
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
	accessKey, secretKey, err := GetSecrets()
	if err != nil {
		return nil, fmt.Errorf("could not get secrets from keyring: %w", err)
	}

	staticProvider := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")

	sConfig, err := sconfig.LoadDefaultConfig(ctx,
		sconfig.WithRegion(cfg.Region),
		sconfig.WithLogger(s3SlogAdapter{log: log}),
		sconfig.WithCredentialsProvider(staticProvider),
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

type Metadata struct {
	Salt     []byte `json:"salt"`
	Verifier []byte `json:"verifier"`
}

func (s *Storage) DownloadMeta(ctx context.Context) (*Metadata, error) {
	reader, err := s.Download(ctx, "meta.json")
	if err != nil {
		return nil, fmt.Errorf("failed to download meta.json: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func (s *Storage) Upload(ctx context.Context, key string, body io.Reader) error {
	s.log.Info("upload object %s into bucket %s", key, s.bucket)

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
