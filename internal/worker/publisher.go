package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

type PublishInput struct {
	TaskType   string
	QueueName  string
	ResourceID string
	Payload    map[string]any
}

type TaskInfo struct {
	ID            string
	Queue         string
	State         string
	MaxRetry      int
	NextProcessAt time.Time
}

type enqueueClient interface {
	Enqueue(ctx context.Context, taskType string, payload []byte, queue string) (TaskInfo, error)
}

type Publisher struct {
	client enqueueClient
}

func NewPublisher(client enqueueClient) *Publisher {
	return &Publisher{client: client}
}

func (p *Publisher) Publish(ctx context.Context, input PublishInput) (TaskInfo, error) {
	payload, err := json.Marshal(input.Payload)
	if err != nil {
		return TaskInfo{}, err
	}
	return p.client.Enqueue(ctx, input.TaskType, payload, input.QueueName)
}

type AsynqClient struct {
	client *asynq.Client
}

func NewAsynqClient(redisOpt asynq.RedisConnOpt) *AsynqClient {
	return &AsynqClient{client: asynq.NewClient(redisOpt)}
}

func (c *AsynqClient) Enqueue(_ context.Context, taskType string, payload []byte, queue string) (TaskInfo, error) {
	info, err := c.client.Enqueue(asynq.NewTask(taskType, payload), asynq.Queue(queue))
	if err != nil {
		return TaskInfo{}, err
	}
	return TaskInfo{
		ID:            info.ID,
		Queue:         info.Queue,
		State:         info.State.String(),
		MaxRetry:      info.MaxRetry,
		NextProcessAt: info.NextProcessAt,
	}, nil
}
