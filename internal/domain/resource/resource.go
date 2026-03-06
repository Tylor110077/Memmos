package resource

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type Type string

const (
	TypeFile Type = "file"
	TypeWeb  Type = "web"
)

type Status string

const (
	StatusUploaded        Status = "uploaded"
	StatusParsing         Status = "parsing"
	StatusNormalizing     Status = "normalizing"
	StatusGraphGenerating Status = "graph_generating"
	StatusCompleted       Status = "completed"
	StatusFailed          Status = "failed"
)

var (
	ErrInvalidGroupID    = errors.New("invalid group id")
	ErrInvalidName       = errors.New("invalid resource name")
	ErrInvalidURL        = errors.New("invalid resource url")
	ErrUnsupportedType   = errors.New("unsupported resource type")
	ErrInvalidTransition = errors.New("invalid resource transition")
	ErrInvalidFailure    = errors.New("invalid failure metadata")
	ErrRetryNotAllowed   = errors.New("retry not allowed")
)

var supportedFileExtensions = map[string]struct{}{
	".pdf":  {},
	".docx": {},
	".pptx": {},
	".xlsx": {},
	".txt":  {},
	".md":   {},
}

type Failure struct {
	Stage   Status
	Message string
}

type Resource struct {
	ID        string
	GroupID   string
	Name      string
	Type      Type
	Status    Status
	SourceURL string
	Failure   *Failure
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewFile(groupID, name string) (*Resource, error) {
	groupID = strings.TrimSpace(groupID)
	name = strings.TrimSpace(name)
	if groupID == "" {
		return nil, ErrInvalidGroupID
	}
	if name == "" {
		return nil, ErrInvalidName
	}
	if _, ok := supportedFileExtensions[strings.ToLower(filepath.Ext(name))]; !ok {
		return nil, ErrUnsupportedType
	}

	now := time.Now().UTC()
	return &Resource{
		GroupID:   groupID,
		Name:      name,
		Type:      TypeFile,
		Status:    StatusUploaded,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func NewWeb(groupID, rawURL string) (*Resource, error) {
	return NewNamedWeb(groupID, rawURL, "")
}

func NewNamedWeb(groupID, rawURL, name string) (*Resource, error) {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, ErrInvalidGroupID
	}

	parsed, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, ErrInvalidURL
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = parsed.String()
	}

	now := time.Now().UTC()
	return &Resource{
		GroupID:   groupID,
		Name:      name,
		Type:      TypeWeb,
		Status:    StatusUploaded,
		SourceURL: parsed.String(),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (r *Resource) MoveTo(next Status, failure Failure) error {
	if !isAllowedTransition(r.Status, next) {
		return ErrInvalidTransition
	}
	if next == StatusFailed {
		if failure.Stage == "" || strings.TrimSpace(failure.Message) == "" {
			return ErrInvalidFailure
		}
		r.Failure = &Failure{
			Stage:   failure.Stage,
			Message: strings.TrimSpace(failure.Message),
		}
	} else {
		r.Failure = nil
	}

	r.Status = next
	r.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *Resource) ResetForRetry() error {
	if r.Status != StatusFailed && r.Status != StatusCompleted {
		return ErrRetryNotAllowed
	}

	r.Status = StatusUploaded
	r.Failure = nil
	r.UpdatedAt = time.Now().UTC()
	return nil
}

func isAllowedTransition(current, next Status) bool {
	switch current {
	case StatusUploaded:
		return next == StatusParsing || next == StatusFailed
	case StatusParsing:
		return next == StatusNormalizing || next == StatusFailed
	case StatusNormalizing:
		return next == StatusGraphGenerating || next == StatusFailed
	case StatusGraphGenerating:
		return next == StatusCompleted || next == StatusFailed
	case StatusCompleted, StatusFailed:
		return false
	default:
		return false
	}
}
