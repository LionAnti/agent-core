package harness

import (
	"context"
	"testing"
	"github.com/LionAnti/agent-core/types"
)

func BenchmarkClassifier(b *testing.B) {
	c := NewClassifier(func(ctx *types.RuleContext) []types.RuleOutput { return nil }, nil)
	msgs := []types.Message{{Role: "user", Content: "What is the weather like today?"}}
	state := &types.HarnessState{CurrentTokens: 500, ContextWindow: 128000}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Classify(context.Background(), msgs, state)
	}
}

func BenchmarkDensityEstimate(b *testing.B) {
	de := &DensityEstimator{}
	msgs := make([]types.Message, 50)
	for i := range msgs { msgs[i] = types.Message{Role: "user", Content: "hello world api key token config server db error deploy version", TokenCount: 20} }
	state := &types.HarnessState{CurrentTokens: 5000, ContextWindow: 128000}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		de.Estimate(context.Background(), msgs, state)
	}
}

func BenchmarkOffloadDecide(b *testing.B) {
	od := &OffloadDecider{mild: 0.5, aggressive: 0.85}
	states := []*types.HarnessState{
		{CurrentTokens: 10000, ContextWindow: 128000},
		{CurrentTokens: 70000, ContextWindow: 128000},
		{CurrentTokens: 110000, ContextWindow: 128000},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		od.Decide(context.Background(), nil, states[i%3])
	}
}
