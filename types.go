package agentcore

import "context"

type Message struct {
    Role       string `json:"role"`
    Content    string `json:"content"`
    TokenCount int    `json:"token_count,omitempty"`
    RecordedAt int64  `json:"recorded_at,omitempty"`
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
    Type           IntentType `json:"type"`
    Strategy       string     `json:"strategy,omitempty"`
    MaxResults     int        `json:"max_results,omitempty"`
    ScoreThreshold float64    `json:"score_threshold,omitempty"`
    Confidence     float64    `json:"confidence,omitempty"`
}

type DensitySignals struct {
    MessageRate     float64 `json:"message_rate"`
    TokenDensity    float64 `json:"token_density"`
    TopicShiftScore float64 `json:"topic_shift_score"`
    EntityCount     int     `json:"entity_count"`
    SentimentDelta  float64 `json:"sentiment_delta"`
    OverallDensity  float64 `json:"overall_density"`
    SmoothedDensity float64 `json:"smoothed_density"`
}

type OffloadDecision struct {
    ShouldOffload    bool    `json:"should_offload"`
    Level            string  `json:"offload_level"`
    CompressionRatio float64 `json:"compression_ratio"`
    EstimatedSavings int     `json:"estimated_savings"`
    Reason           string  `json:"reason,omitempty"`
}

type HarnessState struct {
    CurrentTokens      int     `json:"current_tokens"`
    ContextWindow      int     `json:"context_window"`
    SmoothedDensity    float64 `json:"smoothed_density"`
    PendingToolPairs   int     `json:"pending_tool_pairs"`
    TaskHint           string  `json:"task_hint"`
    LastRecallIntent   string  `json:"last_recall_intent"`
    LastRecallStrategy string  `json:"last_recall_strategy"`
}

type UtilityScore struct {
    RecordID     string  `json:"record_id"`
    AccessCount  int     `json:"access_count"`
    HalfLifeDays float64 `json:"half_life_days"`
    Score        float64 `json:"score"`
    LastAccessAt int64   `json:"last_access_at"`
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
    ShouldCompress bool                `json:"should_compress"`
    Level          CompressionLevel    `json:"level"`
    Strategy       CompressionStrategy `json:"strategy"`
    Reason         string              `json:"reason"`
    CurrentRatio   float64             `json:"current_ratio"`
}

type CompressionResult struct {
    Strategy      CompressionStrategy `json:"strategy"`
    TokenSavings  int                 `json:"token_savings"`
    NewTokenTotal int                 `json:"new_token_total"`
    Content       string              `json:"content,omitempty"`
    ReplaceFrom   int                 `json:"replace_from,omitempty"`
    ReplaceTo     int                 `json:"replace_to,omitempty"`
}

type MemoryLevel int

const (
    MemoryL0 MemoryLevel = 0
    MemoryL1 MemoryLevel = 1
    MemoryL2 MemoryLevel = 2
    MemoryL3 MemoryLevel = 3
)

type L1Memory struct {
    ID           string  `json:"id"`
    TenantID     string  `json:"tenant_id"`
    UserID       string  `json:"user_id"`
    Content      string  `json:"content"`
    Type         string  `json:"type"`
    UtilityScore float64 `json:"utility_score"`
    RecallCount  int     `json:"recall_count"`
    IsArchived   bool    `json:"is_archived"`
    CreatedAt    int64   `json:"created_at,omitempty"`
}

type L2Scene struct {
    ID       string `json:"id"`
    TenantID string `json:"tenant_id"`
    UserID   string `json:"user_id"`
    Filename string `json:"filename"`
    Content  string `json:"content"`
    Summary  string `json:"summary"`
    Heat     int    `json:"heat"`
}

type L3Persona struct {
    ID        string `json:"id"`
    TenantID  string `json:"tenant_id"`
    UserID    string `json:"user_id"`
    Content   string `json:"content"`
    Version   int    `json:"version"`
    IsCurrent bool   `json:"is_current"`
}

type RecallResult struct {
    Memories             []*L1Memory  `json:"memories"`
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
    Name     string     `json:"name"`
    Service  string     `json:"service,omitempty"`
    Phase    string     `json:"phase,omitempty"`
    Domain   RuleDomain `json:"domain"`
    Source   string     `json:"source"`
    Priority int        `json:"priority"`
    Pattern  string     `json:"pattern"`
    Label    string     `json:"label,omitempty"`
    Tags     []string   `json:"tags,omitempty"`
    Score    float64    `json:"score"`
    Enabled  bool       `json:"enabled"`
    TenantID string     `json:"tenant_id,omitempty"`
    Level    RuleLevel  `json:"level"`
}

type RuleOutput struct {
    RuleName string     `json:"rule_name"`
    Label    string     `json:"label"`
    Score    float64    `json:"score"`
    Tags     []string   `json:"tags"`
    Domain   RuleDomain `json:"domain"`
    Level    RuleLevel  `json:"level"`
}

type RuleContext struct {
    TenantID string     `json:"tenant_id"`
    UserID   string     `json:"user_id,omitempty"`
    Service  string     `json:"service,omitempty"`
    Phase    string     `json:"phase,omitempty"`
    Domain   RuleDomain `json:"domain"`
    Query    string     `json:"query,omitempty"`
    Input    string     `json:"input,omitempty"`
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
    Name        string      `json:"name"`
    Type        string      `json:"type"`
    Description string      `json:"description"`
    Required    bool        `json:"required"`
    Default     interface{} `json:"default,omitempty"`
}

type ToolSpec struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Parameters  []ToolParameter `json:"parameters"`
    ToolType    ToolType        `json:"tool_type"`
    TenantID    string          `json:"tenant_id,omitempty"`
    Version     string          `json:"version"`
    Tags        []string        `json:"tags,omitempty"`
    Status      ToolStatus      `json:"status"`
    TimeoutMs   int             `json:"timeout_ms"`
    Visibility  ToolVisibility  `json:"visibility"`
}

type ToolExecuteRequest struct {
    Arguments map[string]interface{} `json:"arguments"`
    SessionID string                 `json:"session_id"`
    Variables map[string]string      `json:"variables,omitempty"`
}

type ToolExecuteResponse struct {
    Result   interface{} `json:"result,omitempty"`
    Error    string      `json:"error,omitempty"`
    Duration int64       `json:"duration_ms"`
}

type LLMProvider struct {
    ID           string   `json:"id"`
    Name         string   `json:"name"`
    ProviderType string   `json:"provider_type"`
    BaseURL      string   `json:"base_url"`
    DefaultModel string   `json:"default_model"`
    Models       []string `json:"models,omitempty"`
    Priority     int      `json:"priority"`
    Weight       int      `json:"weight"`
    TimeoutMs    int      `json:"timeout_ms"`
    MaxTokens    int      `json:"max_tokens"`
    Enabled      bool     `json:"enabled"`
}

type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatRequest struct {
    Model       string        `json:"model,omitempty"`
    Messages    []ChatMessage `json:"messages"`
    Temperature float64       `json:"temperature,omitempty"`
    MaxTokens   int           `json:"max_tokens,omitempty"`
}

type ChatResponse struct {
    Content string `json:"content"`
    Usage   Usage  `json:"usage,omitempty"`
}

type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}

type SessionStatus string

const (
    SessionActive SessionStatus = "active"
    SessionEnded  SessionStatus = "ended"
)

// Session represents a conversation session.
// Send is safe for sequential calls (mutex-protected).
// Do not call Send concurrently from multiple goroutines.
type Session struct {
    ID         string        `json:"id"`
    TenantID   string        `json:"tenant_id"`
    UserID     string        `json:"user_id"`
    Model      string        `json:"model,omitempty"`
    Status     SessionStatus `json:"status"`
    ProviderID string        `json:"provider_id,omitempty"`
    CreatedAt  int64         `json:"created_at,omitempty"`
	internal *sessionInternal `json:"-"`
}

type SessionStats struct {
    TotalMessages    int     `json:"total_messages"`
    TotalTokens      int     `json:"total_tokens"`
    OffloadEvents    int     `json:"offload_events"`
    TokenSaved       int     `json:"token_saved"`
    CompressionRatio float64 `json:"compression_ratio"`
    AvgUtility       float64 `json:"avg_utility"`
    AvgLatencyMs     float64 `json:"avg_latency_ms"`
}

type SendResult struct {
    Response    *ChatResponse      `json:"response"`
    Stats       SessionStats       `json:"stats"`
    Recall      *RecallResult      `json:"recall,omitempty"`
    Offload     *OffloadDecision   `json:"offload,omitempty"`
    Compression *CompressionResult `json:"compression,omitempty"`
    Timing      SendTiming         `json:"timing"`
}


// Send executes the full pipeline: rules -> classifier -> tools -> recall -> density -> offload -> compress -> LLM.
func (s *Session) Send(ctx context.Context, input string, history []Message) (*SendResult, error) {
	if s.internal == nil {
		return nil, ErrSessionNotFound
	}
	return s.internal.Send(ctx, input, history)
}

// Close marks the session as ended.
func (s *Session) Close() {
	if s.internal != nil {
		s.internal.Close()
	}
}

type SendTiming struct {
    RulesPre    int64 `json:"rules_pre_ms"`
    Classifier  int64 `json:"classifier_ms"`
    ToolSelect  int64 `json:"tool_select_ms"`
    Recall      int64 `json:"recall_ms"`
    Density     int64 `json:"density_ms"`
    Offload     int64 `json:"offload_ms"`
    Compression int64 `json:"compression_ms"`
    LLMCall     int64 `json:"llm_call_ms"`
    RulesPost   int64 `json:"rules_post_ms"`
}
type L0Record struct {
    ID         string `json:"id"`
    SessionKey string `json:"session_key"`
    Role       string `json:"role"`
    Content    string `json:"content"`
    RecordedAt int64  `json:"recorded_at"`
}
