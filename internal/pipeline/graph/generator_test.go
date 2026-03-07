package graph

import "testing"

func TestGeneratorBuildsGraphFromNormalizedMarkdown(t *testing.T) {
	generator := NewGenerator()

	doc, err := generator.Generate(Input{
		Title:    "Go Backend",
		Summary:  "Go backend summary",
		Markdown: "# Go Backend\n\nUploads handle files.\nWorkers process graphs.\nQueries read results.",
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if err := doc.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(doc.Nodes) < 2 {
		t.Fatalf("nodes = %d, want >= 2", len(doc.Nodes))
	}
	if len(doc.Edges) < 1 {
		t.Fatalf("edges = %d, want >= 1", len(doc.Edges))
	}
}
