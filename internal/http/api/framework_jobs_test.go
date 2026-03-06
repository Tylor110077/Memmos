package api

import (
	"context"
	"net/http"
	"testing"

	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	appjob "github.com/tylor/goaipj/internal/app/job"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestFrameworkGraphGenerateEndpointCreatesGraphAndTracksJob(t *testing.T) {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	graphRepo := appgraph.NewInMemoryRepository()
	graphService := appgraph.NewService(graphRepo)
	if _, err := graphService.SaveResourceGraph(context.Background(), appgraph.SaveInput{
		GroupID:    group.ID,
		ResourceID: "resource-1",
		Title:      "DI",
		Document: pipelinegraph.Document{
			Summary: "dependency injection",
			Nodes: []pipelinegraph.Node{
				{ID: "root", Name: "Dependency Injection", Type: "topic", Level: 0},
				{ID: "n1", Name: "Provider", Type: "concept", Level: 1},
			},
			Edges: []pipelinegraph.Edge{
				{ID: "e1", SourceID: "root", TargetID: "n1", Relation: "contains"},
			},
		},
	}); err != nil {
		t.Fatalf("SaveResourceGraph() error = %v", err)
	}

	jobService := appjob.NewService(appjob.NewInMemoryRepository())
	server := NewServer(Dependencies{
		GroupService:       groupService,
		GraphService:       graphService,
		JobService:         jobService,
		FrameworkGenerator: pipelinegraph.NewFrameworkGenerator(pipelinegraph.FrameworkRules{MaxTopNodes: 8, MaxDepth: 3}),
	})

	createResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/groups/"+group.ID+"/framework-graph/generate", map[string]any{})
	if createResp.Code != http.StatusAccepted {
		t.Fatalf("generate status = %d body=%s", createResp.Code, createResp.Body.String())
	}

	var created struct {
		GroupID string `json:"group_id"`
		JobID   string `json:"job_id"`
		Status  string `json:"status"`
	}
	decodeJSONResponse(t, createResp, &created)
	if created.GroupID != group.ID {
		t.Fatalf("group_id = %q, want %q", created.GroupID, group.ID)
	}
	if created.JobID == "" {
		t.Fatalf("expected job id")
	}

	jobResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/jobs/"+created.JobID, nil)
	if jobResp.Code != http.StatusOK {
		t.Fatalf("job status = %d body=%s", jobResp.Code, jobResp.Body.String())
	}

	var job struct {
		ID          string  `json:"id"`
		JobType     string  `json:"job_type"`
		Status      string  `json:"status"`
		Attempt     int     `json:"attempt"`
		MaxAttempts int     `json:"max_attempts"`
		LastError   *string `json:"last_error"`
	}
	decodeJSONResponse(t, jobResp, &job)
	if job.ID != created.JobID {
		t.Fatalf("job id = %q, want %q", job.ID, created.JobID)
	}
	if job.JobType != "generate_framework_graph" {
		t.Fatalf("job type = %q", job.JobType)
	}
	if job.Status != "done" {
		t.Fatalf("job status = %q, want done", job.Status)
	}
	if job.Attempt != 1 {
		t.Fatalf("job attempt = %d, want 1", job.Attempt)
	}
	if job.MaxAttempts != 3 {
		t.Fatalf("job max_attempts = %d, want 3", job.MaxAttempts)
	}
	if job.LastError != nil {
		t.Fatalf("expected nil last_error, got %v", *job.LastError)
	}

	frameworkResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups/"+group.ID+"/framework-graph", nil)
	if frameworkResp.Code != http.StatusOK {
		t.Fatalf("framework status = %d body=%s", frameworkResp.Code, frameworkResp.Body.String())
	}

	var framework graphResponse
	decodeJSONResponse(t, frameworkResp, &framework)
	if framework.Graph.ID == "" {
		t.Fatalf("expected generated framework graph")
	}
	if len(framework.Nodes) < 2 {
		t.Fatalf("framework nodes = %d, want at least 2", len(framework.Nodes))
	}
}

func TestFrameworkGraphGenerateEndpointReturnsExistingRunningJob(t *testing.T) {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	jobService := appjob.NewService(appjob.NewInMemoryRepository())
	created, err := jobService.CreateQueuedJob(context.Background(), appjob.CreateQueuedJobInput{
		GroupID:      group.ID,
		JobType:      "generate_framework_graph",
		QueueName:    "graph",
		MaxAttempts:  3,
		Deduplication: group.ID + ":generate_framework_graph",
	})
	if err != nil {
		t.Fatalf("CreateQueuedJob() error = %v", err)
	}

	server := NewServer(Dependencies{
		GroupService:       groupService,
		GraphService:       appgraph.NewService(appgraph.NewInMemoryRepository()),
		JobService:         jobService,
		FrameworkGenerator: pipelinegraph.NewFrameworkGenerator(pipelinegraph.FrameworkRules{MaxTopNodes: 8, MaxDepth: 3}),
	})

	resp := performJSONRequest(t, server, http.MethodPost, "/api/v1/groups/"+group.ID+"/framework-graph/generate", map[string]any{})
	if resp.Code != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", resp.Code, resp.Body.String())
	}

	var body struct {
		GroupID string `json:"group_id"`
		JobID   string `json:"job_id"`
		Status  string `json:"status"`
	}
	decodeJSONResponse(t, resp, &body)
	if body.JobID != created.Job.ID {
		t.Fatalf("job_id = %q, want %q", body.JobID, created.Job.ID)
	}
	if body.Status != string(appjob.StatusQueued) {
		t.Fatalf("status = %q, want queued", body.Status)
	}
}

func TestGetJobEndpointReturnsNotFound(t *testing.T) {
	server := NewServer(Dependencies{
		GroupService: appgroup.NewService(appgroup.NewInMemoryRepository()),
		JobService:   appjob.NewService(appjob.NewInMemoryRepository()),
	})

	resp := performJSONRequest(t, server, http.MethodGet, "/api/v1/jobs/missing", nil)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", resp.Code, resp.Body.String())
	}
}
