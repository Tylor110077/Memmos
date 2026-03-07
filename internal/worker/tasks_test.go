package worker

import (
	"testing"
	"time"
)

func TestDedupeKeyForResourceStageIncludesResourceStageAndVersion(t *testing.T) {
	key := DedupeKeyForResourceStage("resource-1", "normalize_resource", "v2")
	if key != "resource:resource-1:normalize_resource:v2" {
		t.Fatalf("key = %q", key)
	}
}

func TestDedupeKeyForFrameworkGraphSortsVersionSet(t *testing.T) {
	first := DedupeKeyForFrameworkGraph("group-1", []string{"resource-b@2", "resource-a@1"})
	second := DedupeKeyForFrameworkGraph("group-1", []string{"resource-a@1", "resource-b@2"})
	if first != second {
		t.Fatalf("framework dedupe key should be order-independent: %q != %q", first, second)
	}
}

func TestDedupeKeyForExpandNodeBucketsByWindow(t *testing.T) {
	now := time.Date(2026, 3, 7, 10, 3, 0, 0, time.UTC)
	first := DedupeKeyForExpandNode("graph-1", "node-1", now)
	second := DedupeKeyForExpandNode("graph-1", "node-1", now.Add(1*time.Minute))
	if first != second {
		t.Fatalf("expand dedupe key should match inside same window: %q != %q", first, second)
	}

	third := DedupeKeyForExpandNode("graph-1", "node-1", now.Add(5*time.Minute))
	if third == first {
		t.Fatalf("expand dedupe key should change across windows")
	}
}
