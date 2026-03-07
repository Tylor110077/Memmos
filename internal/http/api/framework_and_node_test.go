package api

import (
	"context"
	"net/http"
	"testing"

	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestFrameworkGraphAndNodeDetailEndpoints(t *testing.T) {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	graphService := appgraph.NewService(appgraph.NewInMemoryRepository())

	framework, err := graphService.SaveFrameworkGraph(context.Background(), appgraph.SaveFrameworkInput{
		GroupID: group.ID,
		Title:   "Framework",
		Document: pipelinegraph.Document{
			Summary: "summary",
			Nodes: []pipelinegraph.Node{
				{ID: "root", Name: "Backend", Type: "topic", Level: 0},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveFrameworkGraph() error = %v", err)
	}
	graphService.Repository().AddExample(framework.Graph.ID, appgraph.Example{
		ID:      "ex1",
		NodeID:  "root",
		Content: "Backend example",
	})

	server := NewServer(Dependencies{
		GroupService: groupService,
		GraphService: graphService,
	})

	frameworkResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups/"+group.ID+"/framework-graph", nil)
	if frameworkResp.Code != http.StatusOK {
		t.Fatalf("framework status = %d body=%s", frameworkResp.Code, frameworkResp.Body.String())
	}

	nodeResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/graphs/"+framework.Graph.ID+"/nodes/root", nil)
	if nodeResp.Code != http.StatusOK {
		t.Fatalf("node status = %d body=%s", nodeResp.Code, nodeResp.Body.String())
	}
}

func TestGraphVersionHistoryEndpoints(t *testing.T) {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	graphService := appgraph.NewService(appgraph.NewInMemoryRepository())

	for _, title := range []string{"Framework 1", "Framework 2"} {
		if _, err := graphService.SaveFrameworkGraph(context.Background(), appgraph.SaveFrameworkInput{
			GroupID: group.ID,
			Title:   title,
			Document: pipelinegraph.Document{
				Summary: title,
				Nodes: []pipelinegraph.Node{
					{ID: title, Name: title, Type: "topic", Level: 0},
				},
			},
		}); err != nil {
			t.Fatalf("SaveFrameworkGraph(%s) error = %v", title, err)
		}
	}
	for _, title := range []string{"Resource 1", "Resource 2"} {
		if _, err := graphService.SaveResourceGraph(context.Background(), appgraph.SaveInput{
			GroupID:    group.ID,
			ResourceID: "resource-1",
			Title:      title,
			Document: pipelinegraph.Document{
				Summary: title,
				Nodes: []pipelinegraph.Node{
					{ID: title, Name: title, Type: "topic", Level: 0},
				},
			},
		}); err != nil {
			t.Fatalf("SaveResourceGraph(%s) error = %v", title, err)
		}
	}

	server := NewServer(Dependencies{
		GroupService: groupService,
		GraphService: graphService,
	})

	frameworkResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups/"+group.ID+"/framework-graph/versions", nil)
	if frameworkResp.Code != http.StatusOK {
		t.Fatalf("framework versions status = %d body=%s", frameworkResp.Code, frameworkResp.Body.String())
	}
	var frameworkVersions []graphInfoResponse
	decodeJSONResponse(t, frameworkResp, &frameworkVersions)
	if len(frameworkVersions) != 2 {
		t.Fatalf("framework versions = %d, want 2", len(frameworkVersions))
	}
	if frameworkVersions[0].Version != 2 {
		t.Fatalf("latest framework version = %d, want 2", frameworkVersions[0].Version)
	}

	resourceResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/resources/resource-1/graph/versions", nil)
	if resourceResp.Code != http.StatusOK {
		t.Fatalf("resource versions status = %d body=%s", resourceResp.Code, resourceResp.Body.String())
	}
	var resourceVersions []graphInfoResponse
	decodeJSONResponse(t, resourceResp, &resourceVersions)
	if len(resourceVersions) != 2 {
		t.Fatalf("resource versions = %d, want 2", len(resourceVersions))
	}
	if resourceVersions[0].Version != 2 {
		t.Fatalf("latest resource version = %d, want 2", resourceVersions[0].Version)
	}
}
