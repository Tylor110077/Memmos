package api

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	appgroup "github.com/tylor/goaipj/internal/app/group"
	appresource "github.com/tylor/goaipj/internal/app/resource"
	"github.com/tylor/goaipj/internal/infra/storage"
)

func TestResourceEndpoints(t *testing.T) {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	resourceService := appresource.NewService(
		groupService,
		appresource.NewInMemoryRepository(),
		appresource.NewInMemoryArtifactRepository(),
		appresource.NewInMemoryJobPublisher(),
		storage.NewStore(newHTTPFakeBucketClient(), "bucket"),
	)

	server := NewServer(Dependencies{
		GroupService:    groupService,
		ResourceService: resourceService,
	})

	uploadResp := uploadFile(t, server, "/api/v1/groups/"+group.ID+"/resources/upload", "file", "notes.pdf", []byte("pdf"))
	if uploadResp.Code != http.StatusCreated {
		t.Fatalf("upload status = %d body=%s", uploadResp.Code, uploadResp.Body.String())
	}

	var created resourceWithJobResponse
	decodeJSONResponse(t, uploadResp, &created)
	if created.Resource.Type != "file" {
		t.Fatalf("type = %q, want file", created.Resource.Type)
	}

	webResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/groups/"+group.ID+"/web-resources", map[string]any{
		"url": "https://example.com/page",
	})
	if webResp.Code != http.StatusCreated {
		t.Fatalf("web status = %d body=%s", webResp.Code, webResp.Body.String())
	}

	listResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups/"+group.ID+"/resources", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status = %d", listResp.Code)
	}

	var list []resourceResponse
	decodeJSONResponse(t, listResp, &list)
	if len(list) != 2 {
		t.Fatalf("len(list) = %d, want 2", len(list))
	}

	getResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups/"+group.ID+"/resources/"+created.Resource.ID, nil)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get status = %d", getResp.Code)
	}

	retryResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/resources/"+created.Resource.ID+"/retry", nil)
	if retryResp.Code != http.StatusBadRequest {
		t.Fatalf("retry status = %d, want 400 because uploaded resources cannot retry", retryResp.Code)
	}
}

func uploadFile(t *testing.T, handler http.Handler, path, fieldName, filename string, content []byte) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp
}

type httpFakeBucketClient struct {
	objects map[string][]byte
}

func newHTTPFakeBucketClient() *httpFakeBucketClient {
	return &httpFakeBucketClient{objects: map[string][]byte{}}
}

func (f *httpFakeBucketClient) PutObject(_ context.Context, bucket, key string, reader io.Reader, _ int64, opts storage.PutObjectOptions) (storage.UploadInfo, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return storage.UploadInfo{}, err
	}
	f.objects[bucket+"/"+key] = raw
	return storage.UploadInfo{Key: key, Size: int64(len(raw)), ContentType: opts.ContentType}, nil
}

func (f *httpFakeBucketClient) GetObject(_ context.Context, bucket, key string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(f.objects[bucket+"/"+key])), nil
}
