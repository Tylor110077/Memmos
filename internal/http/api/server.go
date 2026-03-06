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
	"strings"
	"time"

	appgroup "github.com/tylor/goaipj/internal/app/group"
	appresource "github.com/tylor/goaipj/internal/app/resource"
	"github.com/tylor/goaipj/internal/infra/apperror"
)

type Dependencies struct {
	GroupService    *appgroup.Service
	ResourceService *appresource.Service
	Logger          *log.Logger
}

type Server struct {
	groupService    *appgroup.Service
	resourceService *appresource.Service
	logger          *log.Logger
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

func NewServer(deps Dependencies) http.Handler {
	server := &Server{
		groupService:    deps.GroupService,
		resourceService: deps.ResourceService,
		logger:          deps.Logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", server.handleHealthz)
	mux.HandleFunc("/readyz", server.handleReadyz)
	mux.HandleFunc("/api/v1/groups", server.handleGroups)
	mux.HandleFunc("/api/v1/groups/", server.handleGroupByID)
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
	if s.resourceService == nil {
		writeError(w, r, apperror.New(apperror.CodeInternal, "resource service not configured"))
		return
	}

	switch {
	case len(parts) == 2 && parts[0] == "resources" && parts[1] == "upload" && r.Method == http.MethodPost:
		s.handleUploadResource(w, r, groupID)
	case len(parts) == 1 && parts[0] == "resources" && r.Method == http.MethodGet:
		s.handleListResources(w, r, groupID)
	case len(parts) == 2 && parts[0] == "resources" && r.Method == http.MethodGet:
		s.handleGetResource(w, r, groupID, parts[1])
	case len(parts) == 1 && parts[0] == "web-resources" && r.Method == http.MethodPost:
		s.handleCreateWebResource(w, r, groupID)
	default:
		writeError(w, r, apperror.New(apperror.CodeNotFound, "route not found"))
	}
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
