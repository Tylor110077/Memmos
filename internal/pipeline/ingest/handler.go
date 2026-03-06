package ingest

import (
	"context"
	"io"

	"github.com/tylor/goaipj/internal/pipeline/normalize"
)

type Parser interface {
	ExtractText(ctx context.Context, contentType string, body io.Reader) (string, error)
}

type Handler struct {
	parser     Parser
	normalizer *normalize.Service
}

type Input struct {
	Title       string
	ContentType string
	Body        io.Reader
}

type Result struct {
	ExtractedText string
	Normalized    normalize.Result
}

func NewHandler(parser Parser, normalizer *normalize.Service) *Handler {
	return &Handler{
		parser:     parser,
		normalizer: normalizer,
	}
}

func (h *Handler) Process(ctx context.Context, input Input) (Result, error) {
	text, err := h.parser.ExtractText(ctx, input.ContentType, input.Body)
	if err != nil {
		return Result{}, err
	}
	normalized := h.normalizer.Normalize(normalize.NormalizeInput{
		Title: input.Title,
		Text:  text,
	})
	return Result{
		ExtractedText: text,
		Normalized:    normalized,
	}, nil
}
