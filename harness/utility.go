package harness

import (
	"context"
	"math"
	"time"
	"github.com/LionAnti/agent-core/types"
)

type UtilityTracker struct {
	halfLife float64
}

func NewUtilityTracker() *UtilityTracker { return &UtilityTracker{halfLife: 7.0} }

func (ut *UtilityTracker) Score(ctx context.Context, id string, ac int, laa int64) (*types.UtilityScore, error) {
	if ut == nil { return &types.UtilityScore{}, nil }
	ds := 0.0; n := time.Now().UnixMilli()
	if laa>0 { ds = float64(n-laa)/(86400*1000) }
	f := float64(ac); r := math.Exp(-0.693*ds/ut.halfLife)
	return &types.UtilityScore{RecordID: id, AccessCount: ac, HalfLifeDays: ut.halfLife, Score: 0.4*f+0.6*r, LastAccessAt: laa}, nil
}

func (ut *UtilityTracker) Decay(ctx context.Context, scores []types.UtilityScore, ds float64) ([]types.UtilityScore, error) {
	rs := make([]types.UtilityScore, len(scores))
	for i, s := range scores { d := math.Exp(-0.693*ds/s.HalfLifeDays); rs[i]=s; rs[i].Score=s.Score*d }
	return rs, nil
}
