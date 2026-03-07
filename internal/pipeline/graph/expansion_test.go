package graph

import "testing"

func TestExpansionGeneratorProducesSingleDirectlyRelatedNode(t *testing.T) {
	generator := NewExpansionGenerator(ExpansionRules{MaxNewNodes: 1})

	result, err := generator.Generate(ExpansionInput{
		Current:   Node{ID: "n1", Name: "Upload", Type: "concept", Level: 1},
		Neighbors: []Node{{ID: "n2", Name: "Storage", Type: "concept", Level: 1}},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if result.Node.ID == "" || result.Edge.ID == "" {
		t.Fatalf("expected node and edge ids")
	}
	if result.Edge.SourceID != "n1" && result.Edge.TargetID != "n1" {
		t.Fatalf("expansion edge must touch current node")
	}
}

func TestExpansionRulesValidate(t *testing.T) {
	if err := (ExpansionRules{}).Validate(); err == nil {
		t.Fatalf("expected invalid rules")
	}
}
