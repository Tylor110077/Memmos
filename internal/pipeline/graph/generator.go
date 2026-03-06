package graph

import (
	"fmt"
	"strings"
)

type Input struct {
	Title    string
	Summary  string
	Markdown string
}

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(input Input) (Document, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "Untitled"
	}

	lines := splitContent(input.Markdown)
	nodes := []Node{{
		ID:    "root",
		Name:  title,
		Type:  "topic",
		Level: 0,
	}}
	edges := make([]Edge, 0, len(lines))

	for idx, line := range lines {
		nodeID := fmt.Sprintf("n%d", idx+1)
		nodes = append(nodes, Node{
			ID:          nodeID,
			Name:        line,
			Type:        "concept",
			Description: line,
			Meaning:     line,
			Level:       1,
		})
		edges = append(edges, Edge{
			ID:       fmt.Sprintf("e%d", idx+1),
			SourceID: "root",
			TargetID: nodeID,
			Relation: "contains",
		})
	}

	doc := Document{
		Summary: fallbackSummary(input.Summary, lines),
		Nodes:   nodes,
		Edges:   edges,
	}
	return doc, doc.Validate()
}

func splitContent(markdown string) []string {
	raw := strings.Split(markdown, "\n")
	var lines []string
	for _, line := range raw {
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	if len(lines) > 4 {
		lines = lines[:4]
	}
	return lines
}

func fallbackSummary(summary string, lines []string) string {
	summary = strings.TrimSpace(summary)
	if summary != "" {
		return summary
	}
	return strings.Join(lines, " ")
}
