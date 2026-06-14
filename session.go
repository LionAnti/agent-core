package agentcore

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/LionAnti/agent-core/memory"
	"github.com/LionAnti/agent-core/types"
)

const defaultCallTimeout = 30 * time.Second
const maxSessionMessages = 1000

type sessionInternal struct {
	core       *AgentCore
	session    *types.Session
	mu         sync.Mutex
	msgs       []types.Message
	offloadCnt int
	tokenSaved int
}

type SessionInternal = sessionInternal

func (c *AgentCore) openSession(s *types.Session) (*sessionInternal, error) {
	if c==nil { return nil, types.ErrAgentClosed }
	si := &sessionInternal{core: c, session: s}
	c.sessionMu.Lock()
	if c.closed { c.sessionMu.Unlock(); return nil, types.ErrAgentClosed }
	if c.sessions==nil { c.sessions = make(map[string]*sessionInternal) }
	c.sessions[s.ID] = si; c.sessionMu.Unlock()
	return si, nil
}

func (c *AgentCore) removeSession(id string) {
	if c==nil { return }; c.sessionMu.Lock(); delete(c.sessions, id); c.sessionMu.Unlock()
}

func (si *sessionInternal) Send(ctx context.Context, input string, history []types.Message) (*types.SendResult, error) {
	if si==nil { return nil, types.ErrSessionNotFound }
	si.mu.Lock()
	defer func() { si.mu.Unlock(); recover() }()
	if si.session.Status != types.SessionActive { return nil, types.ErrSessionAlreadyEnded }

	var timing types.SendTiming
	result := &types.SendResult{}
	start := time.Now()
	log := si.core.logger
	svcCtx := &types.RuleContext{TenantID: si.session.TenantID, UserID: si.session.UserID}

	cancel := func() {}
	if _, ok := ctx.Deadline(); !ok { var nctx context.Context; nctx, cancel = context.WithTimeout(ctx, defaultCallTimeout); ctx = nctx }
	defer cancel()

	// Stage 1: Pre-rules
	func() {
		t0 := time.Now()
		defer func() { timing.RulesPre = time.Since(t0).Milliseconds(); recover() }()
		si.core.rules.MatchPre(svcCtx)
	}()

	msgs := make([]types.Message, 0, len(history)+1)
	msgs = append(msgs, history...)
	msgs = append(msgs, types.Message{Role: types.RoleUser, Content: input, RecordedAt: now()})

	totalTokens := estimateMessagesTokenCount(msgs)
	hs := si.core.HarnessState(totalTokens)

	// Stage 2: Classifier
	var intent *types.IntentClassification
	func() {
		t0 := time.Now()
		defer func() { timing.Classifier = time.Since(t0).Milliseconds(); recover() }()
		var e error; intent, e = si.core.intentClassifier.Classify(ctx, msgs, hs); _ = e
	}()
	log.Debug("intent", "type", intent.Type)

	// Stage 3: Tool Selection
	var matchedTools []types.ToolSpec
	func() {
		t0 := time.Now()
		defer func() { timing.ToolSelect = time.Since(t0).Milliseconds(); recover() }()
		if tools, e := si.core.gtregistry.List(ctx, si.session.TenantID); e == nil {
			matchedTools, _ = si.core.toolSelector.Select(ctx, intent, tools)
		}
	}()

	// Stage 4: Memory Recall
	func() {
		t0 := time.Now()
		defer func() { timing.Recall = time.Since(t0).Milliseconds(); recover() }()
		if r, e := si.core.memoryRecall.Recall(ctx, intent, si.session.TenantID, si.session.UserID); e == nil {
			result.Recall = r
		}
	}()

	// Stage 5: Density
	var density *types.DensitySignals
	func() {
		t0 := time.Now()
		defer func() { timing.Density = time.Since(t0).Milliseconds(); recover() }()
		density, _ = si.core.densityEstimator.Estimate(ctx, msgs, hs)
	}()

	// Stage 6: Offload
	var decision *types.OffloadDecision
	func() {
		t0 := time.Now()
		defer func() { timing.Offload = time.Since(t0).Milliseconds(); recover() }()
		decision, _ = si.core.offloadDecider.Decide(ctx, density, hs)
	}()
	result.Offload = decision

	// Stage 7: Compression
	if decision != nil && decision.ShouldOffload {
		func() {
			t0 := time.Now()
			defer func() { timing.Compression = time.Since(t0).Milliseconds(); recover() }()
			cd := si.core.compressor.ShouldCompress(ctx, hs)
			if cr, e := si.core.compressor.Compress(ctx, msgs, cd); e==nil&&cr!=nil {
				result.Compression = cr; hs.CurrentTokens = cr.NewTokenTotal
				si.offloadCnt++; si.tokenSaved += cr.TokenSavings
			}
		}()
	}

	// Stage 8: LLM Call
	var resp *types.ChatResponse
	func() {
		t0 := time.Now()
		defer func() { timing.LLMCall = time.Since(t0).Milliseconds(); recover() }()
		resp, _ = si.core.llmClient.Chat(ctx, buildRequest(si.session, msgs, result.Recall, matchedTools))
	}()
	if resp==nil { return nil, fmt.Errorf("llm: no response") }

	// Stage 9: Post-rules
	func() {
		t0 := time.Now()
		defer func() { timing.RulesPost = time.Since(t0).Milliseconds(); recover() }()
		svcCtx.Domain = types.DomainScoring; si.core.rules.MatchPost(svcCtx)
	}()

	result.Response = resp
	result.Stats = types.SessionStats{TotalMessages: len(msgs), TotalTokens: totalTokens, OffloadEvents: si.offloadCnt, TokenSaved: si.tokenSaved}
	result.Timing = timing

	if si.core.l0Store != nil {
		recs := []types.L0Record{
			{ID: newID(), SessionKey: si.session.ID, Role: "user", Content: input, RecordedAt: now()},
			{ID: newID(), SessionKey: si.session.ID, Role: "assistant", Content: resp.Content, RecordedAt: now()},
		}
		if se := si.core.l0Store.Save(ctx, recs); se != nil {
			log.Warn("L0 save failed", "error", se)
		} else if si.core.memStore != nil {
			memory.NewPipeline(si.core.memStore, log).ExtractL1(ctx, recs)
		}
	}

	si.msgs = appendMessage(si.msgs, types.Message{Role: types.RoleAssistant, Content: resp.Content}, maxSessionMessages)
	si.core.metrics.RecordLatency("send", float64(time.Since(start).Milliseconds()))
	if resp.Usage.TotalTokens > 0 { si.core.metrics.RecordTokenUsage(resp.Usage.PromptTokens, resp.Usage.CompletionTokens) }
	return result, nil
}

func (si *sessionInternal) Close() {
	if si==nil { return }
	si.mu.Lock(); defer si.mu.Unlock()
	if si.session.Status==types.SessionEnded { return }
	si.session.Status = types.SessionEnded; si.core.removeSession(si.session.ID)
}

func buildRequest(sess *types.Session, msgs []types.Message, recall *types.RecallResult, tools []types.ToolSpec) *types.ChatRequest {
	cm := make([]types.ChatMessage, 0, len(msgs)+4)
	if recall != nil {
		if recall.PrependContext != "" { cm = append(cm, types.ChatMessage{Role:"system", Content: recall.PrependContext}) }
		if recall.AppendSystemContext != "" { cm = append(cm, types.ChatMessage{Role:"system", Content: recall.AppendSystemContext}) }
	}
	if len(tools) > 0 {
		var sb strings.Builder; sb.WriteString("Available tools:\n")
		for _, t := range tools { sb.WriteString("- "); sb.WriteString(t.Name); sb.WriteString(": "); sb.WriteString(t.Description); sb.WriteString("\n") }
		cm = append(cm, types.ChatMessage{Role:"system", Content: sb.String()})
	}
	for i := range msgs { cm = append(cm, types.ChatMessage{Role: msgs[i].Role, Content: msgs[i].Content}) }
	req := &types.ChatRequest{Messages: cm}
	if sess.Model != "" { req.Model = sess.Model }
	return req
}

func appendMessage(msgs []types.Message, msg types.Message, max int) []types.Message {
	if max<=0||len(msgs)<max { return append(msgs, msg) }
	n := copy(msgs, msgs[1:]); msgs = msgs[:n]; return append(msgs, msg)
}
