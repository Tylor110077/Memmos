package graph

import (
	"context"
	"testing"

	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestExpandNodePersistsExpansionFlags(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	saved, err := service.SaveResourceGraph(ctx, SaveInput{
		GroupID:    "group-1",
		ResourceID: "resource-1",
		Title:      "Resource",
		Document: pipelinegraph.Document{
			Summary: "summary",
			Nodes: []pipelinegraph.Node{
				{ID: "root", Name: "Root", Type: "topic", Level: 0},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveResourceGraph() error = %v", err)
	}

	expanded, err := service.ExpandNode(ctx, saved.Graph.ID, "root", pipelinegraph.Expansion{
		Node: pipelinegraph.Node{
			ID:          "exp1",
			Name:        "Upload Tip",
			Type:        "concept",
			Description: "Related tip",
			Meaning:     "Related tip",
			Level:       1,
		},
		Edge: pipelinegraph.Edge{
			ID:       "edge1",
			SourceID: "root",
			TargetID: "exp1",
			Relation: "expands",
		},
	})
	if err != nil {
		t.Fatalf("ExpandNode() error = %v", err)
	}

	foundNode := false
	foundEdge := false
	for _, node := range expanded.Nodes {
		if node.ID == "exp1" {
			foundNode = node.IsExpansion
		}
	}
	for _, edge := range expanded.Edges {
		if edge.ID == "edge1" {
			foundEdge = edge.IsExpansion
		}
	}
	if !foundNode || !foundEdge {
		t.Fatalf("expected expansion node and edge flags")
	}
}
