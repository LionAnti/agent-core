package provider

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
	"github.com/LionAnti/agent-core/types"
)

type Manager struct {
	store  types.ProviderStore
	logger types.Logger
	mu     sync.RWMutex
	active []*types.LLMProvider
	rng    *rand.Rand
}

func NewManager(store types.ProviderStore, logger types.Logger) *Manager {
	pm := &Manager{store: store, logger: logger, rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
	if err := pm.refresh(); err != nil { logger.Warn("initial provider refresh failed", "error", err) }
	return pm
}

func (pm *Manager) refresh() error {
	if pm == nil { return fmt.Errorf("provider manager: nil receiver") }
	if pm.store == nil { return fmt.Errorf("provider: %w: store nil", types.ErrInvalidConfig) }
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	providers, err := pm.store.ListProviders(ctx)
	if err != nil { return fmt.Errorf("list providers: %w", err) }
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.active = make([]*types.LLMProvider, 0, len(providers))
	for _, p := range providers {
		if p.Enabled { cp := *p; pm.active = append(pm.active, &cp) }
	}
	sort.Slice(pm.active, func(i, j int) bool { return pm.active[i].Priority < pm.active[j].Priority })
	return nil
}

func (pm *Manager) RefreshProviders(ctx context.Context) error {
	if pm == nil { return fmt.Errorf("provider: nil receiver") }
	if pm.store == nil { return fmt.Errorf("provider: %w: store nil", types.ErrInvalidConfig) }
	providers, err := pm.store.ListProviders(ctx)
	if err != nil { return fmt.Errorf("list providers: %w", err) }
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.active = make([]*types.LLMProvider, 0, len(providers))
	for _, p := range providers {
		if p.Enabled { cp := *p; pm.active = append(pm.active, &cp) }
	}
	sort.Slice(pm.active, func(i, j int) bool { return pm.active[i].Priority < pm.active[j].Priority })
	return nil
}

func (pm *Manager) Select(ctx context.Context, preferredID string) (*types.LLMProvider, error) {
	if pm == nil { return nil, fmt.Errorf("provider: %w", types.ErrNoAvailableProviders) }
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	if preferredID != "" {
		for _, p := range pm.active { if p.ID == preferredID { cp := *p; return &cp, nil } }
	}
	if len(pm.active) == 0 { return nil, fmt.Errorf("provider: %w", types.ErrNoAvailableProviders) }
	tw := 0
	for _, p := range pm.active { tw += p.Weight }
	if tw <= 0 { cp := *pm.active[0]; return &cp, nil }
	roll := pm.rng.Intn(tw)
	cum := 0
	for _, p := range pm.active { cum += p.Weight; if roll < cum { cp := *p; return &cp, nil } }
	cp := *pm.active[len(pm.active)-1]
	return &cp, nil
}

func (pm *Manager) All() []*types.LLMProvider {
	if pm == nil { return nil }
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	r := make([]*types.LLMProvider, len(pm.active))
	for i, p := range pm.active { cp := *p; r[i] = &cp }
	return r
}
