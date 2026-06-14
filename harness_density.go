package agentcore

import (
	"context"
	"math"
	"strings"
)

var (
	shiftWords = []string{"but", "however", "actually", "instead"}
	entityWords = []string{"http", "api", "key", "token", "server", "db", "config"}
)

type DefaultDensityEstimator struct{}

func NewDefaultDensityEstimator() *DefaultDensityEstimator {
	return &DefaultDensityEstimator{}
}

func (de *DefaultDensityEstimator) Estimate(ctx context.Context, msgs []Message, state *HarnessState) (*DensitySignals, error) {
	signals := &DensitySignals{}
	msgCount := len(msgs)
	if msgCount == 0 || state == nil {
		return signals, nil
	}
	signals.MessageRate = float64(msgCount)
	signals.TokenDensity = float64(state.CurrentTokens) / float64(max(state.ContextWindow, 1))
	signals.TopicShiftScore = de.detectTopicShift(msgs)
	signals.EntityCount = de.countEntities(msgs)
	combined := signals.TokenDensity*0.4 + signals.TopicShiftScore*0.3 +
		float64(signals.EntityCount)/100*0.2 + signals.MessageRate/100*0.1
	signals.OverallDensity = combined
	alpha := 0.3
	if state.SmoothedDensity == 0 {
		signals.SmoothedDensity = combined
	} else {
		signals.SmoothedDensity = alpha*combined + (1-alpha)*state.SmoothedDensity
	}
	state.SmoothedDensity = signals.SmoothedDensity
	return signals, nil
}

func (de *DefaultDensityEstimator) detectTopicShift(msgs []Message) float64 {
	if len(msgs) < 3 {
		return 0
	}
	last := strings.ToLower(msgs[len(msgs)-1].Content)
	prev := strings.ToLower(msgs[len(msgs)-2].Content)
	count := 0
	for _, w := range shiftWords {
		if strings.Contains(last, w) && !strings.Contains(prev, w) {
			count++
		}
	}
	return math.Min(float64(count)*0.25, 1.0)
}

func (de *DefaultDensityEstimator) countEntities(msgs []Message) int {
	count := 0
	for _, m := range msgs {
		lower := strings.ToLower(m.Content)
		for _, w := range entityWords {
			if strings.Contains(lower, w) {
				count++
			}
		}
	}
	return count
}
