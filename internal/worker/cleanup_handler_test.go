package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	appcleanup "github.com/tylor/goaipj/internal/app/cleanup"
)

func TestCleanupHandlerRunsCleanup(t *testing.T) {
	repo := &cleanupRepo{
		artifacts: []appcleanup.ArtifactRecord{{ID: "a1", Type: "debug", StorageKey: "debug.json", CreatedAt: time.Now().Add(-48 * time.Hour)}},
		objects:   []appcleanup.ObjectRecord{{Key: "debug.json", UpdatedAt: time.Now().Add(-48 * time.Hour)}},
		jobs:      []appcleanup.JobRecord{{ID: "job-1", Status: "done", UpdatedAt: time.Now().Add(-48 * time.Hour)}},
	}
	service := appcleanup.NewService(repo)
	handler := NewCleanupHandler(service)
	payload, _ := json.Marshal(CleanupPayload{RetentionHours: 24})

	if err := handler(context.Background(), payload); err != nil {
		t.Fatalf("handler error = %v", err)
	}
	if len(repo.artifacts) != 0 || len(repo.objects) != 0 || len(repo.jobs) != 0 {
		t.Fatalf("expected cleanup to remove expired items")
	}
}

type cleanupRepo struct {
	artifacts []appcleanup.ArtifactRecord
	objects   []appcleanup.ObjectRecord
	jobs      []appcleanup.JobRecord
}

func (r *cleanupRepo) ListArtifacts(_ context.Context) ([]appcleanup.ArtifactRecord, error) {
	return append([]appcleanup.ArtifactRecord(nil), r.artifacts...), nil
}

func (r *cleanupRepo) DeleteArtifacts(_ context.Context, ids []string) error {
	keep := r.artifacts[:0]
	deleted := map[string]struct{}{}
	for _, id := range ids {
		deleted[id] = struct{}{}
	}
	for _, item := range r.artifacts {
		if _, ok := deleted[item.ID]; ok {
			continue
		}
		keep = append(keep, item)
	}
	r.artifacts = keep
	return nil
}

func (r *cleanupRepo) ListObjects(_ context.Context) ([]appcleanup.ObjectRecord, error) {
	return append([]appcleanup.ObjectRecord(nil), r.objects...), nil
}

func (r *cleanupRepo) DeleteObjects(_ context.Context, keys []string) error {
	keep := r.objects[:0]
	deleted := map[string]struct{}{}
	for _, key := range keys {
		deleted[key] = struct{}{}
	}
	for _, item := range r.objects {
		if _, ok := deleted[item.Key]; ok {
			continue
		}
		keep = append(keep, item)
	}
	r.objects = keep
	return nil
}

func (r *cleanupRepo) ListJobs(_ context.Context) ([]appcleanup.JobRecord, error) {
	return append([]appcleanup.JobRecord(nil), r.jobs...), nil
}

func (r *cleanupRepo) DeleteJobs(_ context.Context, ids []string) error {
	keep := r.jobs[:0]
	deleted := map[string]struct{}{}
	for _, id := range ids {
		deleted[id] = struct{}{}
	}
	for _, item := range r.jobs {
		if _, ok := deleted[item.ID]; ok {
			continue
		}
		keep = append(keep, item)
	}
	r.jobs = keep
	return nil
}
