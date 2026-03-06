package normalize

import "testing"

func TestNormalizeProducesSummaryAndChunks(t *testing.T) {
	normalizer := NewService(ChunkConfig{TargetSize: 32, Overlap: 8})

	result := normalizer.Normalize(NormalizeInput{
		Title: "Go",
		Text:  "Go is expressive, concise, clean, and efficient. Its concurrency mechanisms make it easy to write programs that get the most out of multicore and networked machines.",
	})

	if result.Markdown == "" {
		t.Fatalf("expected markdown")
	}
	if result.Summary == "" {
		t.Fatalf("expected summary")
	}
	if len(result.Chunks) < 2 {
		t.Fatalf("chunks = %d, want at least 2", len(result.Chunks))
	}
}

func TestNormalizeUsesTitleAsHeading(t *testing.T) {
	normalizer := NewService(ChunkConfig{TargetSize: 80, Overlap: 10})
	result := normalizer.Normalize(NormalizeInput{
		Title: "Backend MVP",
		Text:  "A short body.",
	})
	if result.Markdown[:13] != "# Backend MVP" {
		t.Fatalf("markdown = %q", result.Markdown)
	}
}
