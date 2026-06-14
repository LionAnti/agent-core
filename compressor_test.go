package agentcore

import (
    "context"
    "strings"
    "testing"
)

type mockCompressorLLM struct{}

func (m *mockCompressorLLM) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
    return &ChatResponse{
        Content: "Summarized content here",
        Usage:   Usage{PromptTokens: 50, CompletionTokens: 5, TotalTokens: 55},
    }, nil
}

func TestCompressor_ShouldCompress_Low(t *testing.T) {
    c := &CompressorEngine{}
    state := &HarnessState{CurrentTokens: 100, ContextWindow: 128000}
    d := c.ShouldCompress(state)
    if d.ShouldCompress {
        t.Fatal("should not compress at <1% ratio")
    }
}

func TestCompressor_ShouldCompress_Mild(t *testing.T) {
    c := &CompressorEngine{}
    state := &HarnessState{CurrentTokens: 64000, ContextWindow: 128000}
    d := c.ShouldCompress(state)
    if !d.ShouldCompress {
        t.Fatal("should compress at 50% ratio")
    }
    if d.Level != CompressionMild {
        t.Fatalf("expected mild, got %s", d.Level)
    }
}

func TestCompressor_ShouldCompress_Aggressive(t *testing.T) {
    c := &CompressorEngine{}
    state := &HarnessState{CurrentTokens: 110000, ContextWindow: 128000}
    d := c.ShouldCompress(state)
    if !d.ShouldCompress {
        t.Fatal("should compress at 86% ratio")
    }
    if d.Level != CompressionAggressive {
        t.Fatalf("expected aggressive, got %s", d.Level)
    }
}

func TestCompressor_SlidingWindow(t *testing.T) {
    c := &CompressorEngine{}
    msgs := make([]Message, 30)
    for i := range msgs {
        msgs[i] = Message{Role: "user", Content: "message content here", TokenCount: 10}
    }
    result, err := c.slidingCompress(msgs)
    if err != nil {
        t.Fatal(err)
    }
    if result.TokenSavings <= 0 {
        t.Fatal("expected token savings from sliding window")
    }
    if !strings.Contains(result.Content, "[Truncated") {
        t.Fatal("expected truncated markers in output")
    }
}

func TestCompressor_Summary(t *testing.T) {
    c := &CompressorEngine{llm: &mockCompressorLLM{}}
    msgs := []Message{
        {Role: "user", Content: "Hello, can you help me?"},
        {Role: "assistant", Content: "Sure, how can I help?"},
    }
    result, err := c.summaryCompress(context.Background(), msgs)
    if err != nil {
        t.Fatal(err)
    }
    if result.Strategy != CompressionSummary {
        t.Fatalf("expected summary strategy, got %s", result.Strategy)
    }
    if !strings.Contains(result.Content, "[Summary]") {
        t.Fatal("expected summary marker in content")
    }
}

func TestCompressor_EmptyMessages(t *testing.T) {
    c := &CompressorEngine{}
    _, err := c.summaryCompress(context.Background(), nil)
    if err == nil {
        t.Fatal("expected error for empty messages")
    }
    _, err = c.slidingCompress(nil)
    if err == nil {
        t.Fatal("expected error for empty messages")
    }
}

func TestEstimateTokens(t *testing.T) {
    if estimateTokens("") != 0 {
        t.Fatal("empty string should estimate 0 tokens")
    }
    if estimateTokens("abcd") < 1 {
        t.Fatal("4 chars should estimate at least 1 token")
    }
}
