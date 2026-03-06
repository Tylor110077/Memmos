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

func TestExpandNodeEndpoint(t *testing.T) {
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
