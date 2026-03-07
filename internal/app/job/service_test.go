package job

import (
	"context"
	"testing"
)

func TestCreateQueuedJobWritesInitialEvent(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)

	created, err := service.CreateQueuedJob(context.Background(), CreateQueuedJobInput{
		GroupID:    "group-1",
		ResourceID: "resource-1",
		JobType:    "parse_resource",
		QueueName:  "default",
		Payload:    map[string]any{"resource_id": "resource-1"},
	})
	if err != nil {
		t.Fatalf("CreateQueuedJob() error = %v", err)
	}

	if created.Job.Status != StatusQueued {
		t.Fatalf("status = %s, want queued", created.Job.Status)
	}
	if len(created.Events) != 1 {
		t.Fatalf("events = %d, want 1", len(created.Events))
	}
	if created.Events[0].Type != EventQueued {
		t.Fatalf("event type = %s, want queued", created.Events[0].Type)
	}
}

func TestUpdateJobStatusTracksHistory(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)

	created, err := service.CreateQueuedJob(context.Background(), CreateQueuedJobInput{
		GroupID:    "group-1",
		ResourceID: "resource-1",
		JobType:    "parse_resource",
		QueueName:  "default",
	})
	if err != nil {
		t.Fatalf("CreateQueuedJob() error = %v", err)
	}

	if _, err := service.MarkRunning(context.Background(), created.Job.ID); err != nil {
		t.Fatalf("MarkRunning() error = %v", err)
	}
	updated, err := service.MarkFailed(context.Background(), created.Job.ID, "boom")
	if err != nil {
		t.Fatalf("MarkFailed() error = %v", err)
	}

	if updated.Job.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", updated.Job.Status)
	}
	if updated.Job.ErrorMessage != "boom" {
		t.Fatalf("error = %q, want boom", updated.Job.ErrorMessage)
	}

	events, err := repo.ListEvents(context.Background(), created.Job.ID)
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("events = %d, want 3", len(events))
	}
}

func TestCancelAndResumeJob(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)

	created, err := service.CreateQueuedJob(context.Background(), CreateQueuedJobInput{
		GroupID:    "group-1",
		ResourceID: "resource-1",
		JobType:    "parse_resource",
		QueueName:  "default",
	})
	if err != nil {
		t.Fatalf("CreateQueuedJob() error = %v", err)
	}

	cancelled, err := service.CancelJob(context.Background(), created.Job.ID)
	if err != nil {
		t.Fatalf("CancelJob() error = %v", err)
	}
	if cancelled.Job.Status != StatusCancelled {
		t.Fatalf("status = %s, want cancelled", cancelled.Job.Status)
	}

	resumed, err := service.ResumeJob(context.Background(), created.Job.ID)
	if err != nil {
		t.Fatalf("ResumeJob() error = %v", err)
	}
	if resumed.Job.Status != StatusQueued {
		t.Fatalf("status = %s, want queued", resumed.Job.Status)
	}

	events, err := repo.ListEvents(context.Background(), created.Job.ID)
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("events = %d, want 3", len(events))
	}
	if events[1].Type != EventCancelled {
		t.Fatalf("cancel event = %s, want cancelled", events[1].Type)
	}
	if events[2].Type != EventQueued {
		t.Fatalf("resume event = %s, want queued", events[2].Type)
	}
}
