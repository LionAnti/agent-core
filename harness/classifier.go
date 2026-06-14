package harness

import (
	"context"
	"fmt"
	"strings"
	"github.com/agent-core/types"
)

type Classifier struct {
	rules func(*types.RuleContext) []types.RuleOutput
	llm   types.LLMClient
}

func NewClassifier(rf func(*types.RuleContext) []types.RuleOutput, llm types.LLMClient) *Classifier {
	return &Classifier{rules: rf, llm: llm}
}

func (c *Classifier) Classify(ctx context.Context, msgs []types.Message, s *types.HarnessState) (*types.IntentClassification, error) {
	if c==nil||c.rules==nil { return &types.IntentClassification{Type: types.IntentFactual, Confidence: 0.3}, nil }
	if len(msgs)==0 { return &types.IntentClassification{Type: types.IntentUnknown}, nil }
	last := msgs[len(msgs)-1].Content
	outs := c.rules(&types.RuleContext{Domain: types.DomainIntentClassification, Query: last})
	for _, o := range outs { if o.Score >= 0.7 { return &types.IntentClassification{Type: types.IntentType(o.Label), Strategy:"rule", Confidence: o.Score}, nil } }
	if c.llm != nil { return c.llmClassify(ctx, last) }
	return &types.IntentClassification{Type: types.IntentFactual, Strategy: "default", Confidence: 0.5}, nil
}

func (c *Classifier) llmClassify(ctx context.Context, input string) (*types.IntentClassification, error) {
	r, e := c.llm.Chat(ctx, &types.ChatRequest{Messages: []types.ChatMessage{{Role:"system", Content:"You are an intent classifier. Output: label|confidence."}, {Role:"user", Content:"Classify: factual, exploratory, recall, task, greeting. Message: "+input}}})
	if e != nil { return nil, e }
	t := strings.TrimSpace(r.Content); p := strings.SplitN(t, "|", 2)
	l := strings.TrimSpace(p[0]); var cf float64 = 0.7
	if len(p)>1 { fmt.Sscanf(p[1], "%f", &cf) }
	it := types.IntentType(l)
	switch it { case types.IntentFactual, types.IntentExploratory, types.IntentRecall, types.IntentTask, types.IntentGreeting: default: it = types.IntentFactual }
	return &types.IntentClassification{Type: it, Strategy: "llm", Confidence: cf}, nil
}
