package resource

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"sort"
	"sync"
	"time"

	appgroup "github.com/tylor/goaipj/internal/app/group"
	domainresource "github.com/tylor/goaipj/internal/domain/resource"
	"github.com/tylor/goaipj/internal/infra/apperror"
	"github.com/tylor/goaipj/internal/infra/storage"
)

const (
	JobTypeParseResource    = "parse_resource"
	JobTypeFetchWebResource = "fetch_web_resource"
)

type Resource struct {
	ID           string                `json:"id"`
	GroupID      string                `json:"group_id"`
	Name         string                `json:"name"`
	Type         domainresource.Type   `json:"type"`
	Status       domainresource.Status `json:"status"`
	SourceURL    string                `json:"source_url,omitempty"`
	FailedStage  string                `json:"failed_stage,omitempty"`
	ErrorMessage string                `json:"error_message,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

type Artifact struct {
	ID          string
	ResourceID  string
	Type        string
	StorageKey  string
	ContentType string
	CreatedAt   time.Time
}

type Job struct {
	ID         string    `json:"id"`
	ResourceID string    `json:"resource_id"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type ResourceWithJob struct {
	Resource Resource `json:"resource"`
	Job      Job      `json:"job"`
}

type UploadFileInput struct {
	GroupID     string
	Filename    string
	ContentType string
	Body        io.Reader
	Size        int64
}

type CreateWebResourceInput struct {
	GroupID string
	URL     string
	Name    string
}

type PresignUploadInput struct {
	GroupID     string
	Filename    string
	ContentType string
	ExpiresIn   time.Duration
}

type PresignedUpload struct {
	Resource  Resource          `json:"resource"`
	UploadURL string            `json:"upload_url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	ObjectKey string            `json:"object_key"`
	ExpiresAt time.Time         `json:"expires_at"`
}

type GroupLookup interface {
	GetGroup(ctx context.Context, id string) (appgroup.Group, error)
}

type Repository interface {
	Create(ctx context.Context, entity domainresource.Resource) error
	Get(ctx context.Context, id string) (*domainresource.Resource, error)
	Update(ctx context.Context, entity domainresource.Resource) error
	ListByGroup(ctx context.Context, groupID string) ([]domainresource.Resource, error)
	Delete(ctx context.Context, id string) error
}

type ArtifactRepository interface {
	Create(ctx context.Context, artifact Artifact) error
	ListByResource(ctx context.Context, resourceID string) ([]Artifact, error)
	DeleteByResource(ctx context.Context, resourceID string) error
}

type JobPublisher interface {
	Publish(ctx context.Context, jobType string, resourceID string) (Job, error)
}

type ObjectStore interface {
	PutObject(ctx context.Context, input storage.PutObjectInput) (storage.ObjectMeta, error)
	DeleteObject(ctx context.Context, key string) error
	PresignPutObject(ctx context.Context, input storage.PresignPutObjectInput) (storage.PresignedUpload, error)
}

type Service struct {
	groupLookup GroupLookup
	repo        Repository
	artifacts   ArtifactRepository
	jobs        JobPublisher
	store       ObjectStore
}

func NewService(groupLookup GroupLookup, repo Repository, artifacts ArtifactRepository, jobs JobPublisher, store ObjectStore) *Service {
	return &Service{
		groupLookup: groupLookup,
		repo:        repo,
		artifacts:   artifacts,
		jobs:        jobs,
		store:       store,
	}
}

func (s *Service) UploadFile(ctx context.Context, input UploadFileInput) (ResourceWithJob, error) {
	if _, err := s.groupLookup.GetGroup(ctx, input.GroupID); err != nil {
		return ResourceWithJob{}, apperror.New(apperror.CodeNotFound, "group not found")
	}

	entity, err := domainresource.NewFile(input.GroupID, input.Filename)
	if err != nil {
		return ResourceWithJob{}, mapDomainError(err)
	}
	entity.ID = newID()

	if _, err := s.store.PutObject(ctx, storage.PutObjectInput{
		Key:         storage.ResourceObjectKey(entity.GroupID, entity.ID, storage.KindRaw, input.Filename),
		ContentType: input.ContentType,
		Body:        input.Body,
		Size:        input.Size,
	}); err != nil {
		return ResourceWithJob{}, apperror.Wrap(apperror.CodeInternal, "store uploaded file", err)
	}

	if err := s.repo.Create(ctx, *entity); err != nil {
		return ResourceWithJob{}, apperror.Wrap(apperror.CodeInternal, "create resource", err)
	}

	if err := s.artifacts.Create(ctx, Artifact{
		ID:          newID(),
		ResourceID:  entity.ID,
		Type:        "raw_file",
		StorageKey:  storage.ResourceObjectKey(entity.GroupID, entity.ID, storage.KindRaw, input.Filename),
		ContentType: input.ContentType,
		CreatedAt:   time.Now().UTC(),
	}); err != nil {
		return ResourceWithJob{}, apperror.Wrap(apperror.CodeInternal, "create artifact", err)
	}

	job, err := s.jobs.Publish(ctx, JobTypeParseResource, entity.ID)
	if err != nil {
		return ResourceWithJob{}, apperror.Wrap(apperror.CodeInternal, "publish parse job", err)
	}
	return ResourceWithJob{Resource: toDTO(*entity), Job: job}, nil
}

func (s *Service) CreateWebResource(ctx context.Context, input CreateWebResourceInput) (ResourceWithJob, error) {
	if _, err := s.groupLookup.GetGroup(ctx, input.GroupID); err != nil {
		return ResourceWithJob{}, apperror.New(apperror.CodeNotFound, "group not found")
	}

	entity, err := domainresource.NewNamedWeb(input.GroupID, input.URL, input.Name)
	if err != nil {
		return ResourceWithJob{}, mapDomainError(err)
	}
	entity.ID = newID()

	if err := s.repo.Create(ctx, *entity); err != nil {
		return ResourceWithJob{}, apperror.Wrap(apperror.CodeInternal, "create resource", err)
	}

	job, err := s.jobs.Publish(ctx, JobTypeFetchWebResource, entity.ID)
	if err != nil {
		return ResourceWithJob{}, apperror.Wrap(apperror.CodeInternal, "publish fetch job", err)
	}
	return ResourceWithJob{Resource: toDTO(*entity), Job: job}, nil
}

func (s *Service) ListResources(ctx context.Context, groupID string) ([]Resource, error) {
	items, err := s.repo.ListByGroup(ctx, groupID)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "list resources", err)
	}
	out := make([]Resource, 0, len(items))
	for _, item := range items {
		out = append(out, toDTO(item))
	}
	return out, nil
}

func (s *Service) GetResource(ctx context.Context, groupID, resourceID string) (Resource, error) {
	item, err := s.repo.Get(ctx, resourceID)
	if err != nil {
		return Resource{}, apperror.New(apperror.CodeNotFound, "resource not found")
	}
	if groupID != "" && item.GroupID != groupID {
		return Resource{}, apperror.New(apperror.CodeNotFound, "resource not found")
	}
	return toDTO(*item), nil
}

func (s *Service) GetResourceByID(ctx context.Context, resourceID string) (Resource, error) {
	return s.GetResource(ctx, "", resourceID)
}

func (s *Service) RetryResource(ctx context.Context, resourceID string) (ResourceWithJob, error) {
	item, err := s.repo.Get(ctx, resourceID)
	if err != nil {
		return ResourceWithJob{}, apperror.New(apperror.CodeNotFound, "resource not found")
	}
	if err := item.ResetForRetry(); err != nil {
		return ResourceWithJob{}, apperror.Wrap(apperror.CodeInvalidArgument, "resource is not retryable", err)
	}
	if err := s.repo.Update(ctx, *item); err != nil {
		return ResourceWithJob{}, apperror.Wrap(apperror.CodeInternal, "update resource", err)
	}

	jobType := JobTypeParseResource
	if item.Type == domainresource.TypeWeb {
		jobType = JobTypeFetchWebResource
	}
	job, err := s.jobs.Publish(ctx, jobType, item.ID)
	if err != nil {
		return ResourceWithJob{}, apperror.Wrap(apperror.CodeInternal, "publish retry job", err)
	}
	return ResourceWithJob{Resource: toDTO(*item), Job: job}, nil
}

func (s *Service) DeleteResource(ctx context.Context, resourceID string) error {
	item, err := s.repo.Get(ctx, resourceID)
	if err != nil {
		return apperror.New(apperror.CodeNotFound, "resource not found")
	}
	artifacts, err := s.artifacts.ListByResource(ctx, resourceID)
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "list artifacts", err)
	}
	for _, artifact := range artifacts {
		if artifact.StorageKey == "" {
			continue
		}
		if err := s.store.DeleteObject(ctx, artifact.StorageKey); err != nil {
			return apperror.Wrap(apperror.CodeInternal, "delete stored artifact", err)
		}
	}
	if err := s.artifacts.DeleteByResource(ctx, resourceID); err != nil {
		return apperror.Wrap(apperror.CodeInternal, "delete artifacts", err)
	}
	if err := s.repo.Delete(ctx, item.ID); err != nil {
		return apperror.Wrap(apperror.CodeInternal, "delete resource", err)
	}
	return nil
}

func (s *Service) PresignUpload(ctx context.Context, input PresignUploadInput) (PresignedUpload, error) {
	if _, err := s.groupLookup.GetGroup(ctx, input.GroupID); err != nil {
		return PresignedUpload{}, apperror.New(apperror.CodeNotFound, "group not found")
	}
	entity, err := domainresource.NewFile(input.GroupID, input.Filename)
	if err != nil {
		return PresignedUpload{}, mapDomainError(err)
	}
	entity.ID = newID()
	objectKey := storage.ResourceObjectKey(entity.GroupID, entity.ID, storage.KindRaw, input.Filename)
	plan, err := s.store.PresignPutObject(ctx, storage.PresignPutObjectInput{
		Key:         objectKey,
		ContentType: input.ContentType,
		ExpiresIn:   input.ExpiresIn,
	})
	if err != nil {
		return PresignedUpload{}, apperror.Wrap(apperror.CodeInternal, "presign upload", err)
	}
	if err := s.repo.Create(ctx, *entity); err != nil {
		return PresignedUpload{}, apperror.Wrap(apperror.CodeInternal, "create resource", err)
	}
	if err := s.artifacts.Create(ctx, Artifact{
		ID:          newID(),
		ResourceID:  entity.ID,
		Type:        "raw_file",
		StorageKey:  objectKey,
		ContentType: input.ContentType,
		CreatedAt:   time.Now().UTC(),
	}); err != nil {
		return PresignedUpload{}, apperror.Wrap(apperror.CodeInternal, "create artifact", err)
	}
	return PresignedUpload{
		Resource:  toDTO(*entity),
		UploadURL: plan.URL,
		Method:    plan.Method,
		Headers:   plan.Headers,
		ObjectKey: objectKey,
		ExpiresAt: plan.ExpiresAt,
	}, nil
}

func (s *Service) CompletePresignedUpload(ctx context.Context, resourceID string) (ResourceWithJob, error) {
	item, err := s.repo.Get(ctx, resourceID)
	if err != nil {
		return ResourceWithJob{}, apperror.New(apperror.CodeNotFound, "resource not found")
	}
	job, err := s.jobs.Publish(ctx, JobTypeParseResource, item.ID)
	if err != nil {
		return ResourceWithJob{}, apperror.Wrap(apperror.CodeInternal, "publish parse job", err)
	}
	return ResourceWithJob{Resource: toDTO(*item), Job: job}, nil
}

func (s *Service) UpdateStatus(ctx context.Context, resourceID string, next domainresource.Status) (Resource, error) {
	item, err := s.repo.Get(ctx, resourceID)
	if err != nil {
		return Resource{}, apperror.New(apperror.CodeNotFound, "resource not found")
	}
	if err := item.MoveTo(next, domainresource.Failure{}); err != nil {
		return Resource{}, apperror.Wrap(apperror.CodeInvalidArgument, "invalid resource status transition", err)
	}
	if err := s.repo.Update(ctx, *item); err != nil {
		return Resource{}, apperror.Wrap(apperror.CodeInternal, "update resource", err)
	}
	return toDTO(*item), nil
}

func (s *Service) FailResource(ctx context.Context, resourceID string, stage domainresource.Status, message string) (Resource, error) {
	item, err := s.repo.Get(ctx, resourceID)
	if err != nil {
		return Resource{}, apperror.New(apperror.CodeNotFound, "resource not found")
	}
	if err := item.MoveTo(domainresource.StatusFailed, domainresource.Failure{Stage: stage, Message: message}); err != nil {
		return Resource{}, apperror.Wrap(apperror.CodeInvalidArgument, "invalid resource failure transition", err)
	}
	if err := s.repo.Update(ctx, *item); err != nil {
		return Resource{}, apperror.Wrap(apperror.CodeInternal, "update resource", err)
	}
	return toDTO(*item), nil
}

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domainresource.ErrInvalidGroupID),
		errors.Is(err, domainresource.ErrInvalidName),
		errors.Is(err, domainresource.ErrInvalidURL),
		errors.Is(err, domainresource.ErrUnsupportedType),
		errors.Is(err, domainresource.ErrInvalidTransition),
		errors.Is(err, domainresource.ErrInvalidFailure),
		errors.Is(err, domainresource.ErrRetryNotAllowed):
		return apperror.Wrap(apperror.CodeInvalidArgument, err.Error(), err)
	default:
		return apperror.Wrap(apperror.CodeInternal, "resource error", err)
	}
}

func toDTO(item domainresource.Resource) Resource {
	dto := Resource{
		ID:        item.ID,
		GroupID:   item.GroupID,
		Name:      item.Name,
		Type:      item.Type,
		Status:    item.Status,
		SourceURL: item.SourceURL,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
	if item.Failure != nil {
		dto.FailedStage = string(item.Failure.Stage)
		dto.ErrorMessage = item.Failure.Message
	}
	return dto
}

type InMemoryRepository struct {
	mu        sync.RWMutex
	resources map[string]domainresource.Resource
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{resources: map[string]domainresource.Resource{}}
}

func (r *InMemoryRepository) Create(_ context.Context, entity domainresource.Resource) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources[entity.ID] = entity
	return nil
}

func (r *InMemoryRepository) Get(_ context.Context, id string) (*domainresource.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.resources[id]
	if !ok {
		return nil, errors.New("not found")
	}
	copy := item
	return &copy, nil
}

func (r *InMemoryRepository) Update(_ context.Context, entity domainresource.Resource) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources[entity.ID] = entity
	return nil
}

func (r *InMemoryRepository) ListByGroup(_ context.Context, groupID string) ([]domainresource.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]domainresource.Resource, 0)
	for _, item := range r.resources {
		if item.GroupID == groupID {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items, nil
}

func (r *InMemoryRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.resources[id]; !ok {
		return errors.New("not found")
	}
	delete(r.resources, id)
	return nil
}

type InMemoryArtifactRepository struct {
	mu        sync.RWMutex
	artifacts []Artifact
}

func NewInMemoryArtifactRepository() *InMemoryArtifactRepository {
	return &InMemoryArtifactRepository{}
}

func (r *InMemoryArtifactRepository) Create(_ context.Context, artifact Artifact) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.artifacts = append(r.artifacts, artifact)
	return nil
}

func (r *InMemoryArtifactRepository) ListByResource(_ context.Context, resourceID string) ([]Artifact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]Artifact, 0)
	for _, artifact := range r.artifacts {
		if artifact.ResourceID == resourceID {
			items = append(items, artifact)
		}
	}
	return items, nil
}

func (r *InMemoryArtifactRepository) DeleteByResource(_ context.Context, resourceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	filtered := r.artifacts[:0]
	for _, artifact := range r.artifacts {
		if artifact.ResourceID == resourceID {
			continue
		}
		filtered = append(filtered, artifact)
	}
	r.artifacts = filtered
	return nil
}

func (r *InMemoryArtifactRepository) List() []Artifact {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Artifact, len(r.artifacts))
	copy(out, r.artifacts)
	return out
}

type InMemoryJobPublisher struct {
	mu   sync.Mutex
	jobs []Job
}

func NewInMemoryJobPublisher() *InMemoryJobPublisher {
	return &InMemoryJobPublisher{}
}

func (p *InMemoryJobPublisher) Publish(_ context.Context, jobType string, resourceID string) (Job, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	job := Job{
		ID:         newID(),
		ResourceID: resourceID,
		Type:       jobType,
		Status:     "queued",
		CreatedAt:  time.Now().UTC(),
	}
	p.jobs = append(p.jobs, job)
	return job, nil
}

func newID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(raw[:])
}
