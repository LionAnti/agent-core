package agentcore

import (
	"context"
	"fmt"
	"strings"
)

type HybridClassifier struct {
	rules *RulesEngine
	llm   LLMClient
}

func NewHybridClassifier(rules *RulesEngine, llm LLMClient) *HybridClassifier {
	return &HybridClassifier{rules: rules, llm: llm}
}

func (c *HybridClassifier) Classify(ctx context.Context, msgs []Message, state *HarnessState) (*IntentClassification, error) {
	if c == nil || c.rules == nil {
		return &IntentClassification{Type: IntentFactual, Confidence: 0.3}, nil
	}
	if len(msgs) == 0 {
		return &IntentClassification{Type: IntentUnknown, Confidence: 0}, nil
	}
	lastMsg := msgs[len(msgs)-1].Content
	outputs := c.rules.Match(&RuleContext{Domain: DomainIntentClassification, Query: lastMsg})
	for _, o := range outputs {
		if o.Score >= 0.7 {
			return &IntentClassification{
				Type: IntentType(o.Label), Strategy: "rule", Confidence: o.Score,
			}, nil
		}
	}
	if c.llm != nil {
		return c.llmClassify(ctx, lastMsg)
	}
	return &IntentClassification{Type: IntentFactual, Strategy: "default", Confidence: 0.5}, nil
}

func (c *HybridClassifier) llmClassify(ctx context.Context, input string) (*IntentClassification, error) {
	resp, err := c.llm.Chat(ctx, &ChatRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: "You are an intent classifier. Output: label|confidence."},
			{Role: "user", Content: "Classify into: factual, exploratory, recall, task, greeting. Message: " + input},
		},
	})
	if err != nil {
		return nil, err
	}
	text := strings.TrimSpace(resp.Content)
	parts := strings.SplitN(text, "|", 2)
	label := strings.TrimSpace(parts[0])
	var confidence float64 = 0.7
	if len(parts) > 1 {
		fmt.Sscanf(parts[1], "%f", &confidence)
	}
	intent := IntentType(label)
	switch intent {
	case IntentFactual, IntentExploratory, IntentRecall, IntentTask, IntentGreeting:
	default:
		intent = IntentFactual
	}
	return &IntentClassification{Type: intent, Strategy: "llm", Confidence: confidence}, nil
}
