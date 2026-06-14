package memory

import (
	"context"
	"strings"
	"github.com/agent-core/types"
)

type Recall struct {
	store  types.MemoryStore
	vector types.VectorStore
	logger types.Logger
}

func NewRecall(store types.MemoryStore, vector types.VectorStore, logger types.Logger) *Recall {
	return &Recall{store: store, vector: vector, logger: logger}
}

func (mr *Recall) Recall(ctx context.Context, intent *types.IntentClassification, tid, uid string) (*types.RecallResult, error) {
	if mr==nil { return &types.RecallResult{}, nil }
	r := &types.RecallResult{}; if mr.store==nil||intent==nil { return r, nil }
	if intent.Type==types.IntentRecall {
		if ms,e := mr.store.SearchL1(ctx, tid, uid, intent.Strategy, 5); e==nil {
			r.Memories = ms
			for i := range ms { if ie := mr.store.IncrementL1Recall(ctx, ms[i].ID); ie!=nil { mr.logger.Warn("inc L1 recall", "id", ms[i].ID, "error", ie) } }
		}
	}
	if p,e := mr.store.GetCurrentL3(ctx, tid, uid); e==nil&&p!=nil { r.Persona=p; r.AppendSystemContext=p.Content }
	if ss,e := mr.store.GetL2ByUser(ctx, tid, uid); e==nil&&len(ss)>0 { r.Scene=ss[0] }
	if len(r.Memories)>0 {
		var sb strings.Builder; sb.WriteString("Relevant past context:\n")
		for _, m := range r.Memories { sb.WriteString("- "); sb.WriteString(m.Content); sb.WriteString("\n") }
		r.PrependContext = sb.String()
	}
	return r, nil
}
