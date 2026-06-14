package agentcore

import (
	"context"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type ProviderManager struct {
	store  ProviderStore
	logger Logger
	mu     sync.RWMutex
	active []*LLMProvider
	rng    *rand.Rand
}

func NewProviderManager(store ProviderStore, logger Logger) *ProviderManager {
	pm := &ProviderManager{
		store:  store,
		logger: logger,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	pm.refresh()
	return pm
}

func (pm *ProviderManager) refresh() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	providers, err := pm.store.ListProviders(ctx)
	if err != nil {
		pm.logger.Warn("failed to list providers", "error", err)
		return
	}
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.active = make([]*LLMProvider, 0, len(providers))
	for _, p := range providers {
		if p.Enabled {
			cp := *p
			pm.active = append(pm.active, &cp)
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
				cp := *p
				return &cp
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
	if totalWeight <= 0 {
		cp := *pm.active[0]
		return &cp
	}

	roll := pm.rng.Intn(totalWeight)
	cumulative := 0
	for _, p := range pm.active {
		cumulative += p.Weight
		if roll < cumulative {
			cp := *p
			return &cp
		}
	}
	cp := *pm.active[len(pm.active)-1]
	return &cp
}

func (pm *ProviderManager) All() []*LLMProvider {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	result := make([]*LLMProvider, len(pm.active))
	for i, p := range pm.active {
		cp := *p
		result[i] = &cp
	}
	return result
}
