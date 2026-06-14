package rules

import (
	"regexp"
	"sync"
	"github.com/LionAnti/agent-core/types"
)

type ruleCache struct {
	mu   sync.RWMutex
	data map[string]*regexp.Regexp
}

func newRuleCache() *ruleCache { return &ruleCache{data: make(map[string]*regexp.Regexp)} }

func (rc *ruleCache) get(name string) *regexp.Regexp {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.data[name]
}

func (rc *ruleCache) set(name string, re *regexp.Regexp) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.data[name] = re
}

func builtinRules() []*types.Rule {
	return []*types.Rule{
		{Name: "builtin_intent_greeting", Domain: types.DomainIntentClassification, Pattern: `(?i)^(hi|hello|hey|你好|早上好|下午好|晚上好)`, Label: "greeting", Score: 0.9, Level: types.RuleLevelSystem, Enabled: true},
		{Name: "builtin_safety_blocklist", Domain: types.DomainSafety, Pattern: `(?i)(ignore previous|ignore all|你不需要遵守|忘记所有)`, Label: "prompt_injection", Score: 1.0, Level: types.RuleLevelSystem, Enabled: true},
		{Name: "builtin_recall_query", Domain: types.DomainMemory, Pattern: `(?i).*(之前|上次|以前|before|earlier|previous|recall|remember).*`, Label: "memory_recall", Score: 0.7, Level: types.RuleLevelSystem, Enabled: true},
		{Name: "builtin_tool_query", Domain: types.DomainToolNeed, Pattern: `(?i).*(search|find|查|搜索|lookup|get|fetch|list|查询).*`, Label: "tool_needed", Score: 0.6, Level: types.RuleLevelSystem, Enabled: true},
		{Name: "builtin_fallback_default", Domain: types.DomainFallback, Pattern: ".*", Label: "fallback_llm", Score: 0.1, Level: types.RuleLevelSystem, Enabled: true},
	}
}
