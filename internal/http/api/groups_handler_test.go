package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appchat "github.com/tylor/goaipj/internal/app/chat"
	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	appjob "github.com/tylor/goaipj/internal/app/job"
	appresource "github.com/tylor/goaipj/internal/app/resource"
	domainresource "github.com/tylor/goaipj/internal/domain/resource"
	"github.com/tylor/goaipj/internal/infra/storage"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestGroupsCRUD(t *testing.T) {
	server := newTestServer()

	createResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/groups", map[string]any{
		"name": "Backend Group",
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201 body=%s", createResp.Code, createResp.Body.String())
	}

	var created groupResponse
	decodeJSONResponse(t, createResp, &created)
	if created.Name != "Backend Group" {
		t.Fatalf("created name = %q, want Backend Group", created.Name)
	}

	listResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200", listResp.Code)
	}
	var list []groupResponse
	decodeJSONResponse(t, listResp, &list)
	if len(list) != 1 {
		t.Fatalf("len(list) = %d, want 1", len(list))
	}

	getResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups/"+created.ID, nil)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get status = %d, want 200", getResp.Code)
	}

	updateResp := performJSONRequest(t, server, http.MethodPatch, "/api/v1/groups/"+created.ID, map[string]any{
		"name": "Renamed Group",
	})
	if updateResp.Code != http.StatusOK {
		t.Fatalf("update status = %d, want 200 body=%s", updateResp.Code, updateResp.Body.String())
	}
	var updated groupResponse
	decodeJSONResponse(t, updateResp, &updated)
	if updated.Name != "Renamed Group" {
		t.Fatalf("updated name = %q, want Renamed Group", updated.Name)
	}

	deleteResp := performJSONRequest(t, server, http.MethodDelete, "/api/v1/groups/"+created.ID, nil)
	if deleteResp.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", deleteResp.Code)
	}

	missingResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups/"+created.ID, nil)
	if missingResp.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d, want 404", missingResp.Code)
	}
}

func TestGroupsIncludeResourceCounts(t *testing.T) {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend Group"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	resourceRepo := appresource.NewInMemoryRepository()
	resourceService := appresource.NewService(
		groupService,
		resourceRepo,
		appresource.NewInMemoryArtifactRepository(),
		appresource.NewInMemoryJobPublisher(),
		storage.NewStore(newHTTPFakeBucketClient(), "bucket"),
	)

	completed, err := domainresource.NewWeb(group.ID, "https://example.com/completed")
	if err != nil {
		t.Fatalf("NewWeb() error = %v", err)
	}
	completed.ID = "res-completed"
	if err := completed.MoveTo(domainresource.StatusParsing, domainresource.Failure{}); err != nil {
		t.Fatalf("MoveTo(parsing) error = %v", err)
	}
	if err := completed.MoveTo(domainresource.StatusNormalizing, domainresource.Failure{}); err != nil {
		t.Fatalf("MoveTo(normalizing) error = %v", err)
	}
	if err := completed.MoveTo(domainresource.StatusGraphGenerating, domainresource.Failure{}); err != nil {
		t.Fatalf("MoveTo(graph_generating) error = %v", err)
	}
	if err := completed.MoveTo(domainresource.StatusCompleted, domainresource.Failure{}); err != nil {
		t.Fatalf("MoveTo(completed) error = %v", err)
	}
	if err := resourceRepo.Create(context.Background(), *completed); err != nil {
		t.Fatalf("resourceRepo.Create() error = %v", err)
	}

	uploaded, err := domainresource.NewWeb(group.ID, "https://example.com/uploaded")
	if err != nil {
		t.Fatalf("NewWeb() error = %v", err)
	}
	uploaded.ID = "res-uploaded"
	if err := resourceRepo.Create(context.Background(), *uploaded); err != nil {
		t.Fatalf("resourceRepo.Create() error = %v", err)
	}

	server := NewServer(Dependencies{
		GroupService:    groupService,
		ResourceService: resourceService,
	})

	resp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups/"+group.ID, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.Code, resp.Body.String())
	}

	var got groupResponse
	decodeJSONResponse(t, resp, &got)
	if got.ResourceCount != 2 {
		t.Fatalf("resource_count = %d, want 2", got.ResourceCount)
	}
	if got.CompletedResourceCount != 1 {
		t.Fatalf("completed_resource_count = %d, want 1", got.CompletedResourceCount)
	}
}

func TestGroupsValidationAndErrorShape(t *testing.T) {
	server := newTestServer()

	resp := performJSONRequest(t, server, http.MethodPost, "/api/v1/groups", map[string]any{
		"name": "",
	})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 body=%s", resp.Code, resp.Body.String())
	}

	var body errorResponse
	decodeErrorResponse(t, resp, &body)
	if body.Error.Code == "" {
		t.Fatalf("expected error code")
	}
	if body.Error.Message == "" {
		t.Fatalf("expected error message")
	}
	if !strings.Contains(resp.Body.String(), `"trace_id"`) {
		t.Fatalf("expected trace id")
	}
}

func TestDeleteGroupCascadesAssociatedData(t *testing.T) {
	ctx := context.Background()
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(ctx, appgroup.CreateGroupInput{Name: "Cascade Group"})
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
	chatService := appchat.NewService(
		groupService,
		graphService,
		appchat.NewInMemoryChunkRepository(),
		appchat.NewInMemoryConversationRepository(),
		nil,
		nil,
	)
	jobService := appjob.NewService(appjob.NewInMemoryRepository())

	resourceResult, err := resourceService.CreateWebResource(ctx, appresource.CreateWebResourceInput{
		GroupID: group.ID,
		URL:     "https://example.com/resource",
		Name:    "Resource",
	})
	if err != nil {
		t.Fatalf("CreateWebResource() error = %v", err)
	}
	resourceID := resourceResult.Resource.ID

	resourceGraph, err := graphService.SaveResourceGraph(ctx, appgraph.SaveInput{
		GroupID:    group.ID,
		ResourceID: resourceID,
		Title:      "Resource Graph",
		Document: pipelinegraph.Document{
			Summary: "resource summary",
			Nodes: []pipelinegraph.Node{
				{ID: "root", Name: "Root", Type: "topic", Level: 0},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveResourceGraph() error = %v", err)
	}
	if _, err := graphService.SaveFrameworkGraph(ctx, appgraph.SaveFrameworkInput{
		GroupID: group.ID,
		Title:   "Framework",
		Document: pipelinegraph.Document{
			Summary: "framework summary",
			Nodes: []pipelinegraph.Node{
				{ID: "framework-root", Name: "Framework Root", Type: "topic", Level: 0},
			},
		},
	}); err != nil {
		t.Fatalf("SaveFrameworkGraph() error = %v", err)
	}

	if _, err := chatService.IndexResourceChunks(ctx, appchat.IndexChunksInput{
		GroupID:    group.ID,
		ResourceID: resourceID,
		Chunks:     []string{"chunk one"},
	}); err != nil {
		t.Fatalf("IndexResourceChunks() error = %v", err)
	}
	conversation, err := chatService.CreateConversation(ctx, appchat.CreateConversationInput{
		GroupID:       group.ID,
		GraphID:       resourceGraph.Graph.ID,
		CurrentNodeID: "root",
		Title:         "Conversation",
	})
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}

	jobResult, err := jobService.CreateQueuedJob(ctx, appjob.CreateQueuedJobInput{
		GroupID:    group.ID,
		ResourceID: resourceID,
		JobType:    "generate_framework_graph",
		QueueName:  "graph",
	})
	if err != nil {
		t.Fatalf("CreateQueuedJob() error = %v", err)
	}

	server := NewServer(Dependencies{
		ChatService:     chatService,
		GraphService:    graphService,
		GroupService:    groupService,
		JobService:      jobService,
		ResourceService: resourceService,
	})

	deleteResp := performJSONRequest(t, server, http.MethodDelete, "/api/v1/groups/"+group.ID, nil)
	if deleteResp.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d body=%s", deleteResp.Code, deleteResp.Body.String())
	}

	if resources, err := resourceService.ListResources(ctx, group.ID); err != nil || len(resources) != 0 {
		t.Fatalf("resources after delete = %d err=%v, want 0", len(resources), err)
	}
	if graph, err := graphService.GetResourceGraph(ctx, resourceID, appgraph.QueryOptions{IncludeExpansion: true}); err != nil || graph.Graph.ID != "" {
		t.Fatalf("resource graph after delete = %#v err=%v, want empty", graph.Graph, err)
	}
	if graph, err := graphService.GetFrameworkGraph(ctx, group.ID, appgraph.QueryOptions{IncludeExpansion: true}); err != nil || graph.Graph.ID != "" {
		t.Fatalf("framework graph after delete = %#v err=%v, want empty", graph.Graph, err)
	}
	if _, err := chatService.GetConversation(ctx, conversation.ID); err == nil {
		t.Fatalf("expected conversation removed")
	}
	if _, err := jobService.GetJob(ctx, jobResult.Job.ID); err == nil {
		t.Fatalf("expected job removed")
	}
}

func TestHealthEndpoints(t *testing.T) {
	server := newTestServer()

	healthz := performJSONRequest(t, server, http.MethodGet, "/healthz", nil)
	if healthz.Code != http.StatusOK {
		t.Fatalf("healthz = %d, want 200", healthz.Code)
	}

	readyz := performJSONRequest(t, server, http.MethodGet, "/readyz", nil)
	if readyz.Code != http.StatusOK {
		t.Fatalf("readyz = %d, want 200", readyz.Code)
	}
}

func newTestServer() http.Handler {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	return NewServer(Dependencies{
		GroupService: groupService,
	})
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp
}

func decodeJSONResponse(t *testing.T, resp *httptest.ResponseRecorder, dst any) {
	t.Helper()
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err == nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, dst); err != nil {
			t.Fatalf("json.Unmarshal(data) error = %v body=%s", err, resp.Body.String())
		}
		return
	}
	if err := json.Unmarshal(resp.Body.Bytes(), dst); err != nil {
		t.Fatalf("json.Unmarshal() error = %v body=%s", err, resp.Body.String())
	}
}

func decodeErrorResponse(t *testing.T, resp *httptest.ResponseRecorder, dst any) {
	t.Helper()
	var envelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("json.Unmarshal() error = %v body=%s", err, resp.Body.String())
	}
	switch target := dst.(type) {
	case *errorResponse:
		target.Error = envelope.Error
	default:
		t.Fatalf("unsupported decodeErrorResponse target %T", dst)
	}
}
