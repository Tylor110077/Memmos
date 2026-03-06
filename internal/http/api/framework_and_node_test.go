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
