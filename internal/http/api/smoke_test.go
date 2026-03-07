package api

import (
	"context"
	"net/http"
	"testing"

	appchat "github.com/tylor/goaipj/internal/app/chat"
	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	appresource "github.com/tylor/goaipj/internal/app/resource"
	"github.com/tylor/goaipj/internal/infra/storage"
	pipelinechat "github.com/tylor/goaipj/internal/pipeline/chat"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestAPISmokeFlow(t *testing.T) {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	resourceService := appresource.NewService(
		groupService,
		appresource.NewInMemoryRepository(),
		appresource.NewInMemoryArtifactRepository(),
		appresource.NewInMemoryJobPublisher(),
		storage.NewStore(newHTTPFakeBucketClient(), "bucket"),
	)
	graphRepo := appgraph.NewInMemoryRepository()
	graphService := appgraph.NewService(graphRepo)
	chatService := appchat.NewService(
		groupService,
		graphService,
		appchat.NewInMemoryChunkRepository(),
		appchat.NewInMemoryConversationRepository(),
		pipelinechat.NewAnswerer(),
		nil,
	)

	server := NewServer(Dependencies{
		GroupService:       groupService,
		ResourceService:    resourceService,
		GraphService:       graphService,
		ChatService:        chatService,
		ExpansionGenerator: pipelinegraph.NewExpansionGenerator(pipelinegraph.ExpansionRules{MaxNewNodes: 1}),
	})

	createGroupResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/groups", map[string]any{
		"name": "Smoke Group",
	})
	if createGroupResp.Code != http.StatusCreated {
		t.Fatalf("create group status = %d body=%s", createGroupResp.Code, createGroupResp.Body.String())
	}
	var group groupResponse
	decodeJSONResponse(t, createGroupResp, &group)

	listGroupsResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups", nil)
	if listGroupsResp.Code != http.StatusOK {
		t.Fatalf("list groups status = %d body=%s", listGroupsResp.Code, listGroupsResp.Body.String())
	}
	getGroupResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/groups/"+group.ID, nil)
	if getGroupResp.Code != http.StatusOK {
		t.Fatalf("get group status = %d body=%s", getGroupResp.Code, getGroupResp.Body.String())
	}
	updateGroupResp := performJSONRequest(t, server, http.MethodPatch, "/api/v1/groups/"+group.ID, map[string]any{
		"name": "Smoke Group Updated",
	})
	if updateGroupResp.Code != http.StatusOK {
		t.Fatalf("update group status = %d body=%s", updateGroupResp.Code, updateGroupResp.Body.String())
	}

	uploadResp := uploadFile(t, server, "/api/v1/groups/"+group.ID+"/resources/upload", "file", "notes.pdf", []byte("pdf"))
	if uploadResp.Code != http.StatusCreated {
		t.Fatalf("upload status = %d body=%s", uploadResp.Code, uploadResp.Body.String())
	}
	var uploaded resourceWithJobResponse
	decodeJSONResponse(t, uploadResp, &uploaded)

	if _, err := graphService.SaveResourceGraph(context.Background(), appgraph.SaveInput{
		GroupID:    group.ID,
		ResourceID: uploaded.Resource.ID,
		Title:      "Smoke Graph",
		Document: pipelinegraph.Document{
			Summary: "Smoke graph summary",
			Nodes: []pipelinegraph.Node{
				{ID: "root", Name: "Root", Type: "topic", Level: 0},
			},
		},
	}); err != nil {
		t.Fatalf("SaveResourceGraph() error = %v", err)
	}
	if _, err := chatService.IndexResourceChunks(context.Background(), appchat.IndexChunksInput{
		GroupID:    group.ID,
		ResourceID: uploaded.Resource.ID,
		Chunks:     []string{"Root captures the main learning concept."},
	}); err != nil {
		t.Fatalf("IndexResourceChunks() error = %v", err)
	}

	graphResp := performJSONRequest(t, server, http.MethodGet, "/api/v1/resources/"+uploaded.Resource.ID+"/graph", nil)
	if graphResp.Code != http.StatusOK {
		t.Fatalf("graph status = %d body=%s", graphResp.Code, graphResp.Body.String())
	}

	createConversationResp := performJSONRequest(t, server, http.MethodPost, "/api/v1/conversations", map[string]any{
		"group_id":        group.ID,
		"graph_id":        "graph-unknown",
		"current_node_id": "root",
	})
	if createConversationResp.Code == http.StatusCreated {
		t.Fatalf("expected invalid graph id to fail")
	}

	saved, err := graphService.GetResourceGraph(context.Background(), uploaded.Resource.ID, appgraph.QueryOptions{IncludeExpansion: true})
	if err != nil {
		t.Fatalf("GetResourceGraph() error = %v", err)
	}

	createConversationResp = performJSONRequest(t, server, http.MethodPost, "/api/v1/conversations", map[string]any{
		"group_id":        group.ID,
		"graph_id":        saved.Graph.ID,
		"current_node_id": "root",
	})
	if createConversationResp.Code != http.StatusCreated {
		t.Fatalf("create conversation status = %d body=%s", createConversationResp.Code, createConversationResp.Body.String())
	}

	deleteGroupResp := performJSONRequest(t, server, http.MethodDelete, "/api/v1/groups/"+group.ID, nil)
	if deleteGroupResp.Code != http.StatusNoContent {
		t.Fatalf("delete group status = %d body=%s", deleteGroupResp.Code, deleteGroupResp.Body.String())
	}
}
