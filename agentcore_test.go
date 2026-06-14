package agentcore

import (
	"context"
	"testing"
	"github.com/LionAnti/agent-core/types"
)

// ─── Mock implementations ───

type mockLLM struct{}
func (m *mockLLM) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	return &types.ChatResponse{Content: "mock response", Usage: types.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15}}, nil
}

type mockStore struct{}

// ProviderStore
func (m *mockStore) ListProviders(ctx context.Context) ([]*types.LLMProvider, error) {
	return []*types.LLMProvider{{ID: "p1", Name: "Test", DefaultModel: "m1", Enabled: true, Weight: 1, Priority: 1}}, nil
}
func (m *mockStore) GetProvider(ctx context.Context, id string) (*types.LLMProvider, error) { return nil, nil }
func (m *mockStore) CreateProvider(ctx context.Context, p *types.LLMProvider) error { return nil }
func (m *mockStore) UpdateProvider(ctx context.Context, p *types.LLMProvider) error { return nil }
func (m *mockStore) DeleteProvider(ctx context.Context, id string) error { return nil }

// RegistryStore
func (m *mockStore) Get(ctx context.Context, key string) ([]byte, error) { return nil, nil }
func (m *mockStore) Put(ctx context.Context, key string, data []byte) error { return nil }
func (m *mockStore) Delete(ctx context.Context, key string) error { return nil }
func (m *mockStore) List(ctx context.Context, prefix string) ([][]byte, error) { return nil, nil }

// MemoryStore
func (m *mockStore) SaveL1(ctx context.Context, mem *types.L1Memory) error { return nil }
func (m *mockStore) SearchL1(ctx context.Context, tid, uid, q string, limit int) ([]*types.L1Memory, error) { return nil, nil }
func (m *mockStore) GetL1ByID(ctx context.Context, id string) (*types.L1Memory, error) { return nil, nil }
func (m *mockStore) UpdateL1Utility(ctx context.Context, id string, s float64) error { return nil }
func (m *mockStore) IncrementL1Recall(ctx context.Context, id string) error { return nil }
func (m *mockStore) CountActiveL1(ctx context.Context, tid, uid string) (int, error) { return 0, nil }
func (m *mockStore) DeleteL1(ctx context.Context, id string) error { return nil }
func (m *mockStore) ArchiveL1(ctx context.Context, id string, a bool) error { return nil }
func (m *mockStore) SaveL2(ctx context.Context, s *types.L2Scene) error { return nil }
func (m *mockStore) GetL2ByUser(ctx context.Context, tid, uid string) ([]*types.L2Scene, error) { return nil, nil }
func (m *mockStore) IncrementL2Heat(ctx context.Context, id string) error { return nil }
func (m *mockStore) DeleteL2(ctx context.Context, id string) error { return nil }
func (m *mockStore) SaveL3(ctx context.Context, p *types.L3Persona) error { return nil }
func (m *mockStore) GetCurrentL3(ctx context.Context, tid, uid string) (*types.L3Persona, error) { return nil, nil }
func (m *mockStore) ListL3Versions(ctx context.Context, tid, uid string, l int) ([]*types.L3Persona, error) { return nil, nil }

// RuleStore
func (m *mockStore) LoadRules(ctx context.Context) ([]*types.Rule, error) { return nil, nil }
func (m *mockStore) SaveRule(ctx context.Context, rule *types.Rule) error { return nil }
func (m *mockStore) DeleteRule(ctx context.Context, name string) error { return nil }

// L0Store
func (m *mockStore) Save(ctx context.Context, recs []types.L0Record) error { return nil }
func (m *mockStore) GetBySession(ctx context.Context, sk string, l, o int) ([]types.L0Record, error) { return nil, nil }

func TestNew_Valid(t *testing.T) {
	c, err := New(Config{
		LLMClient:     &mockLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	if err != nil { t.Fatal(err) }
	if c == nil { t.Fatal("core should not be nil") }
}

func TestNew_MissingLLM(t *testing.T) {
	_, err := New(Config{
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	if err == nil { t.Fatal("expected error for nil LLMClient") }
}

func TestNew_MissingProviderStore(t *testing.T) {
	_, err := New(Config{
		LLMClient:     &mockLLM{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	if err == nil { t.Fatal("expected error for nil ProviderStore") }
}

func TestNew_MissingRegistryStore(t *testing.T) {
	_, err := New(Config{
		LLMClient:     &mockLLM{},
		ProviderStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	if err == nil { t.Fatal("expected error for nil RegistryStore") }
}

func TestNew_MissingMemoryStore(t *testing.T) {
	_, err := New(Config{
		LLMClient:     &mockLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
	})
	if err == nil { t.Fatal("expected error for nil MemoryStore") }
}

func TestNew_InvalidThresholds(t *testing.T) {
	_, err := New(Config{
		LLMClient:           &mockLLM{},
		ProviderStore:       &mockStore{},
		RegistryStore:       &mockStore{},
		MemoryStore:         &mockStore{},
		MildThreshold:       0.9,
		AggressiveThreshold: 0.5,
	})
	if err == nil { t.Fatal("expected error for invalid thresholds") }
}

func TestSession_Send(t *testing.T) {
	core, err := New(Config{
		LLMClient:     &mockLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	if err != nil { t.Fatal(err) }
	sess, err := core.NewSession("t1", "u1")
	if err != nil { t.Fatal(err) }
	if sess == nil { t.Fatal("session should not be nil") }
	r, err := sess.Send(context.Background(), "hello", nil)
	if err != nil { t.Fatal(err) }
	if r.Response == nil { t.Fatal("response should not be nil") }
	if r.Response.Content != "mock response" { t.Fatalf("expected mock response, got %s", r.Response.Content) }
}

func TestSession_SendAfterClose(t *testing.T) {
	core, _ := New(Config{
		LLMClient:     &mockLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	sess, _ := core.NewSession("t1", "u1")
	sess.Close()
	_, err := sess.Send(context.Background(), "hello", nil)
	if err == nil { t.Fatal("expected error for closed session") }
}

func TestSession_WithModelOption(t *testing.T) {
	core, _ := New(Config{
		LLMClient:     &mockLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	sess, _ := core.NewSession("t1", "u1", WithModel("test-model"))
	if sess.Model != "test-model" { t.Fatalf("expected test-model, got %s", sess.Model) }
}

func TestGracefulShutdown(t *testing.T) {
	core, _ := New(Config{
		LLMClient:     &mockLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	sess, _ := core.NewSession("t1", "u1")
	sess.Send(context.Background(), "msg1", nil) // establish session

	if err := core.Close(context.Background()); err != nil { t.Fatal(err) }
	if n := core.ActiveSessionCount(); n != 0 { t.Fatalf("expected 0 active sessions, got %d", n) }

	_, err := core.NewSession("t1", "u1")
	if err == nil { t.Fatal("expected error creating session after close") }
}

func TestHealthCheck(t *testing.T) {
	core, _ := New(Config{
		LLMClient:     &mockLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	if err := core.HealthCheck(context.Background()); err != nil { t.Fatal(err) }
}

func TestHealthCheck_Failed(t *testing.T) {
	var core *AgentCore
	if err := core.HealthCheck(context.Background()); err == nil { t.Fatal("expected error for nil core") }
}

func TestActiveSessionCount(t *testing.T) {
	core, _ := New(Config{
		LLMClient:     &mockLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	if n := core.ActiveSessionCount(); n != 0 { t.Fatalf("expected 0, got %d", n) }
	core.NewSession("t1", "u1")
	if n := core.ActiveSessionCount(); n != 1 { t.Fatalf("expected 1, got %d", n) }
}

func TestHarnessState(t *testing.T) {
	core, _ := New(Config{
		LLMClient:     &mockLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	hs := core.HarnessState(50000)
	if hs == nil { t.Fatal("harness state should not be nil") }
	if hs.CurrentTokens != 50000 { t.Fatalf("expected 50000, got %d", hs.CurrentTokens) }
	if hs.ContextWindow != 128000 { t.Fatalf("expected 128000, got %d", hs.ContextWindow) }
}

func TestNilSafe_Core(t *testing.T) {
	var c *AgentCore
	if c.ActiveSessionCount() != 0 { t.Fatal("expected 0") }
	if c.Rules() != nil { t.Fatal("expected nil") }
	if c.Registry() != nil { t.Fatal("expected nil") }
	if c.HarnessState(100) != nil { t.Fatal("expected nil") }
	if c.Close(context.Background()) != nil { t.Fatal("expected nil") }
}

func TestSession_NilSafe(t *testing.T) {
	var s *types.Session
	r, err := s.Send(context.Background(), "test", nil)
	if err == nil { t.Fatal("expected error on nil session") }
	if r != nil { t.Fatal("expected nil result") }
	s.Close() // should not panic
}

func TestDefaults(t *testing.T) {
	core, _ := New(Config{
		LLMClient:     &mockLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	hs := core.HarnessState(100)
	if hs.ContextWindow != 128000 { t.Fatalf("expected 128000, got %d", hs.ContextWindow) }
}

func TestRuleCustomization(t *testing.T) {
	core, _ := New(Config{
		LLMClient:     &mockLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	eng := core.Rules()
	eng.AddRule(&types.Rule{Name: "custom", Domain: types.DomainScoring, Pattern: "custom_pattern", Label: "custom", Score: 1.0, Enabled: true, Level: types.RuleLevelUser})
	outs := eng.Match(&types.RuleContext{Domain: types.DomainScoring, Query: "custom_pattern"})
	found := false
	for _, o := range outs { if o.RuleName == "custom" { found = true; break } }
	if !found { t.Fatal("custom rule should match") }
}
