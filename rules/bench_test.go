package rules

import (
	"testing"
	"github.com/LionAnti/agent-core/types"
)

func BenchmarkMatch(b *testing.B) {
	e := NewEngine()
	ctx := &types.RuleContext{Domain: types.DomainIntentClassification, Query: "hello world this is a test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Match(ctx)
	}
}

func BenchmarkMatch_LargeRuleset(b *testing.B) {
	e := NewEngine()
	for i := 0; i < 100; i++ {
		e.AddRule(&types.Rule{
			Name:    "extra",
			Domain:  types.DomainScoring,
			Pattern: `test_pattern`,
			Label:   "extra", Score: 0.5, Enabled: true,
		})
	}
	ctx := &types.RuleContext{Domain: types.DomainScoring, Query: "test_pattern"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Match(ctx)
	}
}

func BenchmarkAddRule(b *testing.B) {
	e := NewEngine()
	r := &types.Rule{Name: "bench", Domain: types.DomainScoring, Pattern: "test", Label: "x", Score: 0.5, Enabled: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.AddRule(r)
	}
}
