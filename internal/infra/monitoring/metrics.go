package monitoring

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Metrics struct {
	mu              sync.RWMutex
	httpRequests    map[string]int64
	httpDurationSum map[string]float64
	taskRuns        map[string]int64
	taskDurationSum map[string]float64
}

func NewMetrics() *Metrics {
	return &Metrics{
		httpRequests:    map[string]int64{},
		httpDurationSum: map[string]float64{},
		taskRuns:        map[string]int64{},
		taskDurationSum: map[string]float64{},
	}
}

func (m *Metrics) ObserveHTTPRequest(method, route string, status int, duration time.Duration) {
	key := fmt.Sprintf(`method=%q,route=%q,status=%q`, method, route, fmt.Sprintf("%d", status))
	m.mu.Lock()
	defer m.mu.Unlock()
	m.httpRequests[key]++
	m.httpDurationSum[key] += duration.Seconds()
}

func (m *Metrics) ObserveTask(name string, success bool, duration time.Duration) {
	status := "success"
	if !success {
		status = "error"
	}
	key := fmt.Sprintf(`task=%q,status=%q`, name, status)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.taskRuns[key]++
	m.taskDurationSum[key] += duration.Seconds()
}

func (m *Metrics) RenderPrometheus() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var b strings.Builder
	b.WriteString("# HELP http_requests_total Count of HTTP requests handled by route.\n")
	b.WriteString("# TYPE http_requests_total counter\n")
	httpKeys := sortedKeys(m.httpRequests)
	for _, key := range httpKeys {
		fmt.Fprintf(&b, "http_requests_total{%s} %d\n", key, m.httpRequests[key])
	}
	b.WriteString("# HELP http_request_duration_seconds_sum Total HTTP request duration in seconds.\n")
	b.WriteString("# TYPE http_request_duration_seconds_sum counter\n")
	for _, key := range httpKeys {
		fmt.Fprintf(&b, "http_request_duration_seconds_sum{%s} %.6f\n", key, m.httpDurationSum[key])
	}
	b.WriteString("# HELP app_task_runs_total Count of instrumented application task runs.\n")
	b.WriteString("# TYPE app_task_runs_total counter\n")
	taskKeys := sortedKeys(m.taskRuns)
	for _, key := range taskKeys {
		fmt.Fprintf(&b, "app_task_runs_total{%s} %d\n", key, m.taskRuns[key])
	}
	b.WriteString("# HELP app_task_duration_seconds_sum Total duration of instrumented application tasks.\n")
	b.WriteString("# TYPE app_task_duration_seconds_sum counter\n")
	for _, key := range taskKeys {
		fmt.Fprintf(&b, "app_task_duration_seconds_sum{%s} %.6f\n", key, m.taskDurationSum[key])
	}
	return b.String()
}

func sortedKeys[T any](items map[string]T) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
