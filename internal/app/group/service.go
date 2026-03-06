package group

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	domaingroup "github.com/tylor/goaipj/internal/domain/group"
	"github.com/tylor/goaipj/internal/infra/apperror"
)

type Group struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateGroupInput struct {
	Name string `json:"name"`
}

type UpdateGroupInput struct {
	Name string `json:"name"`
}

type Repository interface {
	Create(ctx context.Context, entity domaingroup.Group) error
	List(ctx context.Context) ([]domaingroup.Group, error)
	Get(ctx context.Context, id string) (domaingroup.Group, error)
	Update(ctx context.Context, entity domaingroup.Group) error
	Delete(ctx context.Context, id string) error
	ExistsByName(ctx context.Context, name string, excludeID string) (bool, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateGroup(ctx context.Context, input CreateGroupInput) (Group, error) {
	entity, err := domaingroup.New(input.Name)
	if err != nil {
		return Group{}, apperror.Wrap(apperror.CodeInvalidArgument, "group name is invalid", err)
	}

	exists, err := s.repo.ExistsByName(ctx, entity.Name, "")
	if err != nil {
		return Group{}, apperror.Wrap(apperror.CodeInternal, "check group name conflict", err)
	}
	if exists {
		return Group{}, apperror.New(apperror.CodeConflict, "group name already exists")
	}

	entity.ID = newID()
	if err := s.repo.Create(ctx, *entity); err != nil {
		return Group{}, apperror.Wrap(apperror.CodeInternal, "create group", err)
	}
	return toDTO(*entity), nil
}

func (s *Service) ListGroups(ctx context.Context) ([]Group, error) {
	entities, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "list groups", err)
	}

	out := make([]Group, 0, len(entities))
	for _, entity := range entities {
		out = append(out, toDTO(entity))
	}
	return out, nil
}

func (s *Service) GetGroup(ctx context.Context, id string) (Group, error) {
	entity, err := s.repo.Get(ctx, id)
	if err != nil {
		return Group{}, mapRepositoryError(err, "get group")
	}
	return toDTO(entity), nil
}

func (s *Service) UpdateGroup(ctx context.Context, id string, input UpdateGroupInput) (Group, error) {
	entity, err := s.repo.Get(ctx, id)
	if err != nil {
		return Group{}, mapRepositoryError(err, "get group")
	}

	if err := entity.Rename(input.Name); err != nil {
		return Group{}, apperror.Wrap(apperror.CodeInvalidArgument, "group name is invalid", err)
	}

	exists, err := s.repo.ExistsByName(ctx, entity.Name, id)
	if err != nil {
		return Group{}, apperror.Wrap(apperror.CodeInternal, "check group name conflict", err)
	}
	if exists {
		return Group{}, apperror.New(apperror.CodeConflict, "group name already exists")
	}

	if err := s.repo.Update(ctx, entity); err != nil {
		return Group{}, mapRepositoryError(err, "update group")
	}
	return toDTO(entity), nil
}

func (s *Service) DeleteGroup(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return mapRepositoryError(err, "delete group")
	}
	return nil
}

func toDTO(entity domaingroup.Group) Group {
	return Group{
		ID:        entity.ID,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}

var errNotFound = errors.New("group not found")

func mapRepositoryError(err error, message string) error {
	switch {
	case errors.Is(err, errNotFound):
		return apperror.Wrap(apperror.CodeNotFound, "group not found", err)
	default:
		return apperror.Wrap(apperror.CodeInternal, message, err)
	}
}

type InMemoryRepository struct {
	mu     sync.RWMutex
	groups map[string]domaingroup.Group
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		groups: map[string]domaingroup.Group{},
	}
}

func (r *InMemoryRepository) Create(_ context.Context, entity domaingroup.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.groups[entity.ID] = entity
	return nil
}

func (r *InMemoryRepository) List(_ context.Context) ([]domaingroup.Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]domaingroup.Group, 0, len(r.groups))
	for _, entity := range r.groups {
		items = append(items, entity)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items, nil
}

func (r *InMemoryRepository) Get(_ context.Context, id string) (domaingroup.Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entity, ok := r.groups[id]
	if !ok {
		return domaingroup.Group{}, errNotFound
	}
	return entity, nil
}

func (r *InMemoryRepository) Update(_ context.Context, entity domaingroup.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.groups[entity.ID]; !ok {
		return errNotFound
	}
	r.groups[entity.ID] = entity
	return nil
}

func (r *InMemoryRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.groups[id]; !ok {
		return errNotFound
	}
	delete(r.groups, id)
	return nil
}

func (r *InMemoryRepository) ExistsByName(_ context.Context, name string, excludeID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, entity := range r.groups {
		if entity.ID == excludeID {
			continue
		}
		if strings.EqualFold(entity.Name, name) {
			return true, nil
		}
	}
	return false, nil
}

func newID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmtFallbackID()
	}
	return hex.EncodeToString(raw[:])
}

func fmtFallbackID() string {
	return time.Now().UTC().Format("20060102150405.000000000")
}
