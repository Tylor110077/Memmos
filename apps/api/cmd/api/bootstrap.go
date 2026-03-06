package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"time"

	appchat "github.com/tylor/goaipj/internal/app/chat"
	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	appjob "github.com/tylor/goaipj/internal/app/job"
	appresource "github.com/tylor/goaipj/internal/app/resource"
	infraevent "github.com/tylor/goaipj/internal/infra/event"
	"github.com/tylor/goaipj/internal/infra/monitoring"
	"github.com/tylor/goaipj/internal/infra/storage"
	"github.com/tylor/goaipj/internal/http/api"
	pipelinegraph "github.com/tylor/goaipj/internal/pipeline/graph"
)

func newDevelopmentServer(logger *log.Logger) http.Handler {
	groupService := appgroup.NewService(appgroup.NewInMemoryRepository())
	resourceService := appresource.NewService(
		groupService,
		appresource.NewInMemoryRepository(),
		appresource.NewInMemoryArtifactRepository(),
		appresource.NewInMemoryJobPublisher(),
		storage.NewStore(newMemoryBucketClient(), "dev-bucket"),
	)
	graphService := appgraph.NewService(appgraph.NewInMemoryRepository())
	chatService := appchat.NewService(
		groupService,
		graphService,
		appchat.NewInMemoryChunkRepository(),
		appchat.NewInMemoryConversationRepository(),
		nil,
		nil,
	)

	return api.NewServer(api.Dependencies{
		ChatService:        chatService,
		EventBroker:        infraevent.NewInMemoryBroker(),
		ExpansionGenerator: pipelinegraph.NewExpansionGenerator(pipelinegraph.ExpansionRules{MaxNewNodes: 1}),
		FrameworkGenerator: pipelinegraph.NewFrameworkGenerator(pipelinegraph.FrameworkRules{MaxTopNodes: 8, MaxDepth: 3}),
		GroupService:       groupService,
		GraphService:       graphService,
		JobService:         appjob.NewService(appjob.NewInMemoryRepository()),
		Logger:             logger,
		Metrics:            monitoring.NewMetrics(),
		ResourceGraphGenerator: pipelinegraph.NewGenerator(),
		ResourceService:    resourceService,
	})
}

type memoryBucketClient struct {
	objects map[string][]byte
}

func newMemoryBucketClient() *memoryBucketClient {
	return &memoryBucketClient{objects: map[string][]byte{}}
}

func (c *memoryBucketClient) PutObject(_ context.Context, bucket, key string, reader io.Reader, _ int64, opts storage.PutObjectOptions) (storage.UploadInfo, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return storage.UploadInfo{}, err
	}
	c.objects[bucket+"/"+key] = raw
	return storage.UploadInfo{
		Key:         key,
		Size:        int64(len(raw)),
		ContentType: opts.ContentType,
	}, nil
}

func (c *memoryBucketClient) GetObject(_ context.Context, bucket, key string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(c.objects[bucket+"/"+key])), nil
}

func (c *memoryBucketClient) DeleteObject(_ context.Context, bucket, key string) error {
	delete(c.objects, bucket+"/"+key)
	return nil
}

func (c *memoryBucketClient) PresignPutObject(_ context.Context, bucket, key string, _ time.Duration, opts storage.PutObjectOptions) (string, http.Header, error) {
	header := http.Header{}
	if opts.ContentType != "" {
		header.Set("Content-Type", opts.ContentType)
	}
	return "https://uploads.example.test/" + bucket + "/" + key, header, nil
}
