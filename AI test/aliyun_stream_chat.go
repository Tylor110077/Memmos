package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultBaseURL   = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	defaultModel     = "qwen3.5-plus"
	defaultSystemMsg = "You are a concise and helpful AI assistant."
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string    `json:"model"`
	Stream         bool      `json:"stream"`
	Messages       []Message `json:"messages"`
	EnableThinking *bool     `json:"enable_thinking,omitempty"`
	StreamOptions  any       `json:"stream_options,omitempty"`
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content          *string `json:"content"`
			ReasoningContent *string `json:"reasoning_content"`
			Role             string  `json:"role"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	apiKeyPath := filepath.Join(root, "docs", "api.md")
	historyPath := filepath.Join(root, "AI test", "chat-history-go.json")
	baseURL := envOr("ALIYUN_BAILIAN_BASE_URL", defaultBaseURL)
	model := envOr("ALIYUN_BAILIAN_MODEL", defaultModel)
	systemPrompt := envOr("ALIYUN_BAILIAN_SYSTEM_PROMPT", defaultSystemMsg)
	enableThinking := envBoolOr("ALIYUN_BAILIAN_ENABLE_THINKING", false)

	apiKey, err := readAPIKey(apiKeyPath)
	if err != nil {
		return err
	}

	history, err := loadHistory(historyPath, systemPrompt)
	if err != nil {
		return err
	}

	fmt.Println("Aliyun Bailian Go streaming chat ready.")
	fmt.Printf("Model: %s\n", model)
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("History file: %s\n", historyPath)
	fmt.Println("Commands: /history, /clear, /exit")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return err
			}
			fmt.Println()
			return nil
		}

		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		switch text {
		case "/exit":
			return nil
		case "/history":
			if err := printHistory(history); err != nil {
				return err
			}
			continue
		case "/clear":
			history = []Message{{Role: "system", Content: systemPrompt}}
			if err := saveHistory(historyPath, history); err != nil {
				return err
			}
			fmt.Println("History cleared.")
			fmt.Println()
			continue
		}

		history = append(history, Message{Role: "user", Content: text})
		fmt.Print("Assistant: ")

		reply, err := streamChat(os.Stdout, apiKey, baseURL, model, history, enableThinking)
		if err != nil {
			history = history[:len(history)-1]
			fmt.Printf("\n\nRequest error: %v\n\n", err)
			continue
		}

		fmt.Print("\n\n")
		history = append(history, Message{Role: "assistant", Content: reply})
		if err := saveHistory(historyPath, history); err != nil {
			return err
		}
	}
}

func envOr(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func envBoolOr(name string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func readAPIKey(path string) (string, error) {
	if key := strings.TrimSpace(os.Getenv("ALIYUN_BAILIAN_API_KEY")); key != "" {
		return key, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var key string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "<") {
			continue
		}
		key = line
		break
	}
	if key == "" {
		return "", fmt.Errorf("no API key found. Set ALIYUN_BAILIAN_API_KEY or update %s", path)
	}
	return key, nil
}

func loadHistory(path, systemPrompt string) ([]Message, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			initial := []Message{{Role: "system", Content: systemPrompt}}
			if err := saveHistory(path, initial); err != nil {
				return nil, err
			}
			return initial, nil
		}
		return nil, err
	}

	var history []Message
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, err
	}
	if len(history) == 0 {
		return []Message{{Role: "system", Content: systemPrompt}}, nil
	}
	return history, nil
}

func saveHistory(path string, history []Message) error {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func printHistory(history []Message) error {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	fmt.Println()
	return nil
}

func streamChat(stdout io.Writer, apiKey, baseURL, model string, history []Message, enableThinking bool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	reqBody := chatRequest{
		Model:          model,
		Stream:         true,
		Messages:       history,
		EnableThinking: &enableThinking,
		StreamOptions: map[string]bool{
			"include_usage": true,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	reader := bufio.NewReader(resp.Body)
	var reply strings.Builder
	var reasoningStarted bool
	var answerStarted bool

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}

		var chunk streamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return "", fmt.Errorf("decode stream chunk: %w", err)
		}
		if len(chunk.Choices) == 0 {
			continue
		}

		reasoning := chunk.Choices[0].Delta.ReasoningContent
		if reasoning != nil && *reasoning != "" {
			if !reasoningStarted {
				reasoningStarted = true
				if _, err := io.WriteString(stdout, "[thinking] "); err != nil {
					return "", err
				}
			}
			if _, err := io.WriteString(stdout, *reasoning); err != nil {
				return "", err
			}
		}

		content := chunk.Choices[0].Delta.Content
		if content != nil && *content != "" {
			if !answerStarted {
				answerStarted = true
				if reasoningStarted {
					if _, err := io.WriteString(stdout, "\n[answer] "); err != nil {
						return "", err
					}
				}
			}
			reply.WriteString(*content)
			if _, err := io.WriteString(stdout, *content); err != nil {
				return "", err
			}
		}
	}

	finalReply := strings.TrimSpace(reply.String())
	if finalReply == "" && reasoningStarted {
		return "(no final answer content returned)", nil
	}
	return finalReply, nil
}
