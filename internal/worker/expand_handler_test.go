package worker

import (
	"context"
	"encoding/json"
	"testing"

	appgraph "github.com/tylor/goaipj/internal/app/graph"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestExpandNodeHandlerAddsExpansionNode(t *testing.T) {
	graphService := appgraph.NewService(appgraph.NewInMemoryRepository())
	saved, err := graphService.SaveResourceGraph(context.Background(), appgraph.SaveInput{
		GroupID:    "group-1",
		ResourceID: "resource-1",
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

	handler := NewExpandNodeHandler(graphService, pipelinegraph.NewExpansionGenerator(pipelinegraph.ExpansionRules{MaxNewNodes: 1}))
	payload, _ := json.Marshal(ExpandNodePayload{
		GraphID: saved.Graph.ID,
		NodeID:  "root",
	})
	if err := handler(context.Background(), payload); err != nil {
		t.Fatalf("handler error = %v", err)
	}

	detail, err := graphService.GetNodeDetail(context.Background(), saved.Graph.ID, "root")
	if err != nil {
		t.Fatalf("GetNodeDetail() error = %v", err)
	}
	if len(detail.Neighbors) != 1 {
		t.Fatalf("neighbors = %d, want 1", len(detail.Neighbors))
	}
}
