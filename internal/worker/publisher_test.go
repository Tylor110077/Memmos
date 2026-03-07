package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestPublisherEnqueuesTaskWithPayload(t *testing.T) {
	client := &fakeClient{}
	publisher := NewPublisher(client)

	info, err := publisher.Publish(context.Background(), PublishInput{
		TaskType:   TaskParseResource,
		QueueName:  "default",
		ResourceID: "resource-1",
		Payload:    map[string]any{"resource_id": "resource-1"},
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if client.lastType != TaskParseResource {
		t.Fatalf("task type = %q, want %q", client.lastType, TaskParseResource)
	}
	if client.lastQueue != "default" {
		t.Fatalf("queue = %q, want default", client.lastQueue)
	}
	if info.ID != "task-1" {
		t.Fatalf("task id = %q, want task-1", info.ID)
	}

	var payload map[string]any
	if err := json.Unmarshal(client.lastPayload, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if payload["resource_id"] != "resource-1" {
		t.Fatalf("resource_id = %v, want resource-1", payload["resource_id"])
	}
}

type fakeClient struct {
	lastType    string
	lastPayload []byte
	lastQueue   string
}

func (f *fakeClient) Enqueue(_ context.Context, taskType string, payload []byte, queue string) (TaskInfo, error) {
	f.lastType = taskType
	f.lastPayload = payload
	f.lastQueue = queue
	return TaskInfo{ID: "task-1", Queue: queue, State: "active", MaxRetry: 25, NextProcessAt: time.Now().UTC()}, nil
}
