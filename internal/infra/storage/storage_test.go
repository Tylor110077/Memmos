package storage

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestResourceObjectKeyBuilder(t *testing.T) {
	key := ResourceObjectKey("group-1", "resource-2", KindRaw, "notes.pdf")
	want := "groups/group-1/resources/resource-2/raw/notes.pdf"
	if key != want {
		t.Fatalf("key = %q, want %q", key, want)
	}

	artifactKey := ResourceObjectKey("group-1", "resource-2", KindNormalized, "summary.md")
	if artifactKey != "groups/group-1/resources/resource-2/artifacts/normalized/summary.md" {
		t.Fatalf("artifactKey = %q", artifactKey)
	}
}

func TestPutObjectUsesBucketAndNormalizedKey(t *testing.T) {
	client := &fakeBucketClient{}
	store := NewStore(client, "bucket-a")

	meta, err := store.PutObject(
		context.Background(),
		PutObjectInput{
			Key:         ResourceObjectKey("group-1", "resource-2", KindRaw, "slides deck.pptx"),
			ContentType: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
			Body:        bytes.NewReader([]byte("pptx")),
			Size:        int64(len("pptx")),
		},
	)
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	if client.lastBucket != "bucket-a" {
		t.Fatalf("bucket = %q, want bucket-a", client.lastBucket)
	}
	if client.lastKey != "groups/group-1/resources/resource-2/raw/slides_deck.pptx" {
		t.Fatalf("key = %q", client.lastKey)
	}
	if meta.Key != client.lastKey {
		t.Fatalf("meta key = %q, want %q", meta.Key, client.lastKey)
	}
}

func TestGetObjectReturnsReader(t *testing.T) {
	client := &fakeBucketClient{objects: map[string][]byte{
		"bucket-a/groups/group-1/resources/resource-2/raw/notes.pdf": []byte("pdf"),
	}}
	store := NewStore(client, "bucket-a")

	reader, err := store.GetObject(context.Background(), "groups/group-1/resources/resource-2/raw/notes.pdf")
	if err != nil {
		t.Fatalf("GetObject() error = %v", err)
	}
	defer reader.Close()

	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(raw) != "pdf" {
		t.Fatalf("body = %q, want pdf", string(raw))
	}
}

func TestDeleteObjectRemovesStoredKey(t *testing.T) {
	client := &fakeBucketClient{objects: map[string][]byte{
		"bucket-a/groups/group-1/resources/resource-2/raw/notes.pdf": []byte("pdf"),
	}}
	store := NewStore(client, "bucket-a")

	if err := store.DeleteObject(context.Background(), "groups/group-1/resources/resource-2/raw/notes.pdf"); err != nil {
		t.Fatalf("DeleteObject() error = %v", err)
	}
	if _, ok := client.objects["bucket-a/groups/group-1/resources/resource-2/raw/notes.pdf"]; ok {
		t.Fatalf("expected object to be deleted")
	}
}

func TestPresignPutObjectReturnsURLAndHeaders(t *testing.T) {
	client := &fakeBucketClient{}
	store := NewStore(client, "bucket-a")

	plan, err := store.PresignPutObject(context.Background(), PresignPutObjectInput{
		Key:         ResourceObjectKey("group-1", "resource-2", KindRaw, "notes.pdf"),
		ContentType: "application/pdf",
		ExpiresIn:   15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("PresignPutObject() error = %v", err)
	}
	if plan.Method != "PUT" {
		t.Fatalf("method = %q, want PUT", plan.Method)
	}
	if plan.Headers["Content-Type"] != "application/pdf" {
		t.Fatalf("content-type = %q", plan.Headers["Content-Type"])
	}
	parsed, err := url.Parse(plan.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		t.Fatalf("expected absolute URL, got %q", plan.URL)
	}
}

type fakeBucketClient struct {
	lastBucket string
	lastKey    string
	objects    map[string][]byte
}

func (f *fakeBucketClient) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, opts PutObjectOptions) (UploadInfo, error) {
	_ = ctx
	f.lastBucket = bucket
	f.lastKey = key
	if f.objects == nil {
		f.objects = map[string][]byte{}
	}
	raw, err := io.ReadAll(reader)
	if err != nil {
		return UploadInfo{}, err
	}
	f.objects[bucket+"/"+key] = raw
	return UploadInfo{Key: key, Size: int64(len(raw)), ContentType: opts.ContentType}, nil
}

func (f *fakeBucketClient) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	_ = ctx
	return io.NopCloser(bytes.NewReader(f.objects[bucket+"/"+key])), nil
}

func (f *fakeBucketClient) DeleteObject(ctx context.Context, bucket, key string) error {
	_ = ctx
	delete(f.objects, bucket+"/"+key)
	return nil
}

func (f *fakeBucketClient) PresignPutObject(ctx context.Context, bucket, key string, expires time.Duration, opts PutObjectOptions) (string, http.Header, error) {
	_ = ctx
	_ = expires
	header := http.Header{}
	if opts.ContentType != "" {
		header.Set("Content-Type", opts.ContentType)
	}
	return "https://uploads.example.test/" + bucket + "/" + key, header, nil
}
