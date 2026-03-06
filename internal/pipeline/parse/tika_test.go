package parse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTikaClientExtractText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		if got := r.Header.Get("Accept"); got != "text/plain" {
			t.Fatalf("accept = %q, want text/plain", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello from tika"))
	}))
	defer server.Close()

	client := NewTikaClient(server.URL)
	text, err := client.ExtractText(context.Background(), "application/pdf", strings.NewReader("pdf-bytes"))
	if err != nil {
		t.Fatalf("ExtractText() error = %v", err)
	}
	if text != "Hello from tika" {
		t.Fatalf("text = %q, want Hello from tika", text)
	}
}

func TestTikaClientPropagatesFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("bad gateway"))
	}))
	defer server.Close()

	client := NewTikaClient(server.URL)
	if _, err := client.ExtractText(context.Background(), "application/pdf", strings.NewReader("pdf-bytes")); err == nil {
		t.Fatalf("expected error")
	}
}
