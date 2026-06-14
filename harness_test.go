package agentcore

import (
    "context"
    "testing"
)

// mockRulesEngine is a simple rules engine for testing.
func TestHybridClassifier_Fallback(t *testing.T) {
    re := NewRulesEngine()
    c := &HybridClassifier{
        rules: re,
        llm:   nil, // no LLM, should fall back to default
    }
    ic, err := c.Classify(context.Background(), []Message{
        {Role: "user", Content: "hello"},
    }, &HarnessState{ContextWindow: 128000})
    if err != nil {
        t.Fatal(err)
    }
    if ic.Type != IntentGreeting {
        t.Fatalf("expected IntentGreeting for hello, got %s", ic.Type)
    }
    if ic.Confidence < 0.5 {
        t.Fatalf("expected >=0.5 confidence for greeting, got %f", ic.Confidence)
    }
}

func TestDensityEstimator_Empty(t *testing.T) {
    de := &DefaultDensityEstimator{}
    signals, err := de.Estimate(context.Background(), nil, &HarnessState{ContextWindow: 128000})
    if err != nil {
        t.Fatal(err)
    }
    if signals.MessageRate != 0 {
        t.Fatal("expected 0 message rate for nil input")
    }
}

func TestDensityEstimator_SmoothedDensity(t *testing.T) {
    de := &DefaultDensityEstimator{}
    msgs := []Message{
        {Role: "user", Content: "hello"},
        {Role: "assistant", Content: "hi there"},
    }
    // Two independent sessions should not interfere
    state1 := &HarnessState{CurrentTokens: 100, ContextWindow: 1000}
    state2 := &HarnessState{CurrentTokens: 100, ContextWindow: 1000}
    s1, _ := de.Estimate(context.Background(), msgs, state1)
    s2, _ := de.Estimate(context.Background(), msgs, state2)
    // Both first calls should produce the same result
    if s1.SmoothedDensity != s2.SmoothedDensity {
        t.Fatalf("independent sessions should get same first estimate: %f vs %f", s1.SmoothedDensity, s2.SmoothedDensity)
    }
    // Second call on state1 should smooth
    s3, _ := de.Estimate(context.Background(), msgs, state1)
    if s3.SmoothedDensity <= 0 {
        t.Fatalf("expected positive smoothed density, got %f", s3.SmoothedDensity)
    }
}

func TestOffloadDecider_BelowThreshold(t *testing.T) {
    od := &DefaultOffloadDecider{mildThreshold: 0.5, aggressiveThreshold: 0.85}
    decision, err := od.Decide(context.Background(), nil, &HarnessState{
        CurrentTokens: 100,
        ContextWindow: 1000,
    })
    if err != nil {
        t.Fatal(err)
    }
    if decision.ShouldOffload {
        t.Fatal("ShouldOffload should be false at 10% ratio")
    }
}

func TestOffloadDecider_Mild(t *testing.T) {
    od := &DefaultOffloadDecider{mildThreshold: 0.5, aggressiveThreshold: 0.85}
    decision, err := od.Decide(context.Background(), nil, &HarnessState{
        CurrentTokens: 600,
        ContextWindow: 1000,
    })
    if err != nil {
        t.Fatal(err)
    }
    if !decision.ShouldOffload {
        t.Fatal("ShouldOffload should be true at 60% ratio")
    }
    if decision.Level != "mild" {
        t.Fatalf("expected mild, got %s", decision.Level)
    }
}

func TestOffloadDecider_Aggressive(t *testing.T) {
    od := &DefaultOffloadDecider{mildThreshold: 0.5, aggressiveThreshold: 0.85}
    decision, err := od.Decide(context.Background(), nil, &HarnessState{
        CurrentTokens: 900,
        ContextWindow: 1000,
    })
    if err != nil {
        t.Fatal(err)
    }
    if !decision.ShouldOffload {
        t.Fatal("ShouldOffload should be true at 90% ratio")
    }
    if decision.Level != "aggressive" {
        t.Fatalf("expected aggressive, got %s", decision.Level)
    }
}

func TestUtilityTracker_Score(t *testing.T) {
    ut := &DefaultUtilityTracker{defaultHalfLife: 7.0}
    score := ut.Score("test-1", 5, 0)
    if score.Score <= 0 {
        t.Fatalf("expected positive score, got %f", score.Score)
    }
    if score.RecordID != "test-1" {
        t.Fatalf("expected recordID test-1, got %s", score.RecordID)
    }
}

func TestUtilityTracker_Decay(t *testing.T) {
    ut := &DefaultUtilityTracker{defaultHalfLife: 7.0}
    scores := []UtilityScore{
        {RecordID: "a", Score: 1.0, HalfLifeDays: 7.0},
        {RecordID: "b", Score: 0.5, HalfLifeDays: 7.0},
    }
    decayed := ut.Decay(scores, 7.0)
    if decayed[0].Score >= 1.0 {
        t.Fatal("score should have decayed below 1.0")
    }
    if decayed[0].Score <= 0 {
        t.Fatal("decayed score should be positive")
    }
}

func TestRulesEngine_Match(t *testing.T) {
    re := NewRulesEngine()
    outputs := re.Match(&RuleContext{
        Domain: DomainIntentClassification,
        Query:  "hello",
    })
    foundGreeting := false
    for _, o := range outputs {
        if o.Label == "greeting" {
            foundGreeting = true
            break
        }
    }
    if !foundGreeting {
        t.Fatal("expected greeting rule to match 'hello'")
    }
}

func TestRulesEngine_AddRemove(t *testing.T) {
    re := NewRulesEngine()
    re.AddRule(&Rule{
        Name: "test_rule", Domain: DomainScoring,
        Pattern: "test", Label: "test_label", Score: 0.5, Enabled: true,
    })
    outputs := re.Match(&RuleContext{Domain: DomainScoring, Query: "test"})
    if len(outputs) == 0 {
        t.Fatal("expected test rule to match")
    }
    re.RemoveRule("test_rule")
    outputs = re.Match(&RuleContext{Domain: DomainScoring, Query: "test"})
    for _, o := range outputs {
        if o.RuleName == "test_rule" {
            t.Fatal("test rule should have been removed")
        }
    }
}

func TestToolRegistry_NewRegistry(t *testing.T) {
    r := NewToolRegistry(nil, NoopLogger{})
    if r == nil {
        t.Fatal("registry should not be nil")
    }
}

func TestCopyTool_Nil(t *testing.T) {
    if got := copyTool(nil); got != nil {
        t.Fatal("copyTool(nil) should return nil")
    }
}

func TestCopyTool(t *testing.T) {
    original := &ToolSpec{
        Name: "test", Description: "test tool",
        Parameters: []ToolParameter{{Name: "p1"}},
        Tags:       []string{"tag1"},
    }
    copied := copyTool(original)
    copied.Name = "modified"
    if original.Name != "test" {
        t.Fatal("modifying copy should not affect original")
    }
}

func TestEnsureTimeout(t *testing.T) {
    ctx := context.Background()
    wrapped := ensureTimeout(ctx, 10)
    deadline, ok := wrapped.Deadline()
    if !ok {
        t.Fatal("expected deadline after ensureTimeout")
    }
    if deadline.IsZero() {
        t.Fatal("deadline should not be zero")
    }
}

func TestEnsureTimeout_Existing(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5)
    defer cancel()
    wrapped := ensureTimeout(ctx, 30)
    deadline, ok := wrapped.Deadline()
    if !ok {
        t.Fatal("expected deadline")
    }
    _ = deadline
}
