package chat

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"hash/fnv"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	appgraph "github.com/tylor/goaipj/internal/app/graph"
	appgroup "github.com/tylor/goaipj/internal/app/group"
	"github.com/tylor/goaipj/internal/infra/apperror"
	pipelinechat "github.com/tylor/goaipj/internal/pipeline/chat"
)

const embeddingDimensions = 32

type Chunk struct {
	ID         string    `json:"id"`
	GroupID    string    `json:"group_id"`
	ResourceID string    `json:"resource_id"`
	ChunkIndex int       `json:"chunk_index"`
	Content    string    `json:"content"`
	Summary    string    `json:"summary"`
	Embedding  []float64 `json:"embedding"`
	CreatedAt  time.Time `json:"created_at"`
}

type RetrievedChunk struct {
	Chunk Chunk   `json:"chunk"`
	Score float64 `json:"score"`
}

type ContextSnapshot struct {
	GraphID           string   `json:"graph_id"`
	CurrentNodeID     string   `json:"current_node_id"`
	ResourceSummary   string   `json:"resource_summary"`
	NeighborNodeIDs   []string `json:"neighbor_node_ids"`
	RetrievedChunkIDs []string `json:"retrieved_chunk_ids"`
	RecentMessageIDs  []string `json:"recent_message_ids"`
}

type Message struct {
	ID              string          `json:"id"`
	ConversationID  string          `json:"conversation_id"`
	Role            string          `json:"role"`
	Content         string          `json:"content"`
	Examples        []string        `json:"examples"`
	CitedChunkIDs   []string        `json:"cited_chunk_ids"`
	CitedNodeIDs    []string        `json:"cited_node_ids"`
	ContextSnapshot ContextSnapshot `json:"context_snapshot"`
	CreatedAt       time.Time       `json:"created_at"`
}

type Conversation struct {
	ID            string    `json:"id"`
	GroupID       string    `json:"group_id"`
	GraphID       string    `json:"graph_id"`
	CurrentNodeID string    `json:"current_node_id"`
	Title         string    `json:"title"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Messages      []Message `json:"messages"`
}

type IndexChunksInput struct {
	GroupID    string
	ResourceID string
	Chunks     []string
}

type CreateConversationInput struct {
	GroupID       string
	GraphID       string
	CurrentNodeID string
	Title         string
}

type AskInput struct {
	ConversationID string
	Content        string
}

type AskResult struct {
	Conversation     Conversation `json:"conversation"`
	UserMessage      Message      `json:"user_message"`
	AssistantMessage Message      `json:"assistant_message"`
}

type GroupLookup interface {
	GetGroup(ctx context.Context, id string) (appgroup.Group, error)
}

type GraphLookup interface {
	GetGraph(ctx context.Context, graphID string) (appgraph.SavedGraph, error)
	GetNodeDetail(ctx context.Context, graphID, nodeID string) (appgraph.NodeDetail, error)
}

type ChunkRepository interface {
	ReplaceForResource(ctx context.Context, groupID, resourceID string, chunks []Chunk) error
	ListByGroup(ctx context.Context, groupID string) ([]Chunk, error)
}

type ConversationRepository interface {
	CreateConversation(ctx context.Context, conversation Conversation) error
	GetConversation(ctx context.Context, conversationID string) (Conversation, error)
	AppendMessage(ctx context.Context, message Message) error
}

type Answerer interface {
	Answer(input pipelinechat.Input) (pipelinechat.Output, error)
}

type Embedder interface {
	EmbedTexts(ctx context.Context, texts []string) ([][]float64, error)
}

type Service struct {
	groups        GroupLookup
	graphs        GraphLookup
	chunks        ChunkRepository
	conversations ConversationRepository
	answerer      Answerer
	embedder      Embedder
}

func NewService(groups GroupLookup, graphs GraphLookup, chunks ChunkRepository, conversations ConversationRepository, answerer Answerer, embedder Embedder) *Service {
	if answerer == nil {
		answerer = pipelinechat.NewAnswerer()
	}
	if embedder == nil {
		embedder = deterministicEmbedder{}
	}
	return &Service{
		groups:        groups,
		graphs:        graphs,
		chunks:        chunks,
		conversations: conversations,
		answerer:      answerer,
		embedder:      embedder,
	}
}

func (s *Service) IndexResourceChunks(ctx context.Context, input IndexChunksInput) ([]Chunk, error) {
	if _, err := s.groups.GetGroup(ctx, input.GroupID); err != nil {
		return nil, apperror.New(apperror.CodeNotFound, "group not found")
	}
	if strings.TrimSpace(input.ResourceID) == "" {
		return nil, apperror.New(apperror.CodeInvalidArgument, "resource_id is required")
	}
	if len(input.Chunks) == 0 {
		return nil, apperror.New(apperror.CodeInvalidArgument, "chunks are required")
	}

	now := time.Now().UTC()
	indexed := make([]Chunk, 0, len(input.Chunks))
	texts := make([]string, 0, len(input.Chunks))
	for idx, content := range input.Chunks {
		trimmed := strings.TrimSpace(content)
		if trimmed == "" {
			continue
		}
		texts = append(texts, trimmed)
		indexed = append(indexed, Chunk{
			ID:         newID(),
			GroupID:    input.GroupID,
			ResourceID: input.ResourceID,
			ChunkIndex: idx,
			Content:    trimmed,
			Summary:    summarize(trimmed),
			CreatedAt:  now,
		})
	}
	if len(indexed) == 0 {
		return nil, apperror.New(apperror.CodeInvalidArgument, "chunks are required")
	}
	embeddings, err := s.embedder.EmbedTexts(ctx, texts)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "embed chunks", err)
	}
	if len(embeddings) != len(indexed) {
		return nil, apperror.New(apperror.CodeInternal, "embedding count mismatch")
	}
	for i := range indexed {
		indexed[i].Embedding = embeddings[i]
	}
	if err := s.chunks.ReplaceForResource(ctx, input.GroupID, input.ResourceID, indexed); err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "index resource chunks", err)
	}
	return indexed, nil
}

func (s *Service) SearchGroupContext(ctx context.Context, groupID, query string, topK int) ([]RetrievedChunk, error) {
	if _, err := s.groups.GetGroup(ctx, groupID); err != nil {
		return nil, apperror.New(apperror.CodeNotFound, "group not found")
	}
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, apperror.New(apperror.CodeInvalidArgument, "query is required")
	}
	if topK <= 0 {
		topK = 8
	}

	items, err := s.chunks.ListByGroup(ctx, groupID)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "list chunks", err)
	}
	queryVectors, err := s.embedder.EmbedTexts(ctx, []string{trimmed})
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "embed query", err)
	}
	if len(queryVectors) != 1 {
		return nil, apperror.New(apperror.CodeInternal, "query embedding missing")
	}
	queryVector := queryVectors[0]
	results := make([]RetrievedChunk, 0, len(items))
	for _, item := range items {
		results = append(results, RetrievedChunk{
			Chunk: item,
			Score: cosineSimilarity(queryVector, item.Embedding),
		})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			if results[i].Chunk.ResourceID == results[j].Chunk.ResourceID {
				return results[i].Chunk.ChunkIndex < results[j].Chunk.ChunkIndex
			}
			return results[i].Chunk.ResourceID < results[j].Chunk.ResourceID
		}
		return results[i].Score > results[j].Score
	})
	if len(results) > topK {
		results = results[:topK]
	}
	return results, nil
}

func (s *Service) CreateConversation(ctx context.Context, input CreateConversationInput) (Conversation, error) {
	if _, err := s.groups.GetGroup(ctx, input.GroupID); err != nil {
		return Conversation{}, apperror.New(apperror.CodeNotFound, "group not found")
	}
	if strings.TrimSpace(input.GraphID) == "" || strings.TrimSpace(input.CurrentNodeID) == "" {
		return Conversation{}, apperror.New(apperror.CodeInvalidArgument, "graph_id and current_node_id are required")
	}
	graph, err := s.graphs.GetGraph(ctx, input.GraphID)
	if err != nil {
		return Conversation{}, apperror.New(apperror.CodeNotFound, "graph not found")
	}
	if graph.Graph.GroupID != input.GroupID {
		return Conversation{}, apperror.New(apperror.CodeInvalidArgument, "graph does not belong to group")
	}
	detail, err := s.graphs.GetNodeDetail(ctx, input.GraphID, input.CurrentNodeID)
	if err != nil {
		return Conversation{}, apperror.New(apperror.CodeNotFound, "node not found")
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "Ask about " + detail.Node.Name
	}

	now := time.Now().UTC()
	conversation := Conversation{
		ID:            newID(),
		GroupID:       input.GroupID,
		GraphID:       input.GraphID,
		CurrentNodeID: input.CurrentNodeID,
		Title:         title,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.conversations.CreateConversation(ctx, conversation); err != nil {
		return Conversation{}, apperror.Wrap(apperror.CodeInternal, "create conversation", err)
	}
	return conversation, nil
}

func (s *Service) GetConversation(ctx context.Context, conversationID string) (Conversation, error) {
	conversation, err := s.conversations.GetConversation(ctx, conversationID)
	if err != nil {
		return Conversation{}, apperror.New(apperror.CodeNotFound, "conversation not found")
	}
	return conversation, nil
}

func (s *Service) Ask(ctx context.Context, input AskInput) (AskResult, error) {
	question := strings.TrimSpace(input.Content)
	if question == "" {
		return AskResult{}, apperror.New(apperror.CodeInvalidArgument, "content is required")
	}
	conversation, err := s.conversations.GetConversation(ctx, input.ConversationID)
	if err != nil {
		return AskResult{}, apperror.New(apperror.CodeNotFound, "conversation not found")
	}
	graph, err := s.graphs.GetGraph(ctx, conversation.GraphID)
	if err != nil {
		return AskResult{}, apperror.New(apperror.CodeNotFound, "graph not found")
	}
	detail, err := s.graphs.GetNodeDetail(ctx, conversation.GraphID, conversation.CurrentNodeID)
	if err != nil {
		return AskResult{}, apperror.New(apperror.CodeNotFound, "node not found")
	}
	contextChunks, err := s.SearchGroupContext(ctx, conversation.GroupID, question, 8)
	if err != nil {
		return AskResult{}, err
	}

	now := time.Now().UTC()
	userMessage := Message{
		ID:             newID(),
		ConversationID: conversation.ID,
		Role:           "user",
		Content:        question,
		CreatedAt:      now,
	}
	if err := s.conversations.AppendMessage(ctx, userMessage); err != nil {
		return AskResult{}, apperror.Wrap(apperror.CodeInternal, "append user message", err)
	}

	recent := recentMessages(conversation.Messages, 6)
	answerInput := pipelinechat.Input{
		Question: question,
		CurrentNode: pipelinechat.NodeContext{
			ID:          detail.Node.ID,
			Name:        detail.Node.Name,
			Description: detail.Node.Description,
			Meaning:     detail.Node.Meaning,
		},
		ResourceSummary: graph.Graph.Summary,
		Examples:        exampleTexts(detail.Examples),
		Neighbors:       toPipelineNeighbors(detail.Neighbors),
		Chunks:          toPipelineChunks(contextChunks),
		RecentMessages:  toPipelineMessages(recent),
	}
	answer, err := s.answerer.Answer(answerInput)
	if err != nil {
		return AskResult{}, apperror.Wrap(apperror.CodeInternal, "generate answer", err)
	}

	assistantMessage := Message{
		ID:             newID(),
		ConversationID: conversation.ID,
		Role:           "assistant",
		Content:        answer.Answer,
		Examples:       answer.Examples,
		CitedChunkIDs:  append([]string(nil), answer.CitedChunkIDs...),
		CitedNodeIDs:   append([]string(nil), answer.CitedNodeIDs...),
		ContextSnapshot: ContextSnapshot{
			GraphID:           conversation.GraphID,
			CurrentNodeID:     conversation.CurrentNodeID,
			ResourceSummary:   graph.Graph.Summary,
			NeighborNodeIDs:   neighborIDs(detail.Neighbors),
			RetrievedChunkIDs: retrievedChunkIDs(contextChunks),
			RecentMessageIDs:  messageIDs(recent),
		},
		CreatedAt: now,
	}
	if err := s.conversations.AppendMessage(ctx, assistantMessage); err != nil {
		return AskResult{}, apperror.Wrap(apperror.CodeInternal, "append assistant message", err)
	}

	updated, err := s.conversations.GetConversation(ctx, conversation.ID)
	if err != nil {
		return AskResult{}, apperror.Wrap(apperror.CodeInternal, "reload conversation", err)
	}
	return AskResult{
		Conversation:     updated,
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
	}, nil
}

type InMemoryChunkRepository struct {
	mu     sync.RWMutex
	chunks []Chunk
}

func NewInMemoryChunkRepository() *InMemoryChunkRepository {
	return &InMemoryChunkRepository{}
}

func (r *InMemoryChunkRepository) ReplaceForResource(_ context.Context, groupID, resourceID string, chunks []Chunk) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	filtered := r.chunks[:0]
	for _, item := range r.chunks {
		if item.GroupID == groupID && item.ResourceID == resourceID {
			continue
		}
		filtered = append(filtered, item)
	}
	r.chunks = append(filtered, chunks...)
	return nil
}

func (r *InMemoryChunkRepository) ListByGroup(_ context.Context, groupID string) ([]Chunk, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]Chunk, 0, len(r.chunks))
	for _, item := range r.chunks {
		if item.GroupID == groupID {
			items = append(items, item)
		}
	}
	return items, nil
}

type InMemoryConversationRepository struct {
	mu            sync.RWMutex
	conversations map[string]Conversation
}

func NewInMemoryConversationRepository() *InMemoryConversationRepository {
	return &InMemoryConversationRepository{
		conversations: map[string]Conversation{},
	}
}

func (r *InMemoryConversationRepository) CreateConversation(_ context.Context, conversation Conversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conversations[conversation.ID] = conversation
	return nil
}

func (r *InMemoryConversationRepository) GetConversation(_ context.Context, conversationID string) (Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	conversation, ok := r.conversations[conversationID]
	if !ok {
		return Conversation{}, errors.New("conversation not found")
	}
	conversation.Messages = append([]Message(nil), conversation.Messages...)
	return conversation, nil
}

func (r *InMemoryConversationRepository) AppendMessage(_ context.Context, message Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	conversation, ok := r.conversations[message.ConversationID]
	if !ok {
		return errors.New("conversation not found")
	}
	conversation.Messages = append(conversation.Messages, message)
	conversation.UpdatedAt = message.CreatedAt
	r.conversations[message.ConversationID] = conversation
	return nil
}

func toPipelineNeighbors(items []appgraph.Neighbor) []pipelinechat.NeighborContext {
	out := make([]pipelinechat.NeighborContext, 0, len(items))
	for _, item := range items {
		out = append(out, pipelinechat.NeighborContext{
			ID:       item.Node.ID,
			Name:     item.Node.Name,
			Relation: item.Relation,
		})
	}
	return out
}

func toPipelineChunks(items []RetrievedChunk) []pipelinechat.ChunkContext {
	out := make([]pipelinechat.ChunkContext, 0, len(items))
	for _, item := range items {
		out = append(out, pipelinechat.ChunkContext{
			ID:      item.Chunk.ID,
			Content: item.Chunk.Content,
			Summary: item.Chunk.Summary,
		})
	}
	return out
}

func toPipelineMessages(items []Message) []pipelinechat.MessageContext {
	out := make([]pipelinechat.MessageContext, 0, len(items))
	for _, item := range items {
		out = append(out, pipelinechat.MessageContext{
			Role:    item.Role,
			Content: item.Content,
		})
	}
	return out
}

func exampleTexts(items []appgraph.Example) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if text := strings.TrimSpace(item.Content); text != "" {
			out = append(out, text)
		}
	}
	return out
}

func neighborIDs(items []appgraph.Neighbor) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Node.ID)
	}
	return out
}

func retrievedChunkIDs(items []RetrievedChunk) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Chunk.ID)
	}
	return out
}

func recentMessages(messages []Message, limit int) []Message {
	if len(messages) <= limit {
		return append([]Message(nil), messages...)
	}
	return append([]Message(nil), messages[len(messages)-limit:]...)
}

func messageIDs(messages []Message) []string {
	out := make([]string, 0, len(messages))
	for _, item := range messages {
		out = append(out, item.ID)
	}
	return out
}

func summarize(text string) string {
	words := strings.Fields(text)
	if len(words) <= 18 {
		return strings.Join(words, " ")
	}
	return strings.Join(words[:18], " ") + "..."
}

type deterministicEmbedder struct{}

func (deterministicEmbedder) EmbedTexts(_ context.Context, texts []string) ([][]float64, error) {
	out := make([][]float64, 0, len(texts))
	for _, text := range texts {
		out = append(out, embedText(text))
	}
	return out, nil
}

func embedText(text string) []float64 {
	vector := make([]float64, embeddingDimensions)
	for _, token := range tokenize(text) {
		hasher := fnv.New32a()
		_, _ = hasher.Write([]byte(token))
		index := int(hasher.Sum32() % embeddingDimensions)
		vector[index]++
	}
	return vector
}

func tokenize(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r >= 0x4e00 && r <= 0x9fff)
	})
}

func cosineSimilarity(left, right []float64) float64 {
	var dot float64
	var leftNorm float64
	var rightNorm float64
	for i := 0; i < len(left) && i < len(right); i++ {
		dot += left[i] * right[i]
		leftNorm += left[i] * left[i]
		rightNorm += right[i] * right[i]
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm))
}

func newID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(raw[:])
}
