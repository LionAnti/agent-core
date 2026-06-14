package agentcore

import (
	"regexp"
	"sort"
	"sync"
)

type RulesEngine struct {
	mu    sync.RWMutex
	rules []*Rule
	cache *ruleCache
}

func NewRulesEngine() *RulesEngine {
	return &RulesEngine{
		rules: builtinRules(),
		cache: newRuleCache(),
	}
}

func (e *RulesEngine) AddRule(r *Rule) {
	if r.Pattern != "" {
		e.cache.set(r.Name, regexp.MustCompile(r.Pattern))
	}
	e.mu.Lock()
	e.rules = append(e.rules, r)
	e.mu.Unlock()
}

func (e *RulesEngine) RemoveRule(name string) {
	filtered := make([]*Rule, 0, 10)
	e.mu.Lock()
	for _, r := range e.rules {
		if r.Name != name {
			filtered = append(filtered, r)
		}
	}
	e.rules = filtered
	e.mu.Unlock()
}

func (e *RulesEngine) Match(ctx *RuleContext) []RuleOutput {
	var results []RuleOutput
	e.mu.RLock()
	for _, r := range e.rules {
		if !r.Enabled {
			continue
		}
		if r.Domain != "" && ctx != nil && ctx.Domain != "" && r.Domain != ctx.Domain {
			continue
		}
		if r.Pattern != "" {
			re := e.cache.get(r.Name)
			if re == nil {
				re = regexp.MustCompile(r.Pattern)
				e.cache.set(r.Name, re)
			}
			input := ""
			if ctx != nil {
				input = ctx.Query
				if input == "" {
					input = ctx.Input
				}
			}
			if !re.MatchString(input) {
				continue
			}
		}
		results = append(results, RuleOutput{
			RuleName: r.Name, Label: r.Label, Score: r.Score,
			Tags: r.Tags, Domain: r.Domain, Level: r.Level,
		})
	}
	e.mu.RUnlock()

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results
}

func (e *RulesEngine) MatchPre(ctx *RuleContext) []RuleOutput { return e.Match(ctx) }
func (e *RulesEngine) MatchPost(ctx *RuleContext) []RuleOutput { return e.Match(ctx) }

func builtinRules() []*Rule {
	return []*Rule{
		{
			Name: "builtin_intent_greeting", Domain: DomainIntentClassification,
			Pattern: `(?i)^(hi|hello|hey|你好|早上好|下午好|晚上好)`,
			Label: "greeting", Score: 0.9, Level: RuleLevelSystem, Enabled: true,
		},
		{
			Name: "builtin_safety_blocklist", Domain: DomainSafety,
			Pattern: `(?i)(ignore previous|ignore all|你不需要遵守|忘记所有)`,
			Label: "prompt_injection", Score: 1.0, Level: RuleLevelSystem, Enabled: true,
		},
		{
			Name: "builtin_recall_query", Domain: DomainMemory,
			Pattern: `(?i).*(之前|上次|以前|before|earlier|previous|recall|remember).*`,
			Label: "memory_recall", Score: 0.7, Level: RuleLevelSystem, Enabled: true,
		},
		{
			Name: "builtin_tool_query", Domain: DomainToolNeed,
			Pattern: `(?i).*(search|find|查|搜索|lookup|get|fetch|list|查询).*`,
			Label: "tool_needed", Score: 0.6, Level: RuleLevelSystem, Enabled: true,
		},
		{
			Name: "builtin_fallback_default", Domain: DomainFallback,
			Pattern: ".*",
			Label: "fallback_llm", Score: 0.1, Level: RuleLevelSystem, Enabled: true,
		},
	}
}
