package agentcore_test

import (
    "context"
    "fmt"
    "testing"

    agentcore "github.com/agent-core"
)

// mockLLM simulates a basic LLM response.
type mockLLM struct{}

func (m *mockLLM) Chat(ctx context.Context, req *agentcore.ChatRequest) (*agentcore.ChatResponse, error) {
    return &agentcore.ChatResponse{
        Content: "Mock response",
        Usage:   agentcore.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
    }, nil
}

// mockStore implements ProviderStore, RegistryStore, MemoryStore for testing.
type mockStore struct{}

func (m *mockStore) ListProviders(ctx context.Context) ([]*agentcore.LLMProvider, error) {
    return []*agentcore.LLMProvider{
        {ID: "test-provider", Name: "Test", DefaultModel: "test-model", Enabled: true, Weight: 1, Priority: 1},
    }, nil
}
func (m *mockStore) GetProvider(ctx context.Context, id string) (*agentcore.LLMProvider, error) {
    return &agentcore.LLMProvider{ID: id, Enabled: true}, nil
}
func (m *mockStore) CreateProvider(ctx context.Context, p *agentcore.LLMProvider) error { return nil }
func (m *mockStore) UpdateProvider(ctx context.Context, p *agentcore.LLMProvider) error { return nil }
func (m *mockStore) DeleteProvider(ctx context.Context, id string) error                { return nil }
func (m *mockStore) Get(ctx context.Context, key string) ([]byte, error)                { return nil, nil }
func (m *mockStore) Put(ctx context.Context, key string, data []byte) error              { return nil }
func (m *mockStore) Delete(ctx context.Context, key string) error                        { return nil }
func (m *mockStore) List(ctx context.Context, prefix string) ([][]byte, error)           { return nil, nil }
func (m *mockStore) LoadRules(ctx context.Context) ([]*agentcore.Rule, error)            { return nil, nil }
func (m *mockStore) SaveRule(ctx context.Context, rule *agentcore.Rule) error            { return nil }
func (m *mockStore) DeleteRule(ctx context.Context, name string) error                   { return nil }
func (m *mockStore) SaveL1(ctx context.Context, mem *agentcore.L1Memory) error           { return nil }
func (m *mockStore) SearchL1(ctx context.Context, tenantID, userID, query string, limit int) ([]*agentcore.L1Memory, error) {
    return nil, nil
}
func (m *mockStore) GetL1ByID(ctx context.Context, id string) (*agentcore.L1Memory, error) { return nil, nil }
func (m *mockStore) UpdateL1Utility(ctx context.Context, id string, score float64) error { return nil }
func (m *mockStore) IncrementL1Recall(ctx context.Context, id string) error              { return nil }
func (m *mockStore) DeleteL1(ctx context.Context, id string) error                        { return nil }
func (m *mockStore) ArchiveL1(ctx context.Context, id string, archived bool) error          { return nil }
func (m *mockStore) CountActiveL1(ctx context.Context, tenantID, userID string) (int, error) { return 0, nil }
func (m *mockStore) SaveL2(ctx context.Context, scene *agentcore.L2Scene) error          { return nil }
func (m *mockStore) GetL2ByUser(ctx context.Context, tenantID, userID string) ([]*agentcore.L2Scene, error) { return nil, nil }
func (m *mockStore) IncrementL2Heat(ctx context.Context, id string) error                { return nil }
func (m *mockStore) DeleteL2(ctx context.Context, id string) error                        { return nil }
func (m *mockStore) SaveL3(ctx context.Context, persona *agentcore.L3Persona) error      { return nil }
func (m *mockStore) GetCurrentL3(ctx context.Context, tenantID, userID string) (*agentcore.L3Persona, error) { return nil, nil }
func (m *mockStore) ListL3Versions(ctx context.Context, tenantID, userID string, limit int) ([]*agentcore.L3Persona, error) { return nil, nil }

// ─── Example: Basic Usage ───

func Example() {
    core, err := agentcore.New(agentcore.Config{
        LLMClient:     &mockLLM{},
        ProviderStore: &mockStore{},
        RegistryStore: &mockStore{},
        MemoryStore:   &mockStore{},
        Logger:        agentcore.NoopLogger{},
    })
    if err != nil {
        fmt.Printf("FAIL: %v\n", err)
        return
    }
    defer core.Close(context.Background())

    sess, err := core.NewSession("test-tenant", "test-user",
        agentcore.WithModel("test-model"))
    if err != nil {
        fmt.Printf("FAIL: %v\n", err)
        return
    }
    result, err := sess.Send(context.Background(), "Hello", nil)
    if err != nil {
        fmt.Printf("FAIL: %v\n", err)
        return
    }
    fmt.Println(result.Response.Content)
    // Output: Mock response
}

// ─── Test: New with nil LLMClient ───

func TestNewInvalidConfig(t *testing.T) {
    _, err := agentcore.New(agentcore.Config{
        ProviderStore: &mockStore{},
        RegistryStore: &mockStore{},
        MemoryStore:   &mockStore{},
    })
    if err == nil {
        t.Fatal("expected error for nil LLMClient")
    }
}

// ─── Test: Session lifecycle ───

func TestSessionLifecycle(t *testing.T) {
    core, err := agentcore.New(agentcore.Config{
        LLMClient:     &mockLLM{},
        ProviderStore: &mockStore{},
        RegistryStore: &mockStore{},
        MemoryStore:   &mockStore{},
    })
    if err != nil {
        t.Fatal(err)
    }
    defer core.Close(context.Background())

    sess, err := core.NewSession("t", "u", agentcore.WithModel("m"))
    if err != nil {
        t.Fatal(err)
    }
    if sess.Status != agentcore.SessionActive {
        t.Fatal("session should be active")
    }

    result, err := sess.Send(context.Background(), "ping", nil)
    if err != nil {
        t.Fatal(err)
    }
    if result.Response.Content != "Mock response" {
        t.Fatalf("expected 'Mock response', got '%s'", result.Response.Content)
    }
}

// ─── Test: Session already ended ───

func TestSessionAlreadyEnded(t *testing.T) {
    core, _ := agentcore.New(agentcore.Config{
        LLMClient:     &mockLLM{},
        ProviderStore: &mockStore{},
        RegistryStore: &mockStore{},
        MemoryStore:   &mockStore{},
    })
    sess, _ := core.NewSession("t", "u")
    sess.Close()
    _, err := sess.Send(context.Background(), "fail", nil)
    if err == nil {
        t.Fatal("expected error for ended session")
    }
}

// ─── Test: Graceful shutdown ───

func TestGracefulShutdown(t *testing.T) {
    core, _ := agentcore.New(agentcore.Config{
        LLMClient:     &mockLLM{},
        ProviderStore: &mockStore{},
        RegistryStore: &mockStore{},
        MemoryStore:   &mockStore{},
    })
    ctx := context.Background()
    if err := core.Close(ctx); err != nil {
        t.Fatal(err)
    }
    _, err := core.NewSession("t", "u")
    if err == nil {
        t.Fatal("expected error after shutdown")
    }
}

func TestHealthCheck(t *testing.T) {
    core, _ := agentcore.New(agentcore.Config{
        LLMClient:     &mockLLM{},
        ProviderStore: &mockStore{},
        RegistryStore: &mockStore{},
        MemoryStore:   &mockStore{},
    })
    if err := core.HealthCheck(context.Background()); err != nil {
        t.Fatal(err)
    }
}



func TestNewSessionAfterClose(t *testing.T) {
    core, _ := agentcore.New(agentcore.Config{
        LLMClient:     &mockLLM{},
        ProviderStore: &mockStore{},
        RegistryStore: &mockStore{},
        MemoryStore:   &mockStore{},
    })
    core.Close(context.Background())
    _, err := core.NewSession("t", "u")
    if err == nil {
        t.Fatal("expected error creating session after close")
    }
}

func TestThresholdValidation(t *testing.T) {
    _, err := agentcore.New(agentcore.Config{
        LLMClient:           &mockLLM{},
        ProviderStore:       &mockStore{},
        RegistryStore:       &mockStore{},
        MemoryStore:         &mockStore{},
        MildThreshold:       0.9,
        AggressiveThreshold: 0.5,
    })
    if err == nil {
        t.Fatal("expected error for MildThreshold > AggressiveThreshold")
    }
}
