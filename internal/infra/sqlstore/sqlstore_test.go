package sqlstore

import (
	"testing"

	"github.com/tylor/goaipj/internal/infra/sqlstore/gen"
)

func TestNewRepositoriesWiresGeneratedQueries(t *testing.T) {
	repos := NewRepositories(nil)
	if repos == nil {
		t.Fatalf("expected repositories")
	}
	if repos.Groups == nil {
		t.Fatalf("expected group repository")
	}
	if repos.Conversations == nil {
		t.Fatalf("expected conversation repository")
	}
	if repos.Resources == nil {
		t.Fatalf("expected resource repository")
	}
	if repos.Graphs == nil {
		t.Fatalf("expected graph repository")
	}
	if repos.Jobs == nil {
		t.Fatalf("expected job repository")
	}
}

func TestGeneratedQueriesTypeIsExposed(t *testing.T) {
	var _ *gen.Queries
}
