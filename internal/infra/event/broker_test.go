package event

import (
	"context"
	"testing"
	"time"
)

func TestBrokerPublishesByGroup(t *testing.T) {
	broker := NewInMemoryBroker()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	group1 := broker.Subscribe(ctx, "group-1")
	group2 := broker.Subscribe(ctx, "group-2")

	broker.Publish(context.Background(), Envelope{
		Name:    NameConversationMessageCreated,
		GroupID: "group-1",
		Payload: map[string]any{"conversation_id": "c1"},
	})

	select {
	case event := <-group1:
		if event.Version != Version {
			t.Fatalf("version = %q, want %q", event.Version, Version)
		}
		if event.Name != NameConversationMessageCreated {
			t.Fatalf("name = %q", event.Name)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected event for group-1 subscriber")
	}

	select {
	case event := <-group2:
		t.Fatalf("unexpected event for group-2: %#v", event)
	case <-time.After(100 * time.Millisecond):
	}
}
