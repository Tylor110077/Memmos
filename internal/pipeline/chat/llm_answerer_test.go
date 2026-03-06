package chat

import (
	"context"
	"strings"
	"testing"
)

func TestLLMAnswererUsesCustomPromptConfig(t *testing.T) {
	client := &fakeChatClient{returnValue: `{"answer":"ok","examples":["e1"],"cited_chunk_ids":["chunk-1"],"cited_node_ids":["node-1"]}`}
	answerer := NewLLMAnswererWithConfig(client, PromptConfig{
		SystemDirectives: []string{
			"You are a strict graph tutor.",
			"Return compact JSON.",
		},
		OutputKeys: []string{"answer", "examples", "cited_chunk_ids", "cited_node_ids"},
	})

	_, err := answerer.Answer(Input{
		Question: "what",
		CurrentNode: NodeContext{
			ID:   "node-1",
			Name: "Root",
		},
		Chunks: []ChunkContext{
			{ID: "chunk-1", Content: "content"},
		},
	})
	if err != nil {
		t.Fatalf("Answer() error = %v", err)
	}
	if len(client.messages) != 2 {
		t.Fatalf("messages = %d, want 2", len(client.messages))
	}
	systemPrompt := client.messages[0].Content
	if !strings.Contains(systemPrompt, "strict graph tutor") {
		t.Fatalf("system prompt = %q", systemPrompt)
	}
	if !strings.Contains(systemPrompt, "answer, examples, cited_chunk_ids, cited_node_ids") {
		t.Fatalf("expected output keys in system prompt, got %q", systemPrompt)
	}
}

type fakeChatClient struct {
	messages    []CompletionMessage
	returnValue string
}

func (f *fakeChatClient) Chat(_ context.Context, messages []CompletionMessage) (string, error) {
	f.messages = append([]CompletionMessage(nil), messages...)
	return f.returnValue, nil
}
