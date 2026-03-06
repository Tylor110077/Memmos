package migrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPlanIncludesCoreMigrations(t *testing.T) {
	plan, err := LoadPlan(filepath.Join("..", "..", "..", "migrations"))
	if err != nil {
		t.Fatalf("LoadPlan() error = %v", err)
	}

	if len(plan.Migrations) == 0 {
		t.Fatalf("expected migrations")
	}

	first := plan.Migrations[0]
	if first.Version != "0001" {
		t.Fatalf("first version = %q, want 0001", first.Version)
	}
	if !strings.Contains(first.UpSQL, "create extension if not exists vector") {
		t.Fatalf("expected pgvector extension in up migration")
	}
	for _, tableName := range []string{
		"groups",
		"resources",
		"resource_artifacts",
		"resource_chunks",
		"graphs",
		"graph_nodes",
		"graph_edges",
		"node_examples",
		"conversations",
		"conversation_messages",
		"processing_jobs",
		"job_events",
	} {
		if !strings.Contains(first.UpSQL, "create table if not exists "+tableName) {
			t.Fatalf("expected table %q in migration", tableName)
		}
	}
}

func TestLoadPlanRejectsUnpairedMigrationFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "0001_bad.up.sql"), "select 1;")

	if _, err := LoadPlan(dir); err == nil {
		t.Fatalf("expected pairing error")
	}
}

func TestComposeTemplateDeclaresRequiredServices(t *testing.T) {
	content := mustReadFile(t, filepath.Join("..", "..", "..", "docker-compose.yml"))

	for _, fragment := range []string{
		"postgres:",
		"redis:",
		"minio:",
		"tika:",
		"5432:5432",
		"6379:6379",
		"9000:9000",
		"9998:9998",
	} {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected %q in docker-compose.yml", fragment)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	return string(raw)
}
