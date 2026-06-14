package agentcore

import (
	"context"
	"sync"
	"time"
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
	return &sessionInternal{core: c, session: s}
}

func (si *sessionInternal) Send(ctx context.Context, input string, history []Message) (*SendResult, error) {
	si.mu.Lock()
	defer si.mu.Unlock()

	if si.session.Status != SessionActive {
		return nil, ErrSessionAlreadyEnded
	}

	var timing SendTiming
	result := &SendResult{}
	start := time.Now()
	log := si.core.logger
	svcCtx := &RuleContext{TenantID: si.session.TenantID, UserID: si.session.UserID}

	t0 := time.Now()
	preRules := si.core.rules.MatchPre(svcCtx)
	timing.RulesPre = time.Since(t0).Milliseconds()

	msgs := make([]Message, 0, len(history)+1)
	msgs = append(msgs, history...)
	msgs = append(msgs, Message{Role: "user", Content: input, RecordedAt: now()})

	t0 = time.Now()
	totalTokens := estimateMessagesTokenCount(msgs)
	harnessState := si.core.HarnessState(totalTokens)
	intent, err := si.core.intentClassifier.Classify(ctx, msgs, harnessState)
	timing.Classifier = time.Since(t0).Milliseconds()
	if err != nil {
		log.Error("intent classification failed", "error", err)
		return nil, err
	}
	log.Debug("intent", "type", intent.Type, "confidence", intent.Confidence)

	t0 = time.Now()
	tools, _ := si.core.registry.List(ctx, si.session.TenantID)
	matchedTools, _ := si.core.toolSelector.Select(ctx, intent, tools)
	timing.ToolSelect = time.Since(t0).Milliseconds()
	_ = matchedTools
	_ = preRules

	t0 = time.Now()
	recall, err := si.core.memoryRecall.Recall(ctx, intent, si.session.TenantID, si.session.UserID)
	timing.Recall = time.Since(t0).Milliseconds()
	if err != nil {
		log.Warn("memory recall failed", "error", err)
	}
	result.Recall = recall

	t0 = time.Now()
	density, err := si.core.densityEstimator.Estimate(ctx, msgs, harnessState)
	timing.Density = time.Since(t0).Milliseconds()
	if err != nil {
		return nil, err
	}

	t0 = time.Now()
	decision, err := si.core.offloadDecider.Decide(ctx, density, harnessState)
	timing.Offload = time.Since(t0).Milliseconds()
	if err != nil {
		return nil, err
	}
	result.Offload = decision

	if decision.ShouldOffload {
		t0 = time.Now()
		compDecision := si.core.compressor.ShouldCompress(harnessState)
		compResult, err := si.core.compressor.Compress(ctx, msgs, compDecision)
		timing.Compression = time.Since(t0).Milliseconds()
		if err == nil && compResult != nil {
			result.Compression = compResult
			harnessState.CurrentTokens = compResult.NewTokenTotal
			si.offloadCnt++
			si.tokenSaved += compResult.TokenSavings
			log.Info("compression applied", "strategy", compResult.Strategy, "savings", compResult.TokenSavings)
		}
	}

	t0 = time.Now()
	llmReq := buildChatRequest(si.session, msgs, recall)
	resp, err := si.core.llmClient.Chat(ctx, llmReq)
	timing.LLMCall = time.Since(t0).Milliseconds()
	if err != nil {
		log.Error("LLM call failed", "error", err)
		return nil, err
	}

	t0 = time.Now()
	svcCtx.Domain = DomainScoring
	si.core.rules.MatchPost(svcCtx)
	timing.RulesPost = time.Since(t0).Milliseconds()

	result.Response = resp
	result.Stats = SessionStats{
		TotalMessages: len(msgs),
		TotalTokens:   totalTokens,
		OffloadEvents: si.offloadCnt,
		TokenSaved:    si.tokenSaved,
		AvgLatencyMs:  float64(timing.LLMCall),
	}
	result.Timing = timing

	if si.core.l0Store != nil {
		recs := []L0Record{
			{ID: newID(), SessionKey: si.session.ID, Role: "user", Content: input, RecordedAt: now()},
			{ID: newID(), SessionKey: si.session.ID, Role: "assistant", Content: resp.Content, RecordedAt: now()},
		}
		if err := si.core.l0Store.Save(ctx, recs); err != nil {
			log.Warn("failed to save L0", "error", err)
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
	si.session.Status = SessionEnded
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
