package normalize

import (
	"strings"
)

type ChunkConfig struct {
	TargetSize int
	Overlap    int
}

type NormalizeInput struct {
	Title string
	Text  string
}

type Result struct {
	Markdown string
	Summary  string
	Chunks   []string
}

type Service struct {
	config ChunkConfig
}

func NewService(config ChunkConfig) *Service {
	if config.TargetSize <= 0 {
		config.TargetSize = 800
	}
	if config.Overlap < 0 {
		config.Overlap = 0
	}
	return &Service{config: config}
}

func (s *Service) Normalize(input NormalizeInput) Result {
	text := strings.TrimSpace(input.Text)
	title := strings.TrimSpace(input.Title)
	var markdown string
	if title != "" {
		markdown = "# " + title + "\n\n" + text
	} else {
		markdown = text
	}
	return Result{
		Markdown: markdown,
		Summary:  summarize(text),
		Chunks:   chunkText(text, s.config.TargetSize, s.config.Overlap),
	}
}

func summarize(text string) string {
	words := strings.Fields(text)
	if len(words) <= 24 {
		return strings.Join(words, " ")
	}
	return strings.Join(words[:24], " ") + "..."
}

func chunkText(text string, targetSize, overlap int) []string {
	if text == "" {
		return nil
	}
	runes := []rune(text)
	if len(runes) <= targetSize {
		return []string{text}
	}

	step := targetSize - overlap
	if step <= 0 {
		step = targetSize
	}

	var chunks []string
	for start := 0; start < len(runes); start += step {
		end := start + targetSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, strings.TrimSpace(string(runes[start:end])))
		if end == len(runes) {
			break
		}
	}
	return chunks
}
