package graph

import (
	"errors"
	"fmt"
)

type Document struct {
	Summary string `json:"summary"`
	Nodes   []Node `json:"nodes"`
	Edges   []Edge `json:"edges"`
}

type Node struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Meaning     string `json:"meaning,omitempty"`
	Level       int    `json:"level"`
}

type Edge struct {
	ID       string `json:"id"`
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Relation string `json:"relation"`
}

var (
	ErrEmptySummary = errors.New("graph summary is required")
	ErrNoNodes      = errors.New("graph nodes are required")
)

func (d Document) Validate() error {
	if d.Summary == "" {
		return ErrEmptySummary
	}
	if len(d.Nodes) == 0 {
		return ErrNoNodes
	}

	seen := map[string]struct{}{}
	for _, node := range d.Nodes {
		if node.ID == "" || node.Name == "" || node.Type == "" {
			return fmt.Errorf("invalid node: %+v", node)
		}
		seen[node.ID] = struct{}{}
	}

	for _, edge := range d.Edges {
		if edge.ID == "" || edge.SourceID == "" || edge.TargetID == "" || edge.Relation == "" {
			return fmt.Errorf("invalid edge: %+v", edge)
		}
		if _, ok := seen[edge.SourceID]; !ok {
			return fmt.Errorf("edge source %q does not exist", edge.SourceID)
		}
		if _, ok := seen[edge.TargetID]; !ok {
			return fmt.Errorf("edge target %q does not exist", edge.TargetID)
		}
	}
	return nil
}
