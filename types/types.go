package types

import "context"

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

type Message struct {
	Role string `json:"Role"`
	Content string `json:"Content"`
	TokenCount int `json:"TokenCount,omitempty"`
	RecordedAt int64 `json:"RecordedAt,omitempty"`
}

type IntentType string

const (
	IntentFactual     IntentType = "factual"
	IntentExploratory IntentType = "exploratory"
	IntentRecall      IntentType = "recall"
	IntentTask        IntentType = "task"
	IntentGreeting    IntentType = "greeting"
	IntentUnknown     IntentType = "unknown"
)

type IntentClassification struct {
	Type IntentType `json:"Type"`
	Strategy       string `json:"Strategy,omitempty"`
	MaxResults int `json:"MaxResults,omitempty"`
	ScoreThreshold float64 `json:"ScoreThreshold,omitempty"`
	Confidence float64 `json:"Confidence,omitempty"`
}

type DensitySignals struct {
	MessageRate float64 `json:"MessageRate"`
	TokenDensity float64 `json:"TokenDensity"`
	TopicShiftScore float64 `json:"TopicShiftScore"`
	EntityCount int `json:"EntityCount"`
	SentimentDelta float64 `json:"SentimentDelta"`
	OverallDensity float64 `json:"OverallDensity"`
	SmoothedDensity float64 `json:"SmoothedDensity"`
}

type OffloadDecision struct {
	ShouldOffload bool `json:"ShouldOffload"`
	Level       string `json:"Level"`
	CompressionRatio float64 `json:"CompressionRatio"`
	EstimatedSavings int `json:"EstimatedSavings"`
	Reason       string `json:"Reason,omitempty"`
}

type HarnessState struct {
	CurrentTokens int `json:"CurrentTokens"`
	ContextWindow int `json:"ContextWindow"`
	SmoothedDensity float64 `json:"SmoothedDensity"`
	PendingToolPairs int `json:"PendingToolPairs"`
	TaskHint       string `json:"TaskHint"`
	LastRecallIntent       string `json:"LastRecallIntent"`
	LastRecallStrategy       string `json:"LastRecallStrategy"`
}

type UtilityScore struct {
	RecordID       string `json:"RecordID"`
	AccessCount int `json:"AccessCount"`
	HalfLifeDays float64 `json:"HalfLifeDays"`
	Score float64 `json:"Score"`
	LastAccessAt int64 `json:"LastAccessAt"`
}

type CompressionLevel string

const (
	CompressionMild       CompressionLevel = "mild"
	CompressionAggressive CompressionLevel = "aggressive"
)

type CompressionStrategy string

const (
	CompressionSummary       CompressionStrategy = "summary"
	CompressionMermaid       CompressionStrategy = "mermaid"
	CompressionSlidingWindow CompressionStrategy = "sliding_window"
)

type CompressionDecision struct {
	ShouldCompress bool `json:"ShouldCompress"`
	Level CompressionLevel `json:"Level"`
	Strategy CompressionStrategy `json:"Strategy"`
	Reason       string `json:"Reason"`
	CurrentRatio float64 `json:"CurrentRatio"`
}

type CompressionResult struct {
	Strategy CompressionStrategy `json:"Strategy"`
	TokenSavings int `json:"TokenSavings"`
	NewTokenTotal int `json:"NewTokenTotal"`
	Content       string `json:"Content,omitempty"`
	ReplaceFrom int `json:"ReplaceFrom,omitempty"`
	ReplaceTo int `json:"ReplaceTo,omitempty"`
}

type L1Memory struct {
	ID       string `json:"ID"`
	TenantID       string `json:"TenantID"`
	UserID       string `json:"UserID"`
	Content       string `json:"Content"`
	Type       string `json:"Type"`
	UtilityScore float64 `json:"UtilityScore"`
	RecallCount int `json:"RecallCount"`
	IsArchived bool `json:"IsArchived"`
	CreatedAt int64 `json:"CreatedAt,omitempty"`
}

type L2Scene struct {
	ID       string `json:"ID"`
	TenantID       string `json:"TenantID"`
	UserID       string `json:"UserID"`
	Filename       string `json:"Filename"`
	Content       string `json:"Content"`
	Summary       string `json:"Summary"`
	Heat int `json:"Heat"`
}

type L3Persona struct {
	ID       string `json:"ID"`
	TenantID       string `json:"TenantID"`
	UserID       string `json:"UserID"`
	Content       string `json:"Content"`
	Version int `json:"Version"`
	IsCurrent bool `json:"IsCurrent"`
}

type RecallResult struct {
	Memories             []*L1Memory `json:"memories"`
	Scene                *L2Scene    `json:"scene,omitempty"`
	Persona              *L3Persona  `json:"persona,omitempty"`
	AppendSystemContext  string      `json:"append_system_context,omitempty"`
	PrependContext       string      `json:"prepend_context,omitempty"`
}

type RuleDomain string

const (
	DomainToolSelection        RuleDomain = "tool_selection"
	DomainIntentClassification RuleDomain = "intent_classification"
	DomainScoring              RuleDomain = "scoring"
	DomainSafety               RuleDomain = "safety"
	DomainMemory               RuleDomain = "memory"
	DomainRouting              RuleDomain = "routing"
	DomainFallback             RuleDomain = "fallback"
	DomainToolNeed             RuleDomain = "tool_need"
)

type RuleLevel int

const (
	RuleLevelSystem RuleLevel = 0
	RuleLevelUser   RuleLevel = 1
)

type Rule struct {
	Name       string `json:"Name"`
	Service       string `json:"Service,omitempty"`
	Phase       string `json:"Phase,omitempty"`
	Domain RuleDomain `json:"Domain"`
	Source       string `json:"Source"`
	Priority int `json:"Priority"`
	Pattern       string `json:"Pattern"`
	Label       string `json:"Label,omitempty"`
	Tags     []string   `json:"tags,omitempty"`
	Score float64 `json:"Score"`
	Enabled bool `json:"Enabled"`
	TenantID       string `json:"TenantID,omitempty"`
	Level RuleLevel `json:"Level"`
}

type RuleOutput struct {
	RuleName       string `json:"RuleName"`
	Label       string `json:"Label"`
	Score float64 `json:"Score"`
	Tags     []string   `json:"tags"`
	Domain RuleDomain `json:"Domain"`
	Level RuleLevel `json:"Level"`
}

type RuleContext struct {
	TenantID       string `json:"TenantID"`
	UserID       string `json:"UserID,omitempty"`
	Service       string `json:"Service,omitempty"`
	Phase       string `json:"Phase,omitempty"`
	Domain RuleDomain `json:"Domain"`
	Query       string `json:"Query,omitempty"`
	Input       string `json:"Input,omitempty"`
}

type ToolType string

const (
	ToolWebhook ToolType = "webhook"
	ToolMCP     ToolType = "mcp"
	ToolSkill   ToolType = "skill"
	ToolBuiltin ToolType = "builtin"
)

type ToolStatus string

const (
	ToolStatusEnabled  ToolStatus = "enabled"
	ToolStatusDisabled ToolStatus = "disabled"
)

type ToolVisibility string

const (
	ToolVisibilityPrivate ToolVisibility = "private"
	ToolVisibilityPublic  ToolVisibility = "public"
)

type ToolParameter struct {
	Name       string `json:"Name"`
	Type       string `json:"Type"`
	Description       string `json:"Description"`
	Required bool `json:"Required"`
	Default     interface{} `json:"default,omitempty"`
}

type ToolSpec struct {
	Name       string `json:"Name"`
	Description       string `json:"Description"`
	Parameters  []ToolParameter `json:"parameters"`
	ToolType ToolType `json:"ToolType"`
	TenantID       string `json:"TenantID,omitempty"`
	Version       string `json:"Version"`
	Tags        []string        `json:"tags,omitempty"`
	Status ToolStatus `json:"Status"`
	TimeoutMs int `json:"TimeoutMs"`
	Visibility ToolVisibility `json:"Visibility"`
}

type LLMProvider struct {
	ID       string `json:"ID"`
	Name       string `json:"Name"`
	ProviderType       string `json:"ProviderType"`
	BaseURL       string `json:"BaseURL"`
	DefaultModel       string `json:"DefaultModel"`
	Models       []string `json:"models,omitempty"`
	Priority int `json:"Priority"`
	Weight int `json:"Weight"`
	TimeoutMs int `json:"TimeoutMs"`
	MaxTokens int `json:"MaxTokens"`
	Enabled bool `json:"Enabled"`
}

type ChatMessage struct {
	Role       string `json:"Role"`
	Content       string `json:"Content"`
}

type ChatRequest struct {
	Model       string `json:"Model,omitempty"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64 `json:"Temperature,omitempty"`
	MaxTokens int `json:"MaxTokens,omitempty"`
}

type ChatResponse struct {
	Content       string `json:"Content"`
	Usage   Usage  `json:"usage,omitempty"`
}

type Usage struct {
	PromptTokens int `json:"PromptTokens"`
	CompletionTokens int `json:"CompletionTokens"`
	TotalTokens int `json:"TotalTokens"`
}

type SessionStatus string

const (
	SessionActive SessionStatus = "active"
	SessionEnded  SessionStatus = "ended"
)

type Session struct {
	ID       string `json:"ID"`
	TenantID       string `json:"TenantID"`
	UserID       string `json:"UserID"`
	Model       string `json:"Model,omitempty"`
	Status SessionStatus `json:"Status"`
	ProviderID       string `json:"ProviderID,omitempty"`
	CreatedAt int64 `json:"CreatedAt,omitempty"`
	Internal   interface{} `json:"-"`
}

func (s *Session) Send(ctx context.Context, input string, history []Message) (*SendResult, error) {
	if s == nil || s.Internal == nil {
		return nil, ErrSessionNotFound
	}
	si, ok := s.Internal.(interface {
		Send(context.Context, string, []Message) (*SendResult, error)
	})
	if !ok {
		return nil, ErrSessionNotFound
	}
	return si.Send(ctx, input, history)
}

func (s *Session) Close() {
	if s == nil || s.Internal == nil {
		return
	}
	si, ok := s.Internal.(interface{ Close() })
	if !ok {
		return
	}
	si.Close()
}

type SessionStats struct {
	TotalMessages int `json:"TotalMessages"`
	TotalTokens int `json:"TotalTokens"`
	OffloadEvents int `json:"OffloadEvents"`
	TokenSaved int `json:"TokenSaved"`
	CompressionRatio float64 `json:"CompressionRatio"`
	AvgUtility float64 `json:"AvgUtility"`
	AvgLatencyMs float64 `json:"AvgLatencyMs"`
}

type SendResult struct {
	Response    *ChatResponse      `json:"response"`
	Stats       SessionStats       `json:"stats"`
	Recall      *RecallResult      `json:"recall,omitempty"`
	Offload     *OffloadDecision   `json:"offload,omitempty"`
	Compression *CompressionResult `json:"compression,omitempty"`
	Timing      SendTiming         `json:"timing"`
}

type SendTiming struct {
	RulesPre int64 `json:"RulesPre"`
	Classifier int64 `json:"Classifier"`
	ToolSelect int64 `json:"ToolSelect"`
	Recall int64 `json:"Recall"`
	Density int64 `json:"Density"`
	Offload int64 `json:"Offload"`
	Compression int64 `json:"Compression"`
	LLMCall int64 `json:"LLMCall"`
	RulesPost int64 `json:"RulesPost"`
}

type L0Record struct {
	ID       string `json:"ID"`
	SessionKey       string `json:"SessionKey"`
	Role       string `json:"Role"`
	Content       string `json:"Content"`
	RecordedAt int64 `json:"RecordedAt"`
}

type VectorResult struct {
	ID       string `json:"ID"`
	Score float64 `json:"Score"`
	Payload map[string]string `json:"payload,omitempty"`
}

type VectorPoint struct {
	ID       string `json:"ID"`
	Vector  []float32         `json:"vector"`
	Payload map[string]string `json:"payload,omitempty"`
}
