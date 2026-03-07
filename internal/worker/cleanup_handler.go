package worker

import (
	"context"
	"encoding/json"
	"time"

	appcleanup "github.com/tylor/goaipj/internal/app/cleanup"
)

const TaskCleanup = "maintenance:cleanup"

type CleanupPayload struct {
	RetentionHours int `json:"retention_hours"`
}

type cleanupService interface {
	Cleanup(ctx context.Context, cutoff time.Time) (appcleanup.Report, error)
}

func NewCleanupHandler(service cleanupService) Handler {
	return func(ctx context.Context, payload []byte) error {
		var body CleanupPayload
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &body); err != nil {
				return err
			}
		}
		retention := time.Duration(body.RetentionHours) * time.Hour
		if retention <= 0 {
			retention = 24 * time.Hour
		}
		_, err := service.Cleanup(ctx, time.Now().UTC().Add(-retention))
		return err
	}
}
