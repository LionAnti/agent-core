package agentcore

import (
	"regexp"
	"sync"
)

type ruleCache struct {
	mu   sync.RWMutex
	data map[string]*regexp.Regexp
}

func newRuleCache() *ruleCache {
	return &ruleCache{data: make(map[string]*regexp.Regexp)}
}

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
