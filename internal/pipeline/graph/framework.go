package graph

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type FrameworkRules struct {
	MaxTopNodes int
	MaxDepth    int
}

func (r FrameworkRules) Validate() error {
	if r.MaxTopNodes <= 0 || r.MaxDepth <= 0 {
		return errors.New("framework rules must be positive")
	}
	return nil
}

type FrameworkGenerator struct {
	rules FrameworkRules
}

func NewFrameworkGenerator(rules FrameworkRules) *FrameworkGenerator {
	return &FrameworkGenerator{rules: rules}
}

func (g *FrameworkGenerator) Generate(documents []Document) (Document, error) {
	if err := g.rules.Validate(); err != nil {
		return Document{}, err
	}
	counts := map[string]int{}
	for _, doc := range documents {
		for _, node := range doc.Nodes {
			if node.Level == 0 {
				continue
			}
			key := strings.TrimSpace(node.Name)
			if key == "" {
				continue
			}
			counts[key]++
		}
	}

	type pair struct {
		Name  string
		Count int
	}
	var items []pair
	for name, count := range counts {
		items = append(items, pair{Name: name, Count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Name < items[j].Name
		}
		return items[i].Count > items[j].Count
	})
	if len(items) > g.rules.MaxTopNodes {
		items = items[:g.rules.MaxTopNodes]
	}

	nodes := []Node{{ID: "root", Name: "Framework", Type: "topic", Level: 0}}
	edges := make([]Edge, 0, len(items))
	summaryParts := make([]string, 0, len(items))
	for idx, item := range items {
		id := fmt.Sprintf("f%d", idx+1)
		nodes = append(nodes, Node{
			ID:          id,
			Name:        item.Name,
			Type:        "theme",
			Description: item.Name,
			Meaning:     fmt.Sprintf("Referenced by %d resource graphs", item.Count),
			Level:       1,
		})
		edges = append(edges, Edge{
			ID:       fmt.Sprintf("fe%d", idx+1),
			SourceID: "root",
			TargetID: id,
			Relation: "contains",
		})
		summaryParts = append(summaryParts, item.Name)
	}
	if len(nodes) == 1 {
		nodes = append(nodes, Node{ID: "f1", Name: "Overview", Type: "theme", Level: 1})
		edges = append(edges, Edge{ID: "fe1", SourceID: "root", TargetID: "f1", Relation: "contains"})
		summaryParts = append(summaryParts, "Overview")
	}

	doc := Document{
		Summary: strings.Join(summaryParts, ", "),
		Nodes:   nodes,
		Edges:   edges,
	}
	return doc, doc.Validate()
}
