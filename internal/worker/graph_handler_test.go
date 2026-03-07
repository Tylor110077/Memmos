package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	appresource "github.com/tylor/goaipj/internal/app/resource"
	"github.com/tylor/goaipj/internal/domain/resource"
	"github.com/tylor/goaipj/internal/infra/storage"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestGenerateGraphHandlerCompletesResource(t *testing.T) {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	resourceService := appresource.NewService(
		groupService,
		appresource.NewInMemoryRepository(),
		appresource.NewInMemoryArtifactRepository(),
		appresource.NewInMemoryJobPublisher(),
		storage.NewStore(newFakeStorageClient(), "bucket"),
	)
	graphService := appgraph.NewService(appgraph.NewInMemoryRepository())
	handler := NewGenerateGraphHandler(resourceService, graphService, pipelinegraph.NewGenerator())

	created, err := resourceService.CreateWebResource(context.Background(), appresource.CreateWebResourceInput{
		GroupID: group.ID,
		URL:     "https://example.com/page",
	})
	if err != nil {
		t.Fatalf("CreateWebResource() error = %v", err)
	}
	if _, err := resourceService.UpdateStatus(context.Background(), created.Resource.ID, resource.StatusParsing); err != nil {
		t.Fatalf("UpdateStatus(parsing) error = %v", err)
	}
	if _, err := resourceService.UpdateStatus(context.Background(), created.Resource.ID, resource.StatusNormalizing); err != nil {
		t.Fatalf("UpdateStatus(normalizing) error = %v", err)
	}

	payload, _ := json.Marshal(GenerateGraphPayload{
		GroupID:    group.ID,
		ResourceID: created.Resource.ID,
		Title:      "Backend",
		Summary:    "summary",
		Markdown:   "# Backend\n\nUploads.\nWorkers.",
	})

	if err := handler(context.Background(), payload); err != nil {
		t.Fatalf("handler error = %v", err)
	}

	res, err := resourceService.GetResource(context.Background(), group.ID, created.Resource.ID)
	if err != nil {
		t.Fatalf("GetResource() error = %v", err)
	}
	if res.Status != resource.StatusCompleted {
		t.Fatalf("status = %s, want completed", res.Status)
	}
}

type fakeStorageClient struct{}

func newFakeStorageClient() *fakeStorageClient { return &fakeStorageClient{} }

func (f *fakeStorageClient) PutObject(_ context.Context, bucket, key string, reader io.Reader, size int64, opts storage.PutObjectOptions) (storage.UploadInfo, error) {
	_, _, _ = bucket, reader, size
	return storage.UploadInfo{Key: key, Size: size, ContentType: opts.ContentType}, nil
}

func (f *fakeStorageClient) GetObject(_ context.Context, bucket, key string) (io.ReadCloser, error) {
	_, _ = bucket, key
	return io.NopCloser(bytes.NewReader(nil)), nil
}

func (f *fakeStorageClient) DeleteObject(_ context.Context, bucket, key string) error {
	_, _ = bucket, key
	return nil
}

func (f *fakeStorageClient) PresignPutObject(_ context.Context, bucket, key string, _ time.Duration, opts storage.PutObjectOptions) (string, http.Header, error) {
	_ = bucket
	header := http.Header{}
	if opts.ContentType != "" {
		header.Set("Content-Type", opts.ContentType)
	}
	return "https://uploads.example.test/" + key, header, nil
}
