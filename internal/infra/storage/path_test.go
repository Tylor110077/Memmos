package storage

import "testing"

func TestNormalizeObjectName(t *testing.T) {
	got := normalizeObjectName(" Diagram Final 2026 .PDF ")
	if got != "Diagram_Final_2026_.PDF" {
		t.Fatalf("normalizeObjectName() = %q", got)
	}
}

func TestKindPath(t *testing.T) {
	tests := map[Kind]string{
		KindRaw:        "raw",
		KindSnapshot:   "artifacts/snapshot",
		KindExtracted:  "artifacts/extracted",
		KindNormalized: "artifacts/normalized",
		KindDebug:      "artifacts/debug",
	}

	for kind, want := range tests {
		if got := kind.pathSegment(); got != want {
			t.Fatalf("kind %s => %q, want %q", kind, got, want)
		}
	}
}
