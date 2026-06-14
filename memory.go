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

func extractTenantID(sessionKey string) string {
	parts := strings.SplitN(sessionKey, "/", 3)
	if len(parts) >= 1 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

func extractUserID(sessionKey string) string {
	parts := strings.SplitN(sessionKey, "/", 3)
	if len(parts) >= 2 && parts[1] != "" {
		return parts[1]
	}
	return ""
}

func (p *MemoryPipeline) ExtractL1(ctx context.Context, records []L0Record) error {
	if p == nil || p.store == nil {
		return nil
	}
	if len(records) < 3 {
		return nil
	}
	var sb strings.Builder
	for _, r := range records {
		sb.WriteString(r.Role)
		sb.WriteString(": ")
		sb.WriteString(r.Content)
		sb.WriteString("\n")
	}
	content := sb.String()
	if strings.TrimSpace(content) == "" {
		return nil
	}
	sessionKey := records[0].SessionKey
	return p.store.SaveL1(ctx, &L1Memory{
		ID:        newID(),
		TenantID:  extractTenantID(sessionKey),
		UserID:    extractUserID(sessionKey),
		Content:   content,
		Type:      "conversation",
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
	if mr == nil {
		return &RecallResult{}, nil
	}
	result := &RecallResult{}
	if mr.store == nil {
		return result, nil
	}
	if intent == nil {
		return result, nil
	}
	// Only trigger recall search when intent explicitly asks for recall
	if intent.Type == IntentRecall {
		memories, err := mr.store.SearchL1(ctx, tenantID, userID, intent.Strategy, 5)
		if err == nil {
			result.Memories = memories
			for i := range memories {
				if incErr := mr.store.IncrementL1Recall(ctx, memories[i].ID); incErr != nil {
					mr.logger.Warn("failed to increment L1 recall", "id", memories[i].ID, "error", incErr)
				}
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
		sb.WriteString("Relevant past context:" + "\n")
		for _, m := range result.Memories {
			sb.WriteString("- ")
			sb.WriteString(m.Content)
			sb.WriteString("\n")
		}
		result.PrependContext = sb.String()
	}
	return result, nil
}
