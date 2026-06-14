package agentcore

import "regexp"

var ruleCompiled = make(map[string]*regexp.Regexp)

func getCompiled(r *Rule) *regexp.Regexp {
	if r == nil {
		return nil
	}
	return ruleCompiled[r.Name]
}

func setCompiled(r *Rule, re *regexp.Regexp) {
	if r != nil {
		ruleCompiled[r.Name] = re
	}
}
