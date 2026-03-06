package graph

import (
	"context"
	"testing"

	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestSaveFrameworkGraphArchivesPreviousActive(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	first, err := service.SaveFrameworkGraph(ctx, SaveFrameworkInput{
		GroupID: "group-1",
		Title:   "Framework 1",
		Document: pipelinegraph.Document{
			Summary: "summary",
			Nodes: []pipelinegraph.Node{
				{ID: "root", Name: "Backend", Type: "topic", Level: 0},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveFrameworkGraph(first) error = %v", err)
	}

	second, err := service.SaveFrameworkGraph(ctx, SaveFrameworkInput{
		GroupID: "group-1",
		Title:   "Framework 2",
		Document: pipelinegraph.Document{
			Summary: "summary",
			Nodes: []pipelinegraph.Node{
				{ID: "root2", Name: "Platform", Type: "topic", Level: 0},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveFrameworkGraph(second) error = %v", err)
	}

	storedFirst, _ := repo.GetByID(first.Graph.ID)
	if storedFirst.Graph.IsActive {
		t.Fatalf("first framework graph should be archived")
	}
	if !second.Graph.IsActive {
		t.Fatalf("second framework graph should be active")
	}
}

func TestGetNodeDetailReturnsNeighborsAndExamples(t *testing.T) {
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
				{ID: "child", Name: "Child", Type: "concept", Level: 1},
			},
			Edges: []pipelinegraph.Edge{
				{ID: "e1", SourceID: "root", TargetID: "child", Relation: "contains"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveResourceGraph() error = %v", err)
	}
	repo.AddExample(saved.Graph.ID, Example{
		ID:      "ex1",
		NodeID:  "child",
		Content: "Child example",
	})

	detail, err := service.GetNodeDetail(ctx, saved.Graph.ID, "child")
	if err != nil {
		t.Fatalf("GetNodeDetail() error = %v", err)
	}
	if detail.Node.ID != "child" {
		t.Fatalf("node id = %q", detail.Node.ID)
	}
	if len(detail.Neighbors) != 1 {
		t.Fatalf("neighbors = %d, want 1", len(detail.Neighbors))
	}
	if len(detail.Examples) != 1 {
		t.Fatalf("examples = %d, want 1", len(detail.Examples))
	}
}

func TestListFrameworkGraphVersionsReturnsDescendingHistory(t *testing.T) {
	repo := NewInMemoryRepository()
	service := NewService(repo)
	ctx := context.Background()

	for _, title := range []string{"Framework 1", "Framework 2", "Framework 3"} {
		if _, err := service.SaveFrameworkGraph(ctx, SaveFrameworkInput{
			GroupID: "group-1",
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

	versions, err := service.ListFrameworkGraphVersions(ctx, "group-1")
	if err != nil {
		t.Fatalf("ListFrameworkGraphVersions() error = %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("versions = %d, want 3", len(versions))
	}
	if versions[0].Graph.Version != 3 || versions[0].Graph.IsActive != true {
		t.Fatalf("latest version = %d active=%v", versions[0].Graph.Version, versions[0].Graph.IsActive)
	}
	if versions[2].Graph.Version != 1 {
		t.Fatalf("oldest version = %d, want 1", versions[2].Graph.Version)
	}
}
