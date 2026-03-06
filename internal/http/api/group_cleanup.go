package api

import (
	"context"

	appchat "github.com/tylor/goaipj/internal/app/chat"
	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appjob "github.com/tylor/goaipj/internal/app/job"
	appresource "github.com/tylor/goaipj/internal/app/resource"
)

type groupCleanup struct {
	resources *appresource.Service
	graphs    *appgraph.Service
	chat      *appchat.Service
	jobs      *appjob.Service
}

func (c groupCleanup) DeleteGroupResources(ctx context.Context, groupID string) error {
	if c.resources == nil {
		return nil
	}
	return c.resources.DeleteGroupResources(ctx, groupID)
}

func (c groupCleanup) DeleteGroupGraphs(ctx context.Context, groupID string) error {
	if c.graphs == nil {
		return nil
	}
	return c.graphs.DeleteGroupGraphs(ctx, groupID)
}

func (c groupCleanup) DeleteGroupConversations(ctx context.Context, groupID string) error {
	if c.chat == nil {
		return nil
	}
	return c.chat.DeleteGroupConversations(ctx, groupID)
}

func (c groupCleanup) DeleteGroupJobs(ctx context.Context, groupID string) error {
	if c.jobs == nil {
		return nil
	}
	return c.jobs.DeleteGroupJobs(ctx, groupID)
}
