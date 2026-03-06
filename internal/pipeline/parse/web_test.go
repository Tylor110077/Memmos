package parse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebFetcherExtractsTitleTextAndMarkdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<!doctype html>
<html>
  <head><title>Go Guide</title></head>
  <body>
    <header>Noise</header>
    <article>
      <h1>Go Guide</h1>
      <p>Go is concise.</p>
      <p>Concurrency is built in.</p>
    </article>
    <footer>Footer</footer>
  </body>
</html>`))
	}))
	defer server.Close()

	fetcher := NewWebFetcher()
	result, err := fetcher.Fetch(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if result.Title != "Go Guide" {
		t.Fatalf("title = %q, want Go Guide", result.Title)
	}
	if !strings.Contains(result.Text, "Go is concise.") {
		t.Fatalf("text = %q", result.Text)
	}
	if !strings.Contains(result.Markdown, "# Go Guide") {
		t.Fatalf("markdown = %q", result.Markdown)
	}
}

func TestWebFetcherRejectsBadStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	fetcher := NewWebFetcher()
	if _, err := fetcher.Fetch(context.Background(), server.URL); err == nil {
		t.Fatalf("expected error")
	}
}
