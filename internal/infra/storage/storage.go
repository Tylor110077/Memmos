package storage

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

type PutObjectInput struct {
	Key         string
	ContentType string
	Body        io.Reader
	Size        int64
}

type ObjectMeta struct {
	Key         string
	Size        int64
	ContentType string
}

type PresignPutObjectInput struct {
	Key         string
	ContentType string
	ExpiresIn   time.Duration
}

type PresignedUpload struct {
	URL       string
	Method    string
	Headers   map[string]string
	ExpiresAt time.Time
}

type PutObjectOptions struct {
	ContentType string
}

type UploadInfo struct {
	Key         string
	Size        int64
	ContentType string
}

type bucketClient interface {
	PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, opts PutObjectOptions) (UploadInfo, error)
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	PresignPutObject(ctx context.Context, bucket, key string, expires time.Duration, opts PutObjectOptions) (string, http.Header, error)
}

type Store struct {
	client bucketClient
	bucket string
}

func NewStore(client bucketClient, bucket string) *Store {
	return &Store{
		client: client,
		bucket: bucket,
	}
}

func (s *Store) PutObject(ctx context.Context, input PutObjectInput) (ObjectMeta, error) {
	key := normalizeObjectKey(input.Key)
	info, err := s.client.PutObject(ctx, s.bucket, key, input.Body, input.Size, PutObjectOptions{
		ContentType: input.ContentType,
	})
	if err != nil {
		return ObjectMeta{}, err
	}
	return ObjectMeta{
		Key:         info.Key,
		Size:        info.Size,
		ContentType: info.ContentType,
	}, nil
}

func (s *Store) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, s.bucket, normalizeObjectKey(key))
}

func (s *Store) DeleteObject(ctx context.Context, key string) error {
	return s.client.DeleteObject(ctx, s.bucket, normalizeObjectKey(key))
}

func (s *Store) PresignPutObject(ctx context.Context, input PresignPutObjectInput) (PresignedUpload, error) {
	key := normalizeObjectKey(input.Key)
	expires := input.ExpiresIn
	if expires <= 0 {
		expires = 15 * time.Minute
	}
	url, headers, err := s.client.PresignPutObject(ctx, s.bucket, key, expires, PutObjectOptions{
		ContentType: input.ContentType,
	})
	if err != nil {
		return PresignedUpload{}, err
	}
	outHeaders := map[string]string{}
	for name, values := range headers {
		if len(values) == 0 {
			continue
		}
		outHeaders[name] = values[0]
	}
	return PresignedUpload{
		URL:       url,
		Method:    "PUT",
		Headers:   outHeaders,
		ExpiresAt: time.Now().UTC().Add(expires),
	}, nil
}

func normalizeObjectKey(key string) string {
	parts := strings.Split(key, "/")
	for i, part := range parts {
		parts[i] = normalizeObjectName(part)
	}
	return strings.Join(parts, "/")
}
