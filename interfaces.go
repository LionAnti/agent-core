package agentcore

import "context"

// IntentClassifier determines the user's intent from their message and session state.
type IntentClassifier interface {
	Classify(ctx context.Context, msgs []Message, state *HarnessState) (*IntentClassification, error)
}

// ToolSelector picks the most relevant tools based on the classified intent.
type ToolSelector interface {
	Select(ctx context.Context, intent *IntentClassification, tools []ToolSpec) ([]ToolSpec, error)
}

// DensityEstimator measures how densely the context window is being utilized.
type DensityEstimator interface {
	Estimate(ctx context.Context, msgs []Message, state *HarnessState) (*DensitySignals, error)
}

// OffloadDecider determines whether to compress the context based on density signals.
type OffloadDecider interface {
	Decide(ctx context.Context, density *DensitySignals, state *HarnessState) (*OffloadDecision, error)
}

// UtilityTracker scores memory records by access frequency and recency decay.
type UtilityTracker interface {
	Score(ctx context.Context, recordID string, accessCount int, lastAccessAt int64) (*UtilityScore, error)
	Decay(ctx context.Context, scores []UtilityScore, daysSince float64) ([]UtilityScore, error)
}

// Compressor selects and applies a compression strategy to reduce context size.
type Compressor interface {
	Compress(ctx context.Context, msgs []Message, decision *CompressionDecision) (*CompressionResult, error)
	ShouldCompress(ctx context.Context, state *HarnessState) *CompressionDecision
}

// LLMClient wraps an LLM provider (OpenAI, Claude, DeepSeek, etc.) for synchronous chat.
type LLMClient interface {
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}

// StreamLLMClient optionally extends LLMClient with streaming support.
type StreamLLMClient interface {
	LLMClient
	// StreamChat sends a chat request and receives tokens one-by-one via the callback.
	// The callback is called for each text fragment. Returns the final assembled response.
	StreamChat(ctx context.Context, req *ChatRequest, onToken func(fragment string)) (*ChatResponse, error)
}

// MemoryStore persists L1 (facts), L2 (scenes), and L3 (persona) memory.
type MemoryStore interface {
	SaveL1(ctx context.Context, mem *L1Memory) error
	SearchL1(ctx context.Context, tenantID, userID, query string, limit int) ([]*L1Memory, error)
	GetL1ByID(ctx context.Context, id string) (*L1Memory, error)
	UpdateL1Utility(ctx context.Context, id string, score float64) error
	IncrementL1Recall(ctx context.Context, id string) error
	CountActiveL1(ctx context.Context, tenantID, userID string) (int, error)
	DeleteL1(ctx context.Context, id string) error
	ArchiveL1(ctx context.Context, id string, archived bool) error
	SaveL2(ctx context.Context, scene *L2Scene) error
	GetL2ByUser(ctx context.Context, tenantID, userID string) ([]*L2Scene, error)
	IncrementL2Heat(ctx context.Context, id string) error
	DeleteL2(ctx context.Context, id string) error
	SaveL3(ctx context.Context, persona *L3Persona) error
	GetCurrentL3(ctx context.Context, tenantID, userID string) (*L3Persona, error)
	ListL3Versions(ctx context.Context, tenantID, userID string, limit int) ([]*L3Persona, error)
}

// VectorStore enables semantic search over memory embeddings.
type VectorStore interface {
	Search(ctx context.Context, collection string, vector []float32, limit int) ([]VectorResult, error)
	Upsert(ctx context.Context, collection string, points []VectorPoint) error
	Delete(ctx context.Context, collection string, ids []string) error
}

// VectorResult is a single match from a vector similarity search.
type VectorResult struct {
	ID      string            `json:"id"`
	Score   float64           `json:"score"`
	Payload map[string]string `json:"payload,omitempty"`
}

// VectorPoint is a single vector to upsert into the vector store.
type VectorPoint struct {
	ID      string            `json:"id"`
	Vector  []float32         `json:"vector"`
	Payload map[string]string `json:"payload,omitempty"`
}

// MemoryRecall retrieves relevant L1-L3 memory for the current context.
type MemoryRecall interface {
	Recall(ctx context.Context, intent *IntentClassification, tenantID, userID string) (*RecallResult, error)
}

// ProviderStore persists LLM provider configurations.
type ProviderStore interface {
	ListProviders(ctx context.Context) ([]*LLMProvider, error)
	GetProvider(ctx context.Context, id string) (*LLMProvider, error)
	CreateProvider(ctx context.Context, p *LLMProvider) error
	UpdateProvider(ctx context.Context, p *LLMProvider) error
	DeleteProvider(ctx context.Context, id string) error
}

// RuleStore persists user-defined rules; built-in rules are compiled in.
type RuleStore interface {
	LoadRules(ctx context.Context) ([]*Rule, error)
	SaveRule(ctx context.Context, rule *Rule) error
	DeleteRule(ctx context.Context, name string) error
}

// RegistryStore provides KV storage for tool registrations.
type RegistryStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, data []byte) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([][]byte, error)
}

// Logger is the interface users implement for structured logging.
type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
}

// MetricsCollector records performance metrics (optional, noop by default).
type MetricsCollector interface {
	RecordLatency(name string, ms float64)
	RecordTokenUsage(prompt, completion int)
	IncrementCounter(name string, labels ...string)
	SetGauge(name string, value float64, labels ...string)
}

// L0Store persists raw conversation records used for L1 extraction.
type L0Store interface {
	Save(ctx context.Context, records []L0Record) error
	GetBySession(ctx context.Context, sessionKey string, limit, offset int) ([]L0Record, error)
}
