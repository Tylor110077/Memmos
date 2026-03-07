package tx

import (
	"context"
	"errors"
	"testing"
)

func TestManagerWithTxCommitsOnSuccess(t *testing.T) {
	db := &fakeBeginner{}
	manager := NewManager(db)

	err := manager.WithTx(context.Background(), func(ctx context.Context, tx Tx) error {
		if tx == nil {
			t.Fatalf("tx should not be nil")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithTx() error = %v", err)
	}

	if db.tx.commitCount != 1 {
		t.Fatalf("commitCount = %d, want 1", db.tx.commitCount)
	}
	if db.tx.rollbackCount != 0 {
		t.Fatalf("rollbackCount = %d, want 0", db.tx.rollbackCount)
	}
}

func TestManagerWithTxRollsBackOnHandlerError(t *testing.T) {
	db := &fakeBeginner{}
	manager := NewManager(db)
	expected := errors.New("boom")

	err := manager.WithTx(context.Background(), func(ctx context.Context, tx Tx) error {
		return expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("WithTx() error = %v, want %v", err, expected)
	}

	if db.tx.commitCount != 0 {
		t.Fatalf("commitCount = %d, want 0", db.tx.commitCount)
	}
	if db.tx.rollbackCount != 1 {
		t.Fatalf("rollbackCount = %d, want 1", db.tx.rollbackCount)
	}
}

func TestManagerWithTxRollsBackWhenCommitFails(t *testing.T) {
	db := &fakeBeginner{tx: &fakeTx{commitErr: errors.New("commit failed")}}
	manager := NewManager(db)

	err := manager.WithTx(context.Background(), func(ctx context.Context, tx Tx) error {
		return nil
	})
	if err == nil {
		t.Fatalf("expected commit failure")
	}

	if db.tx.rollbackCount != 1 {
		t.Fatalf("rollbackCount = %d, want 1", db.tx.rollbackCount)
	}
}

type fakeBeginner struct {
	tx *fakeTx
}

func (f *fakeBeginner) BeginTx(_ context.Context) (Tx, error) {
	if f.tx == nil {
		f.tx = &fakeTx{}
	}
	return f.tx, nil
}

type fakeTx struct {
	commitCount   int
	rollbackCount int
	commitErr     error
	rollbackErr   error
}

func (f *fakeTx) Commit(context.Context) error {
	f.commitCount++
	return f.commitErr
}

func (f *fakeTx) Rollback(context.Context) error {
	f.rollbackCount++
	return f.rollbackErr
}
