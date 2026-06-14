package agentcore

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	defaultCallTimeout  = 30 * time.Second
)

type sessionInternal struct {
	core       *AgentCore
	session    *Session
	mu         sync.Mutex
	msgs       []Message
	offloadCnt int
	tokenSaved int
}

func (c *AgentCore) openSession(s *Session) *sessionInternal {
	si := &sessionInternal{core: c, session: s}
	c.sessionMu.Lock()
	if c.closed {
		c.sessionMu.Unlock()
		return nil
	}
	c.sessions[s.ID] = si
	c.sessionMu.Unlock()
	return si
}

func (c *AgentCore) removeSession(id string) {
	c.sessionMu.Lock()
	delete(c.sessions, id)
	c.sessionMu.Unlock()
}

func runStage[T any](name string, fn func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in stage %q: %v", name, r)
		}
	}()
	return fn()
}

func runStageVoid(name string, fn func() error) error {
	_, err := runStage(name, func() (struct{}, error) {
		return struct{}{}, fn()
	})
	return err
}

func (si *sessionInternal) Send(ctx context.Context, input string, history []Message) (result *SendResult, err error) {
	si.mu.Lock()
	defer func() {
		si.mu.Unlock()
		if r := recover(); r != nil {
			err = fmt.Errorf("critical panic in session.Send: %v", r)
		}
	}()

	if si.session.Status != SessionActive {
		return nil, ErrSessionAlreadyEnded
	}

	var timing SendTiming
	result = &SendResult{}
	start := time.Now()
	log := si.core.logger
	svcCtx := &RuleContext{TenantID: si.session.TenantID, UserID: si.session.UserID}

	// Ensure context has a timeout for external calls.
	// Use a single reusable timeout context to avoid leaking contexts per stage.
	cancel := ensureTimeout(&ctx, defaultCallTimeout)
	if cancel != nil {
		defer cancel()
	}

	// Stage 1: Pre-rules
	if err := runStageVoid("pre-rules", func() error {
		t0 := time.Now()
		preRules := si.core.rules.MatchPre(svcCtx)
		timing.RulesPre = time.Since(t0).Milliseconds()
		_ = preRules
		return nil
	}); err != nil {
		return nil, fmt.Errorf("pipeline aborted at pre-rules: %w", err)
	}

	msgs := make([]Message, 0, len(history)+1)
	msgs = append(msgs, history...)
	msgs = append(msgs, Message{Role: "user", Content: input, RecordedAt: now()})

	totalTokens := estimateMessagesTokenCount(msgs)
	harnessState := si.core.HarnessState(totalTokens)

	// Stage 2: Intent Classification
	var intent *IntentClassification
	if err := runStageVoid("intent-classifier", func() error {
		t0 := time.Now()
		intent, err = si.core.intentClassifier.Classify(ctx, msgs, harnessState)
		timing.Classifier = time.Since(t0).Milliseconds()
		return err
	}); err != nil {
		log.Error("intent classification failed", "error", err)
		return nil, fmt.Errorf("intent classification: %w", err)
	}
	log.Debug("intent", "type", intent.Type, "confidence", intent.Confidence)

	// Stage 3: Tool Selection
	if err := runStageVoid("tool-select", func() error {
		t0 := time.Now()
		tools, listErr := si.core.registry.List(ctx, si.session.TenantID)
		if listErr != nil {
			return listErr
		}
		matchedTools, selectErr := si.core.toolSelector.Select(ctx, intent, tools)
		timing.ToolSelect = time.Since(t0).Milliseconds()
		_ = matchedTools
		return selectErr
	}); err != nil {
		log.Warn("tool selection failed, continuing", "error", err)
	}

	// Stage 4: Memory Recall
	if err := runStageVoid("memory-recall", func() error {
		t0 := time.Now()
		recall, recallErr := si.core.memoryRecall.Recall(ctx, intent, si.session.TenantID, si.session.UserID)
		timing.Recall = time.Since(t0).Milliseconds()
		if recallErr != nil {
			log.Warn("memory recall failed, continuing", "error", recallErr)
		}
		result.Recall = recall
		return nil
	}); err != nil {
		log.Warn("memory recall panicked, continuing without recall", "error", err)
	}

	// Stage 5: Density Estimation
	if err := runStageVoid("density", func() error {
		t0 := time.Now()
		density, densityErr := si.core.densityEstimator.Estimate(ctx, msgs, harnessState)
		timing.Density = time.Since(t0).Milliseconds()
		if densityErr != nil {
			return densityErr
		}
		_ = density
		return nil
	}); err != nil {
		return nil, fmt.Errorf("density estimation: %w", err)
	}

	// Stage 6: Offload Decision
	var decision *OffloadDecision
	if err := runStageVoid("offload", func() error {
		t0 := time.Now()
		var offloadErr error
		decision, offloadErr = si.core.offloadDecider.Decide(ctx, nil, harnessState)
		timing.Offload = time.Since(t0).Milliseconds()
		return offloadErr
	}); err != nil {
		return nil, fmt.Errorf("offload decision: %w", err)
	}
	result.Offload = decision

	// Stage 7: Compression
	if decision != nil && decision.ShouldOffload {
		if err := runStageVoid("compression", func() error {
			t0 := time.Now()
			compDecision := si.core.compressor.ShouldCompress(ctx, harnessState)
			compResult, compErr := si.core.compressor.Compress(ctx, msgs, compDecision)
			timing.Compression = time.Since(t0).Milliseconds()
			if compErr != nil {
				log.Warn("compression failed, continuing without", "error", compErr)
				return nil
			}
			result.Compression = compResult
			if compResult != nil {
				harnessState.CurrentTokens = compResult.NewTokenTotal
				si.offloadCnt++
				si.tokenSaved += compResult.TokenSavings
				log.Info("compression applied", "strategy", compResult.Strategy, "savings", compResult.TokenSavings)
			}
			return nil
		}); err != nil {
			log.Warn("compression panicked, continuing", "error", err)
		}
	}

	// Stage 8: LLM Call
	var resp *ChatResponse
	if err := runStageVoid("llm-call", func() error {
		t0 := time.Now()
		llmReq := buildChatRequest(si.session, msgs, result.Recall)
		var llmErr error
		resp, llmErr = si.core.llmClient.Chat(ctx, llmReq)
		timing.LLMCall = time.Since(t0).Milliseconds()
		return llmErr
	}); err != nil {
		log.Error("LLM call failed", "error", err)
		return nil, fmt.Errorf("LLM call: %w", err)
	}

	// Stage 9: Post-rules
	if err := runStageVoid("post-rules", func() error {
		t0 := time.Now()
		svcCtx.Domain = DomainScoring
		si.core.rules.MatchPost(svcCtx)
		timing.RulesPost = time.Since(t0).Milliseconds()
		return nil
	}); err != nil {
		log.Warn("post-rules panicked, continuing", "error", err)
	}

	result.Response = resp
	result.Stats = SessionStats{
		TotalMessages: len(msgs),
		TotalTokens:   totalTokens,
		OffloadEvents: si.offloadCnt,
		TokenSaved:    si.tokenSaved,
		AvgLatencyMs:  float64(timing.LLMCall),
	}
	result.Timing = timing

	// Record L0 (best-effort, non-fatal)
	if si.core.l0Store != nil {
		recs := []L0Record{
			{ID: newID(), SessionKey: si.session.ID, Role: "user", Content: input, RecordedAt: now()},
			{ID: newID(), SessionKey: si.session.ID, Role: "assistant", Content: resp.Content, RecordedAt: now()},
		}
		if saveErr := si.core.l0Store.Save(ctx, recs); saveErr != nil {
			log.Warn("failed to save L0 records", "error", saveErr)
		}
	}

	si.msgs = append(si.msgs, Message{Role: "assistant", Content: resp.Content})
	si.core.metrics.RecordLatency("send_total", float64(time.Since(start).Milliseconds()))
	si.core.metrics.RecordTokenUsage(resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	return result, nil
}

func (si *sessionInternal) Close() {
	si.mu.Lock()
	defer si.mu.Unlock()
	if si.session.Status == SessionEnded {
		return
	}
	si.session.Status = SessionEnded
	si.core.removeSession(si.session.ID)
}

func buildChatRequest(sess *Session, msgs []Message, recall *RecallResult) *ChatRequest {
	chatMsgs := make([]ChatMessage, 0, len(msgs)+2)
	if recall != nil {
		if recall.PrependContext != "" {
			chatMsgs = append(chatMsgs, ChatMessage{Role: "system", Content: recall.PrependContext})
		}
		if recall.AppendSystemContext != "" {
			chatMsgs = append(chatMsgs, ChatMessage{Role: "system", Content: recall.AppendSystemContext})
		}
	}
	for i := range msgs {
		chatMsgs = append(chatMsgs, ChatMessage{Role: msgs[i].Role, Content: msgs[i].Content})
	}
	req := &ChatRequest{Messages: chatMsgs}
	if sess.Model != "" {
		req.Model = sess.Model
	}
	return req
}

// ensureTimeout adds a timeout to ctx if it doesn't already have one.
// Returns a cancel function that MUST be called by the caller, or nil if no timeout was added.
func ensureTimeout(ctx *context.Context, timeout time.Duration) (cancel func()) {
	if _, ok := (*ctx).Deadline(); !ok {
		newCtx, c := context.WithTimeout(*ctx, timeout)
		*ctx = newCtx
		return c
	}
	return nil
}
