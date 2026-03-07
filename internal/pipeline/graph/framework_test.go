package graph

import "testing"

func TestFrameworkGeneratorBuildsGroupedOverview(t *testing.T) {
	generator := NewFrameworkGenerator(FrameworkRules{
		MaxTopNodes: 3,
		MaxDepth:    2,
	})

	doc, err := generator.Generate([]Document{
		{
			Summary: "Go backend",
			Nodes: []Node{
				{ID: "a1", Name: "Upload", Type: "concept", Level: 1},
				{ID: "a2", Name: "Worker", Type: "concept", Level: 1},
			},
		},
		{
			Summary: "Infra",
			Nodes: []Node{
				{ID: "b1", Name: "Worker", Type: "concept", Level: 1},
				{ID: "b2", Name: "Graph", Type: "concept", Level: 1},
			},
		},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if err := doc.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(doc.Nodes) == 0 {
		t.Fatalf("expected nodes")
	}
}

func TestFrameworkRulesValidate(t *testing.T) {
	if err := (FrameworkRules{}).Validate(); err == nil {
		t.Fatalf("expected invalid rules")
	}
}
