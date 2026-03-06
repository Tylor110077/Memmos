package graph

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
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
	IsExpansion bool   `json:"is_expansion"`
}

type Edge struct {
	ID          string `json:"id"`
	GraphID     string `json:"graph_id"`
	SourceID    string `json:"source_id"`
	TargetID    string `json:"target_id"`
	Relation    string `json:"relation"`
	IsExpansion bool   `json:"is_expansion"`
}

type SavedGraph struct {
	Graph Graph
	Nodes []Node
	Edges []Edge
}

type Example struct {
	ID      string `json:"id"`
	NodeID  string `json:"node_id"`
	Content string `json:"content"`
}

type Neighbor struct {
	Node     Node   `json:"node"`
	Relation string `json:"relation"`
}

type NodeDetail struct {
	Node      Node       `json:"node"`
	Neighbors []Neighbor `json:"neighbors"`
	Examples  []Example  `json:"examples"`
}

type SaveInput struct {
	GroupID    string
	ResourceID string
	Title      string
	Document   pipelinegraph.Document
}

type SaveFrameworkInput struct {
	GroupID  string
	Title    string
	Document pipelinegraph.Document
}

type QueryOptions struct {
	MaxLevel         int
	IncludeExpansion bool
}

type Repository interface {
	LatestResourceGraph(ctx context.Context, resourceID string) (*SavedGraph, error)
	LatestFrameworkGraph(ctx context.Context, groupID string) (*SavedGraph, error)
	Save(ctx context.Context, graph SavedGraph) error
	ListResourceGraphsByGroup(ctx context.Context, groupID string) ([]SavedGraph, error)
	ListExamples(ctx context.Context, graphID string, nodeID string) ([]Example, error)
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
			IsExpansion: false,
		})
	}
	for _, edge := range input.Document.Edges {
		saved.Edges = append(saved.Edges, Edge{
			ID:          edge.ID,
			GraphID:     graphID,
			SourceID:    edge.SourceID,
			TargetID:    edge.TargetID,
			Relation:    edge.Relation,
			IsExpansion: false,
		})
	}
	if err := s.repo.Save(ctx, saved); err != nil {
		return SavedGraph{}, err
	}
	return saved, nil
}

func (s *Service) SaveFrameworkGraph(ctx context.Context, input SaveFrameworkInput) (SavedGraph, error) {
	if err := input.Document.Validate(); err != nil {
		return SavedGraph{}, err
	}
	now := time.Now().UTC()
	version := 1
	if latest, err := s.repo.LatestFrameworkGraph(ctx, input.GroupID); err == nil && latest != nil {
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
			ID:        graphID,
			GroupID:   input.GroupID,
			Title:     input.Title,
			Summary:   input.Document.Summary,
			Version:   version,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
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

func (s *Service) ExpandNode(ctx context.Context, graphID, nodeID string, expansion pipelinegraph.Expansion) (SavedGraph, error) {
	graph, ok := s.getGraphByID(graphID)
	if !ok {
		return SavedGraph{}, errors.New("graph not found")
	}
	nodeExists := false
	for _, node := range graph.Nodes {
		if node.ID == nodeID {
			nodeExists = true
			break
		}
	}
	if !nodeExists {
		return SavedGraph{}, errors.New("node not found")
	}
	graph.Nodes = append(graph.Nodes, Node{
		ID:          expansion.Node.ID,
		GraphID:     graphID,
		Name:        expansion.Node.Name,
		Type:        expansion.Node.Type,
		Description: expansion.Node.Description,
		Meaning:     expansion.Node.Meaning,
		Level:       expansion.Node.Level,
		IsExpansion: true,
	})
	graph.Edges = append(graph.Edges, Edge{
		ID:          expansion.Edge.ID,
		GraphID:     graphID,
		SourceID:    expansion.Edge.SourceID,
		TargetID:    expansion.Edge.TargetID,
		Relation:    expansion.Edge.Relation,
		IsExpansion: true,
	})
	graph.Graph.UpdatedAt = time.Now().UTC()
	if err := s.repo.Save(ctx, graph); err != nil {
		return SavedGraph{}, err
	}
	return graph, nil
}

func (s *Service) GetResourceGraph(ctx context.Context, resourceID string, opts QueryOptions) (SavedGraph, error) {
	graph, err := s.repo.LatestResourceGraph(ctx, resourceID)
	if err != nil {
		return SavedGraph{}, err
	}
	if graph == nil {
		return SavedGraph{}, nil
	}
	if opts.MaxLevel <= 0 && opts.IncludeExpansion {
		return *graph, nil
	}

	allowed := map[string]struct{}{}
	filteredNodes := make([]Node, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		if opts.MaxLevel > 0 && node.Level > opts.MaxLevel {
			continue
		}
		if !opts.IncludeExpansion && node.IsExpansion {
			continue
		}
		filteredNodes = append(filteredNodes, node)
		allowed[node.ID] = struct{}{}
	}
	filteredEdges := make([]Edge, 0, len(graph.Edges))
	for _, edge := range graph.Edges {
		if !opts.IncludeExpansion && edge.IsExpansion {
			continue
		}
		if _, ok := allowed[edge.SourceID]; !ok {
			continue
		}
		if _, ok := allowed[edge.TargetID]; !ok {
			continue
		}
		filteredEdges = append(filteredEdges, edge)
	}

	out := *graph
	out.Nodes = filteredNodes
	out.Edges = filteredEdges
	return out, nil
}

func (s *Service) GetFrameworkGraph(ctx context.Context, groupID string, opts QueryOptions) (SavedGraph, error) {
	graph, err := s.repo.LatestFrameworkGraph(ctx, groupID)
	if err != nil {
		return SavedGraph{}, err
	}
	if graph == nil {
		return SavedGraph{}, nil
	}
	return s.filterGraph(*graph, opts), nil
}

func (s *Service) ListResourceGraphsByGroup(ctx context.Context, groupID string) ([]SavedGraph, error) {
	return s.repo.ListResourceGraphsByGroup(ctx, groupID)
}

func (s *Service) GetGraph(_ context.Context, graphID string) (SavedGraph, error) {
	graph, ok := s.getGraphByID(graphID)
	if !ok {
		return SavedGraph{}, errors.New("graph not found")
	}
	return graph, nil
}

func (s *Service) GetNodeNeighbors(_ context.Context, graphID, nodeID string) ([]Neighbor, error) {
	graph, ok := s.getGraphByID(graphID)
	if !ok {
		return nil, errors.New("graph not found")
	}
	nodeIndex := map[string]Node{}
	found := false
	for _, node := range graph.Nodes {
		nodeIndex[node.ID] = node
		if node.ID == nodeID {
			found = true
		}
	}
	if !found {
		return nil, errors.New("node not found")
	}
	neighbors := make([]Neighbor, 0)
	for _, edge := range graph.Edges {
		if edge.SourceID == nodeID {
			if neighbor, ok := nodeIndex[edge.TargetID]; ok {
				neighbors = append(neighbors, Neighbor{Node: neighbor, Relation: edge.Relation})
			}
		} else if edge.TargetID == nodeID {
			if neighbor, ok := nodeIndex[edge.SourceID]; ok {
				neighbors = append(neighbors, Neighbor{Node: neighbor, Relation: edge.Relation})
			}
		}
	}
	return neighbors, nil
}

func (s *Service) GetNodeDetail(ctx context.Context, graphID, nodeID string) (NodeDetail, error) {
	graph, ok := s.getGraphByID(graphID)
	if !ok {
		return NodeDetail{}, errors.New("graph not found")
	}
	var current Node
	found := false
	for _, node := range graph.Nodes {
		if node.ID == nodeID {
			current = node
			found = true
			break
		}
	}
	if !found {
		return NodeDetail{}, errors.New("node not found")
	}
	neighbors, err := s.GetNodeNeighbors(ctx, graphID, nodeID)
	if err != nil {
		return NodeDetail{}, err
	}
	examples, err := s.repo.ListExamples(ctx, graphID, nodeID)
	if err != nil {
		return NodeDetail{}, err
	}
	return NodeDetail{
		Node:      current,
		Neighbors: neighbors,
		Examples:  examples,
	}, nil
}

func (s *Service) Repository() *InMemoryRepository {
	if repo, ok := s.repo.(*InMemoryRepository); ok {
		return repo
	}
	return nil
}

func (s *Service) filterGraph(graph SavedGraph, opts QueryOptions) SavedGraph {
	if opts.MaxLevel <= 0 && opts.IncludeExpansion {
		return graph
	}
	allowed := map[string]struct{}{}
	filteredNodes := make([]Node, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		if opts.MaxLevel > 0 && node.Level > opts.MaxLevel {
			continue
		}
		if !opts.IncludeExpansion && node.IsExpansion {
			continue
		}
		filteredNodes = append(filteredNodes, node)
		allowed[node.ID] = struct{}{}
	}
	filteredEdges := make([]Edge, 0, len(graph.Edges))
	for _, edge := range graph.Edges {
		if !opts.IncludeExpansion && edge.IsExpansion {
			continue
		}
		if _, ok := allowed[edge.SourceID]; !ok {
			continue
		}
		if _, ok := allowed[edge.TargetID]; !ok {
			continue
		}
		filteredEdges = append(filteredEdges, edge)
	}
	graph.Nodes = filteredNodes
	graph.Edges = filteredEdges
	return graph
}

type InMemoryRepository struct {
	mu       sync.RWMutex
	graphs   map[string]SavedGraph
	examples map[string][]Example
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		graphs:   map[string]SavedGraph{},
		examples: map[string][]Example{},
	}
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

func (r *InMemoryRepository) LatestFrameworkGraph(_ context.Context, groupID string) (*SavedGraph, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var matches []SavedGraph
	for _, graph := range r.graphs {
		if graph.Graph.GroupID == groupID && graph.Graph.ResourceID == "" {
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

func (r *InMemoryRepository) ListResourceGraphsByGroup(_ context.Context, groupID string) ([]SavedGraph, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]SavedGraph, 0)
	for _, graph := range r.graphs {
		if graph.Graph.GroupID == groupID && graph.Graph.ResourceID != "" && graph.Graph.IsActive {
			items = append(items, graph)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Graph.Version > items[j].Graph.Version })
	return items, nil
}

func (r *InMemoryRepository) ListExamples(_ context.Context, graphID string, nodeID string) ([]Example, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := graphID + ":" + nodeID
	items := append([]Example(nil), r.examples[key]...)
	return items, nil
}

func (r *InMemoryRepository) GetByID(id string) (SavedGraph, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	graph, ok := r.graphs[id]
	return graph, ok
}

func (r *InMemoryRepository) AddExample(graphID string, example Example) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := graphID + ":" + example.NodeID
	r.examples[key] = append(r.examples[key], example)
}

func (s *Service) getGraphByID(graphID string) (SavedGraph, bool) {
	if repo, ok := s.repo.(*InMemoryRepository); ok {
		return repo.GetByID(graphID)
	}
	return SavedGraph{}, false
}

func newID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(raw[:])
}
