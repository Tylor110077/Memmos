package api

import (
	"context"
	"net/http"
	"testing"

	appchat "github.com/tylor/goaipj/internal/app/chat"
	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	pipelinechat "github.com/tylor/goaipj/internal/pipeline/chat"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestConversationEndpoints(t *testing.T) {
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

	messageResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/conversations/"+created.ID+"/messages", map[string]any{
		"content": "这个节点的作用是什么？",
		"stream":  false,
	})
	if messageResp.Code != http.StatusOK {
		t.Fatalf("message status = %d body=%s", messageResp.Code, messageResp.Body.String())
	}

	var asked conversationAskResponse
	decodeJSONResponse(t, messageResp, &asked)
	if asked.AssistantMessage.Content == "" {
		t.Fatalf("expected assistant answer")
	}
	if len(asked.AssistantMessage.CitedChunkIDs) == 0 {
		t.Fatalf("expected cited chunks")
	}

	getResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/conversations/"+created.ID, nil)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get status = %d body=%s", getResp.Code, getResp.Body.String())
	}

	var stored struct {
		Conversation conversationSummaryResponse   `json:"conversation"`
		Messages     []conversationMessageResponse `json:"messages"`
	}
	decodeJSONResponse(t, getResp, &stored)
	if stored.Conversation.ID != created.ID {
		t.Fatalf("conversation id = %q, want %q", stored.Conversation.ID, created.ID)
	}
	if len(stored.Messages) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(stored.Messages))
	}
}
