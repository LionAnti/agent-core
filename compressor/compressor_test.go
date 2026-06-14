package compressor

import (
	"context"
	"strings"
	"testing"
	"github.com/LionAnti/agent-core/types"
)

type mockLLM struct{}

func (m *mockLLM) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	return &types.ChatResponse{Content: "Summarized content here", Usage: types.Usage{TotalTokens: 55}}, nil
}

type mockLogger struct{}

func (mockLogger) Debug(msg string, keys ...any) {}
func (mockLogger) Info(msg string, keys ...any)  {}
func (mockLogger) Warn(msg string, keys ...any)  {}
func (mockLogger) Error(msg string, keys ...any) {}

func TestShouldCompress_Low(t *testing.T) {
	eng := &Engine{}
	d := eng.ShouldCompress(context.Background(), &types.HarnessState{CurrentTokens: 100, ContextWindow: 128000})
	if d.ShouldCompress { t.Fatal("should not compress at <1% ratio") }
}

func TestShouldCompress_Mild(t *testing.T) {
	eng := &Engine{}
	d := eng.ShouldCompress(context.Background(), &types.HarnessState{CurrentTokens: 64000, ContextWindow: 128000})
	if !d.ShouldCompress { t.Fatal("should compress at 50%") }
	if d.Level != types.CompressionMild { t.Fatalf("expected mild, got %s", d.Level) }
}

func TestShouldCompress_Aggressive(t *testing.T) {
	eng := &Engine{}
	d := eng.ShouldCompress(context.Background(), &types.HarnessState{CurrentTokens: 110000, ContextWindow: 128000})
	if !d.ShouldCompress { t.Fatal("should compress at 86%") }
	if d.Level != types.CompressionAggressive { t.Fatalf("expected aggressive, got %s", d.Level) }
}

func TestSlidingWindow(t *testing.T) {
	eng := &Engine{}
	msgs := make([]types.Message, 30)
	for i := range msgs { msgs[i] = types.Message{Role: "user", Content: "hello", TokenCount: 10} }
	r, err := eng.sliding(msgs)
	if err != nil { t.Fatal(err) }
	if r.TokenSavings <= 0 { t.Fatal("expected token savings") }
	if !strings.Contains(r.Content, "[Truncated") { t.Fatal("expected truncated markers") }
}

func TestSummary(t *testing.T) {
	eng := &Engine{llm: &mockLLM{}}
	msgs := []types.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "World"},
	}
	r, err := eng.summary(context.Background(), msgs)
	if err != nil { t.Fatal(err) }
	if r.Strategy != types.CompressionSummary { t.Fatalf("expected summary, got %s", r.Strategy) }
}

func TestMermaid(t *testing.T) {
	eng := &Engine{llm: &mockLLM{}}
	msgs := []types.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "World"},
	}
	r, err := eng.mermaid(context.Background(), msgs)
	if err != nil { t.Fatal(err) }
	if r.Strategy != types.CompressionMermaid { t.Fatalf("expected mermaid, got %s", r.Strategy) }
}

func TestEmptyMsgs(t *testing.T) {
	eng := &Engine{}
	_, err := eng.sliding(nil)
	if err == nil { t.Fatal("expected error") }
}

func TestCompress_Dispatch(t *testing.T) {
	eng := &Engine{llm: &mockLLM{}}
	msgs := []types.Message{{Role: "user", Content: "a"}, {Role: "assistant", Content: "b"}}
	r, err := eng.Compress(context.Background(), msgs, &types.CompressionDecision{Strategy: types.CompressionSummary})
	if err != nil { t.Fatal(err) }
	if r.Strategy != types.CompressionSummary { t.Fatalf("expected summary") }
}
