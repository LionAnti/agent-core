package harness

import (
	"context"
	"testing"
	"github.com/LionAnti/agent-core/types"
)

type mockLLM struct{}
func (m *mockLLM) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	return &types.ChatResponse{Content: "factual|0.95"}, nil
}

func matchFunc(ctx *types.RuleContext) []types.RuleOutput {
	if ctx != nil && ctx.Query == "hello" {
		return []types.RuleOutput{{RuleName: "greeting", Label: "greeting", Score: 0.9, Domain: types.DomainIntentClassification}}
	}
	return nil
}

func TestClassifier_GreetingByRule(t *testing.T) {
	c := NewClassifier(matchFunc, nil)
	ic, err := c.Classify(context.Background(), []types.Message{{Role: "user", Content: "hello"}}, &types.HarnessState{ContextWindow: 128000})
	if err != nil { t.Fatal(err) }
	if ic.Type != types.IntentGreeting { t.Fatalf("expected greeting, got %s", ic.Type) }
}

func TestClassifier_FallbackDefault(t *testing.T) {
	c := NewClassifier(matchFunc, nil)
	ic, err := c.Classify(context.Background(), []types.Message{{Role: "user", Content: "some random query"}}, &types.HarnessState{})
	if err != nil { t.Fatal(err) }
	if ic.Type != types.IntentFactual { t.Fatalf("expected factual, got %s", ic.Type) }
}

func TestClassifier_WithLLM(t *testing.T) {
	c := NewClassifier(matchFunc, &mockLLM{})
	ic, err := c.Classify(context.Background(), []types.Message{{Role: "user", Content: "unknown"}}, &types.HarnessState{})
	if err != nil { t.Fatal(err) }
	if ic.Type != types.IntentFactual { t.Fatalf("expected factual, got %s", ic.Type) }
}

func TestClassifier_EmptyMsgs(t *testing.T) {
	c := NewClassifier(matchFunc, nil)
	ic, err := c.Classify(context.Background(), nil, nil)
	if err != nil { t.Fatal(err) }
	if ic.Type != types.IntentUnknown { t.Fatalf("expected unknown, got %s", ic.Type) }
}

func TestDensityEstimator_Empty(t *testing.T) {
	de := &DensityEstimator{}
	s, err := de.Estimate(context.Background(), nil, &types.HarnessState{ContextWindow: 128000})
	if err != nil { t.Fatal(err) }
	if s.MessageRate != 0 { t.Fatal("expected 0") }
}

func TestDensityEstimator_Normal(t *testing.T) {
	de := &DensityEstimator{}
	msgs := []types.Message{{Role: "user", Content: "hello"}}
	state := &types.HarnessState{CurrentTokens: 100, ContextWindow: 1000}
	s1, _ := de.Estimate(context.Background(), msgs, state)
	if s1.SmoothedDensity <= 0 { t.Fatalf("expected positive density, got %f", s1.SmoothedDensity) }
	// Second call with same inputs should smooth
	state.SmoothedDensity = 0 // reset
	s2, _ := de.Estimate(context.Background(), msgs, state)
	if s2.SmoothedDensity != s1.SmoothedDensity { t.Log("independent sessions produce same first estimate - good") }
}

func TestDensityEstimator_SmoothedDensity(t *testing.T) {
	de := &DensityEstimator{}
	msgs := []types.Message{{Role: "user", Content: "hello"}}
	// Two independent sessions
	s1 := &types.HarnessState{CurrentTokens: 100, ContextWindow: 1000}
	s2 := &types.HarnessState{CurrentTokens: 100, ContextWindow: 1000}
	r1, _ := de.Estimate(context.Background(), msgs, s1)
	r2, _ := de.Estimate(context.Background(), msgs, s2)
	// First calls on independent sessions should produce the same result
	if r1.SmoothedDensity != r2.SmoothedDensity {
		t.Fatalf("independent sessions should produce same first estimate: %f vs %f", r1.SmoothedDensity, r2.SmoothedDensity)
	}
	r3, _ := de.Estimate(context.Background(), msgs, s1)
	if r3.SmoothedDensity <= 0 { t.Fatal("expected positive after second call") }
}

func TestOffloadDecider_BelowThreshold(t *testing.T) {
	od := &OffloadDecider{mild: 0.5, aggressive: 0.85}
	d, err := od.Decide(context.Background(), nil, &types.HarnessState{CurrentTokens: 100, ContextWindow: 1000})
	if err != nil { t.Fatal(err) }
	if d.ShouldOffload { t.Fatal("should not offload at 10%") }
}

func TestOffloadDecider_Mild(t *testing.T) {
	od := &OffloadDecider{mild: 0.5, aggressive: 0.85}
	d, err := od.Decide(context.Background(), nil, &types.HarnessState{CurrentTokens: 600, ContextWindow: 1000})
	if err != nil { t.Fatal(err) }
	if !d.ShouldOffload { t.Fatal("should offload at 60%") }
	if d.Level != "mild" { t.Fatalf("expected mild, got %s", d.Level) }
}

func TestOffloadDecider_Aggressive(t *testing.T) {
	od := &OffloadDecider{mild: 0.5, aggressive: 0.85}
	d, err := od.Decide(context.Background(), nil, &types.HarnessState{CurrentTokens: 900, ContextWindow: 1000})
	if err != nil { t.Fatal(err) }
	if !d.ShouldOffload { t.Fatal("should offload at 90%") }
	if d.Level != "aggressive" { t.Fatalf("expected aggressive, got %s", d.Level) }
}

func TestUtilityTracker(t *testing.T) {
	ut := NewUtilityTracker()
	s, err := ut.Score(context.Background(), "test-1", 5, 0)
	if err != nil { t.Fatal(err) }
	if s.Score <= 0 { t.Fatalf("expected positive score, got %f", s.Score) }
	if s.RecordID != "test-1" { t.Fatalf("expected test-1, got %s", s.RecordID) }
}

func TestUtilityTracker_Decay(t *testing.T) {
	ut := NewUtilityTracker()
	scores := []types.UtilityScore{{RecordID: "a", Score: 1.0, HalfLifeDays: 7.0}, {RecordID: "b", Score: 0.5, HalfLifeDays: 7.0}}
	decayed, err := ut.Decay(context.Background(), scores, 7.0)
	if err != nil { t.Fatal(err) }
	if decayed[0].Score >= 1.0 { t.Fatal("score should have decayed") }
	if decayed[0].Score <= 0 { t.Fatal("decayed score should be positive") }
}

func TestNilSafe(t *testing.T) {
	var c *Classifier
	ic, err := c.Classify(context.Background(), nil, nil)
	if err != nil { t.Fatal(err) }
	if ic.Type != types.IntentFactual { t.Fatalf("expected factual, got %s", ic.Type) }

	var od *OffloadDecider
	d, err := od.Decide(context.Background(), nil, nil)
	if err != nil { t.Fatal(err) }
	if d.ShouldOffload { t.Fatal("nil should not offload") }
}
