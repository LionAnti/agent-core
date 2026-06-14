package agentcore

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type AgentCore struct {
	config           Config
	intentClassifier IntentClassifier
	toolSelector     ToolSelector
	densityEstimator DensityEstimator
	offloadDecider   OffloadDecider
	memoryRecall     MemoryRecall
	compressor       Compressor
	rules            *RulesEngine
	registry         *ToolRegistry
	provider         *ProviderManager
	llmClient        LLMClient
	memStore         MemoryStore
	vecStore         VectorStore
	provStore        ProviderStore
	regStore         RegistryStore
	l0Store          L0Store
	logger           Logger
	metrics          MetricsCollector
	sessionMu        sync.RWMutex
	sessions         map[string]*sessionInternal
	closed           bool
}

type Config struct {
	LLMClient           LLMClient
	ProviderStore       ProviderStore
	RuleStore           RuleStore
	RegistryStore       RegistryStore
	MemoryStore         MemoryStore
	L0Store             L0Store
	VectorStore         VectorStore
	Logger              Logger
	MetricsCollector    MetricsCollector
	ContextWindow       int
	MildThreshold       float64
	AggressiveThreshold float64
	IntentClassifier    IntentClassifier
	ToolSelector        ToolSelector
	DensityEstimator    DensityEstimator
	OffloadDecider      OffloadDecider
	MemoryRecall        MemoryRecall
}

func New(cfg Config) (*AgentCore, error) {
	if cfg.LLMClient == nil {
		return nil, fmt.Errorf("agentcore: %w: LLMClient is required", ErrInvalidConfig)
	}
	if cfg.ProviderStore == nil {
		return nil, fmt.Errorf("agentcore: %w: ProviderStore is required", ErrInvalidConfig)
	}
	if cfg.RegistryStore == nil {
		return nil, fmt.Errorf("agentcore: %w: RegistryStore is required", ErrInvalidConfig)
	}
	if cfg.MemoryStore == nil {
		return nil, fmt.Errorf("agentcore: %w: MemoryStore is required", ErrInvalidConfig)
	}
	if cfg.Logger == nil {
		cfg.Logger = NoopLogger{}
	}
	if cfg.MetricsCollector == nil {
		cfg.MetricsCollector = NoopMetricsCollector{}
	}
	if cfg.ContextWindow <= 0 {
		cfg.ContextWindow = 128000
	}
	if cfg.MildThreshold <= 0 {
		cfg.MildThreshold = 0.5
	}
	if cfg.AggressiveThreshold <= 0 {
		cfg.AggressiveThreshold = 0.85
	}
	if cfg.MildThreshold >= cfg.AggressiveThreshold {
		return nil, fmt.Errorf("agentcore: %w: MildThreshold (%0.2f) must be less than AggressiveThreshold (%0.2f)", ErrInvalidConfig, cfg.MildThreshold, cfg.AggressiveThreshold)
	}

	re := NewRulesEngine()
	if cfg.RuleStore != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		userRules, err := cfg.RuleStore.LoadRules(ctx)
		cancel()
		if err != nil {
			cfg.Logger.Warn("failed to load user rules", "error", err)
		} else {
			for i := range userRules {
				if userRules[i] == nil {
					cfg.Logger.Warn("skipping nil rule from store")
					continue
				}
				if addErr := re.AddRule(userRules[i]); addErr != nil {
					cfg.Logger.Warn("failed to add rule", "name", userRules[i].Name, "error", addErr)
				}
			}
		}
	}

	pm := NewProviderManager(cfg.ProviderStore, cfg.Logger)
	tr := NewToolRegistry(cfg.RegistryStore, cfg.Logger)
	comp := NewCompressor(cfg.LLMClient, cfg.Logger)

	ic := cfg.IntentClassifier
	if ic == nil {
		ic = NewHybridClassifier(re, cfg.LLMClient)
	}
	ts := cfg.ToolSelector
	if ts == nil {
		ts = NewRuleToolSelector(re)
	}
	de := cfg.DensityEstimator
	if de == nil {
		de = NewDefaultDensityEstimator()
	}
	od := cfg.OffloadDecider
	if od == nil {
		od = NewDefaultOffloadDecider(cfg.MildThreshold, cfg.AggressiveThreshold)
	}
	mr := cfg.MemoryRecall
	if mr == nil {
		mr = NewDefaultMemoryRecall(cfg.MemoryStore, cfg.VectorStore, cfg.Logger)
	}

	return &AgentCore{
		config:           cfg,
		intentClassifier: ic,
		toolSelector:     ts,
		densityEstimator: de,
		offloadDecider:   od,
		memoryRecall:     mr,
		compressor:       comp,
		rules:            re,
		registry:         tr,
		provider:         pm,
		llmClient:        cfg.LLMClient,
		memStore:         cfg.MemoryStore,
		vecStore:         cfg.VectorStore,
		provStore:        cfg.ProviderStore,
		regStore:         cfg.RegistryStore,
		l0Store:          cfg.L0Store,
		logger:           cfg.Logger,
		metrics:          cfg.MetricsCollector,
		sessions:         make(map[string]*sessionInternal),
	}, nil
}

func (c *AgentCore) NewSession(tenantID, userID string, opts ...SessionOption) (*Session, error) {
	if c == nil {
		return nil, ErrAgentClosed
	}
	s := &Session{
		ID:        newID(),
		TenantID:  tenantID,
		UserID:    userID,
		Status:    SessionActive,
		CreatedAt: now(),
	}
	for _, opt := range opts {
		opt(s)
	}
	si, err := c.openSession(s)
	if err != nil {
		return nil, err
	}
	s.internal = si
	return s, nil
}

func (c *AgentCore) Provider() *ProviderManager {
	if c == nil { return nil }
	return c.provider
}

func (c *AgentCore) Registry() *ToolRegistry {
	if c == nil { return nil }
	return c.registry
}

func (c *AgentCore) Rules() *RulesEngine {
	if c == nil { return nil }
	return c.rules
}

func (c *AgentCore) HarnessState(currentTokens int) *HarnessState {
	if c == nil { return nil }
	return &HarnessState{
		CurrentTokens: currentTokens,
		ContextWindow: c.config.ContextWindow,
	}
}

func (c *AgentCore) Close(ctx context.Context) error {
	if c == nil {
		return nil
	}
	c.sessionMu.Lock()
	c.closed = true
	sessions := make([]*sessionInternal, 0, len(c.sessions))
	for _, si := range c.sessions {
		sessions = append(sessions, si)
	}
	c.sessionMu.Unlock()

	for _, si := range sessions {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		si.Close()
	}
	if c.logger != nil {
		c.logger.Info("agentcore closed", "sessions_closed", len(sessions))
	}
	return nil
}

func (c *AgentCore) HealthCheck(ctx context.Context) error {
	if c == nil || c.closed {
		return ErrAgentClosed
	}
	if c.llmClient == nil {
		return fmt.Errorf("agentcore: %w: LLMClient is nil", ErrInvalidConfig)
	}
	if c.provider == nil {
		return fmt.Errorf("agentcore: %w: ProviderManager is nil", ErrInvalidConfig)
	}
	return nil
}

func (c *AgentCore) ActiveSessionCount() int {
	if c == nil { return 0 }
	c.sessionMu.RLock()
	defer c.sessionMu.RUnlock()
	return len(c.sessions)
}
