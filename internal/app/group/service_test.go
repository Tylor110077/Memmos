package group

import (
	"context"
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
