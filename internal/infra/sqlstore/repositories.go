package sqlstore

import (
	"database/sql"

	"github.com/tylor/goaipj/internal/infra/sqlstore/gen"
)

type Repositories struct {
	Queries       *gen.Queries
	Conversations *ConversationRepository
	Groups        *GroupRepository
	Resources     *ResourceRepository
	Graphs        *GraphRepository
	Jobs          *JobRepository
}

type GroupRepository struct {
	queries *gen.Queries
}

type ResourceRepository struct {
	queries *gen.Queries
}

type GraphRepository struct {
	queries *gen.Queries
}

type JobRepository struct {
	queries *gen.Queries
}

type ConversationRepository struct {
	queries *gen.Queries
}

func NewRepositories(db *sql.DB) *Repositories {
	queries := gen.New(db)
	return &Repositories{
		Queries:       queries,
		Conversations: &ConversationRepository{queries: queries},
		Groups:        &GroupRepository{queries: queries},
		Resources:     &ResourceRepository{queries: queries},
		Graphs:        &GraphRepository{queries: queries},
		Jobs:          &JobRepository{queries: queries},
	}
}
