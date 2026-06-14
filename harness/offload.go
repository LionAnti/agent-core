package harness

import (
	"context"
	"github.com/agent-core/types"
)

type OffloadDecider struct {
	mild, aggressive float64
}

func NewOffloadDecider(mild, aggressive float64) *OffloadDecider {
	return &OffloadDecider{mild: mild, aggressive: aggressive}
}

func (od *OffloadDecider) Decide(ctx context.Context, density *types.DensitySignals, state *types.HarnessState) (*types.OffloadDecision, error) {
	if od==nil||state==nil { return &types.OffloadDecision{ShouldOffload: false}, nil }
	r := float64(state.CurrentTokens)/float64(state.ContextWindow)
	if r >= od.aggressive { return &types.OffloadDecision{ShouldOffload: true, Level:"aggressive", CompressionRatio: r, EstimatedSavings: int(float64(state.CurrentTokens)*0.6), Reason:"context full"}, nil }
	if r >= od.mild { return &types.OffloadDecision{ShouldOffload: true, Level:"mild", CompressionRatio: r, EstimatedSavings: int(float64(state.CurrentTokens)*0.3), Reason:"context moderate"}, nil }
	return &types.OffloadDecision{ShouldOffload: false, CompressionRatio: r}, nil
}
