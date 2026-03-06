package storage

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

type minioBucketClient struct {
	client *minio.Client
}

func NewMinIOBucketClient(cfg MinIOConfig) (*minioBucketClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &minioBucketClient{client: client}, nil
}

func (m *minioBucketClient) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, opts PutObjectOptions) (UploadInfo, error) {
	info, err := m.client.PutObject(ctx, bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: opts.ContentType,
	})
	if err != nil {
		return UploadInfo{}, err
	}
	return UploadInfo{
		Key:         info.Key,
		Size:        info.Size,
		ContentType: opts.ContentType,
	}, nil
}

func (m *minioBucketClient) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return m.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
}
