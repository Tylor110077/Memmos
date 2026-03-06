package cleanup

import (
	"context"
	"testing"
	"time"
)

func TestCleanupRemovesExpiredDebugArtifactsOrphanObjectsAndFinishedJobs(t *testing.T) {
	now := time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC)
	repo := &fakeRepository{
		artifacts: []ArtifactRecord{
			{ID: "a-keep", Type: "raw_file", StorageKey: "groups/g/resources/r/raw/file.pdf", CreatedAt: now.Add(-48 * time.Hour)},
			{ID: "a-debug", Type: "debug", StorageKey: "groups/g/resources/r/artifacts/debug/out.json", CreatedAt: now.Add(-48 * time.Hour)},
		},
		objects: []ObjectRecord{
			{Key: "groups/g/resources/r/raw/file.pdf", UpdatedAt: now.Add(-48 * time.Hour)},
			{Key: "groups/g/resources/r/artifacts/debug/out.json", UpdatedAt: now.Add(-48 * time.Hour)},
			{Key: "groups/g/resources/orphan/artifacts/debug/tmp.json", UpdatedAt: now.Add(-48 * time.Hour)},
		},
		jobs: []JobRecord{
			{ID: "job-done", Status: "done", UpdatedAt: now.Add(-72 * time.Hour)},
			{ID: "job-failed", Status: "failed", UpdatedAt: now.Add(-72 * time.Hour)},
			{ID: "job-running", Status: "running", UpdatedAt: now.Add(-72 * time.Hour)},
		},
	}
	service := NewService(repo)

	report, err := service.Cleanup(context.Background(), now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}
	if report.DeletedArtifacts != 1 {
		t.Fatalf("deleted artifacts = %d, want 1", report.DeletedArtifacts)
	}
	if report.DeletedObjects != 2 {
		t.Fatalf("deleted objects = %d, want 2", report.DeletedObjects)
	}
	if report.DeletedJobs != 2 {
		t.Fatalf("deleted jobs = %d, want 2", report.DeletedJobs)
	}
}

type fakeRepository struct {
	artifacts []ArtifactRecord
	objects   []ObjectRecord
	jobs      []JobRecord
}

func (f *fakeRepository) ListArtifacts(_ context.Context) ([]ArtifactRecord, error) {
	return append([]ArtifactRecord(nil), f.artifacts...), nil
}

func (f *fakeRepository) DeleteArtifacts(_ context.Context, ids []string) error {
	keep := f.artifacts[:0]
	deleted := map[string]struct{}{}
	for _, id := range ids {
		deleted[id] = struct{}{}
	}
	for _, item := range f.artifacts {
		if _, ok := deleted[item.ID]; ok {
			continue
		}
		keep = append(keep, item)
	}
	f.artifacts = keep
	return nil
}

func (f *fakeRepository) ListObjects(_ context.Context) ([]ObjectRecord, error) {
	return append([]ObjectRecord(nil), f.objects...), nil
}

func (f *fakeRepository) DeleteObjects(_ context.Context, keys []string) error {
	keep := f.objects[:0]
	deleted := map[string]struct{}{}
	for _, key := range keys {
		deleted[key] = struct{}{}
	}
	for _, item := range f.objects {
		if _, ok := deleted[item.Key]; ok {
			continue
		}
		keep = append(keep, item)
	}
	f.objects = keep
	return nil
}

func (f *fakeRepository) ListJobs(_ context.Context) ([]JobRecord, error) {
	return append([]JobRecord(nil), f.jobs...), nil
}

func (f *fakeRepository) DeleteJobs(_ context.Context, ids []string) error {
	keep := f.jobs[:0]
	deleted := map[string]struct{}{}
	for _, id := range ids {
		deleted[id] = struct{}{}
	}
	for _, item := range f.jobs {
		if _, ok := deleted[item.ID]; ok {
			continue
		}
		keep = append(keep, item)
	}
	f.jobs = keep
	return nil
}
