package agentcore

import (
	"context"
	"strings"
	"time"
)

type MemoryPipeline struct {
	store  MemoryStore
	logger Logger
}

func NewMemoryPipeline(store MemoryStore, logger Logger) *MemoryPipeline {
	return &MemoryPipeline{store: store, logger: logger}
}

func (p *MemoryPipeline) ExtractL1(ctx context.Context, records []L0Record) error {
	if len(records) < 3 {
		return nil
	}
	var sb strings.Builder
	for _, r := range records {
		sb.WriteString(r.Role + ": " + r.Content + "\n")
	}
	content := sb.String()
	if strings.TrimSpace(content) == "" {
		return nil
	}
	return p.store.SaveL1(ctx, &L1Memory{
		ID: newID(), Content: content, Type: "conversation",
		CreatedAt: time.Now().UnixMilli(),
	})
}

type DefaultMemoryRecall struct {
	store  MemoryStore
	vector VectorStore
	logger Logger
}

func NewDefaultMemoryRecall(store MemoryStore, vector VectorStore, logger Logger) *DefaultMemoryRecall {
	return &DefaultMemoryRecall{store: store, vector: vector, logger: logger}
}

func (mr *DefaultMemoryRecall) Recall(ctx context.Context, intent *IntentClassification, tenantID, userID string) (*RecallResult, error) {
	result := &RecallResult{}
	if intent == nil {
		return result, nil
	}
	if intent.Type == IntentRecall || intent.Strategy != "" {
		memories, err := mr.store.SearchL1(ctx, tenantID, userID, intent.Strategy, 5)
		if err == nil {
			result.Memories = memories
			for i := range memories {
				mr.store.IncrementL1Recall(ctx, memories[i].ID)
			}
		}
	}
	persona, err := mr.store.GetCurrentL3(ctx, tenantID, userID)
	if err == nil && persona != nil {
		result.Persona = persona
		result.AppendSystemContext = persona.Content
	}
	scenes, err := mr.store.GetL2ByUser(ctx, tenantID, userID)
	if err == nil && len(scenes) > 0 {
		result.Scene = scenes[0]
	}
	if len(result.Memories) > 0 {
		var sb strings.Builder
		sb.WriteString("Relevant past context:\n")
		for _, m := range result.Memories {
			sb.WriteString("- " + m.Content + "\n")
		}
		result.PrependContext = sb.String()
	}
	return result, nil
}
