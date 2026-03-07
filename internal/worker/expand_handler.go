package worker

import (
	"context"
	"encoding/json"

	appgraph "github.com/tylor/goaipj/internal/app/graph"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

type ExpandNodePayload struct {
	GraphID string `json:"graph_id"`
	NodeID  string `json:"node_id"`
}

type nodeDetailGetter interface {
	GetNodeDetail(ctx context.Context, graphID, nodeID string) (appgraph.NodeDetail, error)
	ExpandNode(ctx context.Context, graphID, nodeID string, expansion pipelinegraph.Expansion) (appgraph.SavedGraph, error)
}

type expansionGenerator interface {
	Generate(input pipelinegraph.ExpansionInput) (pipelinegraph.Expansion, error)
}

func NewExpandNodeHandler(graphs nodeDetailGetter, generator expansionGenerator) Handler {
	return func(ctx context.Context, payload []byte) error {
		var body ExpandNodePayload
		if err := json.Unmarshal(payload, &body); err != nil {
			return err
		}
		detail, err := graphs.GetNodeDetail(ctx, body.GraphID, body.NodeID)
		if err != nil {
			return err
		}
		neighbors := make([]pipelinegraph.Node, 0, len(detail.Neighbors))
		for _, neighbor := range detail.Neighbors {
			neighbors = append(neighbors, pipelinegraph.Node{
				ID:    neighbor.Node.ID,
				Name:  neighbor.Node.Name,
				Type:  neighbor.Node.Type,
				Level: neighbor.Node.Level,
			})
		}
		expansion, err := generator.Generate(pipelinegraph.ExpansionInput{
			Current: pipelinegraph.Node{
				ID:    detail.Node.ID,
				Name:  detail.Node.Name,
				Type:  detail.Node.Type,
				Level: detail.Node.Level,
			},
			Neighbors: neighbors,
		})
		if err != nil {
			return err
		}
		_, err = graphs.ExpandNode(ctx, body.GraphID, body.NodeID, expansion)
		return err
	}
}
