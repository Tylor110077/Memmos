package ingest

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/tylor/goaipj/internal/pipeline/normalize"
)

func TestHandlerParsesAndNormalizes(t *testing.T) {
	handler := NewHandler(
		&fakeParser{text: "Go makes concurrency manageable."},
		normalize.NewService(normalize.ChunkConfig{TargetSize: 16, Overlap: 4}),
	)

	result, err := handler.Process(context.Background(), Input{
		Title:       "Go",
		ContentType: "application/pdf",
		Body:        strings.NewReader("pdf"),
	})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if result.ExtractedText != "Go makes concurrency manageable." {
		t.Fatalf("text = %q", result.ExtractedText)
	}
	if result.Normalized.Summary == "" {
		t.Fatalf("expected summary")
	}
	if len(result.Normalized.Chunks) == 0 {
		t.Fatalf("expected chunks")
	}
}

type fakeParser struct {
	text string
}

func (f *fakeParser) ExtractText(_ context.Context, _ string, _ io.Reader) (string, error) {
	return f.text, nil
}
