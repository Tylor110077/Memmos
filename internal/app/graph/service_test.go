package graph

import (
	"context"
	"testing"

	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestPersistArchivesPreviousActiveGraph(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	first, err := service.SaveResourceGraph(ctx, SaveInput{
		GroupID:    "group-1",
		ResourceID: "resource-1",
		Title:      "First",
		Document: pipelinegraph.Document{
			Summary: "summary",
			Nodes: []pipelinegraph.Node{
				{ID: "root", Name: "Root", Type: "topic", Level: 0},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveResourceGraph(first) error = %v", err)
	}

	second, err := service.SaveResourceGraph(ctx, SaveInput{
		GroupID:    "group-1",
		ResourceID: "resource-1",
		Title:      "Second",
		Document: pipelinegraph.Document{
			Summary: "summary",
			Nodes: []pipelinegraph.Node{
				{ID: "root2", Name: "Root 2", Type: "topic", Level: 0},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveResourceGraph(second) error = %v", err)
	}

	storedFirst, ok := repo.GetByID(first.Graph.ID)
	if !ok {
		t.Fatalf("first graph not found in repo")
	}
	if storedFirst.Graph.IsActive {
		t.Fatalf("first graph should have been archived")
	}
	if !second.Graph.IsActive {
		t.Fatalf("second graph should be active")
	}
	if second.Graph.Version != 2 {
		t.Fatalf("version = %d, want 2", second.Graph.Version)
	}
}

func TestGetResourceGraphFiltersByLevel(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	saved, err := service.SaveResourceGraph(ctx, SaveInput{
		GroupID:    "group-1",
		ResourceID: "resource-1",
		Title:      "Graph",
		Document: pipelinegraph.Document{
			Summary: "summary",
			Nodes: []pipelinegraph.Node{
				{ID: "root", Name: "Root", Type: "topic", Level: 0},
				{ID: "n1", Name: "Child", Type: "concept", Level: 1},
				{ID: "n2", Name: "Deep", Type: "concept", Level: 2},
			},
			Edges: []pipelinegraph.Edge{
				{ID: "e1", SourceID: "root", TargetID: "n1", Relation: "contains"},
				{ID: "e2", SourceID: "n1", TargetID: "n2", Relation: "contains"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveResourceGraph() error = %v", err)
	}

	got, err := service.GetResourceGraph(ctx, saved.Graph.ResourceID, QueryOptions{MaxLevel: 1})
	if err != nil {
		t.Fatalf("GetResourceGraph() error = %v", err)
	}
	if len(got.Nodes) != 2 {
		t.Fatalf("nodes = %d, want 2", len(got.Nodes))
	}
	if len(got.Edges) != 1 {
		t.Fatalf("edges = %d, want 1", len(got.Edges))
	}
}
