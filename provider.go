package agentcore

import (
    "context"
    "math/rand"
    "sort"
    "sync"
)

type ProviderManager struct {
    store  ProviderStore
    logger Logger
    mu     sync.RWMutex
    active []*LLMProvider
}

func NewProviderManager(store ProviderStore, logger Logger) *ProviderManager {
    pm := &ProviderManager{store: store, logger: logger}
    pm.refresh()
    return pm
}

func (pm *ProviderManager) refresh() {
    providers, err := pm.store.ListProviders(context.Background())
    if err != nil {
        pm.logger.Warn("failed to list providers", "error", err)
        return
    }
    pm.mu.Lock()
    defer pm.mu.Unlock()
    pm.active = make([]*LLMProvider, 0, len(providers))
    for _, p := range providers {
        if p.Enabled {
            pm.active = append(pm.active, p)
        }
    }
    sort.Slice(pm.active, func(i, j int) bool {
        return pm.active[i].Priority < pm.active[j].Priority
    })
}

func (pm *ProviderManager) Select(ctx context.Context, preferredID string) *LLMProvider {
    pm.mu.RLock()
    defer pm.mu.RUnlock()

    if preferredID != "" {
        for _, p := range pm.active {
            if p.ID == preferredID {
                return p
            }
        }
    }

    if len(pm.active) == 0 {
        return nil
    }

    totalWeight := 0
    for _, p := range pm.active {
        totalWeight += p.Weight
    }
    if totalWeight == 0 {
        return pm.active[0]
    }

    roll := rand.Intn(totalWeight)
    cumulative := 0
    for _, p := range pm.active {
        cumulative += p.Weight
        if roll < cumulative {
            return p
        }
    }
    return pm.active[len(pm.active)-1]
}

func (pm *ProviderManager) All() []*LLMProvider {
    pm.mu.RLock()
    defer pm.mu.RUnlock()
    result := make([]*LLMProvider, len(pm.active))
    copy(result, pm.active)
    return result
}
