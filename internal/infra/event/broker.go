package event

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const Version = "v1"

const (
	NameResourceStatusChanged      = "resource.status.changed"
	NameResourceCompleted          = "resource.completed"
	NameResourceFailed             = "resource.failed"
	NameFrameworkGraphUpdated      = "framework_graph.updated"
	NameGraphNodeExpanded          = "graph.node.expanded"
	NameConversationMessageCreated = "conversation.message.created"
)

type Envelope struct {
	ID         string         `json:"id"`
	Version    string         `json:"version"`
	Name       string         `json:"name"`
	GroupID    string         `json:"group_id"`
	OccurredAt time.Time      `json:"occurred_at"`
	Payload    map[string]any `json:"payload"`
}

type Broker interface {
	Publish(ctx context.Context, event Envelope)
	Subscribe(ctx context.Context, groupID string) <-chan Envelope
}

type InMemoryBroker struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan Envelope]struct{}
}

func NewInMemoryBroker() *InMemoryBroker {
	return &InMemoryBroker{
		subscribers: map[string]map[chan Envelope]struct{}{},
	}
}

func (b *InMemoryBroker) Publish(_ context.Context, event Envelope) {
	if event.ID == "" {
		event.ID = newID()
	}
	if event.Version == "" {
		event.Version = Version
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}

	b.mu.RLock()
	targets := make([]chan Envelope, 0, len(b.subscribers[event.GroupID]))
	for ch := range b.subscribers[event.GroupID] {
		targets = append(targets, ch)
	}
	b.mu.RUnlock()

	for _, ch := range targets {
		select {
		case ch <- event:
		default:
		}
	}
}

func (b *InMemoryBroker) Subscribe(ctx context.Context, groupID string) <-chan Envelope {
	ch := make(chan Envelope, 16)

	b.mu.Lock()
	if b.subscribers[groupID] == nil {
		b.subscribers[groupID] = map[chan Envelope]struct{}{}
	}
	b.subscribers[groupID][ch] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		b.mu.Lock()
		if items, ok := b.subscribers[groupID]; ok {
			delete(items, ch)
			if len(items) == 0 {
				delete(b.subscribers, groupID)
			}
		}
		b.mu.Unlock()
		close(ch)
	}()

	return ch
}

func newID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(raw[:])
}
