package agentcore

import (
	"context"
	"math"
	"strings"
)

type DefaultDensityEstimator struct {
	prevDensity float64
	alpha       float64
}

func NewDefaultDensityEstimator() *DefaultDensityEstimator {
	return &DefaultDensityEstimator{alpha: 0.3}
}

func (de *DefaultDensityEstimator) Estimate(ctx context.Context, msgs []Message, state *HarnessState) (*DensitySignals, error) {
	signals := &DensitySignals{}
	msgCount := len(msgs)
	if msgCount == 0 {
		return signals, nil
	}
	signals.MessageRate = float64(msgCount)
	signals.TokenDensity = float64(state.CurrentTokens) / float64(max(state.ContextWindow, 1))
	signals.TopicShiftScore = de.detectTopicShift(msgs)
	signals.EntityCount = de.countEntities(msgs)
	combined := signals.TokenDensity*0.4 + signals.TopicShiftScore*0.3 +
		float64(signals.EntityCount)/100*0.2 + signals.MessageRate/100*0.1
	signals.OverallDensity = combined
	if de.prevDensity == 0 {
		signals.SmoothedDensity = combined
	} else {
		signals.SmoothedDensity = de.alpha*combined + (1-de.alpha)*de.prevDensity
	}
	de.prevDensity = signals.SmoothedDensity
	return signals, nil
}

func (de *DefaultDensityEstimator) detectTopicShift(msgs []Message) float64 {
	if len(msgs) < 3 {
		return 0
	}
	last := strings.ToLower(msgs[len(msgs)-1].Content)
	prev := strings.ToLower(msgs[len(msgs)-2].Content)
	shiftWords := []string{"but", "however", "actually", "instead"}
	count := 0
	for _, w := range shiftWords {
		if strings.Contains(last, w) && !strings.Contains(prev, w) {
			count++
		}
	}
	return math.Min(float64(count)*0.25, 1.0)
}

func (de *DefaultDensityEstimator) countEntities(msgs []Message) int {
	words := []string{"http", "api", "key", "token", "server", "db", "config"}
	count := 0
	for _, m := range msgs {
		lower := strings.ToLower(m.Content)
		for _, w := range words {
			if strings.Contains(lower, w) {
				count++
			}
		}
	}
	return count
}
