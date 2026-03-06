package observability

import (
	"context"
	"sync"
	"time"
)

type Attribute struct {
	Key   string
	Value string
}

type tracerContextKey struct{}
type errorReporterContextKey struct{}

type Span interface {
	End()
}

type Tracer interface {
	Start(ctx context.Context, name string, attrs ...Attribute) (context.Context, Span)
}

type ErrorReport struct {
	Message string
	Code    string
	TraceID string
	Route   string
	At      time.Time
}

type ErrorReporter interface {
	Report(ctx context.Context, report ErrorReport)
}

type noopSpan struct{}

func (noopSpan) End() {}

type noopTracer struct{}

func (noopTracer) Start(ctx context.Context, _ string, _ ...Attribute) (context.Context, Span) {
	return ctx, noopSpan{}
}

type noopErrorReporter struct{}

func (noopErrorReporter) Report(_ context.Context, _ ErrorReport) {}

func NewNoopTracer() Tracer {
	return noopTracer{}
}

func NewNoopErrorReporter() ErrorReporter {
	return noopErrorReporter{}
}

func ContextWithTracer(ctx context.Context, tracer Tracer) context.Context {
	return context.WithValue(ctx, tracerContextKey{}, tracer)
}

func TracerFromContext(ctx context.Context) (Tracer, bool) {
	tracer, ok := ctx.Value(tracerContextKey{}).(Tracer)
	return tracer, ok
}

func ContextWithErrorReporter(ctx context.Context, reporter ErrorReporter) context.Context {
	return context.WithValue(ctx, errorReporterContextKey{}, reporter)
}

func ErrorReporterFromContext(ctx context.Context) (ErrorReporter, bool) {
	reporter, ok := ctx.Value(errorReporterContextKey{}).(ErrorReporter)
	return reporter, ok
}

type FinishedSpan struct {
	Name      string
	StartedAt time.Time
	EndedAt   time.Time
	Attrs     []Attribute
}

type spanContextKey struct{}

type inMemorySpan struct {
	tracer    *InMemoryTracer
	name      string
	startedAt time.Time
	attrs     []Attribute
	once      sync.Once
}

func (s *inMemorySpan) End() {
	s.once.Do(func() {
		s.tracer.mu.Lock()
		defer s.tracer.mu.Unlock()
		s.tracer.finished = append(s.tracer.finished, FinishedSpan{
			Name:      s.name,
			StartedAt: s.startedAt,
			EndedAt:   time.Now().UTC(),
			Attrs:     append([]Attribute(nil), s.attrs...),
		})
	})
}

type InMemoryTracer struct {
	mu       sync.Mutex
	finished []FinishedSpan
}

func NewInMemoryTracer() *InMemoryTracer {
	return &InMemoryTracer{}
}

func (t *InMemoryTracer) Start(ctx context.Context, name string, attrs ...Attribute) (context.Context, Span) {
	span := &inMemorySpan{
		tracer:    t,
		name:      name,
		startedAt: time.Now().UTC(),
		attrs:     append([]Attribute(nil), attrs...),
	}
	return context.WithValue(ctx, spanContextKey{}, name), span
}

func (t *InMemoryTracer) FinishedSpans() []FinishedSpan {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]FinishedSpan(nil), t.finished...)
}

type InMemoryErrorReporter struct {
	mu      sync.Mutex
	reports []ErrorReport
}

func NewInMemoryErrorReporter() *InMemoryErrorReporter {
	return &InMemoryErrorReporter{}
}

func (r *InMemoryErrorReporter) Report(_ context.Context, report ErrorReport) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reports = append(r.reports, report)
}

func (r *InMemoryErrorReporter) Reports() []ErrorReport {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]ErrorReport(nil), r.reports...)
}
