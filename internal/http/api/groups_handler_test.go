package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appgroup "github.com/tylor/goaipj/internal/app/group"
)

func TestGroupsCRUD(t *testing.T) {
	server := newTestServer()

	createResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/groups", map[string]any{
		"name": "Backend Group",
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201 body=%s", createResp.Code, createResp.Body.String())
	}

	var created groupResponse
	decodeJSONResponse(t, createResp, &created)
	if created.Name != "Backend Group" {
		t.Fatalf("created name = %q, want Backend Group", created.Name)
	}

	listResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200", listResp.Code)
	}
	var list []groupResponse
	decodeJSONResponse(t, listResp, &list)
	if len(list) != 1 {
		t.Fatalf("len(list) = %d, want 1", len(list))
	}

	getResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups/"+created.ID, nil)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get status = %d, want 200", getResp.Code)
	}

	updateResp := performJSONRequest(t, server, http.MethodPatch, "/api/v1/groups/"+created.ID, map[string]any{
		"name": "Renamed Group",
	})
	if updateResp.Code != http.StatusOK {
		t.Fatalf("update status = %d, want 200 body=%s", updateResp.Code, updateResp.Body.String())
	}
	var updated groupResponse
	decodeJSONResponse(t, updateResp, &updated)
	if updated.Name != "Renamed Group" {
		t.Fatalf("updated name = %q, want Renamed Group", updated.Name)
	}

	deleteResp := performJSONRequest(t, server, http.MethodDelete, "/api/v1/groups/"+created.ID, nil)
	if deleteResp.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", deleteResp.Code)
	}

	missingResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups/"+created.ID, nil)
	if missingResp.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d, want 404", missingResp.Code)
	}
}

func TestGroupsValidationAndErrorShape(t *testing.T) {
	server := newTestServer()

	resp := performJSONRequest(t, server, http.MethodPost, "/api/v1/groups", map[string]any{
		"name": "",
	})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 body=%s", resp.Code, resp.Body.String())
	}

	var body errorResponse
	decodeErrorResponse(t, resp, &body)
	if body.Error.Code == "" {
		t.Fatalf("expected error code")
	}
	if body.Error.Message == "" {
		t.Fatalf("expected error message")
	}
	if !strings.Contains(resp.Body.String(), `"trace_id"`) {
		t.Fatalf("expected trace id")
	}
}

func TestHealthEndpoints(t *testing.T) {
	server := newTestServer()

	healthz := performJSONRequest(t, server, http.MethodGet, "/healthz", nil)
	if healthz.Code != http.StatusOK {
		t.Fatalf("healthz = %d, want 200", healthz.Code)
	}

	readyz := performJSONRequest(t, server, http.MethodGet, "/readyz", nil)
	if readyz.Code != http.StatusOK {
		t.Fatalf("readyz = %d, want 200", readyz.Code)
	}
}

func newTestServer() http.Handler {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	return NewServer(Dependencies{
		GroupService: groupService,
	})
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp
}

func decodeJSONResponse(t *testing.T, resp *httptest.ResponseRecorder, dst any) {
	t.Helper()
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err == nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, dst); err != nil {
			t.Fatalf("json.Unmarshal(data) error = %v body=%s", err, resp.Body.String())
		}
		return
	}
	if err := json.Unmarshal(resp.Body.Bytes(), dst); err != nil {
		t.Fatalf("json.Unmarshal() error = %v body=%s", err, resp.Body.String())
	}
}

func decodeErrorResponse(t *testing.T, resp *httptest.ResponseRecorder, dst any) {
	t.Helper()
	var envelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("json.Unmarshal() error = %v body=%s", err, resp.Body.String())
	}
	switch target := dst.(type) {
	case *errorResponse:
		target.Error = envelope.Error
	default:
		t.Fatalf("unsupported decodeErrorResponse target %T", dst)
	}
}
