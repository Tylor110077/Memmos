package api

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	appgroup "github.com/tylor/goaipj/internal/app/group"
	"github.com/tylor/goaipj/internal/infra/monitoring"
)

func TestReadinessChecksDependencies(t *testing.T) {
	server := NewServer(Dependencies{
		GroupService: appgroup.NewService(appgroup.NewInMemoryRepository()),
		Readiness: []DependencyCheck{
			{Name: "database", Check: func(context.Context) error { return nil }},
			{Name: "redis", Check: func(context.Context) error { return context.DeadlineExceeded }},
		},
	})

	healthz := performJSONRequest(t, server, http.MethodGet, "/healthz", nil)
	if healthz.Code != http.StatusOK {
		t.Fatalf("healthz = %d", healthz.Code)
	}

	readyz := performJSONRequest(t, server, http.MethodGet, "/readyz", nil)
	if readyz.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz = %d body=%s", readyz.Code, readyz.Body.String())
	}
	if !strings.Contains(readyz.Body.String(), `"redis":"down"`) {
		t.Fatalf("expected failed dependency in body=%s", readyz.Body.String())
	}
}

func TestMetricsEndpointExposesHTTPAndTaskMetrics(t *testing.T) {
	metrics := monitoring.NewMetrics()
	server := NewServer(Dependencies{
		GroupService: appgroup.NewService(appgroup.NewInMemoryRepository()),
		Metrics:      metrics,
	})

	resp := performJSONRequest(t, server, http.MethodGet, "/healthz", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("healthz = %d", resp.Code)
	}
	metrics.ObserveTask("graph_generate", true, 20*time.Millisecond)

	metricsResp := performJSONRequest(t, server, http.MethodGet, "/metrics", nil)
	if metricsResp.Code != http.StatusOK {
		t.Fatalf("metrics status = %d body=%s", metricsResp.Code, metricsResp.Body.String())
	}
	body := metricsResp.Body.String()
	if !strings.Contains(body, "http_requests_total") {
		t.Fatalf("expected http metric, body=%s", body)
	}
	if !strings.Contains(body, "app_task_runs_total") {
		t.Fatalf("expected task metric, body=%s", body)
	}
	if !strings.Contains(body, `/healthz`) {
		t.Fatalf("expected /healthz route metric, body=%s", body)
	}
}
