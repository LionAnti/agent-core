package agentcore

import "context"

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
		return nil, ErrInvalidConfig
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

	re := NewRulesEngine()
	if cfg.RuleStore != nil {
		userRules, err := cfg.RuleStore.LoadRules(context.Background())
		if err != nil {
			cfg.Logger.Warn("failed to load user rules", "error", err)
		} else {
			for i := range userRules {
				re.AddRule(userRules[i])
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
	}, nil
}

func (c *AgentCore) NewSession(tenantID, userID string, opts ...SessionOption) *Session {
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
	return s
}

func (c *AgentCore) Provider() *ProviderManager { return c.provider }
func (c *AgentCore) Registry() *ToolRegistry     { return c.registry }
func (c *AgentCore) Rules() *RulesEngine         { return c.rules }

func (c *AgentCore) HarnessState(currentTokens int) *HarnessState {
	return &HarnessState{
		CurrentTokens: currentTokens,
		ContextWindow: c.config.ContextWindow,
	}
}
