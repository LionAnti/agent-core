package agentcore

import (
	"context"
	"math"
)

type DefaultUtilityTracker struct {
	defaultHalfLife float64
}

func NewDefaultUtilityTracker() *DefaultUtilityTracker {
	return &DefaultUtilityTracker{defaultHalfLife: 7.0}
}

func (ut *DefaultUtilityTracker) Score(ctx context.Context, recordID string, accessCount int, lastAccessAt int64) (*UtilityScore, error) {
	if ut == nil {
		return &UtilityScore{}, nil
	}
	n := now()
	daysSince := 0.0
	if lastAccessAt > 0 {
		daysSince = float64(n-lastAccessAt) / (86400 * 1000)
	}
	freq := float64(accessCount)
	recency := math.Exp(-0.693 * daysSince / ut.defaultHalfLife)
	score := 0.4*freq + 0.6*recency
	return &UtilityScore{
		RecordID: recordID, AccessCount: accessCount,
		HalfLifeDays: ut.defaultHalfLife, Score: score, LastAccessAt: lastAccessAt,
	}, nil
}

func (ut *DefaultUtilityTracker) Decay(ctx context.Context, scores []UtilityScore, daysSince float64) ([]UtilityScore, error) {
	results := make([]UtilityScore, len(scores))
	for i, s := range scores {
		decay := math.Exp(-0.693 * daysSince / s.HalfLifeDays)
		results[i] = s
		results[i].Score = s.Score * decay
	}
	return results, nil
}
