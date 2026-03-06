package group

import (
	"errors"
	"strings"
	"time"
)

const maxNameLength = 64

var ErrInvalidName = errors.New("invalid group name")

type Group struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func New(name string) (*Group, error) {
	normalized, err := normalizeName(name)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Group{
		Name:      normalized,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (g *Group) Rename(name string) error {
	normalized, err := normalizeName(name)
	if err != nil {
		return err
	}

	g.Name = normalized
	g.UpdatedAt = time.Now().UTC()
	return nil
}

func normalizeName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || len([]rune(trimmed)) > maxNameLength {
		return "", ErrInvalidName
	}
	return trimmed, nil
}
