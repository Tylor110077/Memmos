package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDevelopmentServerWiresCoreRoutes(t *testing.T) {
	server := newDevelopmentServer(log.New(io.Discard, "", 0))

	createGroupResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/groups", map[string]any{
		"name": "Bootstrap Group",
	})
	if createGroupResp.Code != http.StatusCreated {
		t.Fatalf("create group status = %d body=%s", createGroupResp.Code, createGroupResp.Body.String())
	}

	var created struct {
		ID string `json:"id"`
	}
	decodeJSONData(t, createGroupResp.Body.Bytes(), &created)
	if created.ID == "" {
		t.Fatalf("expected group id")
	}

	resourcesResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups/"+created.ID+"/resources", nil)
	if resourcesResp.Code != http.StatusOK {
		t.Fatalf("resources status = %d body=%s", resourcesResp.Code, resourcesResp.Body.String())
	}

	frameworkResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/groups/"+created.ID+"/framework-graph/generate", map[string]any{})
	if frameworkResp.Code != http.StatusAccepted {
		t.Fatalf("framework generate status = %d body=%s", frameworkResp.Code, frameworkResp.Body.String())
	}

	jobResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/jobs/missing", nil)
	if jobResp.Code != http.StatusNotFound {
		t.Fatalf("job status = %d body=%s", jobResp.Code, jobResp.Body.String())
	}

	metricsResp := performJSONRequest(t, server, http.MethodGet, "/metrics", nil)
	if metricsResp.Code != http.StatusOK {
		t.Fatalf("metrics status = %d body=%s", metricsResp.Code, metricsResp.Body.String())
	}
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader = http.NoBody
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
		reader = bytesReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp
}

func decodeJSONData(t *testing.T, raw []byte, dst any) {
	t.Helper()
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("json.Unmarshal() error = %v body=%s", err, string(raw))
	}
	if err := json.Unmarshal(envelope.Data, dst); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v body=%s", err, string(raw))
	}
}

func bytesReader(raw []byte) io.Reader {
	return bytes.NewReader(raw)
}
