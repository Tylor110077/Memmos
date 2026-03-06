package resource

import (
	"bytes"
	"context"
	"io"
	"testing"

	appgroup "github.com/tylor/goaipj/internal/app/group"
	"github.com/tylor/goaipj/internal/domain/resource"
	"github.com/tylor/goaipj/internal/infra/storage"
)

func TestUploadFileCreatesResourceArtifactAndJob(t *testing.T) {
	groupService := newGroupService(t)
	repo := NewInMemoryRepository()
	artifacts := NewInMemoryArtifactRepository()
	jobs := NewInMemoryJobPublisher()
	store := storage.NewStore(newFakeBucketClient(), "test-bucket")
	service := NewService(groupService, repo, artifacts, jobs, store)

	created, err := service.UploadFile(context.Background(), UploadFileInput{
		GroupID:     existingGroupID(t, groupService),
		Filename:    "notes.pdf",
		ContentType: "application/pdf",
		Body:        bytes.NewReader([]byte("pdf")),
		Size:        3,
	})
	if err != nil {
		t.Fatalf("UploadFile() error = %v", err)
	}

	if created.Resource.Type != resource.TypeFile {
		t.Fatalf("type = %s, want file", created.Resource.Type)
	}
	if created.Resource.Status != resource.StatusUploaded {
		t.Fatalf("status = %s, want uploaded", created.Resource.Status)
	}
	if created.Job.Type != JobTypeParseResource {
		t.Fatalf("job type = %s, want %s", created.Job.Type, JobTypeParseResource)
	}
	if len(artifacts.List()) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(artifacts.List()))
	}
}

func TestCreateWebResourceQueuesFetchJob(t *testing.T) {
	groupService := newGroupService(t)
	service := NewService(
		groupService,
		NewInMemoryRepository(),
		NewInMemoryArtifactRepository(),
		NewInMemoryJobPublisher(),
		storage.NewStore(newFakeBucketClient(), "test-bucket"),
	)

	created, err := service.CreateWebResource(context.Background(), CreateWebResourceInput{
		GroupID: existingGroupID(t, groupService),
		URL:     "https://example.com/page",
	})
	if err != nil {
		t.Fatalf("CreateWebResource() error = %v", err)
	}

	if created.Resource.Type != resource.TypeWeb {
		t.Fatalf("type = %s, want web", created.Resource.Type)
	}
	if created.Job.Type != JobTypeFetchWebResource {
		t.Fatalf("job type = %s, want %s", created.Job.Type, JobTypeFetchWebResource)
	}
}

func TestListGetAndRetryResource(t *testing.T) {
	groupService := newGroupService(t)
	repo := NewInMemoryRepository()
	jobs := NewInMemoryJobPublisher()
	service := NewService(
		groupService,
		repo,
		NewInMemoryArtifactRepository(),
		jobs,
		storage.NewStore(newFakeBucketClient(), "test-bucket"),
	)

	created, err := service.CreateWebResource(context.Background(), CreateWebResourceInput{
		GroupID: existingGroupID(t, groupService),
		URL:     "https://example.com/page",
	})
	if err != nil {
		t.Fatalf("CreateWebResource() error = %v", err)
	}

	item, err := repo.Get(context.Background(), created.Resource.ID)
	if err != nil {
		t.Fatalf("repo.Get() error = %v", err)
	}
	if err := item.MoveTo(resource.StatusFailed, resource.Failure{Stage: resource.StatusParsing, Message: "fetch failed"}); err != nil {
		t.Fatalf("MoveTo(failed) error = %v", err)
	}
	if err := repo.Update(context.Background(), *item); err != nil {
		t.Fatalf("repo.Update() error = %v", err)
	}

	list, err := service.ListResources(context.Background(), created.Resource.GroupID)
	if err != nil {
		t.Fatalf("ListResources() error = %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len(list) = %d, want 1", len(list))
	}

	got, err := service.GetResource(context.Background(), created.Resource.GroupID, created.Resource.ID)
	if err != nil {
		t.Fatalf("GetResource() error = %v", err)
	}
	if got.ID != created.Resource.ID {
		t.Fatalf("id = %s, want %s", got.ID, created.Resource.ID)
	}

	retried, err := service.RetryResource(context.Background(), created.Resource.ID)
	if err != nil {
		t.Fatalf("RetryResource() error = %v", err)
	}
	if retried.Resource.Status != resource.StatusUploaded {
		t.Fatalf("status = %s, want uploaded", retried.Resource.Status)
	}
	if retried.Job.Type != JobTypeFetchWebResource {
		t.Fatalf("job type = %s, want %s", retried.Job.Type, JobTypeFetchWebResource)
	}
}

func TestUpdateProcessingStatusAndFailure(t *testing.T) {
	groupService := newGroupService(t)
	repo := NewInMemoryRepository()
	service := NewService(
		groupService,
		repo,
		NewInMemoryArtifactRepository(),
		NewInMemoryJobPublisher(),
		storage.NewStore(newFakeBucketClient(), "test-bucket"),
	)

	created, err := service.CreateWebResource(context.Background(), CreateWebResourceInput{
		GroupID: existingGroupID(t, groupService),
		URL:     "https://example.com/page",
	})
	if err != nil {
		t.Fatalf("CreateWebResource() error = %v", err)
	}

	for _, next := range []resource.Status{resource.StatusParsing, resource.StatusNormalizing, resource.StatusGraphGenerating, resource.StatusCompleted} {
		updated, err := service.UpdateStatus(context.Background(), created.Resource.ID, next)
		if err != nil {
			t.Fatalf("UpdateStatus(%s) error = %v", next, err)
		}
		if updated.Status != next {
			t.Fatalf("status = %s, want %s", updated.Status, next)
		}
	}

	retried, err := service.RetryResource(context.Background(), created.Resource.ID)
	if err != nil {
		t.Fatalf("RetryResource() error = %v", err)
	}

	failed, err := service.FailResource(context.Background(), retried.Resource.ID, resource.StatusParsing, "fetch timeout")
	if err != nil {
		t.Fatalf("FailResource() error = %v", err)
	}
	if failed.Status != resource.StatusFailed {
		t.Fatalf("status = %s, want failed", failed.Status)
	}
	if failed.ErrorMessage != "fetch timeout" {
		t.Fatalf("error = %q, want fetch timeout", failed.ErrorMessage)
	}
}

func TestDeleteResourceRemovesStoredArtifacts(t *testing.T) {
	groupService := newGroupService(t)
	repo := NewInMemoryRepository()
	artifacts := NewInMemoryArtifactRepository()
	bucket := newFakeBucketClient()
	service := NewService(
		groupService,
		repo,
		artifacts,
		NewInMemoryJobPublisher(),
		storage.NewStore(bucket, "test-bucket"),
	)

	created, err := service.UploadFile(context.Background(), UploadFileInput{
		GroupID:     existingGroupID(t, groupService),
		Filename:    "notes.pdf",
		ContentType: "application/pdf",
		Body:        bytes.NewReader([]byte("pdf")),
		Size:        3,
	})
	if err != nil {
		t.Fatalf("UploadFile() error = %v", err)
	}
	if len(artifacts.List()) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(artifacts.List()))
	}
	if len(bucket.objects) != 1 {
		t.Fatalf("object count = %d, want 1", len(bucket.objects))
	}

	if err := service.DeleteResource(context.Background(), created.Resource.ID); err != nil {
		t.Fatalf("DeleteResource() error = %v", err)
	}

	if _, err := repo.Get(context.Background(), created.Resource.ID); err == nil {
		t.Fatalf("expected resource to be deleted")
	}
	if len(artifacts.List()) != 0 {
		t.Fatalf("artifact count = %d, want 0", len(artifacts.List()))
	}
	if len(bucket.objects) != 0 {
		t.Fatalf("object count = %d, want 0", len(bucket.objects))
	}
}

func newGroupService(t *testing.T) *appgroup.Service {
	t.Helper()
	repo := appgroup.NewInMemoryRepository()
	service := appgroup.NewService(repo)
	_, err := service.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	return service
}

func existingGroupID(t *testing.T, service *appgroup.Service) string {
	t.Helper()
	groups, err := service.ListGroups(context.Background())
	if err != nil {
		t.Fatalf("ListGroups() error = %v", err)
	}
	return groups[0].ID
}

type fakeBucketClient struct {
	objects map[string][]byte
}

func newFakeBucketClient() *fakeBucketClient {
	return &fakeBucketClient{objects: map[string][]byte{}}
}

func (f *fakeBucketClient) PutObject(_ context.Context, bucket, key string, reader io.Reader, _ int64, opts storage.PutObjectOptions) (storage.UploadInfo, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return storage.UploadInfo{}, err
	}
	f.objects[bucket+"/"+key] = raw
	return storage.UploadInfo{Key: key, Size: int64(len(raw)), ContentType: opts.ContentType}, nil
}

func (f *fakeBucketClient) GetObject(_ context.Context, bucket, key string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(f.objects[bucket+"/"+key])), nil
}

func (f *fakeBucketClient) DeleteObject(_ context.Context, bucket, key string) error {
	delete(f.objects, bucket+"/"+key)
	return nil
}
