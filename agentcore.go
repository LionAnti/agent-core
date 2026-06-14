package agentcore

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/LionAnti/agent-core/compressor"
	"github.com/LionAnti/agent-core/harness"
	"github.com/LionAnti/agent-core/memory"
	"github.com/LionAnti/agent-core/provider"
	"github.com/LionAnti/agent-core/registry"
	"github.com/LionAnti/agent-core/rules"
	"github.com/LionAnti/agent-core/types"
)

type AgentCore struct {
	config           Config
	intentClassifier types.IntentClassifier
	toolSelector     types.ToolSelector
	densityEstimator types.DensityEstimator
	offloadDecider   types.OffloadDecider
	memoryRecall     types.MemoryRecall
	compressor       types.Compressor
	rules            *rules.Engine
	gtregistry       *registry.ToolRegistry
	provider         *provider.Manager
	llmClient        types.LLMClient
	memStore         types.MemoryStore
	vecStore         types.VectorStore
	provStore        types.ProviderStore
	regStore         types.RegistryStore
	l0Store          types.L0Store
	logger           types.Logger
	metrics          types.MetricsCollector
	sessionMu        sync.RWMutex
	sessions         map[string]*sessionInternal
	closed           bool
}

type Config struct {
	LLMClient           types.LLMClient
	ProviderStore       types.ProviderStore
	RuleStore           types.RuleStore
	RegistryStore       types.RegistryStore
	MemoryStore         types.MemoryStore
	L0Store             types.L0Store
	VectorStore         types.VectorStore
	Logger              types.Logger
	MetricsCollector    types.MetricsCollector
	ContextWindow       int
	MildThreshold       float64
	AggressiveThreshold float64
	IntentClassifier    types.IntentClassifier
	ToolSelector        types.ToolSelector
	DensityEstimator    types.DensityEstimator
	OffloadDecider      types.OffloadDecider
	MemoryRecall        types.MemoryRecall
}

func New(cfg Config) (*AgentCore, error) {
	if cfg.LLMClient == nil { return nil, fmt.Errorf("agentcore: %w: LLMClient required", types.ErrInvalidConfig) }
	if cfg.ProviderStore == nil { return nil, fmt.Errorf("agentcore: %w: ProviderStore required", types.ErrInvalidConfig) }
	if cfg.RegistryStore == nil { return nil, fmt.Errorf("agentcore: %w: RegistryStore required", types.ErrInvalidConfig) }
	if cfg.MemoryStore == nil { return nil, fmt.Errorf("agentcore: %w: MemoryStore required", types.ErrInvalidConfig) }
	if cfg.Logger == nil { cfg.Logger = &noopLogger{} }
	if cfg.MetricsCollector == nil { cfg.MetricsCollector = &noopMetricsCollector{} }
	if cfg.ContextWindow <= 0 { cfg.ContextWindow = 128000 }
	if cfg.MildThreshold <= 0 { cfg.MildThreshold = 0.5 }
	if cfg.AggressiveThreshold <= 0 { cfg.AggressiveThreshold = 0.85 }
	if cfg.MildThreshold >= cfg.AggressiveThreshold { return nil, fmt.Errorf("agentcore: thresholds invalid: %0.2f >= %0.2f", cfg.MildThreshold, cfg.AggressiveThreshold) }

	re := rules.NewEngine()
	if cfg.RuleStore != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		ur, err := cfg.RuleStore.LoadRules(ctx); cancel()
		if err != nil { cfg.Logger.Warn("failed to load rules", "error", err) } else {
			for i := range ur { if ur[i]==nil { continue }; if ae := re.AddRule(ur[i]); ae!=nil { cfg.Logger.Warn("add rule", "error", ae) } }
		}
	}
	pm := provider.NewManager(cfg.ProviderStore, cfg.Logger)
	tr := registry.NewToolRegistry(cfg.RegistryStore, cfg.Logger)
	comp := compressor.NewEngine(cfg.LLMClient, cfg.Logger)

	ic := cfg.IntentClassifier; if ic == nil { ic = harness.NewClassifier(re.Match, cfg.LLMClient) }
	ts := cfg.ToolSelector; if ts == nil { ts = registry.NewRuleToolSelector(re.Match) }
	de := cfg.DensityEstimator; if de == nil { de = harness.NewDensityEstimator() }
	od := cfg.OffloadDecider; if od == nil { od = harness.NewOffloadDecider(cfg.MildThreshold, cfg.AggressiveThreshold) }
	mr := cfg.MemoryRecall; if mr == nil { mr = memory.NewRecall(cfg.MemoryStore, cfg.VectorStore, cfg.Logger) }

	return &AgentCore{
		config: cfg, intentClassifier: ic, toolSelector: ts, densityEstimator: de,
		offloadDecider: od, memoryRecall: mr, compressor: comp, rules: re,
		gtregistry: tr, provider: pm, llmClient: cfg.LLMClient,
		memStore: cfg.MemoryStore, vecStore: cfg.VectorStore, provStore: cfg.ProviderStore,
		regStore: cfg.RegistryStore, l0Store: cfg.L0Store,
		logger: cfg.Logger, metrics: cfg.MetricsCollector,
		sessions: make(map[string]*sessionInternal),
	}, nil
}

func (c *AgentCore) NewSession(tenantID, userID string, opts ...SessionOption) (*types.Session, error) {
	if c == nil { return nil, types.ErrAgentClosed }
	s := &types.Session{ID: newID(), TenantID: tenantID, UserID: userID, Status: types.SessionActive, CreatedAt: now()}
	for _, o := range opts { o(s) }
	si, err := c.openSession(s)
	if err != nil { return nil, err }
	s.Internal = si
	return s, nil
}

func (c *AgentCore) Provider() *provider.Manager     { if c==nil { return nil }; return c.provider }
func (c *AgentCore) Registry() *registry.ToolRegistry { if c==nil { return nil }; return c.gtregistry }
func (c *AgentCore) Rules() *rules.Engine             { if c==nil { return nil }; return c.rules }

func (c *AgentCore) HarnessState(ct int) *types.HarnessState {
	if c==nil { return nil }
	return &types.HarnessState{CurrentTokens: ct, ContextWindow: c.config.ContextWindow}
}

func (c *AgentCore) Close(ctx context.Context) error {
	if c==nil { return nil }
	c.sessionMu.Lock(); c.closed = true
	ss := make([]*sessionInternal, 0, len(c.sessions))
	for _, si := range c.sessions { ss = append(ss, si) }
	c.sessionMu.Unlock()
	for _, si := range ss { select { case <-ctx.Done(): return ctx.Err(); default: }; si.Close() }
	if c.logger != nil { c.logger.Info("agentcore closed", "count", len(ss)) }
	return nil
}

func (c *AgentCore) HealthCheck(ctx context.Context) error {
	if c==nil||c.closed { return types.ErrAgentClosed }
	if c.llmClient==nil { return fmt.Errorf("health: %w: LLM nil", types.ErrInvalidConfig) }
	return nil
}

func (c *AgentCore) ActiveSessionCount() int {
	if c==nil { return 0 }
	c.sessionMu.RLock(); defer c.sessionMu.RUnlock()
	return len(c.sessions)
}
