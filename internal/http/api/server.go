package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	appgroup "github.com/tylor/goaipj/internal/app/group"
	"github.com/tylor/goaipj/internal/infra/apperror"
)

type Dependencies struct {
	GroupService *appgroup.Service
	Logger       *log.Logger
}

type Server struct {
	groupService *appgroup.Service
	logger       *log.Logger
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

func NewServer(deps Dependencies) http.Handler {
	server := &Server{
		groupService: deps.GroupService,
		logger:       deps.Logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", server.handleHealthz)
	mux.HandleFunc("/readyz", server.handleReadyz)
	mux.HandleFunc("/api/v1/groups", server.handleGroups)
	mux.HandleFunc("/api/v1/groups/", server.handleGroupByID)
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
	if id == "" || strings.Contains(id, "/") {
		writeError(w, r, apperror.New(apperror.CodeNotFound, "group not found"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		group, err := s.groupService.GetGroup(r.Context(), id)
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

		group, err := s.groupService.UpdateGroup(r.Context(), id, appgroup.UpdateGroupInput{Name: req.Name})
		if err != nil {
			writeError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, toGroupResponse(group))
	case http.MethodDelete:
		if err := s.groupService.DeleteGroup(r.Context(), id); err != nil {
			writeError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func toGroupResponse(group appgroup.Group) groupResponse {
	return groupResponse{
		ID:        group.ID,
		Name:      group.Name,
		CreatedAt: group.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: group.UpdatedAt.Format(time.RFC3339Nano),
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
