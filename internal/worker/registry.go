package worker

import (
	"context"
	"encoding/json"

	appcleanup "github.com/tylor/goaipj/internal/app/cleanup"
	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appresource "github.com/tylor/goaipj/internal/app/resource"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

const (
	TaskParseResource    = "resource:parse"
	TaskFetchWebResource = "resource:fetch_web"
	TaskGenerateGraph    = "graph:generate"
	TaskExpandNode       = "graph:expand_node"
)

type Handler func(ctx context.Context, payload []byte) error

type Registrar interface {
	Handle(taskType string, handler Handler)
}

type Dependencies struct {
	ResourceService    *appresource.Service
	GraphService       *appgraph.Service
	GraphGenerator     *pipelinegraph.Generator
	ExpansionGenerator expansionGenerator
	CleanupService     *appcleanup.Service
}

type Registry struct {
	deps Dependencies
}

func NewRegistry(deps ...Dependencies) *Registry {
	registry := &Registry{}
	if len(deps) > 0 {
		registry.deps = deps[0]
	}
	return registry
}

func (r *Registry) RegisterAll(registrar Registrar) {
	registrar.Handle(TaskParseResource, decodeNoopHandler())
	registrar.Handle(TaskFetchWebResource, decodeNoopHandler())
	if r.deps.ResourceService != nil && r.deps.GraphService != nil && r.deps.GraphGenerator != nil {
		registrar.Handle(TaskGenerateGraph, NewGenerateGraphHandler(r.deps.ResourceService, r.deps.GraphService, r.deps.GraphGenerator))
	} else {
		registrar.Handle(TaskGenerateGraph, decodeNoopHandler())
	}
	if r.deps.GraphService != nil && r.deps.ExpansionGenerator != nil {
		registrar.Handle(TaskExpandNode, NewExpandNodeHandler(r.deps.GraphService, r.deps.ExpansionGenerator))
	} else {
		registrar.Handle(TaskExpandNode, decodeNoopHandler())
	}
	if r.deps.CleanupService != nil {
		registrar.Handle(TaskCleanup, NewCleanupHandler(r.deps.CleanupService))
		return
	}
	registrar.Handle(TaskCleanup, decodeNoopHandler())
}

func decodeNoopHandler() Handler {
	return func(ctx context.Context, payload []byte) error {
		_ = ctx
		var body map[string]any
		if len(payload) == 0 {
			return nil
		}
		return json.Unmarshal(payload, &body)
	}
}
