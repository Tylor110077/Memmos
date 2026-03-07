package api

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	appchat "github.com/tylor/goaipj/internal/app/chat"
	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	"github.com/tylor/goaipj/internal/infra/llm"
	pipelinechat "github.com/tylor/goaipj/internal/pipeline/chat"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestConversationEndpointsWithAliyunBailian(t *testing.T) {
	apiKey := strings.TrimSpace(os.Getenv("ALIYUN_BAILIAN_API_KEY"))
	if apiKey == "" {
		if path := strings.TrimSpace(os.Getenv("ALIYUN_BAILIAN_API_KEY_FILE")); path != "" {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read api key file: %v", err)
			}
			apiKey = strings.TrimSpace(string(raw))
		}
	}
	if apiKey == "" {
		t.Skip("ALIYUN_BAILIAN_API_KEY or ALIYUN_BAILIAN_API_KEY_FILE is required")
	}

	client := llm.NewAliyunBailianClient(llm.Options{
		APIKey:         apiKey,
		BaseURL:        os.Getenv("ALIYUN_BAILIAN_BASE_URL"),
		ChatModel:      os.Getenv("ALIYUN_BAILIAN_MODEL"),
		EmbeddingModel: os.Getenv("ALIYUN_BAILIAN_EMBEDDING_MODEL"),
	})

	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend Integration"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	graphRepo := appgraph.NewInMemoryRepository()
	graphService := appgraph.NewService(graphRepo)
	saved, err := graphService.SaveResourceGraph(context.Background(), appgraph.SaveInput{
		GroupID:    group.ID,
		ResourceID: "resource-real-1",
		Title:      "Backend Pipeline",
		Document: pipelinegraph.Document{
			Summary: "The backend parses uploaded resources, normalizes text, generates graphs, and lets users ask focused questions around nodes.",
			Nodes: []pipelinegraph.Node{
				{ID: "node-parser", Name: "Parser", Type: "component", Level: 0, Description: "Extracts text from resources."},
				{ID: "node-worker", Name: "Worker", Type: "component", Level: 1, Description: "Runs async jobs."},
			},
			Edges: []pipelinegraph.Edge{
				{ID: "edge-parser-worker", SourceID: "node-parser", TargetID: "node-worker", Relation: "feeds"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveResourceGraph() error = %v", err)
	}
	graphRepo.AddExample(saved.Graph.ID, appgraph.Example{
		ID:      "example-parser-1",
		NodeID:  "node-parser",
		Content: "A PDF upload is parsed into normalized markdown before graph generation.",
	})

	chatService := appchat.NewService(
		groupService,
		graphService,
		appchat.NewInMemoryChunkRepository(),
		appchat.NewInMemoryConversationRepository(),
		pipelinechat.NewLLMAnswerer(client),
		client,
	)
	if _, err := chatService.IndexResourceChunks(context.Background(), appchat.IndexChunksInput{
		GroupID:    group.ID,
		ResourceID: saved.Graph.ResourceID,
		Chunks: []string{
			"The parser extracts text from uploaded files and creates normalized markdown.",
			"The worker consumes queued processing jobs asynchronously and triggers graph generation.",
			"Focused node conversations use retrieved group context and neighbor nodes.",
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
		"title":           "Real provider integration",
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", createResp.Code, createResp.Body.String())
	}
	var created conversationResponse
	decodeJSONResponse(t, createResp, &created)

	messageResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/conversations/"+created.ID+"/messages", map[string]any{
		"content": "Parser 在这条资源处理链路里的作用是什么？请顺带给一个例子。",
		"stream":  false,
	})
	if messageResp.Code != http.StatusOK {
		t.Fatalf("message status = %d body=%s", messageResp.Code, messageResp.Body.String())
	}
	var asked conversationAskResponse
	decodeJSONResponse(t, messageResp, &asked)
	if strings.TrimSpace(asked.AssistantMessage.Content) == "" {
		t.Fatalf("expected assistant answer")
	}
	if len(asked.AssistantMessage.CitedChunkIDs) == 0 {
		t.Fatalf("expected cited chunk ids")
	}
	if len(asked.AssistantMessage.CitedNodeIDs) == 0 {
		t.Fatalf("expected cited node ids")
	}
	if len(asked.AssistantMessage.Examples) == 0 {
		t.Fatalf("expected examples")
	}

	getResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/conversations/"+created.ID, nil)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get status = %d body=%s", getResp.Code, getResp.Body.String())
	}
	var stored conversationResponse
	decodeJSONResponse(t, getResp, &stored)
	if len(stored.Messages) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(stored.Messages))
	}
	if stored.Messages[1].CreatedAt == "" {
		t.Fatalf("expected assistant created_at")
	}
	if _, err := time.Parse(time.RFC3339Nano, stored.Messages[1].CreatedAt); err != nil {
		t.Fatalf("assistant created_at parse error = %v", err)
	}
}
