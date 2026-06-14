package rules

import (
	"fmt"
	"regexp"
	"sort"
	"sync"
	"github.com/LionAnti/agent-core/types"
)

type Engine struct {
	mu    sync.RWMutex
	rules []*types.Rule
	cache *ruleCache
}

func NewEngine() *Engine {
	return &Engine{rules: builtinRules(), cache: newRuleCache()}
}

func (e *Engine) AddRule(r *types.Rule) error {
	if e == nil { return fmt.Errorf("rules engine: nil receiver") }
	if r == nil { return fmt.Errorf("rules engine: nil rule") }
	if r.Pattern != "" {
		re, err := regexp.Compile(r.Pattern)
		if err != nil { return fmt.Errorf("compile rule %q: %w", r.Name, err) }
		e.cache.set(r.Name, re)
	}
	e.mu.Lock()
	e.rules = append(e.rules, r)
	e.mu.Unlock()
	return nil
}

func (e *Engine) RemoveRule(name string) {
	if e == nil { return }
	e.mu.Lock()
	filtered := make([]*types.Rule, 0, len(e.rules))
	for _, r := range e.rules {
		if r.Name != name { filtered = append(filtered, r) }
	}
	e.rules = filtered
	e.mu.Unlock()
}

func (e *Engine) Match(ctx *types.RuleContext) []types.RuleOutput {
	if e == nil { return nil }
	e.mu.RLock()
	defer e.mu.RUnlock()
	results := make([]types.RuleOutput, 0, len(e.rules))
	for _, r := range e.rules {
		if !r.Enabled { continue }
		if r.Domain != "" && ctx != nil && ctx.Domain != "" && r.Domain != ctx.Domain { continue }
		if r.Pattern != "" {
			re := e.cache.get(r.Name)
			if re == nil { re = regexp.MustCompile(r.Pattern); e.cache.set(r.Name, re) }
			input := ""
			if ctx != nil { input = ctx.Query; if input == "" { input = ctx.Input } }
			if !re.MatchString(input) { continue }
		}
		results = append(results, types.RuleOutput{
			RuleName: r.Name, Label: r.Label, Score: r.Score,
			Tags: r.Tags, Domain: r.Domain, Level: r.Level,
		})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	return results
}

func (e *Engine) MatchPre(ctx *types.RuleContext) []types.RuleOutput { return e.Match(ctx) }
func (e *Engine) MatchPost(ctx *types.RuleContext) []types.RuleOutput { return e.Match(ctx) }

func (e *Engine) ListRules() []*types.Rule {
	if e == nil { return nil }
	e.mu.RLock()
	defer e.mu.RUnlock()
	r := make([]*types.Rule, len(e.rules))
	for i, rl := range e.rules { cp := *rl; r[i] = &cp }
	return r
}

func (e *Engine) CountRules() int {
	if e == nil { return 0 }
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.rules)
}
