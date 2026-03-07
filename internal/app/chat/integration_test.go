package chat

import (
	"context"
	"testing"

	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	pipelinechat "github.com/tylor/goaipj/internal/pipeline/chat"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestConversationFlowPersistsAnswerAndCitations(t *testing.T) {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	graphRepo := appgraph.NewInMemoryRepository()
	graphService := appgraph.NewService(graphRepo)
	saved, err := graphService.SaveResourceGraph(context.Background(), appgraph.SaveInput{
		GroupID:    group.ID,
		ResourceID: "resource-1",
		Title:      "Pipeline",
		Document: pipelinegraph.Document{
			Summary: "Uploads are parsed and converted into graph context.",
			Nodes: []pipelinegraph.Node{
				{ID: "node-parser", Name: "Parser", Type: "component", Level: 0},
				{ID: "node-worker", Name: "Worker", Type: "component", Level: 1},
			},
			Edges: []pipelinegraph.Edge{
				{ID: "edge-1", SourceID: "node-parser", TargetID: "node-worker", Relation: "feeds"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveResourceGraph() error = %v", err)
	}
	graphRepo.AddExample(saved.Graph.ID, appgraph.Example{
		ID:      "ex-1",
		NodeID:  "node-parser",
		Content: "A PDF is parsed into markdown before graph generation.",
	})

	service := NewService(
		groupService,
		graphService,
		NewInMemoryChunkRepository(),
		NewInMemoryConversationRepository(),
		pipelinechat.NewAnswerer(),
		nil,
	)
	indexed, err := service.IndexResourceChunks(context.Background(), IndexChunksInput{
		GroupID:    group.ID,
		ResourceID: saved.Graph.ResourceID,
		Chunks: []string{
			"The parser extracts text from uploaded files and normalizes content.",
			"The worker consumes async jobs and writes graph results.",
		},
	})
	if err != nil {
		t.Fatalf("IndexResourceChunks() error = %v", err)
	}

	conversation, err := service.CreateConversation(context.Background(), CreateConversationInput{
		GroupID:       group.ID,
		GraphID:       saved.Graph.ID,
		CurrentNodeID: "node-parser",
	})
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}

	result, err := service.Ask(context.Background(), AskInput{
		ConversationID: conversation.ID,
		Content:        "Parser 在链路里的作用是什么？",
	})
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}

	if len(result.AssistantMessage.CitedChunkIDs) == 0 {
		t.Fatalf("expected cited chunks")
	}
	if result.AssistantMessage.CitedChunkIDs[0] != indexed[0].ID {
		t.Fatalf("first cited chunk = %q, want %q", result.AssistantMessage.CitedChunkIDs[0], indexed[0].ID)
	}
	if len(result.AssistantMessage.CitedNodeIDs) == 0 || result.AssistantMessage.CitedNodeIDs[0] != "node-parser" {
		t.Fatalf("cited node ids = %#v", result.AssistantMessage.CitedNodeIDs)
	}

	stored, err := service.GetConversation(context.Background(), conversation.ID)
	if err != nil {
		t.Fatalf("GetConversation() error = %v", err)
	}
	if len(stored.Messages) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(stored.Messages))
	}
	if stored.Messages[1].ContextSnapshot.CurrentNodeID != "node-parser" {
		t.Fatalf("current_node_id = %q", stored.Messages[1].ContextSnapshot.CurrentNodeID)
	}
	if len(stored.Messages[1].ContextSnapshot.RetrievedChunkIDs) == 0 {
		t.Fatalf("expected retrieved chunk ids in snapshot")
	}
}
