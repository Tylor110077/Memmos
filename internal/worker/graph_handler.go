package worker

import (
	"context"
	"encoding/json"

	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appresource "github.com/tylor/goaipj/internal/app/resource"
	domainresource "github.com/tylor/goaipj/internal/domain/resource"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

type GenerateGraphPayload struct {
	GroupID    string `json:"group_id"`
	ResourceID string `json:"resource_id"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
	Markdown   string `json:"markdown"`
}

type resourceStatusUpdater interface {
	UpdateStatus(ctx context.Context, resourceID string, next domainresource.Status) (appresource.Resource, error)
	FailResource(ctx context.Context, resourceID string, stage domainresource.Status, message string) (appresource.Resource, error)
}

type resourceGraphSaver interface {
	SaveResourceGraph(ctx context.Context, input appgraph.SaveInput) (appgraph.SavedGraph, error)
}

type graphGenerator interface {
	Generate(input pipelinegraph.Input) (pipelinegraph.Document, error)
}

func NewGenerateGraphHandler(resources resourceStatusUpdater, graphs resourceGraphSaver, generator graphGenerator) Handler {
	return func(ctx context.Context, payload []byte) error {
		var body GenerateGraphPayload
		if err := json.Unmarshal(payload, &body); err != nil {
			return err
		}
		if _, err := resources.UpdateStatus(ctx, body.ResourceID, domainresource.StatusGraphGenerating); err != nil {
			return err
		}
		document, err := generator.Generate(pipelinegraph.Input{
			Title:    body.Title,
			Summary:  body.Summary,
			Markdown: body.Markdown,
		})
		if err != nil {
			_, _ = resources.FailResource(ctx, body.ResourceID, domainresource.StatusGraphGenerating, err.Error())
			return err
		}
		if _, err := graphs.SaveResourceGraph(ctx, appgraph.SaveInput{
			GroupID:    body.GroupID,
			ResourceID: body.ResourceID,
			Title:      body.Title,
			Document:   document,
		}); err != nil {
			_, _ = resources.FailResource(ctx, body.ResourceID, domainresource.StatusGraphGenerating, err.Error())
			return err
		}
		_, err = resources.UpdateStatus(ctx, body.ResourceID, domainresource.StatusCompleted)
		return err
	}
}
