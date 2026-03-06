package worker

import (
	"context"
	"encoding/json"
)

const (
	TaskParseResource    = "resource:parse"
	TaskFetchWebResource = "resource:fetch_web"
	TaskGenerateGraph    = "graph:generate"
)

type Handler func(ctx context.Context, payload []byte) error

type Registrar interface {
	Handle(taskType string, handler Handler)
}

type Registry struct{}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) RegisterAll(registrar Registrar) {
	registrar.Handle(TaskParseResource, decodeNoopHandler())
	registrar.Handle(TaskFetchWebResource, decodeNoopHandler())
	registrar.Handle(TaskGenerateGraph, decodeNoopHandler())
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
