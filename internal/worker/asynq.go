package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

type AsynqRegistrar struct {
	mux *asynq.ServeMux
}

func NewAsynqRegistrar(mux *asynq.ServeMux) *AsynqRegistrar {
	return &AsynqRegistrar{mux: mux}
}

func (r *AsynqRegistrar) Handle(taskType string, handler Handler) {
	r.mux.HandleFunc(taskType, func(ctx context.Context, task *asynq.Task) error {
		return handler(ctx, task.Payload())
	})
}
