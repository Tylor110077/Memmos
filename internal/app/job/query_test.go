package job

import (
	"context"
	"testing"
)

func TestGetJobReturnsCreatedRecord(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)

	created, err := service.CreateQueuedJob(context.Background(), CreateQueuedJobInput{
		GroupID:     "group-1",
		JobType:     "generate_framework_graph",
		QueueName:   "graph",
		MaxAttempts: 3,
	})
	if err != nil {
		t.Fatalf("CreateQueuedJob() error = %v", err)
	}

	job, err := service.GetJob(context.Background(), created.Job.ID)
	if err != nil {
		t.Fatalf("GetJob() error = %v", err)
	}
	if job.ID != created.Job.ID {
		t.Fatalf("job id = %q, want %q", job.ID, created.Job.ID)
	}
	if job.MaxAttempts != 3 {
		t.Fatalf("max attempts = %d, want 3", job.MaxAttempts)
	}
}

func TestFindActiveJobByDeduplication(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)

	created, err := service.CreateQueuedJob(context.Background(), CreateQueuedJobInput{
		GroupID:       "group-1",
		JobType:       "generate_framework_graph",
		QueueName:     "graph",
		MaxAttempts:   3,
		Deduplication: "group-1:generate_framework_graph",
	})
	if err != nil {
		t.Fatalf("CreateQueuedJob() error = %v", err)
	}

	job, ok, err := service.FindActiveJob(context.Background(), "group-1:generate_framework_graph")
	if err != nil {
		t.Fatalf("FindActiveJob() error = %v", err)
	}
	if !ok {
		t.Fatalf("expected active job")
	}
	if job.ID != created.Job.ID {
		t.Fatalf("job id = %q, want %q", job.ID, created.Job.ID)
	}
}
