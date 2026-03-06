package api

import (
	"context"
	"net/http"
	"strings"
	"testing"

	appchat "github.com/tylor/goaipj/internal/app/chat"
	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	"github.com/tylor/goaipj/internal/infra/observability"
	pipelinechat "github.com/tylor/goaipj/internal/pipeline/chat"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestTraceCapturesHTTPAndSSEFlow(t *testing.T) {
	tracer := observability.NewInMemoryTracer()
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Trace Group"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	graphRepo := appgraph.NewInMemoryRepository()
	graphService := appgraph.NewService(graphRepo)
	saved, err := graphService.SaveResourceGraph(context.Background(), appgraph.SaveInput{
		GroupID:    group.ID,
		ResourceID: "resource-1",
		Title:      "Trace Graph",
		Document: pipelinegraph.Document{
			Summary: "trace summary",
			Nodes: []pipelinegraph.Node{
				{ID: "root", Name: "Root", Type: "topic", Level: 0},
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
		Chunks:     []string{"traceable content"},
	}); err != nil {
		t.Fatalf("IndexResourceChunks() error = %v", err)
	}

	server := NewServer(Dependencies{
		GroupService: groupService,
		GraphService: graphService,
		ChatService:  chatService,
		Tracer:       tracer,
	})

	createResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/conversations", map[string]any{
		"group_id":        group.ID,
		"graph_id":        saved.Graph.ID,
		"current_node_id": "root",
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", createResp.Code, createResp.Body.String())
	}
	var created conversationResponse
	decodeJSONResponse(t, createResp, &created)

	streamResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/conversations/"+created.ID+"/messages", map[string]any{
		"content": "讲讲这个节点",
		"stream":  true,
	})
	if streamResp.Code != http.StatusOK {
		t.Fatalf("stream status = %d body=%s", streamResp.Code, streamResp.Body.String())
	}

	spans := tracer.FinishedSpans()
	assertSpanNames(t, spans, []string{
		"http.request",
		"chat.ask",
		"chat.answer.generate",
		"sse.message.stream",
	})
}

func TestErrorReporterCapturesInternalErrorOnly(t *testing.T) {
	reporter := observability.NewInMemoryErrorReporter()
	server := NewServer(Dependencies{
		GroupService:   appgroup.NewService(appgroup.NewInMemoryRepository()),
		ErrorReporter:  reporter,
	})

	internalResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups/group-1/framework-graph", nil)
	if internalResp.Code != http.StatusInternalServerError {
		t.Fatalf("internal status = %d body=%s", internalResp.Code, internalResp.Body.String())
	}
	reports := reporter.Reports()
	if len(reports) != 1 {
		t.Fatalf("reports = %d, want 1", len(reports))
	}
	if !strings.Contains(reports[0].Message, "graph service not configured") {
		t.Fatalf("report message = %q", reports[0].Message)
	}

	notFoundResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups/missing", nil)
	if notFoundResp.Code != http.StatusNotFound {
		t.Fatalf("not found status = %d body=%s", notFoundResp.Code, notFoundResp.Body.String())
	}
	if len(reporter.Reports()) != 1 {
		t.Fatalf("expected 404 not to be reported, reports=%d", len(reporter.Reports()))
	}
}

func assertSpanNames(t *testing.T, spans []observability.FinishedSpan, want []string) {
	t.Helper()
	seen := map[string]bool{}
	for _, span := range spans {
		seen[span.Name] = true
	}
	for _, name := range want {
		if !seen[name] {
			t.Fatalf("missing span %q in %#v", name, spans)
		}
	}
}
