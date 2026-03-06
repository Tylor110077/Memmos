package chat

import "testing"

func TestAnswererBuildsFocusedResponse(t *testing.T) {
	answerer := NewAnswerer()

	result, err := answerer.Answer(Input{
		Question: "这个节点的作用是什么？",
		CurrentNode: NodeContext{
			ID:          "node-root",
			Name:        "Parser",
			Description: "Parses uploaded documents.",
			Meaning:     "Turns files into normalized text.",
		},
		Neighbors: []NeighborContext{
			{ID: "node-worker", Name: "Worker", Relation: "processed_by"},
		},
		ResourceSummary: "The backend parses documents and schedules follow-up graph generation jobs.",
		Examples:        []string{"A PDF upload is parsed before graph generation starts."},
		Chunks: []ChunkContext{
			{ID: "chunk-1", Content: "The parser extracts text from uploaded PDF files."},
			{ID: "chunk-2", Content: "Workers consume processing jobs asynchronously."},
		},
	})
	if err != nil {
		t.Fatalf("Answer() error = %v", err)
	}

	if result.Answer == "" {
		t.Fatalf("expected answer")
	}
	if len(result.Examples) == 0 {
		t.Fatalf("expected examples")
	}
	if len(result.CitedChunkIDs) == 0 || result.CitedChunkIDs[0] != "chunk-1" {
		t.Fatalf("cited chunks = %#v, want chunk-1 first", result.CitedChunkIDs)
	}
	if len(result.CitedNodeIDs) < 2 {
		t.Fatalf("cited nodes = %#v, want current node and neighbor", result.CitedNodeIDs)
	}
}
