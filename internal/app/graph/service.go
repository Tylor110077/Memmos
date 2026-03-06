package graph

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sort"
	"sync"
	"time"

	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

type Graph struct {
	ID         string    `json:"id"`
	GroupID    string    `json:"group_id"`
	ResourceID string    `json:"resource_id"`
	Title      string    `json:"title"`
	Summary    string    `json:"summary"`
	Version    int       `json:"version"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Node struct {
	ID          string `json:"id"`
	GraphID     string `json:"graph_id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Meaning     string `json:"meaning,omitempty"`
	Level       int    `json:"level"`
}

type Edge struct {
	ID       string `json:"id"`
	GraphID  string `json:"graph_id"`
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Relation string `json:"relation"`
}

type SavedGraph struct {
	Graph Graph
	Nodes []Node
	Edges []Edge
}

type SaveInput struct {
	GroupID    string
	ResourceID string
	Title      string
	Document   pipelinegraph.Document
}

type Repository interface {
	LatestResourceGraph(ctx context.Context, resourceID string) (*SavedGraph, error)
	Save(ctx context.Context, graph SavedGraph) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SaveResourceGraph(ctx context.Context, input SaveInput) (SavedGraph, error) {
	if err := input.Document.Validate(); err != nil {
		return SavedGraph{}, err
	}

	now := time.Now().UTC()
	version := 1
	if latest, err := s.repo.LatestResourceGraph(ctx, input.ResourceID); err == nil && latest != nil {
		latest.Graph.IsActive = false
		latest.Graph.UpdatedAt = now
		if err := s.repo.Save(ctx, *latest); err != nil {
			return SavedGraph{}, err
		}
		version = latest.Graph.Version + 1
	}

	graphID := newID()
	saved := SavedGraph{
		Graph: Graph{
			ID:         graphID,
			GroupID:    input.GroupID,
			ResourceID: input.ResourceID,
			Title:      input.Title,
			Summary:    input.Document.Summary,
			Version:    version,
			IsActive:   true,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		Nodes: make([]Node, 0, len(input.Document.Nodes)),
		Edges: make([]Edge, 0, len(input.Document.Edges)),
	}
	for _, node := range input.Document.Nodes {
		saved.Nodes = append(saved.Nodes, Node{
			ID:          node.ID,
			GraphID:     graphID,
			Name:        node.Name,
			Type:        node.Type,
			Description: node.Description,
			Meaning:     node.Meaning,
			Level:       node.Level,
		})
	}
	for _, edge := range input.Document.Edges {
		saved.Edges = append(saved.Edges, Edge{
			ID:       edge.ID,
			GraphID:  graphID,
			SourceID: edge.SourceID,
			TargetID: edge.TargetID,
			Relation: edge.Relation,
		})
	}
	if err := s.repo.Save(ctx, saved); err != nil {
		return SavedGraph{}, err
	}
	return saved, nil
}

type InMemoryRepository struct {
	mu     sync.RWMutex
	graphs map[string]SavedGraph
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{graphs: map[string]SavedGraph{}}
}

func (r *InMemoryRepository) LatestResourceGraph(_ context.Context, resourceID string) (*SavedGraph, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var matches []SavedGraph
	for _, graph := range r.graphs {
		if graph.Graph.ResourceID == resourceID {
			matches = append(matches, graph)
		}
	}
	if len(matches) == 0 {
		return nil, nil
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].Graph.Version > matches[j].Graph.Version })
	copy := matches[0]
	return &copy, nil
}

func (r *InMemoryRepository) Save(_ context.Context, graph SavedGraph) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.graphs[graph.Graph.ID] = graph
	return nil
}

func (r *InMemoryRepository) GetByID(id string) (SavedGraph, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	graph, ok := r.graphs[id]
	return graph, ok
}

func newID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(raw[:])
}
