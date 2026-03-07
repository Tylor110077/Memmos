package monitoring

import (
	"strings"
	"testing"
	"time"
)

func TestMetricsRenderPrometheus(t *testing.T) {
	metrics := NewMetrics()
	metrics.ObserveHTTPRequest("GET", "/healthz", 200, 10*time.Millisecond)
	metrics.ObserveTask("chat_answer", true, 25*time.Millisecond)

	body := metrics.RenderPrometheus()
	if !strings.Contains(body, "http_requests_total") {
		t.Fatalf("expected http metrics, body=%s", body)
	}
	if !strings.Contains(body, `route="/healthz"`) {
		t.Fatalf("expected route label, body=%s", body)
	}
	if !strings.Contains(body, `task="chat_answer"`) {
		t.Fatalf("expected task metric, body=%s", body)
	}
}
