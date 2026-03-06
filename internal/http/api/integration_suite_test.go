package api

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	appchat "github.com/tylor/goaipj/internal/app/chat"
	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	appresource "github.com/tylor/goaipj/internal/app/resource"
	domainresource "github.com/tylor/goaipj/internal/domain/resource"
	"github.com/tylor/goaipj/internal/infra/storage"
	pipelinechat "github.com/tylor/goaipj/internal/pipeline/chat"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
	"github.com/tylor/goaipj/internal/worker"
)

func TestResourceProcessingFlowIntegration(t *testing.T) {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	resourceService := appresource.NewService(
		groupService,
		appresource.NewInMemoryRepository(),
		appresource.NewInMemoryArtifactRepository(),
		appresource.NewInMemoryJobPublisher(),
		storage.NewStore(newHTTPFakeBucketClient(), "bucket"),
	)
	graphService := appgraph.NewService(appgraph.NewInMemoryRepository())

	created, err := resourceService.CreateWebResource(context.Background(), appresource.CreateWebResourceInput{
		GroupID: group.ID,
		URL:     "https://example.com/resource",
	})
	if err != nil {
		t.Fatalf("CreateWebResource() error = %v", err)
	}
	if created.Job.ID == "" {
		t.Fatalf("expected created job")
	}
	if _, err := resourceService.UpdateStatus(context.Background(), created.Resource.ID, "parsing"); err != nil {
		t.Fatalf("UpdateStatus(parsing) error = %v", err)
	}
	if _, err := resourceService.UpdateStatus(context.Background(), created.Resource.ID, "normalizing"); err != nil {
		t.Fatalf("UpdateStatus(normalizing) error = %v", err)
	}

	handler := worker.NewGenerateGraphHandler(resourceService, graphService, pipelinegraph.NewGenerator())
	payload, err := json.Marshal(worker.GenerateGraphPayload{
		GroupID:    group.ID,
		ResourceID: created.Resource.ID,
		Title:      "Processed Resource",
		Summary:    "Parsed summary",
		Markdown:   "# Processed Resource\n\nParser converts uploads into graph-ready content.",
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if err := handler(context.Background(), payload); err != nil {
		t.Fatalf("handler() error = %v", err)
	}

	resourceItem, err := resourceService.GetResource(context.Background(), group.ID, created.Resource.ID)
	if err != nil {
		t.Fatalf("GetResource() error = %v", err)
	}
	if resourceItem.Status != domainresource.StatusCompleted {
		t.Fatalf("status = %s, want completed", resourceItem.Status)
	}

	server := NewServer(Dependencies{
		GroupService:    groupService,
		ResourceService: resourceService,
		GraphService:    graphService,
	})

	graphResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/resources/"+created.Resource.ID+"/graph", nil)
	if graphResp.Code != http.StatusOK {
		t.Fatalf("graph status = %d body=%s", graphResp.Code, graphResp.Body.String())
	}
}

func TestExpandNodeIntegration(t *testing.T) {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	resourceService := appresource.NewService(
		groupService,
		appresource.NewInMemoryRepository(),
		appresource.NewInMemoryArtifactRepository(),
		appresource.NewInMemoryJobPublisher(),
		storage.NewStore(newHTTPFakeBucketClient(), "bucket"),
	)
	resourceCreated, err := resourceService.CreateWebResource(context.Background(), appresource.CreateWebResourceInput{
		GroupID: group.ID,
		URL:     "https://example.com/page",
	})
	if err != nil {
		t.Fatalf("CreateWebResource() error = %v", err)
	}
	graphService := appgraph.NewService(appgraph.NewInMemoryRepository())
	saved, err := graphService.SaveResourceGraph(context.Background(), appgraph.SaveInput{
		GroupID:    group.ID,
		ResourceID: resourceCreated.Resource.ID,
		Title:      "Resource",
		Document: pipelinegraph.Document{
			Summary: "summary",
			Nodes: []pipelinegraph.Node{
				{ID: "root", Name: "Backend", Type: "topic", Level: 0},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveResourceGraph() error = %v", err)
	}

	server := NewServer(Dependencies{
		GroupService:       groupService,
		GraphService:       graphService,
		ResourceService:    resourceService,
		ExpansionGenerator: pipelinegraph.NewExpansionGenerator(pipelinegraph.ExpansionRules{MaxNewNodes: 1}),
	})

	resp := performJSONRequest(t, server, http.MethodPost, "/api/v1/graphs/"+saved.Graph.ID+"/nodes/root/expand", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.Code, resp.Body.String())
	}
	graphResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/resources/"+resourceCreated.Resource.ID+"/graph?include_expansion=true", nil)
	if graphResp.Code != http.StatusOK {
		t.Fatalf("graph status = %d body=%s", graphResp.Code, graphResp.Body.String())
	}
	var body graphResponse
	decodeJSONResponse(t, graphResp, &body)
	if len(body.Nodes) != 2 {
		t.Fatalf("nodes = %d, want 2", len(body.Nodes))
	}
}

func TestConversationFlowIntegration(t *testing.T) {
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
	})
	if messageResp.Code != http.StatusOK {
		t.Fatalf("message status = %d body=%s", messageResp.Code, messageResp.Body.String())
	}
	var asked conversationAskResponse
	decodeJSONResponse(t, messageResp, &asked)
	if len(asked.AssistantMessage.CitedChunkIDs) == 0 {
		t.Fatalf("expected cited chunks")
	}
	if len(asked.AssistantMessage.CitedNodeIDs) == 0 {
		t.Fatalf("expected cited nodes")
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
	if len(stored.Messages[1].CitedChunkIDs) == 0 {
		t.Fatalf("expected cited chunk ids to persist")
	}
}
