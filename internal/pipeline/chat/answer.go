package chat

import (
	"errors"
	"strings"
)

type NodeContext struct {
	ID          string
	Name        string
	Description string
	Meaning     string
}

type NeighborContext struct {
	ID       string
	Name     string
	Relation string
}

type ChunkContext struct {
	ID      string
	Content string
	Summary string
}

type MessageContext struct {
	Role    string
	Content string
}

type Input struct {
	Question        string
	CurrentNode     NodeContext
	Neighbors       []NeighborContext
	ResourceSummary string
	Examples        []string
	Chunks          []ChunkContext
	RecentMessages  []MessageContext
}

type Output struct {
	Answer        string
	Examples      []string
	CitedChunkIDs []string
	CitedNodeIDs  []string
}

type Answerer struct{}

func NewAnswerer() *Answerer {
	return &Answerer{}
}

func (a *Answerer) Answer(input Input) (Output, error) {
	if strings.TrimSpace(input.Question) == "" {
		return Output{}, errors.New("question is required")
	}
	if strings.TrimSpace(input.CurrentNode.Name) == "" {
		return Output{}, errors.New("current node is required")
	}

	parts := []string{
		input.CurrentNode.Name + " is the current focus node.",
	}
	if text := firstNonEmpty(input.CurrentNode.Meaning, input.CurrentNode.Description); text != "" {
		parts = append(parts, text)
	}
	if input.ResourceSummary != "" {
		parts = append(parts, "Resource context: "+input.ResourceSummary)
	}
	if len(input.Neighbors) > 0 {
		related := make([]string, 0, len(input.Neighbors))
		for i, neighbor := range input.Neighbors {
			if i == 2 {
				break
			}
			if neighbor.Relation != "" {
				related = append(related, neighbor.Name+" ("+neighbor.Relation+")")
				continue
			}
			related = append(related, neighbor.Name)
		}
		if len(related) > 0 {
			parts = append(parts, "Direct neighbors: "+strings.Join(related, ", ")+".")
		}
	}
	if len(input.Chunks) > 0 {
		parts = append(parts, "Relevant background: "+truncate(firstNonEmpty(input.Chunks[0].Summary, input.Chunks[0].Content), 140))
	}

	examples := append([]string(nil), input.Examples...)
	if len(examples) == 0 && len(input.Chunks) > 0 {
		for i, chunk := range input.Chunks {
			if i == 2 {
				break
			}
			examples = append(examples, truncate(firstNonEmpty(chunk.Summary, chunk.Content), 100))
		}
	}
	if len(examples) > 2 {
		examples = examples[:2]
	}

	citedChunkIDs := make([]string, 0, len(input.Chunks))
	for i, chunk := range input.Chunks {
		if i == 3 {
			break
		}
		citedChunkIDs = append(citedChunkIDs, chunk.ID)
	}
	citedNodeIDs := []string{input.CurrentNode.ID}
	for i, neighbor := range input.Neighbors {
		if i == 2 {
			break
		}
		citedNodeIDs = append(citedNodeIDs, neighbor.ID)
	}

	return Output{
		Answer:        strings.Join(parts, " "),
		Examples:      examples,
		CitedChunkIDs: citedChunkIDs,
		CitedNodeIDs:  citedNodeIDs,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}
