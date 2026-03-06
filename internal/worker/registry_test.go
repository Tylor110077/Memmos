package worker

import (
	"context"
	"testing"
)

func TestRegistryRegistersExpectedTaskTypes(t *testing.T) {
	recorder := &fakeRegistrar{handlers: map[string]Handler{}}
	registry := NewRegistry()

	registry.RegisterAll(recorder)

	for _, taskType := range []string{
		TaskParseResource,
		TaskFetchWebResource,
		TaskGenerateGraph,
	} {
		if _, ok := recorder.handlers[taskType]; !ok {
			t.Fatalf("task type %q not registered", taskType)
		}
	}
}

func TestNoopHandlersExecute(t *testing.T) {
	recorder := &fakeRegistrar{handlers: map[string]Handler{}}
	registry := NewRegistry()
	registry.RegisterAll(recorder)

	for taskType, handler := range recorder.handlers {
		if err := handler(context.Background(), []byte(`{"resource_id":"r1"}`)); err != nil {
			t.Fatalf("handler %q error = %v", taskType, err)
		}
	}
}

type fakeRegistrar struct {
	handlers map[string]Handler
}

func (f *fakeRegistrar) Handle(taskType string, handler Handler) {
	f.handlers[taskType] = handler
}
