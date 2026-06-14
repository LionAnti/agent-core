package agentcore

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    "sync"
)

type ToolRegistry struct {
    store  RegistryStore
    logger Logger
    mu     sync.RWMutex
    cache  map[string]*ToolSpec
}

func NewToolRegistry(store RegistryStore, logger Logger) *ToolRegistry {
    return &ToolRegistry{
        store:  store,
        logger: logger,
        cache:  make(map[string]*ToolSpec),
    }
}

func (r *ToolRegistry) Register(ctx context.Context, tool *ToolSpec) error {
    key := fmt.Sprintf("tools/%s/%s", tool.TenantID, tool.Name)
    data, err := json.Marshal(tool)
    if err != nil {
        return fmt.Errorf("marshal tool: %w", err)
    }
    if err := r.store.Put(ctx, key, data); err != nil {
        return fmt.Errorf("store tool: %w", err)
    }
    r.mu.Lock()
    r.cache[key] = tool
    r.mu.Unlock()
    return nil
}

func (r *ToolRegistry) Unregister(ctx context.Context, tenantID, name string) error {
    key := fmt.Sprintf("tools/%s/%s", tenantID, name)
    if err := r.store.Delete(ctx, key); err != nil {
        return fmt.Errorf("delete tool: %w", err)
    }
    r.mu.Lock()
    delete(r.cache, key)
    r.mu.Unlock()
    return nil
}

func (r *ToolRegistry) Get(ctx context.Context, tenantID, name string) (*ToolSpec, error) {
    key := fmt.Sprintf("tools/%s/%s", tenantID, name)
    r.mu.RLock()
    cached, ok := r.cache[key]
    r.mu.RUnlock()
    if ok {
        return cached, nil
    }
    data, err := r.store.Get(ctx, key)
    if err != nil {
        return nil, fmt.Errorf("get tool: %w", err)
    }
    var tool ToolSpec
    if err := json.Unmarshal(data, &tool); err != nil {
        return nil, fmt.Errorf("unmarshal tool: %w", err)
    }
    r.mu.Lock()
    r.cache[key] = &tool
    r.mu.Unlock()
    return &tool, nil
}

func (r *ToolRegistry) List(ctx context.Context, tenantID string) ([]ToolSpec, error) {
    prefix := fmt.Sprintf("tools/%s/", tenantID)
    dataList, err := r.store.List(ctx, prefix)
    if err != nil {
        return nil, fmt.Errorf("list tools: %w", err)
    }
    tools := make([]ToolSpec, 0, len(dataList))
    for _, data := range dataList {
        var t ToolSpec
        if err := json.Unmarshal(data, &t); err != nil {
            r.logger.Warn("unmarshal tool", "error", err)
            continue
        }
        if t.Status == ToolStatusEnabled {
            tools = append(tools, t)
        }
    }
    return tools, nil
}

func (r *ToolRegistry) Search(ctx context.Context, tenantID, query string) ([]ToolSpec, error) {
    tools, err := r.List(ctx, tenantID)
    if err != nil {
        return nil, err
    }
    q := strings.ToLower(query)
    var results []ToolSpec
    for _, t := range tools {
        if strings.Contains(strings.ToLower(t.Name), q) ||
            strings.Contains(strings.ToLower(t.Description), q) {
            results = append(results, t)
            continue
        }
        for _, tag := range t.Tags {
            if strings.Contains(strings.ToLower(tag), q) {
                results = append(results, t)
                break
            }
        }
    }
    return results, nil
}

// ─── Tool Selector ───
type RuleToolSelector struct {
    engine *RulesEngine
}

func NewRuleToolSelector(engine *RulesEngine) *RuleToolSelector {
    return &RuleToolSelector{engine: engine}
}

func (s *RuleToolSelector) Select(ctx context.Context, intent *IntentClassification, tools []ToolSpec) ([]ToolSpec, error) {
    if len(tools) == 0 || intent == nil {
        return tools, nil
    }
    ctx2 := &RuleContext{Domain: DomainToolSelection, Input: string(intent.Type)}
    _ = s.engine.Match(ctx2)

    var results []ToolSpec
    for _, t := range tools {
        t := t
        if matchesToolIntent(&t, intent) {
            results = append(results, t)
        }
    }
    if len(results) == 0 {
        return tools, nil
    }
    return results, nil
}

func matchesToolIntent(tool *ToolSpec, intent *IntentClassification) bool {
    switch intent.Type {
    case IntentRecall:
        for _, tag := range tool.Tags {
            if tag == "memory" || tag == "search" || tag == "query" {
                return true
            }
        }
    case IntentTask:
        for _, tag := range tool.Tags {
            if tag == "action" || tag == "execute" || tag == "write" {
                return true
            }
        }
    case IntentFactual:
        for _, tag := range tool.Tags {
            if tag == "read" || tag == "get" || tag == "search" {
                return true
            }
        }
    }
    return true
}
