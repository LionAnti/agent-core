package harness

import (
	"context"
	"math"
	"strings"
	"github.com/LionAnti/agent-core/types"
)

var shiftWords = []string{"but","however","actually","instead"}
var entityWords = []string{"http","api","key","token","server","db","config"}

type DensityEstimator struct{}

func NewDensityEstimator() *DensityEstimator { return &DensityEstimator{} }

func (de *DensityEstimator) Estimate(ctx context.Context, msgs []types.Message, state *types.HarnessState) (*types.DensitySignals, error) {
	s := &types.DensitySignals{}
	if len(msgs)==0||state==nil { return s, nil }
	s.MessageRate = float64(len(msgs))
	s.TokenDensity = float64(state.CurrentTokens)/float64(max(state.ContextWindow,1))
	s.TopicShiftScore = de.detectTopicShift(msgs)
	s.EntityCount = de.countEntities(msgs)
	c := s.TokenDensity*0.4 + s.TopicShiftScore*0.3 + float64(s.EntityCount)/100*0.2 + s.MessageRate/100*0.1
	s.OverallDensity = c
	a := 0.3
	if state.SmoothedDensity==0 { s.SmoothedDensity = c } else { s.SmoothedDensity = a*c + (1-a)*state.SmoothedDensity }
	state.SmoothedDensity = s.SmoothedDensity
	return s, nil
}

func (de *DensityEstimator) detectTopicShift(msgs []types.Message) float64 {
	if len(msgs)<3 { return 0 }
	l := strings.ToLower(msgs[len(msgs)-1].Content); p := strings.ToLower(msgs[len(msgs)-2].Content)
	c := 0
	for _, w := range shiftWords { if strings.Contains(l,w)&&!strings.Contains(p,w) { c++ } }
	return math.Min(float64(c)*0.25, 1.0)
}

func (de *DensityEstimator) countEntities(msgs []types.Message) int {
	c := 0
	for _, m := range msgs {
		lower := strings.ToLower(m.Content)
		for _, w := range entityWords { if strings.Contains(lower, w) { c++ } }
	}
	return c
}
