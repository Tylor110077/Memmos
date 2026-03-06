package group

import (
	"context"
	"errors"
	"strings"
	"testing"

	domaingroup "github.com/tylor/goaipj/internal/domain/group"
)

func TestServiceCreateListGetUpdateDelete(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	created, err := svc.CreateGroup(ctx, CreateGroupInput{Name: "MVP"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	if created.Name != "MVP" {
		t.Fatalf("created name = %q, want MVP", created.Name)
	}

	groups, err := svc.ListGroups(ctx)
	if err != nil {
		t.Fatalf("ListGroups() error = %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("len(groups) = %d, want 1", len(groups))
	}

	got, err := svc.GetGroup(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetGroup() error = %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("got id = %q, want %q", got.ID, created.ID)
	}

	updated, err := svc.UpdateGroup(ctx, created.ID, UpdateGroupInput{Name: "Renamed"})
	if err != nil {
		t.Fatalf("UpdateGroup() error = %v", err)
	}
	if updated.Name != "Renamed" {
		t.Fatalf("updated name = %q, want Renamed", updated.Name)
	}

	if err := svc.DeleteGroup(ctx, created.ID); err != nil {
		t.Fatalf("DeleteGroup() error = %v", err)
	}

	if _, err := svc.GetGroup(ctx, created.ID); err == nil {
		t.Fatalf("expected not found after delete")
	}
}

func TestServiceReturnsConflictWhenNameAlreadyExists(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	if _, err := svc.CreateGroup(ctx, CreateGroupInput{Name: "Duplicate"}); err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	if _, err := svc.CreateGroup(ctx, CreateGroupInput{Name: "Duplicate"}); err == nil {
		t.Fatalf("expected conflict error")
	}
}

func TestServiceRejectsUnknownID(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)

	if _, err := svc.GetGroup(context.Background(), "missing"); err == nil {
		t.Fatalf("expected not found")
	}
}

func TestServiceRejectsInvalidNamesOnCreateAndUpdate(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	if _, err := svc.CreateGroup(ctx, CreateGroupInput{Name: "   "}); err == nil {
		t.Fatalf("expected create validation error")
	}

	tooLong := strings.Repeat("知", 65)
	if _, err := svc.CreateGroup(ctx, CreateGroupInput{Name: tooLong}); err == nil {
		t.Fatalf("expected max length validation error")
	}

	created, err := svc.CreateGroup(ctx, CreateGroupInput{Name: "Valid"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	if _, err := svc.UpdateGroup(ctx, created.ID, UpdateGroupInput{Name: ""}); err == nil {
		t.Fatalf("expected update validation error")
	}
}

func TestServiceUpdateRejectsConflictingName(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	first, err := svc.CreateGroup(ctx, CreateGroupInput{Name: "Alpha"})
	if err != nil {
		t.Fatalf("CreateGroup(first) error = %v", err)
	}
	if _, err := svc.CreateGroup(ctx, CreateGroupInput{Name: "Beta"}); err != nil {
		t.Fatalf("CreateGroup(second) error = %v", err)
	}

	if _, err := svc.UpdateGroup(ctx, first.ID, UpdateGroupInput{Name: "beta"}); err == nil {
		t.Fatalf("expected name conflict on update")
	}
}

func TestServiceDeleteGroupCascadesCleanup(t *testing.T) {
	repo := NewInMemoryRepository()
	cleanup := &stubCleanup{}
	svc := NewService(repo, cleanup)
	ctx := context.Background()

	created, err := svc.CreateGroup(ctx, CreateGroupInput{Name: "Delete Me"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	if err := svc.DeleteGroup(ctx, created.ID); err != nil {
		t.Fatalf("DeleteGroup() error = %v", err)
	}

	wantCalls := []string{"resources", "graphs", "conversations", "jobs"}
	if strings.Join(cleanup.calls, ",") != strings.Join(wantCalls, ",") {
		t.Fatalf("cleanup calls = %v, want %v", cleanup.calls, wantCalls)
	}
	if cleanup.groupIDs["resources"] != created.ID {
		t.Fatalf("resources cleanup group_id = %q, want %q", cleanup.groupIDs["resources"], created.ID)
	}
	if _, err := svc.GetGroup(ctx, created.ID); err == nil {
		t.Fatalf("expected group removed after delete")
	}
}

func TestServiceDeleteGroupStopsWhenCleanupFails(t *testing.T) {
	repo := NewInMemoryRepository()
	cleanup := &stubCleanup{failAt: "graphs"}
	svc := NewService(repo, cleanup)
	ctx := context.Background()

	created, err := svc.CreateGroup(ctx, CreateGroupInput{Name: "Delete Me"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	if err := svc.DeleteGroup(ctx, created.ID); err == nil {
		t.Fatalf("expected cleanup failure")
	}
	if _, err := svc.GetGroup(ctx, created.ID); err != nil {
		t.Fatalf("expected group to remain when cleanup fails: %v", err)
	}
	if strings.Join(cleanup.calls, ",") != "resources,graphs" {
		t.Fatalf("cleanup calls = %v, want resources,graphs", cleanup.calls)
	}
}

func TestInMemoryRepositoryMatchesDomainEntity(t *testing.T) {
	entity, err := domaingroup.New("Repo")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	repo := NewInMemoryRepository()
	if err := repo.Create(context.Background(), *entity); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

type stubCleanup struct {
	calls    []string
	groupIDs map[string]string
	failAt   string
}

func (s *stubCleanup) DeleteGroupResources(_ context.Context, groupID string) error {
	return s.record("resources", groupID)
}

func (s *stubCleanup) DeleteGroupGraphs(_ context.Context, groupID string) error {
	return s.record("graphs", groupID)
}

func (s *stubCleanup) DeleteGroupConversations(_ context.Context, groupID string) error {
	return s.record("conversations", groupID)
}

func (s *stubCleanup) DeleteGroupJobs(_ context.Context, groupID string) error {
	return s.record("jobs", groupID)
}

func (s *stubCleanup) record(name string, groupID string) error {
	s.calls = append(s.calls, name)
	if s.groupIDs == nil {
		s.groupIDs = map[string]string{}
	}
	s.groupIDs[name] = groupID
	if s.failAt == name {
		return errors.New(name + " cleanup failed")
	}
	return nil
}
