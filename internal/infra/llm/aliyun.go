package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pipelinechat "github.com/tylor/goaipj/internal/pipeline/chat"
)

const (
	defaultBaseURL        = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	defaultChatModel      = "qwen3.5-plus"
	defaultEmbeddingModel = "text-embedding-v4"
)

type Client struct {
	apiKey         string
	baseURL        string
	chatModel      string
	embeddingModel string
	httpClient     *http.Client
}

type Options struct {
	APIKey         string
	BaseURL        string
	ChatModel      string
	EmbeddingModel string
	HTTPClient     *http.Client
}

func NewAliyunBailianClient(opts Options) *Client {
	baseURL := strings.TrimSpace(opts.BaseURL)
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	chatModel := strings.TrimSpace(opts.ChatModel)
	if chatModel == "" {
		chatModel = defaultChatModel
	}
	embeddingModel := strings.TrimSpace(opts.EmbeddingModel)
	if embeddingModel == "" {
		embeddingModel = defaultEmbeddingModel
	}
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 90 * time.Second}
	}
	return &Client{
		apiKey:         strings.TrimSpace(opts.APIKey),
		baseURL:        strings.TrimRight(baseURL, "/"),
		chatModel:      chatModel,
		embeddingModel: embeddingModel,
		httpClient:     httpClient,
	}
}

func (c *Client) Chat(ctx context.Context, messages []pipelinechat.CompletionMessage) (string, error) {
	type request struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		Stream bool `json:"stream"`
	}
	body := request{Model: c.chatModel, Stream: false}
	for _, message := range messages {
		body.Messages = append(body.Messages, struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{Role: message.Role, Content: message.Content})
	}
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/chat/completions", body, &response); err != nil {
		return "", err
	}
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("chat response missing choices")
	}
	return response.Choices[0].Message.Content, nil
}

func (c *Client) EmbedTexts(ctx context.Context, texts []string) ([][]float64, error) {
	type request struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}
	var response struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/embeddings", request{
		Model: c.embeddingModel,
		Input: texts,
	}, &response); err != nil {
		return nil, err
	}
	out := make([][]float64, 0, len(response.Data))
	for _, item := range response.Data {
		out = append(out, item.Embedding)
	}
	return out, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, requestBody any, responseBody any) error {
	if c.apiKey == "" {
		return fmt.Errorf("aliyun bailian api key is required")
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if err := json.Unmarshal(raw, responseBody); err != nil {
		return err
	}
	return nil
}
