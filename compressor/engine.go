package compressor

import (
	"context"
	"fmt"
	"strings"

	"github.com/LionAnti/agent-core/types"
)

type Engine struct {
	llm    types.LLMClient
	logger types.Logger
}

func NewEngine(llm types.LLMClient, logger types.Logger) *Engine {
	return &Engine{llm: llm, logger: logger}
}

func (eng *Engine) ShouldCompress(ctx context.Context, state *types.HarnessState) *types.CompressionDecision {
	if state == nil { return &types.CompressionDecision{ShouldCompress: false} }
	r := float64(state.CurrentTokens) / float64(state.ContextWindow)
	if r >= 0.85 { return &types.CompressionDecision{ShouldCompress: true, Level: types.CompressionAggressive, Strategy: types.CompressionSlidingWindow, Reason: "context full", CurrentRatio: r} }
	if r >= 0.50 {
		if state.PendingToolPairs >= 4 && strings.Contains(strings.ToLower(state.TaskHint), "long") {
			return &types.CompressionDecision{ShouldCompress: true, Level: types.CompressionMild, Strategy: types.CompressionMermaid, Reason: "mermaid", CurrentRatio: r}
		}
		return &types.CompressionDecision{ShouldCompress: true, Level: types.CompressionMild, Strategy: types.CompressionSummary, Reason: "summary", CurrentRatio: r}
	}
	return &types.CompressionDecision{ShouldCompress: false, CurrentRatio: r}
}

func (eng *Engine) Compress(ctx context.Context, msgs []types.Message, d *types.CompressionDecision) (*types.CompressionResult, error) {
	if d == nil { return nil, fmt.Errorf("nil decision") }
	switch d.Strategy {
	case types.CompressionSummary: return eng.summary(ctx, msgs)
	case types.CompressionMermaid: return eng.mermaid(ctx, msgs)
	case types.CompressionSlidingWindow: return eng.sliding(msgs)
	default: return eng.sliding(msgs)
	}
}

func (eng *Engine) summary(ctx context.Context, msgs []types.Message) (*types.CompressionResult, error) {
	if len(msgs) < 2 { return nil, fmt.Errorf("need >=2 msgs") }
	n := len(msgs); end := n * 60 / 100
	if end < 1 { end = 1 }; if end >= n { end = n - 1 }
	var sb strings.Builder
	for i := 0; i < end; i++ {
		sb.WriteString(msgs[i].Role); sb.WriteString(": "); sb.WriteString(msgs[i].Content); sb.WriteString("\n")
	}
	t := sb.String(); ot := len(t) / 4
	resp, err := eng.llm.Chat(ctx, &types.ChatRequest{Messages: []types.ChatMessage{
		{Role: "system", Content: "You are a compression expert. Summarize preserving ALL key facts."},
		{Role: "user", Content: t},
	}})
	if err != nil { return nil, fmt.Errorf("summary: %w", err) }
	tt := resp.Usage.TotalTokens; if tt <= 0 { tt = len(resp.Content) / 4 }
	s := ot - tt; if s < 0 { s = 0 }
	return &types.CompressionResult{Strategy: types.CompressionSummary, TokenSavings: s, NewTokenTotal: tt, Content: fmt.Sprintf("[Summary]:\n%s\n[End]", resp.Content), ReplaceFrom: 0, ReplaceTo: end}, nil
}

func (eng *Engine) mermaid(ctx context.Context, msgs []types.Message) (*types.CompressionResult, error) {
	if len(msgs) < 2 { return nil, fmt.Errorf("need >=2 msgs") }
	n := len(msgs); end := n * 60 / 100
	if end < 1 { end = 1 }; if end >= n { end = n - 1 }
	var sb strings.Builder
	for i := 0; i < end; i++ {
		c := msgs[i].Content; if len(c) > 500 { c = c[:500] + "..." }
		sb.WriteString(msgs[i].Role); sb.WriteString(": "); sb.WriteString(c); sb.WriteString("\n")
	}
	t := sb.String(); ot := len(t) / 4
	resp, err := eng.llm.Chat(ctx, &types.ChatRequest{Messages: []types.ChatMessage{
		{Role: "system", Content: "You convert conversations to mermaid mindmaps."},
		{Role: "user", Content: fmt.Sprintf("Convert this conversation into a mermaid mindmap:\n%s", t)},
	}})
	if err != nil { return nil, fmt.Errorf("mermaid: %w", err) }
	tt := resp.Usage.TotalTokens; if tt <= 0 { tt = len(resp.Content) / 4 }
	s := ot - tt; if s < 0 { s = 0 }
	return &types.CompressionResult{Strategy: types.CompressionMermaid, TokenSavings: s, NewTokenTotal: tt, Content: fmt.Sprintf("[Mermaid]:\n%s\n[End]", resp.Content), ReplaceFrom: 0, ReplaceTo: end}, nil
}

func (eng *Engine) sliding(msgs []types.Message) (*types.CompressionResult, error) {
	if len(msgs) < 2 { return nil, fmt.Errorf("need >=2 msgs") }
	k := 20; if k >= len(msgs) { k = 1 }
	tr := msgs[:len(msgs)-k]; kp := msgs[len(msgs)-k:]
	ot := 0; for _, m := range msgs { if m.TokenCount > 0 { ot += m.TokenCount } else { ot += len(m.Content)/4 } }
	nt := 0; for _, m := range kp { if m.TokenCount > 0 { nt += m.TokenCount } else { nt += len(m.Content)/4 } }
	s := ot - nt; if s < 0 { s = 0 }
	var sb strings.Builder
	for _, m := range tr { sb.WriteString(fmt.Sprintf("[Truncated %s message]\n", m.Role)) }
	return &types.CompressionResult{Strategy: types.CompressionSlidingWindow, TokenSavings: s, NewTokenTotal: nt, Content: sb.String(), ReplaceFrom: 0, ReplaceTo: len(tr)}, nil
}
