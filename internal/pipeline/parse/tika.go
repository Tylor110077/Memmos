package parse

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type TikaClient struct {
	baseURL string
	client  *http.Client
}

func NewTikaClient(baseURL string) *TikaClient {
	return &TikaClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *TikaClient) ExtractText(ctx context.Context, contentType string, body io.Reader) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+"/tika", body)
	if err != nil {
		return "", err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "text/plain")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("tika request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return string(raw), nil
}
