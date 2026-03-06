package sqlstore

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/tylor/goaipj/internal/infra/sqlstore/gen"
)

func TestRepositoriesPersistResourceGraphAndConversationData(t *testing.T) {
	testDB := startTestPostgres(t)
	db := openTestDB(t, testDB.DSN)
	t.Cleanup(func() { _ = db.Close() })

	queries := gen.New(db)
	ctx := context.Background()
	now := time.Now().UTC()

	if _, err := queries.CreateGroup(ctx, gen.CreateGroupParams{
		ID:        "group-1",
		Name:      "Backend",
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: sql.NullTime{},
	}); err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	resourceRow, err := queries.CreateResource(ctx, gen.CreateResourceParams{
		ID:           "resource-1",
		GroupID:      "group-1",
		Name:         "Architecture Notes",
		Type:         "web",
		Status:       "uploaded",
		FailedStage:  sql.NullString{},
		ErrorMessage: sql.NullString{},
		CreatedAt:    now,
		UpdatedAt:    now,
		DeletedAt:    sql.NullTime{},
	})
	if err != nil {
		t.Fatalf("CreateResource() error = %v", err)
	}
	if resourceRow.ID != "resource-1" {
		t.Fatalf("resource id = %q", resourceRow.ID)
	}

	graphRow, err := queries.CreateGraph(ctx, gen.CreateGraphParams{
		ID:         "graph-1",
		GroupID:    "group-1",
		ResourceID: sql.NullString{String: "resource-1", Valid: true},
		GraphType:  "resource",
		Version:    1,
		Status:     "active",
		Title:      sql.NullString{String: "Architecture", Valid: true},
		Summary:    sql.NullString{String: "summary", Valid: true},
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
		ArchivedAt: sql.NullTime{},
	})
	if err != nil {
		t.Fatalf("CreateGraph() error = %v", err)
	}
	if graphRow.ID != "graph-1" {
		t.Fatalf("graph id = %q", graphRow.ID)
	}

	nodeRow, err := queries.CreateGraphNode(ctx, gen.CreateGraphNodeParams{
		ID:           "node-1",
		GraphID:      "graph-1",
		ParentNodeID: sql.NullString{},
		Name:         "Parser",
		Description:  sql.NullString{String: "Parses uploads", Valid: true},
		Meaning:      sql.NullString{String: "Turns resources into text", Valid: true},
		NodeType:     "component",
		SourceType:   "resource",
		Level:        0,
		IsExpansion:  false,
		Metadata:     []byte(`{}`),
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		t.Fatalf("CreateGraphNode() error = %v", err)
	}
	if nodeRow.ID != "node-1" {
		t.Fatalf("node id = %q", nodeRow.ID)
	}

	if _, err := queries.CreateConversation(ctx, gen.CreateConversationParams{
		ID:        "conv-1",
		GroupID:   "group-1",
		GraphID:   sql.NullString{String: "graph-1", Valid: true},
		NodeID:    sql.NullString{String: "node-1", Valid: true},
		Title:     sql.NullString{String: "Ask parser", Valid: true},
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}

	if _, err := queries.CreateConversationMessage(ctx, gen.CreateConversationMessageParams{
		ID:              "msg-1",
		ConversationID:  "conv-1",
		Role:            "assistant",
		Content:         "Parser extracts text first.",
		CitedChunkIds:   []string{"chunk-1"},
		CitedNodeIds:    []string{"node-1"},
		ContextSnapshot: []byte(`{"current_node_id":"node-1"}`),
		CreatedAt:       now,
	}); err != nil {
		t.Fatalf("CreateConversationMessage() error = %v", err)
	}

	storedConversation, err := queries.GetConversation(ctx, "conv-1")
	if err != nil {
		t.Fatalf("GetConversation() error = %v", err)
	}
	if !storedConversation.NodeID.Valid || storedConversation.NodeID.String != "node-1" {
		t.Fatalf("node_id = %#v", storedConversation.NodeID)
	}

	messages, err := queries.ListConversationMessages(ctx, "conv-1")
	if err != nil {
		t.Fatalf("ListConversationMessages() error = %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(messages))
	}
	if messages[0].Content != "Parser extracts text first." {
		t.Fatalf("content = %q", messages[0].Content)
	}

	nodes, err := queries.ListNodesByGraph(ctx, "graph-1")
	if err != nil {
		t.Fatalf("ListNodesByGraph() error = %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("len(nodes) = %d, want 1", len(nodes))
	}
}

type testPostgres struct {
	DSN    string
	logBuf *bytes.Buffer
}

func startTestPostgres(t *testing.T) testPostgres {
	t.Helper()

	if _, err := exec.LookPath("postgres"); err != nil {
		t.Skip("postgres binary not available")
	}
	if _, err := exec.LookPath("initdb"); err != nil {
		t.Skip("initdb binary not available")
	}

	baseDir, err := os.MkdirTemp("/tmp", "goaipj-pg-*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(baseDir) })
	dataDir := filepath.Join(baseDir, "data")
	socketDir := filepath.Join(baseDir, "sock")
	if err := os.MkdirAll(socketDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(socketDir) error = %v", err)
	}
	port := freePort(t)

	initdb := exec.Command("initdb", "-D", dataDir, "-A", "trust", "-U", "postgres")
	initdb.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")
	if output, err := initdb.CombinedOutput(); err != nil {
		t.Fatalf("initdb error = %v output=%s", err, string(output))
	}

	logBuf := &bytes.Buffer{}
	postgres := exec.Command("postgres",
		"-D", dataDir,
		"-k", socketDir,
		"-p", port,
		"-c", "listen_addresses=",
	)
	postgres.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")
	postgres.Stdout = logBuf
	postgres.Stderr = logBuf
	if err := postgres.Start(); err != nil {
		t.Fatalf("postgres.Start() error = %v", err)
	}
	t.Cleanup(func() {
		if postgres.Process != nil {
			_ = postgres.Process.Kill()
			_, _ = postgres.Process.Wait()
		}
	})

	dsn := fmt.Sprintf("host=%s port=%s user=postgres dbname=postgres sslmode=disable", socketDir, port)
	db := openTestDB(t, dsn)
	defer db.Close()
	waitForReady(t, db, logBuf)
	applyTestSchema(t, db)

	return testPostgres{
		DSN:    dsn,
		logBuf: logBuf,
	}
}

func openTestDB(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	return db
}

func waitForReady(t *testing.T, db *sql.DB, logBuf *bytes.Buffer) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		if err := db.Ping(); err == nil {
			return
		} else {
			lastErr = err
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("postgres did not become ready: %v logs=%s", lastErr, strings.TrimSpace(logBuf.String()))
}

func applyTestSchema(t *testing.T, db *sql.DB) {
	t.Helper()
	schema := []string{
		`create table groups (
			id text primary key,
			name varchar(64) not null,
			created_at timestamptz not null,
			updated_at timestamptz not null,
			deleted_at timestamptz
		);`,
		`create table resources (
			id text primary key,
			group_id text not null references groups(id) on delete cascade,
			name text not null,
			type text not null,
			status text not null,
			failed_stage text,
			error_message text,
			created_at timestamptz not null,
			updated_at timestamptz not null,
			deleted_at timestamptz
		);`,
		`create table graphs (
			id text primary key,
			group_id text not null references groups(id) on delete cascade,
			resource_id text references resources(id) on delete cascade,
			graph_type text not null,
			version integer not null,
			status text not null,
			title text,
			summary text,
			is_active boolean not null,
			created_at timestamptz not null,
			updated_at timestamptz not null,
			archived_at timestamptz
		);`,
		`create table graph_nodes (
			id text primary key,
			graph_id text not null references graphs(id) on delete cascade,
			parent_node_id text references graph_nodes(id) on delete set null,
			name text not null,
			description text,
			meaning text,
			node_type text not null,
			source_type text not null,
			level integer not null,
			is_expansion boolean not null,
			metadata jsonb not null default '{}'::jsonb,
			created_at timestamptz not null,
			updated_at timestamptz not null
		);`,
		`create table graph_edges (
			id text primary key,
			graph_id text not null references graphs(id) on delete cascade,
			from_node_id text not null references graph_nodes(id) on delete cascade,
			to_node_id text not null references graph_nodes(id) on delete cascade,
			relation_type text not null,
			description text,
			is_expansion boolean not null,
			metadata jsonb not null default '{}'::jsonb,
			created_at timestamptz not null
		);`,
		`create table conversations (
			id text primary key,
			group_id text not null references groups(id) on delete cascade,
			graph_id text references graphs(id) on delete set null,
			node_id text references graph_nodes(id) on delete set null,
			title text,
			created_at timestamptz not null,
			updated_at timestamptz not null
		);`,
		`create table conversation_messages (
			id text primary key,
			conversation_id text not null references conversations(id) on delete cascade,
			role text not null,
			content text not null,
			cited_chunk_ids text[] not null default '{}',
			cited_node_ids text[] not null default '{}',
			context_snapshot jsonb not null default '{}'::jsonb,
			created_at timestamptz not null
		);`,
	}
	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("apply schema error = %v stmt=%s", err, compactSQL(stmt))
		}
	}
}

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	defer l.Close()
	_, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		t.Fatalf("SplitHostPort() error = %v", err)
	}
	return port
}

func compactSQL(input string) string {
	return strings.Join(strings.Fields(input), " ")
}
