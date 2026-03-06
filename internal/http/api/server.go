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

	appchat "github.com/tylor/goaipj/internal/app/chat"
	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	appjob "github.com/tylor/goaipj/internal/app/job"
	appresource "github.com/tylor/goaipj/internal/app/resource"
	domainresource "github.com/tylor/goaipj/internal/domain/resource"
	"github.com/tylor/goaipj/internal/infra/apperror"
	infraevent "github.com/tylor/goaipj/internal/infra/event"
	"github.com/tylor/goaipj/internal/infra/monitoring"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

type DependencyCheck struct {
	Name  string
	Check func(ctx context.Context) error
}

type Dependencies struct {
	ExpansionGenerator appgraphExpansionGenerator
	ChatService        *appchat.Service
	EventBroker        infraevent.Broker
	FrameworkGenerator frameworkGraphGenerator
	GroupService       *appgroup.Service
	GraphService       *appgraph.Service
	JobService         *appjob.Service
	Metrics            *monitoring.Metrics
	Readiness          []DependencyCheck
	ResourceService    *appresource.Service
	Logger             *log.Logger
}

type Server struct {
	expansionGenerator appgraphExpansionGenerator
	chatService        *appchat.Service
	eventBroker        infraevent.Broker
	frameworkGenerator frameworkGraphGenerator
	graphService       *appgraph.Service
	groupService       *appgroup.Service
	jobService         *appjob.Service
	metrics            *monitoring.Metrics
	readiness          []DependencyCheck
	resourceService    *appresource.Service
	logger             *log.Logger
}

type groupResponse struct {
	Description            *string `json:"description"`
	ID                     string  `json:"id"`
	Name                   string  `json:"name"`
	ResourceCount          int     `json:"resource_count"`
	CompletedResourceCount int     `json:"completed_resource_count"`
	CreatedAt              string  `json:"created_at"`
	UpdatedAt              string  `json:"updated_at"`
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type envelopeResponse struct {
	Data  any            `json:"data"`
	Error any            `json:"error"`
	Meta  map[string]any `json:"meta"`
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

type conversationSummaryResponse struct {
	ID            string                        `json:"id"`
	GroupID       string                        `json:"group_id"`
	GraphID       string                        `json:"graph_id"`
	CurrentNodeID string                        `json:"current_node_id"`
	Title         string                        `json:"title"`
	CreatedAt     string                        `json:"created_at"`
	UpdatedAt     string                        `json:"updated_at"`
}

type conversationResponse struct {
	ID            string                        `json:"id"`
	GroupID       string                        `json:"group_id"`
	GraphID       string                        `json:"graph_id"`
	CurrentNodeID string                        `json:"current_node_id"`
	Title         string                        `json:"title"`
	CreatedAt     string                        `json:"created_at"`
	UpdatedAt     string                        `json:"updated_at"`
	Messages      []conversationMessageResponse `json:"messages"`
}

type conversationDetailResponse struct {
	Conversation conversationSummaryResponse    `json:"conversation"`
	Messages      []conversationMessageResponse `json:"messages"`
}

type conversationMessageResponse struct {
	ID              string                  `json:"id"`
	ConversationID  string                  `json:"conversation_id"`
	Role            string                  `json:"role"`
	Content         string                  `json:"content"`
	Examples        []string                `json:"examples,omitempty"`
	CitedChunkIDs   []string                `json:"cited_chunk_ids"`
	CitedNodeIDs    []string                `json:"cited_node_ids"`
	ContextSnapshot conversationContextData `json:"context_snapshot"`
	CreatedAt       string                  `json:"created_at"`
}

type conversationContextData struct {
	GraphID           string   `json:"graph_id"`
	CurrentNodeID     string   `json:"current_node_id"`
	ResourceSummary   string   `json:"resource_summary"`
	NeighborNodeIDs   []string `json:"neighbor_node_ids"`
	RetrievedChunkIDs []string `json:"retrieved_chunk_ids"`
	RecentMessageIDs  []string `json:"recent_message_ids"`
}

type conversationAskResponse struct {
	ConversationID   string                      `json:"conversation_id"`
	UserMessage      conversationMessageResponse `json:"user_message"`
	AssistantMessage conversationMessageResponse `json:"assistant_message"`
}

type appgraphExpansionGenerator interface {
	Generate(input pipelinegraph.ExpansionInput) (pipelinegraph.Expansion, error)
}

type frameworkGraphGenerator interface {
	Generate(documents []pipelinegraph.Document) (pipelinegraph.Document, error)
}

func NewServer(deps Dependencies) http.Handler {
	eventBroker := deps.EventBroker
	if eventBroker == nil {
		eventBroker = infraevent.NewInMemoryBroker()
	}
	server := &Server{
		expansionGenerator: deps.ExpansionGenerator,
		chatService:        deps.ChatService,
		eventBroker:        eventBroker,
		frameworkGenerator: deps.FrameworkGenerator,
		graphService:       deps.GraphService,
		groupService:       deps.GroupService,
		jobService:         deps.JobService,
		metrics:            deps.Metrics,
		readiness:          deps.Readiness,
		resourceService:    deps.ResourceService,
		logger:             deps.Logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", server.handleHealthz)
	mux.HandleFunc("/readyz", server.handleReadyz)
	mux.HandleFunc("/metrics", server.handleMetrics)
	mux.HandleFunc("/api/v1/groups", server.handleGroups)
	mux.HandleFunc("/api/v1/groups/", server.handleGroupByID)
	mux.HandleFunc("/api/v1/conversations", server.handleConversations)
	mux.HandleFunc("/api/v1/conversations/", server.handleConversationByID)
	mux.HandleFunc("/api/v1/events/groups/", server.handleGroupEvents)
	mux.HandleFunc("/api/v1/graphs/", server.handleGraphsRoot)
	mux.HandleFunc("/api/v1/jobs/", server.handleJobsRoot)
	mux.HandleFunc("/api/v1/resources/", server.handleResourcesRoot)
	return traceMiddleware(metricsMiddleware(server.metrics, mux))
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"checks": s.checkStatuses(context.Background()),
	})
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	checks := s.checkStatuses(ctx)
	status := "ready"
	code := http.StatusOK
	for _, value := range checks {
		if value != "up" {
			status = "not_ready"
			code = http.StatusServiceUnavailable
			break
		}
	}
	writeJSON(w, code, map[string]any{
		"status": status,
		"checks": checks,
	})
}

func (s *Server) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	if s.metrics == nil {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "")
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, s.metrics.RenderPrometheus())
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
			resp = append(resp, s.toGroupResponse(r.Context(), item))
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
		writeJSON(w, http.StatusCreated, s.toGroupResponse(r.Context(), group))
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
		writeJSON(w, http.StatusOK, s.toGroupResponse(r.Context(), group))
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
		writeJSON(w, http.StatusOK, s.toGroupResponse(r.Context(), group))
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
	case len(parts) == 2 && parts[0] == "framework-graph" && parts[1] == "generate" && r.Method == http.MethodPost:
		s.handleGenerateFrameworkGraph(w, r, groupID)
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

func (s *Server) handleConversations(w http.ResponseWriter, r *http.Request) {
	if s.chatService == nil {
		writeError(w, r, apperror.New(apperror.CodeInternal, "chat service not configured"))
		return
	}
	switch r.Method {
	case http.MethodPost:
		var req struct {
			GroupID       string `json:"group_id"`
			GraphID       string `json:"graph_id"`
			CurrentNodeID string `json:"current_node_id"`
			Title         string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "invalid json body"))
			return
		}
		conversation, err := s.chatService.CreateConversation(r.Context(), appchat.CreateConversationInput{
			GroupID:       req.GroupID,
			GraphID:       req.GraphID,
			CurrentNodeID: req.CurrentNodeID,
			Title:         req.Title,
		})
		if err != nil {
			writeError(w, r, err)
			return
		}
		writeJSON(w, http.StatusCreated, toConversationResponse(conversation))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGroupEvents(w http.ResponseWriter, r *http.Request) {
	if s.eventBroker == nil {
		writeError(w, r, apperror.New(apperror.CodeInternal, "event broker not configured"))
		return
	}
	groupID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/events/groups/"), "/")
	if groupID == "" {
		writeError(w, r, apperror.New(apperror.CodeNotFound, "group not found"))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	flusher, _ := w.(http.Flusher)
	sub := s.eventBroker.Subscribe(r.Context(), groupID)
	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-sub:
			if !ok {
				return
			}
			writeSSEJSON(w, event.Name, event)
			if flusher != nil {
				flusher.Flush()
			}
		}
	}
}

func (s *Server) handleConversationByID(w http.ResponseWriter, r *http.Request) {
	if s.chatService == nil {
		writeError(w, r, apperror.New(apperror.CodeInternal, "chat service not configured"))
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/conversations/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && r.Method == http.MethodGet {
		conversation, err := s.chatService.GetConversation(r.Context(), parts[0])
		if err != nil {
			writeError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, toConversationDetailResponse(conversation))
		return
	}
	if len(parts) == 2 && parts[1] == "messages" && r.Method == http.MethodPost {
		var req struct {
			Content string `json:"content"`
			Stream  bool   `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "invalid json body"))
			return
		}
		result, err := s.chatService.Ask(r.Context(), appchat.AskInput{
			ConversationID: parts[0],
			Content:        req.Content,
		})
		s.observeTask("chat_answer", err == nil, startedAt(r))
		if err != nil {
			writeError(w, r, err)
			return
		}
		s.publishGroupEvent(r.Context(), infraevent.Envelope{
			Name:    infraevent.NameConversationMessageCreated,
			GroupID: result.Conversation.GroupID,
			Payload: map[string]any{
				"conversation_id": result.Conversation.ID,
				"message_id":      result.AssistantMessage.ID,
				"node_id":         result.Conversation.CurrentNodeID,
			},
		})
		if req.Stream {
			s.writeConversationStream(w, result)
			return
		}
		writeJSON(w, http.StatusOK, conversationAskResponse{
			ConversationID:   result.Conversation.ID,
			UserMessage:      toConversationMessageResponse(result.UserMessage),
			AssistantMessage: toConversationMessageResponse(result.AssistantMessage),
		})
		return
	}
	writeError(w, r, apperror.New(apperror.CodeNotFound, "route not found"))
}

func (s *Server) handleJobsRoot(w http.ResponseWriter, r *http.Request) {
	if s.jobService == nil {
		writeError(w, r, apperror.New(apperror.CodeInternal, "job service not configured"))
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	jobID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/jobs/"), "/")
	if jobID == "" {
		writeError(w, r, apperror.New(apperror.CodeNotFound, "job not found"))
		return
	}
	job, err := s.jobService.GetJob(r.Context(), jobID)
	if err != nil {
		writeError(w, r, apperror.New(apperror.CodeNotFound, "job not found"))
		return
	}
	writeJSON(w, http.StatusOK, toJobStatusResponse(job))
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
		s.publishGroupEvent(r.Context(), infraevent.Envelope{
			Name:    infraevent.NameResourceStatusChanged,
			GroupID: result.Resource.GroupID,
			Payload: map[string]any{
				"resource_id": result.Resource.ID,
				"status":      result.Resource.Status,
				"job_id":      result.Job.ID,
			},
		})
		writeJSON(w, http.StatusOK, toResourceWithJobResponse(result))
		return
	}
	if len(parts) == 1 && r.Method == http.MethodDelete {
		s.handleDeleteResource(w, r, parts[0])
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
	s.publishGroupEvent(r.Context(), infraevent.Envelope{
		Name:    infraevent.NameResourceStatusChanged,
		GroupID: result.Resource.GroupID,
		Payload: map[string]any{
			"resource_id": result.Resource.ID,
			"status":      result.Resource.Status,
			"job_id":      result.Job.ID,
		},
	})
	writeJSON(w, http.StatusCreated, toResourceWithJobResponse(result))
}

func (s *Server) handleCreateWebResource(w http.ResponseWriter, r *http.Request, groupID string) {
		var req struct {
			URL  string `json:"url"`
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, apperror.New(apperror.CodeInvalidArgument, "invalid json body"))
		return
	}
		result, err := s.resourceService.CreateWebResource(r.Context(), appresource.CreateWebResourceInput{
			GroupID: groupID,
			URL:     req.URL,
			Name:    req.Name,
		})
	if err != nil {
		writeError(w, r, err)
		return
	}
	s.publishGroupEvent(r.Context(), infraevent.Envelope{
		Name:    infraevent.NameResourceStatusChanged,
		GroupID: result.Resource.GroupID,
		Payload: map[string]any{
			"resource_id": result.Resource.ID,
			"status":      result.Resource.Status,
			"job_id":      result.Job.ID,
		},
	})
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

func (s *Server) handleDeleteResource(w http.ResponseWriter, r *http.Request, resourceID string) {
	item, err := s.resourceService.GetResourceByID(r.Context(), resourceID)
	if err != nil {
		writeError(w, r, apperror.New(apperror.CodeNotFound, "resource not found"))
		return
	}
	if err := s.resourceService.DeleteResource(r.Context(), resourceID); err != nil {
		writeError(w, r, err)
		return
	}
	if s.graphService != nil {
		if err := s.graphService.DeleteResourceGraphs(r.Context(), resourceID); err != nil {
			writeError(w, r, apperror.Wrap(apperror.CodeInternal, "delete resource graphs", err))
			return
		}
	}
	s.publishGroupEvent(r.Context(), infraevent.Envelope{
		Name:    "resource.deleted",
		GroupID: item.GroupID,
		Payload: map[string]any{
			"resource_id": resourceID,
		},
	})
	w.WriteHeader(http.StatusNoContent)
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

func (s *Server) handleGenerateFrameworkGraph(w http.ResponseWriter, r *http.Request, groupID string) {
	if s.groupService == nil {
		writeError(w, r, apperror.New(apperror.CodeInternal, "group service not configured"))
		return
	}
	if s.graphService == nil {
		writeError(w, r, apperror.New(apperror.CodeInternal, "graph service not configured"))
		return
	}
	if s.frameworkGenerator == nil {
		writeError(w, r, apperror.New(apperror.CodeInternal, "framework generator not configured"))
		return
	}
	if _, err := s.groupService.GetGroup(r.Context(), groupID); err != nil {
		writeError(w, r, err)
		return
	}

	const jobType = "generate_framework_graph"
	dedupeKey := groupID + ":" + jobType
	if s.jobService != nil {
		if existing, ok, err := s.jobService.FindActiveJob(r.Context(), dedupeKey); err == nil && ok {
			writeJSON(w, http.StatusAccepted, map[string]any{
				"group_id": groupID,
				"job_id":   existing.ID,
				"status":   string(existing.Status),
			})
			return
		}
	}

	var jobResult appjob.Result
	if s.jobService != nil {
		created, err := s.jobService.CreateQueuedJob(r.Context(), appjob.CreateQueuedJobInput{
			GroupID:       groupID,
			JobType:       jobType,
			QueueName:     "graph",
			MaxAttempts:   3,
			Deduplication: dedupeKey,
			Payload: map[string]any{
				"group_id": groupID,
			},
		})
		if err != nil {
			writeError(w, r, apperror.Wrap(apperror.CodeInternal, "create framework graph job", err))
			return
		}
		jobResult = created
		if _, err := s.jobService.MarkRunning(r.Context(), created.Job.ID); err != nil {
			writeError(w, r, apperror.Wrap(apperror.CodeInternal, "mark framework graph job running", err))
			return
		}
	}

	resourceGraphs, err := s.graphService.ListResourceGraphsByGroup(r.Context(), groupID)
	if err != nil {
		s.failFrameworkJob(r.Context(), jobResult.Job.ID, err)
		writeError(w, r, apperror.Wrap(apperror.CodeInternal, "list resource graphs", err))
		return
	}
	documents := make([]pipelinegraph.Document, 0, len(resourceGraphs))
	for _, graph := range resourceGraphs {
		documents = append(documents, toFrameworkSourceDocument(graph))
	}
	document, err := s.frameworkGenerator.Generate(documents)
	if err != nil {
		s.failFrameworkJob(r.Context(), jobResult.Job.ID, err)
		writeError(w, r, apperror.Wrap(apperror.CodeInternal, "generate framework graph", err))
		return
	}
	saved, err := s.graphService.SaveFrameworkGraph(r.Context(), appgraph.SaveFrameworkInput{
		GroupID:  groupID,
		Title:    "Framework",
		Document: document,
	})
	if err != nil {
		s.failFrameworkJob(r.Context(), jobResult.Job.ID, err)
		writeError(w, r, apperror.Wrap(apperror.CodeInternal, "save framework graph", err))
		return
	}
	if s.jobService != nil {
		if _, err := s.jobService.MarkDone(r.Context(), jobResult.Job.ID); err != nil {
			writeError(w, r, apperror.Wrap(apperror.CodeInternal, "mark framework graph job done", err))
			return
		}
	}
	s.publishGroupEvent(r.Context(), infraevent.Envelope{
		Name:    "framework_graph.updated",
		GroupID: groupID,
		Payload: map[string]any{
			"group_id": groupID,
			"graph_id": saved.Graph.ID,
		},
	})
	status := "completed"
	if s.jobService != nil {
		status = string(appjob.StatusDone)
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"group_id": groupID,
		"job_id":   jobResult.Job.ID,
		"status":   status,
	})
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
		s.observeTask("graph_expand", false, startedAt(r))
		writeError(w, r, apperror.New(apperror.CodeInvalidArgument, err.Error()))
		return
	}
	if _, err := s.graphService.ExpandNode(r.Context(), graphID, nodeID, expansion); err != nil {
		s.observeTask("graph_expand", false, startedAt(r))
		writeError(w, r, apperror.New(apperror.CodeInternal, err.Error()))
		return
	}
	s.observeTask("graph_expand", true, startedAt(r))
	if graph, err := s.graphService.GetGraph(r.Context(), graphID); err == nil {
		s.publishGroupEvent(r.Context(), infraevent.Envelope{
			Name:    infraevent.NameGraphNodeExpanded,
			GroupID: graph.Graph.GroupID,
			Payload: map[string]any{
				"graph_id":       graphID,
				"source_node_id": nodeID,
				"new_node_id":    expansion.Node.ID,
				"edge_id":        expansion.Edge.ID,
			},
		})
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

func (s *Server) toGroupResponse(ctx context.Context, group appgroup.Group) groupResponse {
	resp := groupResponse{
		Description:            nil,
		ID:                     group.ID,
		Name:                   group.Name,
		CreatedAt:              group.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:              group.UpdatedAt.Format(time.RFC3339Nano),
	}
	if s.resourceService == nil {
		return resp
	}
	items, err := s.resourceService.ListResources(ctx, group.ID)
	if err != nil {
		return resp
	}
	resp.ResourceCount = len(items)
	for _, item := range items {
		if item.Status == domainresource.StatusCompleted {
			resp.CompletedResourceCount++
		}
	}
	return resp
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

func toJobStatusResponse(job appjob.Job) map[string]any {
	return map[string]any{
		"id":           job.ID,
		"job_type":     job.JobType,
		"status":       string(job.Status),
		"attempt":      job.Attempts,
		"max_attempts": job.MaxAttempts,
		"last_error":   nullableString(job.ErrorMessage),
		"started_at":   formatOptionalTime(job.StartedAt),
		"finished_at":  formatOptionalTime(job.FinishedAt),
		"created_at":   job.CreatedAt.Format(time.RFC3339Nano),
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

func toConversationResponse(conversation appchat.Conversation) conversationResponse {
	resp := conversationResponse{
		ID:            conversation.ID,
		GroupID:       conversation.GroupID,
		GraphID:       conversation.GraphID,
		CurrentNodeID: conversation.CurrentNodeID,
		Title:         conversation.Title,
		CreatedAt:     conversation.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:     conversation.UpdatedAt.Format(time.RFC3339Nano),
		Messages:      make([]conversationMessageResponse, 0, len(conversation.Messages)),
	}
	for _, message := range conversation.Messages {
		resp.Messages = append(resp.Messages, toConversationMessageResponse(message))
	}
	return resp
}

func toConversationDetailResponse(conversation appchat.Conversation) conversationDetailResponse {
	resp := conversationDetailResponse{
		Conversation: conversationSummaryResponse{
			ID:            conversation.ID,
			GroupID:       conversation.GroupID,
			GraphID:       conversation.GraphID,
			CurrentNodeID: conversation.CurrentNodeID,
			Title:         conversation.Title,
			CreatedAt:     conversation.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:     conversation.UpdatedAt.Format(time.RFC3339Nano),
		},
		Messages: make([]conversationMessageResponse, 0, len(conversation.Messages)),
	}
	for _, message := range conversation.Messages {
		resp.Messages = append(resp.Messages, toConversationMessageResponse(message))
	}
	return resp
}

func toFrameworkSourceDocument(graph appgraph.SavedGraph) pipelinegraph.Document {
	document := pipelinegraph.Document{
		Summary: graph.Graph.Summary,
		Nodes:   make([]pipelinegraph.Node, 0, len(graph.Nodes)),
		Edges:   make([]pipelinegraph.Edge, 0, len(graph.Edges)),
	}
	for _, node := range graph.Nodes {
		document.Nodes = append(document.Nodes, pipelinegraph.Node{
			ID:          node.ID,
			Name:        node.Name,
			Type:        node.Type,
			Description: node.Description,
			Meaning:     node.Meaning,
			Level:       node.Level,
		})
	}
	for _, edge := range graph.Edges {
		document.Edges = append(document.Edges, pipelinegraph.Edge{
			ID:       edge.ID,
			SourceID: edge.SourceID,
			TargetID: edge.TargetID,
			Relation: edge.Relation,
		})
	}
	return document
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func formatOptionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format(time.RFC3339Nano)
}

func (s *Server) failFrameworkJob(ctx context.Context, jobID string, err error) {
	if s.jobService == nil || jobID == "" {
		return
	}
	_, _ = s.jobService.MarkFailed(ctx, jobID, err.Error())
}

func (s *Server) writeConversationStream(w http.ResponseWriter, result appchat.AskResult) {
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	flusher, _ := w.(http.Flusher)
	for _, delta := range splitStreamDeltas(result.AssistantMessage.Content, 48) {
		writeSSEJSON(w, "assistant.delta", map[string]any{
			"conversation_id": result.Conversation.ID,
			"delta":           delta,
		})
		if flusher != nil {
			flusher.Flush()
		}
	}
	writeSSEJSON(w, "assistant.message", conversationAskResponse{
		ConversationID:   result.Conversation.ID,
		UserMessage:      toConversationMessageResponse(result.UserMessage),
		AssistantMessage: toConversationMessageResponse(result.AssistantMessage),
	})
	writeSSEJSON(w, "done", map[string]bool{"done": true})
	if flusher != nil {
		flusher.Flush()
	}
}

func toConversationMessageResponse(message appchat.Message) conversationMessageResponse {
	return conversationMessageResponse{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		Role:           message.Role,
		Content:        message.Content,
		Examples:       append([]string(nil), message.Examples...),
		CitedChunkIDs:  append([]string(nil), message.CitedChunkIDs...),
		CitedNodeIDs:   append([]string(nil), message.CitedNodeIDs...),
		ContextSnapshot: conversationContextData{
			GraphID:           message.ContextSnapshot.GraphID,
			CurrentNodeID:     message.ContextSnapshot.CurrentNodeID,
			ResourceSummary:   message.ContextSnapshot.ResourceSummary,
			NeighborNodeIDs:   append([]string(nil), message.ContextSnapshot.NeighborNodeIDs...),
			RetrievedChunkIDs: append([]string(nil), message.ContextSnapshot.RetrievedChunkIDs...),
			RecentMessageIDs:  append([]string(nil), message.ContextSnapshot.RecentMessageIDs...),
		},
		CreatedAt: message.CreatedAt.Format(time.RFC3339Nano),
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelopeResponse{
		Data:  payload,
		Error: nil,
		Meta:  map[string]any{},
	})
}

func writeSSEJSON(w http.ResponseWriter, event string, payload any) {
	_, _ = io.WriteString(w, "event: "+event+"\n")
	raw, err := json.Marshal(payload)
	if err != nil {
		raw = []byte(`{"error":"marshal_sse_payload"}`)
	}
	_, _ = io.WriteString(w, "data: ")
	_, _ = w.Write(raw)
	_, _ = io.WriteString(w, "\n\n")
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	traceID, _ := TraceIDFromContext(r.Context())
	resp := errorResponse{}
	resp.Error.Code = apperror.CodeOf(err)
	resp.Error.Message = err.Error()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apperror.StatusCode(err))
	_ = json.NewEncoder(w).Encode(envelopeResponse{
		Data:  nil,
		Error: resp.Error,
		Meta: map[string]any{
			"trace_id": traceID,
		},
	})
}

type contextKey string

const traceIDKey contextKey = "trace_id"
const requestStartedAtKey contextKey = "request_started_at"

func traceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = newTraceID()
		}
		ctx := context.WithValue(r.Context(), traceIDKey, traceID)
		ctx = context.WithValue(ctx, requestStartedAtKey, time.Now().UTC())
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

func splitStreamDeltas(content string, chunkSize int) []string {
	if chunkSize <= 0 {
		chunkSize = 32
	}
	runes := []rune(content)
	if len(runes) == 0 {
		return nil
	}
	out := make([]string, 0, (len(runes)+chunkSize-1)/chunkSize)
	for start := 0; start < len(runes); start += chunkSize {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[start:end]))
	}
	return out
}

func (s *Server) publishGroupEvent(ctx context.Context, event infraevent.Envelope) {
	if s.eventBroker == nil {
		return
	}
	s.eventBroker.Publish(ctx, event)
}

func (s *Server) checkStatuses(ctx context.Context) map[string]string {
	checks := map[string]string{}
	for _, check := range s.readiness {
		if check.Name == "" || check.Check == nil {
			continue
		}
		if err := check.Check(ctx); err != nil {
			checks[check.Name] = "down"
			continue
		}
		checks[check.Name] = "up"
	}
	return checks
}

func (s *Server) observeTask(name string, success bool, started time.Time) {
	if s.metrics == nil || started.IsZero() {
		return
	}
	s.metrics.ObserveTask(name, success, time.Since(started))
}

func metricsMiddleware(metrics *monitoring.Metrics, next http.Handler) http.Handler {
	if metrics == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		started := time.Now()
		next.ServeHTTP(recorder, r)
		metrics.ObserveHTTPRequest(r.Method, routePattern(r.URL.Path), recorder.status, time.Since(started))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func routePattern(path string) string {
	switch {
	case path == "/healthz":
		return "/healthz"
	case path == "/readyz":
		return "/readyz"
	case path == "/metrics":
		return "/metrics"
	case strings.HasPrefix(path, "/api/v1/groups/") && strings.Contains(path, "/resources/upload"):
		return "/api/v1/groups/{groupId}/resources/upload"
	case strings.HasPrefix(path, "/api/v1/groups/") && strings.Contains(path, "/web-resources"):
		return "/api/v1/groups/{groupId}/web-resources"
	case strings.HasPrefix(path, "/api/v1/groups/") && strings.Contains(path, "/framework-graph"):
		return "/api/v1/groups/{groupId}/framework-graph"
	case strings.HasPrefix(path, "/api/v1/groups/") && strings.Contains(path, "/resources/"):
		return "/api/v1/groups/{groupId}/resources/{resourceId}"
	case path == "/api/v1/groups":
		return "/api/v1/groups"
	case strings.HasPrefix(path, "/api/v1/groups/"):
		return "/api/v1/groups/{groupId}"
	case path == "/api/v1/conversations":
		return "/api/v1/conversations"
	case strings.HasPrefix(path, "/api/v1/conversations/") && strings.Contains(path, "/messages"):
		return "/api/v1/conversations/{conversationId}/messages"
	case strings.HasPrefix(path, "/api/v1/conversations/"):
		return "/api/v1/conversations/{conversationId}"
	case strings.HasPrefix(path, "/api/v1/events/groups/"):
		return "/api/v1/events/groups/{groupId}"
	case strings.HasPrefix(path, "/api/v1/graphs/") && strings.Contains(path, "/expand"):
		return "/api/v1/graphs/{graphId}/nodes/{nodeId}/expand"
	case strings.HasPrefix(path, "/api/v1/graphs/") && strings.Contains(path, "/nodes/"):
		return "/api/v1/graphs/{graphId}/nodes/{nodeId}"
	case strings.HasPrefix(path, "/api/v1/resources/") && strings.Contains(path, "/retry"):
		return "/api/v1/resources/{resourceId}/retry"
	case strings.HasPrefix(path, "/api/v1/resources/") && strings.Contains(path, "/graph"):
		return "/api/v1/resources/{resourceId}/graph"
	default:
		return path
	}
}

func startedAt(r *http.Request) time.Time {
	value, ok := r.Context().Value(requestStartedAtKey).(time.Time)
	if !ok {
		return time.Time{}
	}
	return value
}
