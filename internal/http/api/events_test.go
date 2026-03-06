package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appchat "github.com/tylor/goaipj/internal/app/chat"
	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	infraevent "github.com/tylor/goaipj/internal/infra/event"
	pipelinechat "github.com/tylor/goaipj/internal/pipeline/chat"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestGroupEventsEndpointStreamsEnvelope(t *testing.T) {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	broker := infraevent.NewInMemoryBroker()

	reqCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := NewServer(Dependencies{
		GroupService: groupService,
		EventBroker:  broker,
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/groups/"+group.ID, nil).WithContext(reqCtx)
	resp := httptest.NewRecorder()
	done := make(chan struct{}, 1)
	go func() {
		handler.ServeHTTP(resp, req)
		done <- struct{}{}
	}()
	time.Sleep(50 * time.Millisecond)
	broker.Publish(context.Background(), infraevent.Envelope{
		Name:    infraevent.NameGraphNodeExpanded,
		GroupID: group.ID,
		Payload: map[string]any{"graph_id": "g1", "node_id": "n1"},
	})
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for handler to finish")
	}
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.Code, resp.Body.String())
	}
	bodyText := resp.Body.String()
	if !strings.Contains(bodyText, "event: "+infraevent.NameGraphNodeExpanded) {
		t.Fatalf("expected event name in body=%s", bodyText)
	}
	idx := strings.Index(bodyText, "data: ")
	if idx < 0 {
		t.Fatalf("expected data payload in body=%s", bodyText)
	}
	raw := bodyText[idx+len("data: "):]
	raw = strings.TrimSpace(strings.SplitN(raw, "\n", 2)[0])
	var body infraevent.Envelope
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v raw=%s", err, raw)
	}
	if body.Version != infraevent.Version {
		t.Fatalf("version = %q", body.Version)
	}
	if body.Name != infraevent.NameGraphNodeExpanded {
		t.Fatalf("name = %q", body.Name)
	}
	if body.GroupID != group.ID {
		t.Fatalf("group_id = %q", body.GroupID)
	}
}

func TestConversationMessagePublishesGroupEvent(t *testing.T) {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	graphRepo := appgraph.NewInMemoryRepository()
	graphService := appgraph.NewService(graphRepo)
	saved, err := graphService.SaveResourceGraph(context.Background(), appgraph.SaveInput{
		GroupID:    group.ID,
		ResourceID: "resource-1",
		Title:      "Parser Graph",
		Document: pipelinegraph.Document{
			Summary: "Parser and worker process uploaded resources.",
			Nodes: []pipelinegraph.Node{
				{ID: "node-parser", Name: "Parser", Type: "component", Level: 0},
				{ID: "node-worker", Name: "Worker", Type: "component", Level: 1},
			},
			Edges: []pipelinegraph.Edge{
				{ID: "edge-1", SourceID: "node-parser", TargetID: "node-worker", Relation: "handled_by"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveResourceGraph() error = %v", err)
	}
	chatService := appchat.NewService(
		groupService,
		graphService,
		appchat.NewInMemoryChunkRepository(),
		appchat.NewInMemoryConversationRepository(),
		pipelinechat.NewAnswerer(),
		nil,
	)
	if _, err := chatService.IndexResourceChunks(context.Background(), appchat.IndexChunksInput{
		GroupID:    group.ID,
		ResourceID: saved.Graph.ResourceID,
		Chunks: []string{
			"The parser extracts text before background jobs generate graphs.",
		},
	}); err != nil {
		t.Fatalf("IndexResourceChunks() error = %v", err)
	}

	broker := infraevent.NewInMemoryBroker()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sub := broker.Subscribe(ctx, group.ID)

	server := NewServer(Dependencies{
		GroupService: groupService,
		GraphService: graphService,
		ChatService:  chatService,
		EventBroker:  broker,
	})

	createResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/conversations", map[string]any{
		"group_id":        group.ID,
		"graph_id":        saved.Graph.ID,
		"current_node_id": "node-parser",
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", createResp.Code, createResp.Body.String())
	}
	var created conversationResponse
	decodeJSONResponse(t, createResp, &created)

	messageResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/conversations/"+created.ID+"/messages", map[string]any{
		"content": "这个节点的作用是什么？",
	})
	if messageResp.Code != http.StatusOK {
		t.Fatalf("message status = %d body=%s", messageResp.Code, messageResp.Body.String())
	}

	select {
	case event := <-sub:
		if event.Name != infraevent.NameConversationMessageCreated {
			t.Fatalf("name = %q", event.Name)
		}
		if event.Payload["conversation_id"] != created.ID {
			t.Fatalf("conversation_id = %#v", event.Payload["conversation_id"])
		}
	case <-time.After(time.Second):
		t.Fatalf("expected published conversation event")
	}
}
