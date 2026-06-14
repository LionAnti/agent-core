package agentcore

import "context"

type DefaultOffloadDecider struct {
	mildThreshold       float64
	aggressiveThreshold float64
}

func NewDefaultOffloadDecider(mild, aggressive float64) *DefaultOffloadDecider {
	return &DefaultOffloadDecider{mildThreshold: mild, aggressiveThreshold: aggressive}
}

func (od *DefaultOffloadDecider) Decide(ctx context.Context, density *DensitySignals, state *HarnessState) (*OffloadDecision, error) {
	ratio := float64(state.CurrentTokens) / float64(state.ContextWindow)
	if ratio >= od.aggressiveThreshold {
		return &OffloadDecision{
			ShouldOffload: true, Level: "aggressive",
			CompressionRatio: ratio,
			EstimatedSavings: int(float64(state.CurrentTokens) * 0.6),
			Reason:           "context aggressively full",
		}, nil
	}
	if ratio >= od.mildThreshold {
		return &OffloadDecision{
			ShouldOffload: true, Level: "mild",
			CompressionRatio: ratio,
			EstimatedSavings: int(float64(state.CurrentTokens) * 0.3),
			Reason:           "context moderately utilized",
		}, nil
	}
	return &OffloadDecision{ShouldOffload: false, CompressionRatio: ratio}, nil
}
