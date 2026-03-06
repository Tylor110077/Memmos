package api

import (
	"context"
	"net/http"
	"testing"

	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	appresource "github.com/tylor/goaipj/internal/app/resource"
	"github.com/tylor/goaipj/internal/infra/storage"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestGetResourceGraphEndpoint(t *testing.T) {
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
	if _, err := graphService.SaveResourceGraph(context.Background(), appgraph.SaveInput{
		GroupID:    group.ID,
		ResourceID: resourceCreated.Resource.ID,
		Title:      "Backend",
		Document: pipelinegraph.Document{
			Summary: "summary",
			Nodes: []pipelinegraph.Node{
				{ID: "root", Name: "Backend", Type: "topic", Level: 0},
				{ID: "n1", Name: "Child", Type: "concept", Level: 1},
			},
			Edges: []pipelinegraph.Edge{
				{ID: "e1", SourceID: "root", TargetID: "n1", Relation: "contains"},
			},
		},
	}); err != nil {
		t.Fatalf("SaveResourceGraph() error = %v", err)
	}

	server := NewServer(Dependencies{
		GroupService:    groupService,
		ResourceService: resourceService,
		GraphService:    graphService,
	})

	resp := performJSONRequest(t, server, http.MethodGet, "/api/v1/resources/"+resourceCreated.Resource.ID+"/graph?max_level=1", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.Code, resp.Body.String())
	}

	var body graphResponse
	decodeJSONResponse(t, resp, &body)
	if body.Graph.ResourceID != resourceCreated.Resource.ID {
		t.Fatalf("resource id = %q", body.Graph.ResourceID)
	}
	if len(body.Nodes) != 2 {
		t.Fatalf("nodes = %d, want 2", len(body.Nodes))
	}
}
