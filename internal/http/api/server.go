package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	appresource "github.com/tylor/goaipj/internal/app/resource"
	"github.com/tylor/goaipj/internal/infra/apperror"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

type Dependencies struct {
	ExpansionGenerator appgraphExpansionGenerator
	GroupService       *appgroup.Service
	GraphService       *appgraph.Service
	ResourceService    *appresource.Service
	Logger             *log.Logger
}

type Server struct {
	expansionGenerator appgraphExpansionGenerator
	graphService       *appgraph.Service
	groupService       *appgroup.Service
	resourceService    *appresource.Service
	logger             *log.Logger
}

type groupResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type errorResponse struct {
	TraceID string `json:"trace_id"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type resourceResponse struct {
	ID           string `json:"id"`
	GroupID      string `json:"group_id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	SourceURL    string `json:"source_url,omitempty"`
	FailedStage  string `json:"failed_stage,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type jobResponse struct {
	ID         string `json:"id"`
	ResourceID string `json:"resource_id"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

type resourceWithJobResponse struct {
	Resource resourceResponse `json:"resource"`
	Job      jobResponse      `json:"job"`
}

type graphInfoResponse struct {
	ID         string `json:"id"`
	GroupID    string `json:"group_id"`
	ResourceID string `json:"resource_id"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
	Version    int    `json:"version"`
	IsActive   bool   `json:"is_active"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type graphNodeResponse struct {
	ID          string `json:"id"`
	GraphID     string `json:"graph_id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Meaning     string `json:"meaning,omitempty"`
	Level       int    `json:"level"`
	IsExpansion bool   `json:"is_expansion"`
}

type graphEdgeResponse struct {
	ID          string `json:"id"`
	GraphID     string `json:"graph_id"`
	SourceID    string `json:"source_id"`
	TargetID    string `json:"target_id"`
	Relation    string `json:"relation"`
	IsExpansion bool   `json:"is_expansion"`
}

type graphResponse struct {
	Graph graphInfoResponse   `json:"graph"`
	Nodes []graphNodeResponse `json:"nodes"`
	Edges []graphEdgeResponse `json:"edges"`
}

type nodeDetailResponse struct {
	Node      graphNodeResponse   `json:"node"`
	Neighbors []nodeNeighborEntry `json:"neighbors"`
	Examples  []nodeExampleEntry  `json:"examples"`
}

type nodeNeighborEntry struct {
	Node     graphNodeResponse `json:"node"`
	Relation string            `json:"relation"`
}

type nodeExampleEntry struct {
	ID      string `json:"id"`
	NodeID  string `json:"node_id"`
	Content string `json:"content"`
}

type expandNodeResponse struct {
	Job jobResponse `json:"job"`
}

type appgraphExpansionGenerator interface {
	Generate(input pipelinegraph.ExpansionInput) (pipelinegraph.Expansion, error)
}

func NewServer(deps Dependencies) http.Handler {
	server := &Server{
		expansionGenerator: deps.ExpansionGenerator,
		graphService:       deps.GraphService,
		groupService:       deps.GroupService,
		resourceService:    deps.ResourceService,
		logger:             deps.Logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", server.handleHealthz)
	mux.HandleFunc("/readyz", server.handleReadyz)
	mux.HandleFunc("/api/v1/groups", server.handleGroups)
	mux.HandleFunc("/api/v1/groups/", server.handleGroupByID)
	mux.HandleFunc("/api/v1/graphs/", server.handleGraphsRoot)
	mux.HandleFunc("/api/v1/resources/", server.handleResourcesRoot)
	return traceMiddleware(mux)
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) handleGroups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		groups, err := s.groupService.ListGroups(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}

		resp := make([]groupResponse, 0, len(groups))
		for _, item := range groups {
			resp = append(resp, toGroupResponse(item))
		}
		writeJSON(w, http.StatusOK, resp)
	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "invalid json body"))
			return
		}

		group, err := s.groupService.CreateGroup(r.Context(), appgroup.CreateGroupInput{Name: req.Name})
		if err != nil {
			writeError(w, r, err)
			return
		}
		writeJSON(w, http.StatusCreated, toGroupResponse(group))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGroupByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/groups/")
	parts := strings.Split(strings.Trim(id, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, r, apperror.New(apperror.CodeNotFound, "group not found"))
		return
	}
	groupID := parts[0]
	if len(parts) > 1 {
		s.handleGroupSubresource(w, r, groupID, parts[1:])
		return
	}

	switch r.Method {
	case http.MethodGet:
		group, err := s.groupService.GetGroup(r.Context(), groupID)
		if err != nil {
			writeError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, toGroupResponse(group))
	case http.MethodPatch:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "invalid json body"))
			return
		}

		group, err := s.groupService.UpdateGroup(r.Context(), groupID, appgroup.UpdateGroupInput{Name: req.Name})
		if err != nil {
			writeError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, toGroupResponse(group))
	case http.MethodDelete:
		if err := s.groupService.DeleteGroup(r.Context(), groupID); err != nil {
			writeError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGroupSubresource(w http.ResponseWriter, r *http.Request, groupID string, parts []string) {
	switch {
	case len(parts) == 2 && parts[0] == "resources" && parts[1] == "upload" && r.Method == http.MethodPost:
		if s.resourceService == nil {
			writeError(w, r, apperror.New(apperror.CodeInternal, "resource service not configured"))
			return
		}
		s.handleUploadResource(w, r, groupID)
	case len(parts) == 1 && parts[0] == "resources" && r.Method == http.MethodGet:
		if s.resourceService == nil {
			writeError(w, r, apperror.New(apperror.CodeInternal, "resource service not configured"))
			return
		}
		s.handleListResources(w, r, groupID)
	case len(parts) == 2 && parts[0] == "resources" && r.Method == http.MethodGet:
		if s.resourceService == nil {
			writeError(w, r, apperror.New(apperror.CodeInternal, "resource service not configured"))
			return
		}
		s.handleGetResource(w, r, groupID, parts[1])
	case len(parts) == 1 && parts[0] == "web-resources" && r.Method == http.MethodPost:
		if s.resourceService == nil {
			writeError(w, r, apperror.New(apperror.CodeInternal, "resource service not configured"))
			return
		}
		s.handleCreateWebResource(w, r, groupID)
	case len(parts) == 1 && parts[0] == "framework-graph" && r.Method == http.MethodGet:
		s.handleGetFrameworkGraph(w, r, groupID)
	default:
		writeError(w, r, apperror.New(apperror.CodeNotFound, "route not found"))
	}
}

func (s *Server) handleGraphsRoot(w http.ResponseWriter, r *http.Request) {
	if s.graphService == nil {
		writeError(w, r, apperror.New(apperror.CodeInternal, "graph service not configured"))
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/graphs/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 3 && parts[1] == "nodes" && r.Method == http.MethodGet {
		s.handleGetNodeDetail(w, r, parts[0], parts[2])
		return
	}
	if len(parts) == 4 && parts[1] == "nodes" && parts[3] == "expand" && r.Method == http.MethodPost {
		s.handleExpandNode(w, r, parts[0], parts[2])
		return
	}
	writeError(w, r, apperror.New(apperror.CodeNotFound, "route not found"))
}

func (s *Server) handleResourcesRoot(w http.ResponseWriter, r *http.Request) {
	if s.resourceService == nil {
		writeError(w, r, apperror.New(apperror.CodeInternal, "resource service not configured"))
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/resources/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 2 && parts[1] == "retry" && r.Method == http.MethodPost {
		result, err := s.resourceService.RetryResource(r.Context(), parts[0])
		if err != nil {
			writeError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, toResourceWithJobResponse(result))
		return
	}
	if len(parts) == 2 && parts[1] == "graph" && r.Method == http.MethodGet {
		s.handleGetResourceGraph(w, r, parts[0])
		return
	}
	writeError(w, r, apperror.New(apperror.CodeNotFound, "route not found"))
}

func (s *Server) handleUploadResource(w http.ResponseWriter, r *http.Request, groupID string) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "invalid multipart form"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "file is required"))
		return
	}
	defer file.Close()

	buffered, err := io.ReadAll(file)
	if err != nil {
		writeError(w, r, apperror.Wrap(apperror.CodeInternal, "read upload", err))
		return
	}

	result, err := s.resourceService.UploadFile(r.Context(), appresource.UploadFileInput{
		GroupID:     groupID,
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		Body:        bytes.NewReader(buffered),
		Size:        int64(len(buffered)),
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toResourceWithJobResponse(result))
}

func (s *Server) handleCreateWebResource(w http.ResponseWriter, r *http.Request, groupID string) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "invalid json body"))
		return
	}
	result, err := s.resourceService.CreateWebResource(r.Context(), appresource.CreateWebResourceInput{
		GroupID: groupID,
		URL:     req.URL,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toResourceWithJobResponse(result))
}

func (s *Server) handleListResources(w http.ResponseWriter, r *http.Request, groupID string) {
	resources, err := s.resourceService.ListResources(r.Context(), groupID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	resp := make([]resourceResponse, 0, len(resources))
	for _, item := range resources {
		resp = append(resp, toResourceResponse(item))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetResource(w http.ResponseWriter, r *http.Request, groupID, resourceID string) {
	item, err := s.resourceService.GetResource(r.Context(), groupID, resourceID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toResourceResponse(item))
}

func (s *Server) handleGetResourceGraph(w http.ResponseWriter, r *http.Request, resourceID string) {
	if s.graphService == nil {
		writeError(w, r, apperror.New(apperror.CodeInternal, "graph service not configured"))
		return
	}
	maxLevel := 0
	if raw := r.URL.Query().Get("max_level"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "invalid max_level"))
			return
		}
		maxLevel = parsed
	}
	includeExpansion := r.URL.Query().Get("include_expansion") == "true"
	graph, err := s.graphService.GetResourceGraph(r.Context(), resourceID, appgraph.QueryOptions{
		MaxLevel:         maxLevel,
		IncludeExpansion: includeExpansion,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	if graph.Graph.ID == "" {
		writeError(w, r, apperror.New(apperror.CodeNotFound, "graph not found"))
		return
	}
	writeJSON(w, http.StatusOK, toGraphResponse(graph))
}

func (s *Server) handleGetFrameworkGraph(w http.ResponseWriter, r *http.Request, groupID string) {
	if s.graphService == nil {
		writeError(w, r, apperror.New(apperror.CodeInternal, "graph service not configured"))
		return
	}
	graph, err := s.graphService.GetFrameworkGraph(r.Context(), groupID, appgraph.QueryOptions{
		IncludeExpansion: true,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	if graph.Graph.ID == "" {
		writeError(w, r, apperror.New(apperror.CodeNotFound, "framework graph not found"))
		return
	}
	writeJSON(w, http.StatusOK, toGraphResponse(graph))
}

func (s *Server) handleGetNodeDetail(w http.ResponseWriter, r *http.Request, graphID, nodeID string) {
	detail, err := s.graphService.GetNodeDetail(r.Context(), graphID, nodeID)
	if err != nil {
		writeError(w, r, apperror.New(apperror.CodeNotFound, err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, toNodeDetailResponse(detail))
}

func (s *Server) handleExpandNode(w http.ResponseWriter, r *http.Request, graphID, nodeID string) {
	if s.graphService == nil || s.expansionGenerator == nil {
		writeError(w, r, apperror.New(apperror.CodeInternal, "expansion service not configured"))
		return
	}
	detail, err := s.graphService.GetNodeDetail(r.Context(), graphID, nodeID)
	if err != nil {
		writeError(w, r, apperror.New(apperror.CodeNotFound, err.Error()))
		return
	}
	neighbors := make([]pipelinegraph.Node, 0, len(detail.Neighbors))
	for _, neighbor := range detail.Neighbors {
		neighbors = append(neighbors, pipelinegraph.Node{
			ID:    neighbor.Node.ID,
			Name:  neighbor.Node.Name,
			Type:  neighbor.Node.Type,
			Level: neighbor.Node.Level,
		})
	}
	expansion, err := s.expansionGenerator.Generate(pipelinegraph.ExpansionInput{
		Current: pipelinegraph.Node{
			ID:    detail.Node.ID,
			Name:  detail.Node.Name,
			Type:  detail.Node.Type,
			Level: detail.Node.Level,
		},
		Neighbors: neighbors,
	})
	if err != nil {
		writeError(w, r, apperror.New(apperror.CodeInvalidArgument, err.Error()))
		return
	}
	if _, err := s.graphService.ExpandNode(r.Context(), graphID, nodeID, expansion); err != nil {
		writeError(w, r, apperror.New(apperror.CodeInternal, err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, expandNodeResponse{
		Job: jobResponse{
			ID:        newTraceID(),
			Type:      "expand_node",
			Status:    "completed",
			CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
		},
	})
}

func toGroupResponse(group appgroup.Group) groupResponse {
	return groupResponse{
		ID:        group.ID,
		Name:      group.Name,
		CreatedAt: group.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: group.UpdatedAt.Format(time.RFC3339Nano),
	}
}

func toResourceResponse(item appresource.Resource) resourceResponse {
	return resourceResponse{
		ID:           item.ID,
		GroupID:      item.GroupID,
		Name:         item.Name,
		Type:         string(item.Type),
		Status:       string(item.Status),
		SourceURL:    item.SourceURL,
		FailedStage:  item.FailedStage,
		ErrorMessage: item.ErrorMessage,
		CreatedAt:    item.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:    item.UpdatedAt.Format(time.RFC3339Nano),
	}
}

func toJobResponse(job appresource.Job) jobResponse {
	return jobResponse{
		ID:         job.ID,
		ResourceID: job.ResourceID,
		Type:       job.Type,
		Status:     job.Status,
		CreatedAt:  job.CreatedAt.Format(time.RFC3339Nano),
	}
}

func toResourceWithJobResponse(item appresource.ResourceWithJob) resourceWithJobResponse {
	return resourceWithJobResponse{
		Resource: toResourceResponse(item.Resource),
		Job:      toJobResponse(item.Job),
	}
}

func toGraphResponse(item appgraph.SavedGraph) graphResponse {
	nodes := make([]graphNodeResponse, 0, len(item.Nodes))
	for _, node := range item.Nodes {
		nodes = append(nodes, graphNodeResponse{
			ID:          node.ID,
			GraphID:     node.GraphID,
			Name:        node.Name,
			Type:        node.Type,
			Description: node.Description,
			Meaning:     node.Meaning,
			Level:       node.Level,
			IsExpansion: node.IsExpansion,
		})
	}
	edges := make([]graphEdgeResponse, 0, len(item.Edges))
	for _, edge := range item.Edges {
		edges = append(edges, graphEdgeResponse{
			ID:          edge.ID,
			GraphID:     edge.GraphID,
			SourceID:    edge.SourceID,
			TargetID:    edge.TargetID,
			Relation:    edge.Relation,
			IsExpansion: edge.IsExpansion,
		})
	}
	return graphResponse{
		Graph: graphInfoResponse{
			ID:         item.Graph.ID,
			GroupID:    item.Graph.GroupID,
			ResourceID: item.Graph.ResourceID,
			Title:      item.Graph.Title,
			Summary:    item.Graph.Summary,
			Version:    item.Graph.Version,
			IsActive:   item.Graph.IsActive,
			CreatedAt:  item.Graph.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:  item.Graph.UpdatedAt.Format(time.RFC3339Nano),
		},
		Nodes: nodes,
		Edges: edges,
	}
}

func toNodeDetailResponse(detail appgraph.NodeDetail) nodeDetailResponse {
	resp := nodeDetailResponse{
		Node: toGraphNodeResponse(detail.Node),
	}
	for _, neighbor := range detail.Neighbors {
		resp.Neighbors = append(resp.Neighbors, nodeNeighborEntry{
			Node:     toGraphNodeResponse(neighbor.Node),
			Relation: neighbor.Relation,
		})
	}
	for _, example := range detail.Examples {
		resp.Examples = append(resp.Examples, nodeExampleEntry{
			ID:      example.ID,
			NodeID:  example.NodeID,
			Content: example.Content,
		})
	}
	return resp
}

func toGraphNodeResponse(node appgraph.Node) graphNodeResponse {
	return graphNodeResponse{
		ID:          node.ID,
		GraphID:     node.GraphID,
		Name:        node.Name,
		Type:        node.Type,
		Description: node.Description,
		Meaning:     node.Meaning,
		Level:       node.Level,
		IsExpansion: node.IsExpansion,
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	traceID, _ := TraceIDFromContext(r.Context())
	resp := errorResponse{TraceID: traceID}
	resp.Error.Code = apperror.CodeOf(err)
	resp.Error.Message = err.Error()

	writeJSON(w, apperror.StatusCode(err), resp)
}

type contextKey string

const traceIDKey contextKey = "trace_id"

func traceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = newTraceID()
		}
		ctx := context.WithValue(r.Context(), traceIDKey, traceID)
		w.Header().Set("X-Trace-ID", traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func TraceIDFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(traceIDKey).(string)
	return value, ok
}

func newTraceID() string {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "trace-fallback"
	}
	return hex.EncodeToString(raw[:])
}
