package cleanup

import (
	"context"
	"time"
)

type ArtifactRecord struct {
	ID         string
	Type       string
	StorageKey string
	CreatedAt  time.Time
}

type ObjectRecord struct {
	Key       string
	UpdatedAt time.Time
}

type JobRecord struct {
	ID        string
	Status    string
	UpdatedAt time.Time
}

type Repository interface {
	ListArtifacts(ctx context.Context) ([]ArtifactRecord, error)
	DeleteArtifacts(ctx context.Context, ids []string) error
	ListObjects(ctx context.Context) ([]ObjectRecord, error)
	DeleteObjects(ctx context.Context, keys []string) error
	ListJobs(ctx context.Context) ([]JobRecord, error)
	DeleteJobs(ctx context.Context, ids []string) error
}

type Report struct {
	DeletedArtifacts int `json:"deleted_artifacts"`
	DeletedObjects   int `json:"deleted_objects"`
	DeletedJobs      int `json:"deleted_jobs"`
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Cleanup(ctx context.Context, cutoff time.Time) (Report, error) {
	artifacts, err := s.repo.ListArtifacts(ctx)
	if err != nil {
		return Report{}, err
	}
	referenced := map[string]struct{}{}
	var artifactDeleteIDs []string
	for _, artifact := range artifacts {
		if artifact.StorageKey != "" {
			referenced[artifact.StorageKey] = struct{}{}
		}
		if artifact.Type == "debug" && artifact.CreatedAt.Before(cutoff) {
			artifactDeleteIDs = append(artifactDeleteIDs, artifact.ID)
		}
	}
	if len(artifactDeleteIDs) > 0 {
		if err := s.repo.DeleteArtifacts(ctx, artifactDeleteIDs); err != nil {
			return Report{}, err
		}
	}

	objects, err := s.repo.ListObjects(ctx)
	if err != nil {
		return Report{}, err
	}
	var objectDeleteKeys []string
	for _, object := range objects {
		if _, ok := referenced[object.Key]; !ok && object.UpdatedAt.Before(cutoff) {
			objectDeleteKeys = append(objectDeleteKeys, object.Key)
			continue
		}
		for _, artifact := range artifacts {
			if artifact.StorageKey == object.Key && artifact.Type == "debug" && artifact.CreatedAt.Before(cutoff) {
				objectDeleteKeys = append(objectDeleteKeys, object.Key)
				break
			}
		}
	}
	if len(objectDeleteKeys) > 0 {
		if err := s.repo.DeleteObjects(ctx, objectDeleteKeys); err != nil {
			return Report{}, err
		}
	}

	jobs, err := s.repo.ListJobs(ctx)
	if err != nil {
		return Report{}, err
	}
	var jobDeleteIDs []string
	for _, job := range jobs {
		if job.UpdatedAt.After(cutoff) {
			continue
		}
		switch job.Status {
		case "done", "failed", "cancelled":
			jobDeleteIDs = append(jobDeleteIDs, job.ID)
		}
	}
	if len(jobDeleteIDs) > 0 {
		if err := s.repo.DeleteJobs(ctx, jobDeleteIDs); err != nil {
			return Report{}, err
		}
	}

	return Report{
		DeletedArtifacts: len(artifactDeleteIDs),
		DeletedObjects:   len(objectDeleteKeys),
		DeletedJobs:      len(jobDeleteIDs),
	}, nil
}
