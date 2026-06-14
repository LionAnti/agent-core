package agentcore

import (
	"context"
	"fmt"
	"strings"
)

type CompressorEngine struct {
	llm    LLMClient
	logger Logger
}

func NewCompressor(llm LLMClient, logger Logger) *CompressorEngine {
	return &CompressorEngine{llm: llm, logger: logger}
}

func (ce *CompressorEngine) ShouldCompress(ctx context.Context, state *HarnessState) *CompressionDecision {
	if state == nil {
		return &CompressionDecision{ShouldCompress: false}
	}
	ratio := float64(state.CurrentTokens) / float64(state.ContextWindow)
	if ratio >= 0.85 {
		return &CompressionDecision{
			ShouldCompress: true, Level: CompressionAggressive,
			Strategy: CompressionSlidingWindow, Reason: "context nearly full", CurrentRatio: ratio,
		}
	}
	if ratio >= 0.50 {
		if state.PendingToolPairs >= 4 && strings.Contains(strings.ToLower(state.TaskHint), "long") {
			return &CompressionDecision{
				ShouldCompress: true, Level: CompressionMild,
				Strategy: CompressionMermaid, Reason: "mermaid compression", CurrentRatio: ratio,
			}
		}
		return &CompressionDecision{
			ShouldCompress: true, Level: CompressionMild,
			Strategy: CompressionSummary, Reason: "summary compression", CurrentRatio: ratio,
		}
	}
	return &CompressionDecision{ShouldCompress: false, CurrentRatio: ratio}
}

func (ce *CompressorEngine) Compress(ctx context.Context, msgs []Message, decision *CompressionDecision) (*CompressionResult, error) {
	if decision == nil {
		return nil, fmt.Errorf("nil compression decision")
	}
	switch decision.Strategy {
	case CompressionSummary:
		return ce.summaryCompress(ctx, msgs)
	case CompressionMermaid:
		return ce.mermaidCompress(ctx, msgs)
	case CompressionSlidingWindow:
		return ce.slidingCompress(msgs)
	default:
		return ce.slidingCompress(msgs)
	}
}

func (ce *CompressorEngine) summaryCompress(ctx context.Context, msgs []Message) (*CompressionResult, error) {
	if len(msgs) < 2 {
		return nil, fmt.Errorf("need at least 2 messages to compress, got %d", len(msgs))
	}
	n := len(msgs)
	compressEnd := n * 60 / 100
	if compressEnd < 1 {
		compressEnd = 1
	}
	if compressEnd >= n {
		compressEnd = n - 1
	}

	var sb strings.Builder
	for i := 0; i < compressEnd; i++ {
		sb.WriteString(fmt.Sprintf("%s: %s\n", msgs[i].Role, msgs[i].Content))
	}
	text := sb.String()
	origTokens := estimateTokens(text)

	resp, err := ce.llm.Chat(ctx, &ChatRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a compression expert. Summarize preserving ALL key facts."},
			{Role: "user", Content: text},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("summary failed: %w", err)
	}
	totalTokens := resp.Usage.TotalTokens
	if totalTokens <= 0 {
		totalTokens = estimateTokens(resp.Content)
	}
	savings := origTokens - totalTokens
	if savings < 0 {
		savings = 0
	}
	return &CompressionResult{
		Strategy: CompressionSummary, TokenSavings: savings,
		NewTokenTotal: totalTokens,
		Content:       fmt.Sprintf("[Summary]:\n%s\n[End]", resp.Content),
		ReplaceFrom: 0, ReplaceTo: compressEnd,
	}, nil
}

func (ce *CompressorEngine) mermaidCompress(ctx context.Context, msgs []Message) (*CompressionResult, error) {
	if len(msgs) < 2 {
		return nil, fmt.Errorf("need at least 2 messages to compress, got %d", len(msgs))
	}
	n := len(msgs)
	compressEnd := n * 60 / 100
	if compressEnd < 1 {
		compressEnd = 1
	}
	if compressEnd >= n {
		compressEnd = n - 1
	}

	var sb strings.Builder
	for i := 0; i < compressEnd; i++ {
		content := msgs[i].Content
		if len(content) > 500 {
			content = content[:500] + "..."
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n", msgs[i].Role, content))
	}
	text := sb.String()
	origTokens := estimateTokens(text)

	resp, err := ce.llm.Chat(ctx, &ChatRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a technical architect. Convert conversations into mermaid mindmap diagrams."},
			{Role: "user", Content: fmt.Sprintf("Convert this conversation into a mermaid mindmap:\n%s", text)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("mermaid failed: %w", err)
	}
	totalTokens := resp.Usage.TotalTokens
	if totalTokens <= 0 {
		totalTokens = estimateTokens(resp.Content)
	}
	savings := origTokens - totalTokens
	if savings < 0 {
		savings = 0
	}
	return &CompressionResult{
		Strategy: CompressionMermaid, TokenSavings: savings,
		NewTokenTotal: totalTokens,
		Content:       fmt.Sprintf("[Mermaid]:\n%s\n[End]", resp.Content),
		ReplaceFrom: 0, ReplaceTo: compressEnd,
	}, nil
}

func (ce *CompressorEngine) slidingCompress(msgs []Message) (*CompressionResult, error) {
	if len(msgs) < 2 {
		return nil, fmt.Errorf("need at least 2 messages to compress, got %d", len(msgs))
	}
	keepRecent := 20
	if keepRecent >= len(msgs) {
		keepRecent = 1
	}
	truncated := msgs[:len(msgs)-keepRecent]
	kept := msgs[len(msgs)-keepRecent:]

	origTokens := 0
	for _, m := range msgs {
		if m.TokenCount > 0 {
			origTokens += m.TokenCount
		} else {
			origTokens += len(m.Content) / 4
		}
	}
	newTokens := 0
	for _, m := range kept {
		if m.TokenCount > 0 {
			newTokens += m.TokenCount
		} else {
			newTokens += len(m.Content) / 4
		}
	}
	savings := origTokens - newTokens
	if savings < 0 {
		savings = 0
	}
	var sb strings.Builder
	for _, m := range truncated {
		sb.WriteString(fmt.Sprintf("[Truncated %s message]\n", m.Role))
	}
	content := sb.String()
	return &CompressionResult{
		Strategy: CompressionSlidingWindow, TokenSavings: savings,
		NewTokenTotal: newTokens, Content: content,
		ReplaceFrom: 0, ReplaceTo: len(truncated),
	}, nil
}
