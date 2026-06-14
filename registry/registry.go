package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"github.com/agent-core/types"
)

type ToolRegistry struct {
	store  types.RegistryStore
	logger types.Logger
	mu     sync.RWMutex
	cache  map[string]*types.ToolSpec
}

func NewToolRegistry(store types.RegistryStore, logger types.Logger) *ToolRegistry {
	return &ToolRegistry{store: store, logger: logger, cache: make(map[string]*types.ToolSpec)}
}

func (r *ToolRegistry) Register(ctx context.Context, tool *types.ToolSpec) error {
	if r == nil { return fmt.Errorf("registry: nil receiver") }
	k := fmt.Sprintf("tools/%s/%s", tool.TenantID, tool.Name)
	d, e := json.Marshal(tool); if e != nil { return fmt.Errorf("marshal: %w", e) }
	if r.store == nil { return fmt.Errorf("registry: %w: store nil", types.ErrInvalidConfig) }
	if e = r.store.Put(ctx, k, d); e != nil { return fmt.Errorf("put: %w", e) }
	r.mu.Lock(); r.cache[k] = copyTool(tool); r.mu.Unlock()
	return nil
}

func (r *ToolRegistry) Unregister(ctx context.Context, tenantID, name string) error {
	if r == nil { return fmt.Errorf("registry: nil receiver") }
	k := fmt.Sprintf("tools/%s/%s", tenantID, name)
	if r.store == nil { return fmt.Errorf("registry: %w: store nil", types.ErrInvalidConfig) }
	if e := r.store.Delete(ctx, k); e != nil { return fmt.Errorf("delete: %w", e) }
	r.mu.Lock(); delete(r.cache, k); r.mu.Unlock()
	return nil
}

func (r *ToolRegistry) Get(ctx context.Context, tenantID, name string) (*types.ToolSpec, error) {
	if r == nil { return nil, fmt.Errorf("registry: nil receiver") }
	k := fmt.Sprintf("tools/%s/%s", tenantID, name)
	r.mu.RLock(); c, ok := r.cache[k]; r.mu.RUnlock()
	if ok { return copyTool(c), nil }
	if r.store == nil { return nil, fmt.Errorf("registry: %w: store nil", types.ErrInvalidConfig) }
	d, e := r.store.Get(ctx, k); if e != nil { return nil, fmt.Errorf("get: %w", e) }
	var t types.ToolSpec
	if e = json.Unmarshal(d, &t); e != nil { return nil, fmt.Errorf("unmarshal: %w", e) }
	tc := copyTool(&t); r.mu.Lock(); r.cache[k] = tc; r.mu.Unlock()
	return tc, nil
}

func (r *ToolRegistry) List(ctx context.Context, tenantID string) ([]types.ToolSpec, error) {
	if r == nil { return nil, fmt.Errorf("registry: nil receiver") }
	if r.store == nil { return nil, fmt.Errorf("registry: %w: store nil", types.ErrInvalidConfig) }
	dl, e := r.store.List(ctx, fmt.Sprintf("tools/%s/", tenantID))
	if e != nil { return nil, fmt.Errorf("list: %w", e) }
	ts := make([]types.ToolSpec, 0, len(dl))
	for _, d := range dl {
		var t types.ToolSpec
		if e := json.Unmarshal(d, &t); e != nil { r.logger.Warn("unmarshal", "error", e); continue }
		if t.Status == types.ToolStatusEnabled { ts = append(ts, t) }
	}
	return ts, nil
}

func (r *ToolRegistry) Search(ctx context.Context, tenantID, query string) ([]types.ToolSpec, error) {
	if r == nil { return nil, nil }
	ts, e := r.List(ctx, tenantID); if e != nil { return nil, e }
	q := strings.ToLower(query)
	rs := make([]types.ToolSpec, 0, len(ts))
	for _, t := range ts {
		if strings.Contains(strings.ToLower(t.Name), q) || strings.Contains(strings.ToLower(t.Description), q) { rs = append(rs, t); continue }
		for _, tag := range t.Tags { if strings.Contains(strings.ToLower(tag), q) { rs = append(rs, t); break } }
	}
	return rs, nil
}

func copyTool(t *types.ToolSpec) *types.ToolSpec {
	if t == nil { return nil }
	c := *t; c.Parameters = make([]types.ToolParameter, len(t.Parameters)); copy(c.Parameters, t.Parameters)
	c.Tags = make([]string, len(t.Tags)); copy(c.Tags, t.Tags); return &c
}

type RuleToolSelector struct {
	match func(*types.RuleContext) []types.RuleOutput
}

func NewRuleToolSelector(m func(*types.RuleContext) []types.RuleOutput) *RuleToolSelector {
	return &RuleToolSelector{match: m}
}

func (s *RuleToolSelector) Select(ctx context.Context, intent *types.IntentClassification, tools []types.ToolSpec) ([]types.ToolSpec, error) {
	if s == nil || len(tools)==0 || intent==nil { return tools, nil }
	rs := make([]types.ToolSpec, 0, len(tools))
	for _, t := range tools { if matchesToolIntent(&t, intent) { rs = append(rs, t) } }
	if len(rs)==0 { return tools, nil }
	return rs, nil
}

func matchesToolIntent(t *types.ToolSpec, i *types.IntentClassification) bool {
	switch i.Type {
	case types.IntentRecall:
		for _, tag := range t.Tags { if tag=="memory"||tag=="search"||tag=="query" { return true } }
	case types.IntentTask:
		for _, tag := range t.Tags { if tag=="action"||tag=="execute"||tag=="write" { return true } }
	case types.IntentFactual:
		for _, tag := range t.Tags { if tag=="read"||tag=="get"||tag=="search" { return true } }
	}
	return true
}
