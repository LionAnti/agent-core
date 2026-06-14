package agentcore

import (
	"context"
	"fmt"
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
	if err := pm.refresh(); err != nil {
		logger.Warn("initial provider refresh failed, will retry on next Select", "error", err)
	}
	return pm
}

func (pm *ProviderManager) refresh() error {
	if pm == nil {
		return fmt.Errorf("provider manager: nil receiver")
	}
	if pm.store == nil {
		return fmt.Errorf("provider manager: %w: store is nil", ErrInvalidConfig)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	providers, err := pm.store.ListProviders(ctx)
	if err != nil {
		return fmt.Errorf("list providers: %w", err)
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
	return nil
}

func (pm *ProviderManager) RefreshProviders(ctx context.Context) error {
	if pm == nil {
		return fmt.Errorf("provider manager: nil receiver")
	}
	if pm.store == nil {
		return fmt.Errorf("provider manager: %w: store is nil", ErrInvalidConfig)
	}
	providers, err := pm.store.ListProviders(ctx)
	if err != nil {
		return fmt.Errorf("list providers: %w", err)
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
	return nil
}

func (pm *ProviderManager) Select(ctx context.Context, preferredID string) (*LLMProvider, error) {
	if pm == nil {
		return nil, fmt.Errorf("provider manager: %w", ErrNoAvailableProviders)
	}
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if preferredID != "" {
		for _, p := range pm.active {
			if p.ID == preferredID {
				cp := *p
				return &cp, nil
			}
		}
	}
	if len(pm.active) == 0 {
		return nil, fmt.Errorf("provider: %w", ErrNoAvailableProviders)
	}
	totalWeight := 0
	for _, p := range pm.active {
		totalWeight += p.Weight
	}
	if totalWeight <= 0 {
		cp := *pm.active[0]
		return &cp, nil
	}
	roll := pm.rng.Intn(totalWeight)
	cumulative := 0
	for _, p := range pm.active {
		cumulative += p.Weight
		if roll < cumulative {
			cp := *p
			return &cp, nil
		}
	}
	cp := *pm.active[len(pm.active)-1]
	return &cp, nil
}

func (pm *ProviderManager) All() []*LLMProvider {
	if pm == nil {
		return nil
	}
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	result := make([]*LLMProvider, len(pm.active))
	for i, p := range pm.active {
		cp := *p
		result[i] = &cp
	}
	return result
}
