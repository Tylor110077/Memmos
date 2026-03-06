package job

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"sync"
	"time"
)

type Status string

const (
	StatusQueued  Status = "queued"
	StatusRunning Status = "running"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
	StatusCancelled Status = "cancelled"
)

type EventType string

const (
	EventQueued  EventType = "queued"
	EventRunning EventType = "running"
	EventDone    EventType = "done"
	EventFailed  EventType = "failed"
	EventCancelled EventType = "cancelled"
)

type Job struct {
	ID           string         `json:"id"`
	GroupID      string         `json:"group_id"`
	ResourceID   string         `json:"resource_id"`
	JobType      string         `json:"job_type"`
	QueueName    string         `json:"queue_name"`
	Status       Status         `json:"status"`
	Payload      map[string]any `json:"payload,omitempty"`
	DedupeKey    string         `json:"dedupe_key,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
	Attempts     int            `json:"attempts"`
	MaxAttempts  int            `json:"max_attempts"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	StartedAt    *time.Time     `json:"started_at,omitempty"`
	FinishedAt   *time.Time     `json:"finished_at,omitempty"`
}

type Event struct {
	ID        string    `json:"id"`
	JobID     string    `json:"job_id"`
	Type      EventType `json:"type"`
	Message   string    `json:"message,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Result struct {
	Job    Job
	Events []Event
}

type CreateQueuedJobInput struct {
	GroupID       string
	ResourceID    string
	JobType       string
	QueueName     string
	Payload       map[string]any
	MaxAttempts   int
	Deduplication string
}

type Repository interface {
	CreateJob(ctx context.Context, job Job) error
	UpdateJob(ctx context.Context, job Job) error
	GetJob(ctx context.Context, jobID string) (Job, error)
	ListJobs(ctx context.Context) ([]Job, error)
	CreateEvent(ctx context.Context, event Event) error
	ListEvents(ctx context.Context, jobID string) ([]Event, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateQueuedJob(ctx context.Context, input CreateQueuedJobInput) (Result, error) {
	now := time.Now().UTC()
	maxAttempts := input.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	job := Job{
		ID:          newID(),
		GroupID:     input.GroupID,
		ResourceID:  input.ResourceID,
		JobType:     input.JobType,
		QueueName:   input.QueueName,
		Status:      StatusQueued,
		Payload:     input.Payload,
		DedupeKey:   input.Deduplication,
		MaxAttempts: maxAttempts,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateJob(ctx, job); err != nil {
		return Result{}, err
	}
	event := Event{
		ID:        newID(),
		JobID:     job.ID,
		Type:      EventQueued,
		Message:   "job queued",
		CreatedAt: now,
	}
	if err := s.repo.CreateEvent(ctx, event); err != nil {
		return Result{}, err
	}
	return Result{Job: job, Events: []Event{event}}, nil
}

func (s *Service) GetJob(ctx context.Context, jobID string) (Job, error) {
	return s.repo.GetJob(ctx, jobID)
}

func (s *Service) FindActiveJob(ctx context.Context, dedupeKey string) (Job, bool, error) {
	jobs, err := s.repo.ListJobs(ctx)
	if err != nil {
		return Job{}, false, err
	}
	for _, job := range jobs {
		if job.DedupeKey != dedupeKey {
			continue
		}
		if job.Status != StatusQueued && job.Status != StatusRunning {
			continue
		}
		return job, true, nil
	}
	return Job{}, false, nil
}

func (s *Service) MarkRunning(ctx context.Context, jobID string) (Result, error) {
	return s.transition(ctx, jobID, StatusRunning, EventRunning, "", func(job *Job, now time.Time) {
		job.StartedAt = &now
		job.Attempts++
	})
}

func (s *Service) MarkDone(ctx context.Context, jobID string) (Result, error) {
	return s.transition(ctx, jobID, StatusDone, EventDone, "", func(job *Job, now time.Time) {
		job.FinishedAt = &now
	})
}

func (s *Service) MarkFailed(ctx context.Context, jobID string, message string) (Result, error) {
	return s.transition(ctx, jobID, StatusFailed, EventFailed, message, func(job *Job, now time.Time) {
		job.ErrorMessage = message
		job.FinishedAt = &now
	})
}

func (s *Service) CancelJob(ctx context.Context, jobID string) (Result, error) {
	job, err := s.repo.GetJob(ctx, jobID)
	if err != nil {
		return Result{}, err
	}
	if job.Status != StatusQueued && job.Status != StatusRunning {
		return Result{}, errors.New("job is not cancellable")
	}
	return s.transition(ctx, jobID, StatusCancelled, EventCancelled, "job cancelled", func(job *Job, now time.Time) {
		job.FinishedAt = &now
	})
}

func (s *Service) ResumeJob(ctx context.Context, jobID string) (Result, error) {
	job, err := s.repo.GetJob(ctx, jobID)
	if err != nil {
		return Result{}, err
	}
	if job.Status != StatusCancelled && job.Status != StatusFailed {
		return Result{}, errors.New("job is not resumable")
	}
	return s.transition(ctx, jobID, StatusQueued, EventQueued, "job resumed", func(job *Job, _ time.Time) {
		job.FinishedAt = nil
		job.StartedAt = nil
		job.ErrorMessage = ""
	})
}

func (s *Service) transition(ctx context.Context, jobID string, status Status, eventType EventType, message string, mutate func(job *Job, now time.Time)) (Result, error) {
	job, err := s.repo.GetJob(ctx, jobID)
	if err != nil {
		return Result{}, err
	}
	now := time.Now().UTC()
	job.Status = status
	job.UpdatedAt = now
	if mutate != nil {
		mutate(&job, now)
	}
	if err := s.repo.UpdateJob(ctx, job); err != nil {
		return Result{}, err
	}
	event := Event{
		ID:        newID(),
		JobID:     job.ID,
		Type:      eventType,
		Message:   message,
		CreatedAt: now,
	}
	if err := s.repo.CreateEvent(ctx, event); err != nil {
		return Result{}, err
	}
	return Result{Job: job, Events: []Event{event}}, nil
}

type InMemoryRepository struct {
	mu     sync.RWMutex
	jobs   map[string]Job
	events map[string][]Event
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		jobs:   map[string]Job{},
		events: map[string][]Event{},
	}
}

func (r *InMemoryRepository) CreateJob(_ context.Context, job Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[job.ID] = job
	return nil
}

func (r *InMemoryRepository) UpdateJob(_ context.Context, job Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.jobs[job.ID]; !ok {
		return errors.New("job not found")
	}
	r.jobs[job.ID] = job
	return nil
}

func (r *InMemoryRepository) GetJob(_ context.Context, jobID string) (Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.jobs[jobID]
	if !ok {
		return Job{}, errors.New("job not found")
	}
	return job, nil
}

func (r *InMemoryRepository) ListJobs(_ context.Context) ([]Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]Job, 0, len(r.jobs))
	for _, job := range r.jobs {
		items = append(items, job)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items, nil
}

func (r *InMemoryRepository) CreateEvent(_ context.Context, event Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events[event.JobID] = append(r.events[event.JobID], event)
	return nil
}

func (r *InMemoryRepository) ListEvents(_ context.Context, jobID string) ([]Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := append([]Event(nil), r.events[jobID]...)
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items, nil
}

func newID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(raw[:])
}
