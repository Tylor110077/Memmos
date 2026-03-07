package parse

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type WebResult struct {
	Title    string
	Text     string
	Markdown string
	HTML     string
}

type WebFetcher struct {
	client *http.Client
}

func NewWebFetcher() *WebFetcher {
	return &WebFetcher{
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (f *WebFetcher) Fetch(ctx context.Context, url string) (WebResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return WebResult{}, err
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return WebResult{}, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return WebResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return WebResult{}, fmt.Errorf("web fetch failed: status=%d", resp.StatusCode)
	}

	title, text := extractMainContent(raw)
	md := strings.TrimSpace("# " + title + "\n\n" + text)
	return WebResult{
		Title:    title,
		Text:     text,
		Markdown: md,
		HTML:     string(raw),
	}, nil
}

func extractMainContent(raw []byte) (string, string) {
	doc, err := html.Parse(bytes.NewReader(raw))
	if err != nil {
		return "", strings.TrimSpace(string(raw))
	}

	title := findFirstText(doc, "title")
	article := findNode(doc, "article")
	if article == nil {
		article = findNode(doc, "main")
	}
	if article == nil {
		article = findNode(doc, "body")
	}
	text := normalizeText(nodeText(article))
	if title == "" {
		title = firstNonEmptyLine(text)
	}
	return title, text
}

func findNode(node *html.Node, tag string) *html.Node {
	if node == nil {
		return nil
	}
	if node.Type == html.ElementNode && node.Data == tag {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findNode(child, tag); found != nil {
			return found
		}
	}
	return nil
}

func findFirstText(node *html.Node, tag string) string {
	found := findNode(node, tag)
	return normalizeText(nodeText(found))
}

func nodeText(node *html.Node) string {
	if node == nil {
		return ""
	}
	if node.Type == html.TextNode {
		return node.Data
	}
	var parts []string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		parts = append(parts, nodeText(child))
	}
	return strings.Join(parts, " ")
}

func normalizeText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func firstNonEmptyLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}
