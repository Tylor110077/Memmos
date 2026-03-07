package api

import (
	"context"
	"net/http"
	"strings"
	"testing"

	appchat "github.com/tylor/goaipj/internal/app/chat"
	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	pipelinechat "github.com/tylor/goaipj/internal/pipeline/chat"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestConversationMessageStreamEndpoint(t *testing.T) {
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
	graphRepo.AddExample(saved.Graph.ID, appgraph.Example{
		ID:      "example-1",
		NodeID:  "node-parser",
		Content: "Parser converts uploaded files into normalized text.",
	})

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
			"The worker consumes asynchronous processing jobs.",
		},
	}); err != nil {
		t.Fatalf("IndexResourceChunks() error = %v", err)
	}

	server := NewServer(Dependencies{
		GroupService: groupService,
		GraphService: graphService,
		ChatService:  chatService,
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

	streamResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/conversations/"+created.ID+"/messages", map[string]any{
		"content": "这个节点的作用是什么？",
		"stream":  true,
	})
	if streamResp.Code != http.StatusOK {
		t.Fatalf("stream status = %d body=%s", streamResp.Code, streamResp.Body.String())
	}
	if contentType := streamResp.Header().Get("Content-Type"); !strings.Contains(contentType, "text/event-stream") {
		t.Fatalf("content-type = %q, want text/event-stream", contentType)
	}
	body := streamResp.Body.String()
	if !strings.Contains(body, "event: assistant.delta") {
		t.Fatalf("expected assistant.delta events, body=%s", body)
	}
	if !strings.Contains(body, "event: assistant.message") {
		t.Fatalf("expected assistant.message event, body=%s", body)
	}
	if !strings.Contains(body, "event: done") {
		t.Fatalf("expected done event, body=%s", body)
	}
}
