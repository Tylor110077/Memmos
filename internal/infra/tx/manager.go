package tx

import "context"

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Beginner interface {
	BeginTx(ctx context.Context) (Tx, error)
}

type Manager struct {
	beginner Beginner
}

func NewManager(beginner Beginner) *Manager {
	return &Manager{beginner: beginner}
}

func (m *Manager) WithTx(ctx context.Context, fn func(ctx context.Context, tx Tx) error) (err error) {
	tx, err := m.beginner.BeginTx(ctx)
	if err != nil {
		return err
	}

	rolledBack := false
	defer func() {
		if err == nil || rolledBack {
			return
		}
		_ = tx.Rollback(ctx)
	}()

	if err = fn(ctx, tx); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		rolledBack = true
		return err
	}
	return nil
}
