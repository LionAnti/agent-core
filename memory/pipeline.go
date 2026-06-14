package memory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
	"github.com/agent-core/types"
)

type Pipeline struct {
	store  types.MemoryStore
	logger types.Logger
}

func NewPipeline(store types.MemoryStore, logger types.Logger) *Pipeline {
	return &Pipeline{store: store, logger: logger}
}

func extractTID(k string) string {
	p := strings.SplitN(k, "/", 3)
	if len(p)>=1&&p[0]!="" { return p[0] }; return ""
}

func extractUID(k string) string {
	p := strings.SplitN(k, "/", 3)
	if len(p)>=2&&p[1]!="" { return p[1] }; return ""
}

func nid() string {
	b := make([]byte,16)
	if _,e := rand.Read(b); e!=nil { return time.Now().Format("150405.000") }
	return hex.EncodeToString(b)
}

func (p *Pipeline) ExtractL1(ctx context.Context, records []types.L0Record) error {
	if p==nil||p.store==nil||len(records)<3 { return nil }
	var sb strings.Builder
	for _, r := range records { sb.WriteString(r.Role); sb.WriteString(": "); sb.WriteString(r.Content); sb.WriteString("\n") }
	c := sb.String()
	if strings.TrimSpace(c)=="" { return nil }
	return p.store.SaveL1(ctx, &types.L1Memory{ID: nid(), TenantID: extractTID(records[0].SessionKey), UserID: extractUID(records[0].SessionKey), Content: c, Type:"conversation", CreatedAt: time.Now().UnixMilli()})
}
