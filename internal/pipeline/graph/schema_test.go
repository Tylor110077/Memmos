package graph

import "testing"

func TestDocumentValidateAcceptsConsistentGraph(t *testing.T) {
	doc := Document{
		Summary: "Backend overview",
		Nodes: []Node{
			{ID: "root", Name: "Backend", Type: "topic", Level: 0},
			{ID: "n1", Name: "Upload", Type: "concept", Level: 1},
		},
		Edges: []Edge{
			{ID: "e1", SourceID: "root", TargetID: "n1", Relation: "contains"},
		},
	}

	if err := doc.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestDocumentValidateRejectsDanglingEdges(t *testing.T) {
	doc := Document{
		Summary: "Backend overview",
		Nodes: []Node{
			{ID: "root", Name: "Backend", Type: "topic", Level: 0},
		},
		Edges: []Edge{
			{ID: "e1", SourceID: "root", TargetID: "missing", Relation: "contains"},
		},
	}

	if err := doc.Validate(); err == nil {
		t.Fatalf("expected dangling edge error")
	}
}
