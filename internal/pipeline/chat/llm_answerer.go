package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type CompletionMessage struct {
	Role    string
	Content string
}

type ChatClient interface {
	Chat(ctx context.Context, messages []CompletionMessage) (string, error)
}

type LLMAnswerer struct {
	client ChatClient
}

func NewLLMAnswerer(client ChatClient) *LLMAnswerer {
	return &LLMAnswerer{client: client}
}

func (a *LLMAnswerer) Answer(input Input) (Output, error) {
	if a.client == nil {
		return Output{}, errors.New("chat client is required")
	}
	if strings.TrimSpace(input.Question) == "" {
		return Output{}, errors.New("question is required")
	}
	systemPrompt := strings.Join([]string{
		"You are a focused learning assistant.",
		"Answer only from the provided node and group context.",
		"Return strict JSON with keys: answer, examples, cited_chunk_ids, cited_node_ids.",
		"Do not invent chunk ids or node ids outside the provided allowlists.",
	}, "\n")
	userPrompt := buildPrompt(input)
	raw, err := a.client.Chat(context.Background(), []CompletionMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		return Output{}, err
	}
	var parsed struct {
		Answer        string   `json:"answer"`
		Examples      []string `json:"examples"`
		CitedChunkIDs []string `json:"cited_chunk_ids"`
		CitedNodeIDs  []string `json:"cited_node_ids"`
	}
	if err := json.Unmarshal([]byte(extractJSONObject(raw)), &parsed); err != nil {
		return Output{}, fmt.Errorf("decode answer json: %w", err)
	}
	if strings.TrimSpace(parsed.Answer) == "" {
		return Output{}, errors.New("empty answer")
	}
	chunkAllow := map[string]struct{}{}
	for _, chunk := range input.Chunks {
		chunkAllow[chunk.ID] = struct{}{}
	}
	nodeAllow := map[string]struct{}{input.CurrentNode.ID: {}}
	for _, neighbor := range input.Neighbors {
		nodeAllow[neighbor.ID] = struct{}{}
	}
	parsed.CitedChunkIDs = filterAllowed(parsed.CitedChunkIDs, chunkAllow)
	parsed.CitedNodeIDs = filterAllowed(parsed.CitedNodeIDs, nodeAllow)
	if len(parsed.CitedChunkIDs) == 0 {
		for i, chunk := range input.Chunks {
			if i == 3 {
				break
			}
			parsed.CitedChunkIDs = append(parsed.CitedChunkIDs, chunk.ID)
		}
	}
	if len(parsed.CitedNodeIDs) == 0 {
		parsed.CitedNodeIDs = append(parsed.CitedNodeIDs, input.CurrentNode.ID)
		for i, neighbor := range input.Neighbors {
			if i == 2 {
				break
			}
			parsed.CitedNodeIDs = append(parsed.CitedNodeIDs, neighbor.ID)
		}
	}
	parsed.Examples = trimExamples(parsed.Examples)
	if len(parsed.Examples) == 0 {
		parsed.Examples = fallbackExamples(input)
	}
	return Output{
		Answer:        strings.TrimSpace(parsed.Answer),
		Examples:      parsed.Examples,
		CitedChunkIDs: parsed.CitedChunkIDs,
		CitedNodeIDs:  parsed.CitedNodeIDs,
	}, nil
}

func buildPrompt(input Input) string {
	var builder strings.Builder
	builder.WriteString("Question:\n")
	builder.WriteString(input.Question)
	builder.WriteString("\n\nCurrent node:\n")
	builder.WriteString(fmt.Sprintf("- id: %s\n- name: %s\n- description: %s\n- meaning: %s\n", input.CurrentNode.ID, input.CurrentNode.Name, input.CurrentNode.Description, input.CurrentNode.Meaning))
	builder.WriteString("\nNeighbors:\n")
	for _, neighbor := range input.Neighbors {
		builder.WriteString(fmt.Sprintf("- id: %s, name: %s, relation: %s\n", neighbor.ID, neighbor.Name, neighbor.Relation))
	}
	builder.WriteString("\nResource summary:\n")
	builder.WriteString(input.ResourceSummary)
	builder.WriteString("\n\nExamples:\n")
	for _, example := range input.Examples {
		builder.WriteString("- " + example + "\n")
	}
	builder.WriteString("\nRetrieved chunks:\n")
	for _, chunk := range input.Chunks {
		builder.WriteString(fmt.Sprintf("- id: %s, summary: %s, content: %s\n", chunk.ID, chunk.Summary, chunk.Content))
	}
	builder.WriteString("\nAllowed chunk ids:\n")
	builder.WriteString(strings.Join(chunkIDs(input.Chunks), ", "))
	builder.WriteString("\nAllowed node ids:\n")
	builder.WriteString(strings.Join(nodeIDs(input), ", "))
	builder.WriteString("\nOutput JSON only.\n")
	return builder.String()
}

func chunkIDs(chunks []ChunkContext) []string {
	out := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		out = append(out, chunk.ID)
	}
	return out
}

func nodeIDs(input Input) []string {
	out := []string{input.CurrentNode.ID}
	for _, neighbor := range input.Neighbors {
		out = append(out, neighbor.ID)
	}
	return out
}

func filterAllowed(values []string, allow map[string]struct{}) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := allow[value]; !ok {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func trimExamples(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
		if len(out) == 3 {
			break
		}
	}
	return out
}

func fallbackExamples(input Input) []string {
	if len(input.Examples) > 0 {
		return trimExamples(input.Examples)
	}
	out := make([]string, 0, len(input.Chunks))
	for _, chunk := range input.Chunks {
		text := strings.TrimSpace(firstNonEmpty(chunk.Summary, chunk.Content))
		if text == "" {
			continue
		}
		out = append(out, truncate(text, 120))
		if len(out) == 2 {
			break
		}
	}
	return out
}

func extractJSONObject(raw string) string {
	trimmed := strings.TrimSpace(raw)
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		return trimmed[start : end+1]
	}
	return trimmed
}
