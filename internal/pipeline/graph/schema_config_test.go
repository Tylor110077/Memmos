package graph

import "testing"

func TestDocumentValidateWithConfigAllowsEmptySummaryWhenDisabled(t *testing.T) {
	doc := Document{
		Nodes: []Node{
			{ID: "root", Name: "Root", Type: "topic", Level: 0},
		},
	}

	if err := doc.ValidateWithConfig(SchemaConfig{RequireSummary: false}); err != nil {
		t.Fatalf("ValidateWithConfig() error = %v", err)
	}
}

func TestDocumentValidateWithConfigRequiresEdgesWhenEnabled(t *testing.T) {
	doc := Document{
		Summary: "summary",
		Nodes: []Node{
			{ID: "root", Name: "Root", Type: "topic", Level: 0},
		},
	}

	if err := doc.ValidateWithConfig(SchemaConfig{RequireSummary: true, RequireEdges: true}); err == nil {
		t.Fatalf("expected missing edges to fail")
	}
}
