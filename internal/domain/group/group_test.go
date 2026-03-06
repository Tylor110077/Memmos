package group

import "testing"

func TestNewValidatesName(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if _, err := New("   "); err == nil {
			t.Fatalf("expected error for empty name")
		}
	})

	t.Run("too long", func(t *testing.T) {
		if _, err := New("12345678901234567890123456789012345678901234567890123456789012345"); err == nil {
			t.Fatalf("expected error for too long name")
		}
	})
}

func TestNewNormalizesName(t *testing.T) {
	entity, err := New("  Backend MVP  ")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if entity.Name != "Backend MVP" {
		t.Fatalf("Name = %q, want Backend MVP", entity.Name)
	}
}

func TestRenameValidatesName(t *testing.T) {
	entity, err := New("Initial")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := entity.Rename(""); err == nil {
		t.Fatalf("expected rename validation error")
	}
}
