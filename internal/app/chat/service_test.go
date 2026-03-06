package chat

import (
	"context"
	"testing"

	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	pipelinechat "github.com/tylor/goaipj/internal/pipeline/chat"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func TestIndexAndSearchGroupContext(t *testing.T) {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	group, err := groupService.CreateGroup(context.Background(), appgroup.CreateGroupInput{Name: "Backend"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	service := NewService(groupService, appgraph.NewService(appgraph.NewInMemoryRepository()), NewInMemoryChunkRepository(), NewInMemoryConversationRepository(), pipelinechat.NewAnswerer())

	indexed, err := service.IndexResourceChunks(context.Background(), IndexChunksInput{
		GroupID:    group.ID,
		ResourceID: "resource-1",
		Chunks: []string{
			"The parser extracts text from uploaded documents before normalization.",
			"The worker consumes queued jobs and generates resource graphs.",
			"The frontend renders graph nodes in the browser.",
		},
	})
	if err != nil {
		t.Fatalf("IndexResourceChunks() error = %v", err)
	}
	if len(indexed) != 3 {
		t.Fatalf("len(indexed) = %d, want 3", len(indexed))
	}
	if len(indexed[0].Embedding) == 0 {
		t.Fatalf("expected embedding to be written")
	}

	results, err := service.SearchGroupContext(context.Background(), group.ID, "worker graph jobs", 2)
	if err != nil {
		t.Fatalf("SearchGroupContext() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Chunk.ID != indexed[1].ID {
		t.Fatalf("top result = %q, want worker chunk %q", results[0].Chunk.ID, indexed[1].ID)
	}
}

func TestConversationAskStoresHistoryAndCitations(t *testing.T) {
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
		Title:      "Parser Graph",
		Document: pipelinegraph.Document{
			Summary: "Parser and worker collaborate to process uploaded resources.",
			Nodes: []pipelinegraph.Node{
				{ID: "node-parser", Name: "Parser", Type: "component", Level: 0, Description: "Parses files"},
				{ID: "node-worker", Name: "Worker", Type: "component", Level: 1, Description: "Runs background jobs"},
			},
			Edges: []pipelinegraph.Edge{
				{ID: "edge-1", SourceID: "node-parser", TargetID: "node-worker", Relation: "handled_by"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveResourceGraph() error = %v", err)
	}
	graphRepo.AddExample(saved.Graph.ID, appgraph.Example{
		ID:      "example-1",
		NodeID:  "node-parser",
		Content: "A PDF is parsed into markdown before graph generation.",
	})

	service := NewService(groupService, graphService, NewInMemoryChunkRepository(), NewInMemoryConversationRepository(), pipelinechat.NewAnswerer())
	if _, err := service.IndexResourceChunks(context.Background(), IndexChunksInput{
		GroupID:    group.ID,
		ResourceID: saved.Graph.ResourceID,
		Chunks: []string{
			"The parser extracts text and produces normalized markdown for graph generation.",
			"The worker asynchronously consumes jobs from the processing queue.",
		},
	}); err != nil {
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

	reply, err := service.Ask(context.Background(), AskInput{
		ConversationID: conversation.ID,
		Content:        "这个节点的作用是什么？",
	})
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}

	if reply.AssistantMessage.Content == "" {
		t.Fatalf("expected assistant answer")
	}
	if len(reply.AssistantMessage.Examples) == 0 {
		t.Fatalf("expected assistant examples")
	}
	if len(reply.AssistantMessage.CitedChunkIDs) == 0 {
		t.Fatalf("expected cited chunk ids")
	}
	if len(reply.AssistantMessage.CitedNodeIDs) == 0 || reply.AssistantMessage.CitedNodeIDs[0] != "node-parser" {
		t.Fatalf("cited node ids = %#v", reply.AssistantMessage.CitedNodeIDs)
	}

	stored, err := service.GetConversation(context.Background(), conversation.ID)
	if err != nil {
		t.Fatalf("GetConversation() error = %v", err)
	}
	if len(stored.Messages) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(stored.Messages))
	}
	if stored.Messages[1].ContextSnapshot.ResourceSummary == "" {
		t.Fatalf("expected context snapshot to persist resource summary")
	}
}
