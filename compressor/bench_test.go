package compressor

import (
	"context"
	"testing"
	"github.com/LionAnti/agent-core/types"
)

type benchLLM struct{}
func (benchLLM) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	return &types.ChatResponse{Content: "summary"}, nil
}

func BenchmarkShouldCompress(b *testing.B) {
	e := &Engine{}
	states := []*types.HarnessState{
		{CurrentTokens: 100, ContextWindow: 128000},
		{CurrentTokens: 64000, ContextWindow: 128000},
		{CurrentTokens: 110000, ContextWindow: 128000},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ShouldCompress(context.Background(), states[i%3])
	}
}

func BenchmarkSlidingCompress(b *testing.B) {
	e := &Engine{}
	msgs := make([]types.Message, 100)
	for i := range msgs { msgs[i] = types.Message{Role: "user", Content: "hello world", TokenCount: 10} }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.sliding(msgs)
	}
}

func BenchmarkSummaryCompress(b *testing.B) {
	e := &Engine{llm: benchLLM{}}
	msgs := make([]types.Message, 20)
	for i := range msgs { msgs[i] = types.Message{Role: "user", Content: "hello world test data for summary", TokenCount: 20} }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.summary(context.Background(), msgs)
	}
}
