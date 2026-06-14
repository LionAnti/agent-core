package agentcore

import "context"

type IntentClassifier interface {
    Classify(ctx context.Context, msgs []Message, state *HarnessState) (*IntentClassification, error)
}

type ToolSelector interface {
    Select(ctx context.Context, intent *IntentClassification, tools []ToolSpec) ([]ToolSpec, error)
}

type DensityEstimator interface {
    Estimate(ctx context.Context, msgs []Message, state *HarnessState) (*DensitySignals, error)
}

type OffloadDecider interface {
    Decide(ctx context.Context, density *DensitySignals, state *HarnessState) (*OffloadDecision, error)
}

type UtilityTracker interface {
    Score(ctx context.Context, recordID string, accessCount int, lastAccessAt int64) (*UtilityScore, error)
    Decay(ctx context.Context, scores []UtilityScore, daysSince float64) ([]UtilityScore, error)
}

type Compressor interface {
    Compress(ctx context.Context, msgs []Message, decision *CompressionDecision) (*CompressionResult, error)
    ShouldCompress(state *HarnessState) *CompressionDecision
}

type LLMClient interface {
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}

type MemoryStore interface {
    SaveL1(ctx context.Context, mem *L1Memory) error
    SearchL1(ctx context.Context, tenantID, userID, query string, limit int) ([]*L1Memory, error)
    GetL1ByID(ctx context.Context, id string) (*L1Memory, error)
    UpdateL1Utility(ctx context.Context, id string, score float64) error
    IncrementL1Recall(ctx context.Context, id string) error
    CountActiveL1(ctx context.Context, tenantID, userID string) (int, error)
    SaveL2(ctx context.Context, scene *L2Scene) error
    GetL2ByUser(ctx context.Context, tenantID, userID string) ([]*L2Scene, error)
    IncrementL2Heat(ctx context.Context, id string) error
    SaveL3(ctx context.Context, persona *L3Persona) error
    GetCurrentL3(ctx context.Context, tenantID, userID string) (*L3Persona, error)
    ListL3Versions(ctx context.Context, tenantID, userID string, limit int) ([]*L3Persona, error)
}

type VectorStore interface {
    Search(ctx context.Context, collection string, vector []float32, limit int) ([]VectorResult, error)
    Upsert(ctx context.Context, collection string, points []VectorPoint) error
    Delete(ctx context.Context, collection string, ids []string) error
}

type VectorResult struct {
    ID      string            `json:"id"`
    Score   float64           `json:"score"`
    Payload map[string]string `json:"payload,omitempty"`
}

type VectorPoint struct {
    ID      string            `json:"id"`
    Vector  []float32         `json:"vector"`
    Payload map[string]string `json:"payload,omitempty"`
}

type MemoryRecall interface {
    Recall(ctx context.Context, intent *IntentClassification, tenantID, userID string) (*RecallResult, error)
}

type ProviderStore interface {
    ListProviders(ctx context.Context) ([]*LLMProvider, error)
    GetProvider(ctx context.Context, id string) (*LLMProvider, error)
    CreateProvider(ctx context.Context, p *LLMProvider) error
    UpdateProvider(ctx context.Context, p *LLMProvider) error
    DeleteProvider(ctx context.Context, id string) error
}

type RuleStore interface {
    LoadRules(ctx context.Context) ([]*Rule, error)
    SaveRule(ctx context.Context, rule *Rule) error
    DeleteRule(ctx context.Context, name string) error
}

type RegistryStore interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Put(ctx context.Context, key string, data []byte) error
    Delete(ctx context.Context, key string) error
    List(ctx context.Context, prefix string) ([][]byte, error)
}

type Logger interface {
    Debug(msg string, keysAndValues ...any)
    Info(msg string, keysAndValues ...any)
    Warn(msg string, keysAndValues ...any)
    Error(msg string, keysAndValues ...any)
}

type MetricsCollector interface {
    RecordLatency(name string, ms float64)
    RecordTokenUsage(prompt, completion int)
    IncrementCounter(name string, labels ...string)
    SetGauge(name string, value float64, labels ...string)
}

type L0Store interface {
    Save(ctx context.Context, records []L0Record) error
    GetBySession(ctx context.Context, sessionKey string, limit, offset int) ([]L0Record, error)
}
