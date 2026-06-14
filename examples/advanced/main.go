package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/LionAnti/agent-core"
	"github.com/LionAnti/agent-core/types"
)

// ─── Full implementation with L0 store ───

type fullLLM struct{}

func (m *fullLLM) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	time.Sleep(10 * time.Millisecond) // simulate latency
	return &types.ChatResponse{Content: fmt.Sprintf("You said: %s", req.Messages[len(req.Messages)-1].Content)}, nil
}

type fullProviderStore struct{}
func (m *fullProviderStore) ListProviders(ctx context.Context) ([]*types.LLMProvider, error) {
	return []*types.LLMProvider{{ID: "default", Name: "Default", DefaultModel: "gpt-4", Enabled: true, Weight: 1, Priority: 1}}, nil
}
func (m *fullProviderStore) GetProvider(_ context.Context, id string) (*types.LLMProvider, error) { return nil, nil }
func (m *fullProviderStore) CreateProvider(_ context.Context, _ *types.LLMProvider) error { return nil }
func (m *fullProviderStore) UpdateProvider(_ context.Context, _ *types.LLMProvider) error { return nil }
func (m *fullProviderStore) DeleteProvider(_ context.Context, _ string) error { return nil }

type fullRegistryStore struct{ data map[string][]byte }
func (m *fullRegistryStore) Get(_ context.Context, key string) ([]byte, error) {
	if m.data == nil { return nil, nil }
	v, _ := m.data[key]; return v, nil
}
func (m *fullRegistryStore) Put(_ context.Context, key string, data []byte) error {
	if m.data == nil { m.data = make(map[string][]byte) }; m.data[key] = data; return nil
}
func (m *fullRegistryStore) Delete(_ context.Context, key string) error {
	if m.data != nil { delete(m.data, key) }; return nil
}
func (m *fullRegistryStore) List(_ context.Context, prefix string) ([][]byte, error) {
	var r [][]byte
	if m.data != nil {
		for k, v := range m.data {
			if len(k) >= len(prefix) && k[:len(prefix)] == prefix { r = append(r, v) }
		}
	}
	return r, nil
}

type fullMemoryStore struct{}
func (m *fullMemoryStore) SaveL1(_ context.Context, mem *types.L1Memory) error {
	fmt.Printf("[L1 saved] %s...\n", mem.Content[:min(40, len(mem.Content))]); return nil
}
func (m *fullMemoryStore) SearchL1(_ context.Context, tid, uid, q string, l int) ([]*types.L1Memory, error) { return nil, nil }
func (m *fullMemoryStore) GetL1ByID(_ context.Context, id string) (*types.L1Memory, error) { return nil, nil }
func (m *fullMemoryStore) UpdateL1Utility(_ context.Context, id string, s float64) error { return nil }
func (m *fullMemoryStore) IncrementL1Recall(_ context.Context, id string) error { return nil }
func (m *fullMemoryStore) CountActiveL1(_ context.Context, tid, uid string) (int, error) { return 0, nil }
func (m *fullMemoryStore) DeleteL1(_ context.Context, id string) error { return nil }
func (m *fullMemoryStore) ArchiveL1(_ context.Context, id string, a bool) error { return nil }
func (m *fullMemoryStore) SaveL2(_ context.Context, s *types.L2Scene) error { return nil }
func (m *fullMemoryStore) GetL2ByUser(_ context.Context, tid, uid string) ([]*types.L2Scene, error) { return nil, nil }
func (m *fullMemoryStore) IncrementL2Heat(_ context.Context, id string) error { return nil }
func (m *fullMemoryStore) DeleteL2(_ context.Context, id string) error { return nil }
func (m *fullMemoryStore) SaveL3(_ context.Context, p *types.L3Persona) error { return nil }
func (m *fullMemoryStore) GetCurrentL3(_ context.Context, tid, uid string) (*types.L3Persona, error) { return nil, nil }
func (m *fullMemoryStore) ListL3Versions(_ context.Context, tid, uid string, l int) ([]*types.L3Persona, error) { return nil, nil }

type fullL0Store struct{ records []types.L0Record }

func (m *fullL0Store) Save(_ context.Context, recs []types.L0Record) error {
	m.records = append(m.records, recs...)
	for _, r := range recs { fmt.Printf("[L0] %s: %s\n", r.Role, r.Content) }
	return nil
}
func (m *fullL0Store) GetBySession(_ context.Context, sk string, l, o int) ([]types.L0Record, error) { return nil, nil }

func main() {
	ctx := context.Background()

	// Register a tool
	regStore := &fullRegistryStore{}
	regStore.Put(ctx, "tools/demo-tenant/my-search", toJSON(types.ToolSpec{
		Name: "my-search", Description: "Search the web", ToolType: types.ToolBuiltin,
		Status: types.ToolStatusEnabled, Version: "1.0", Tags: []string{"search"},
	}))

	core, err := agentcore.New(agentcore.Config{
		LLMClient:     &fullLLM{},
		ProviderStore: &fullProviderStore{},
		RegistryStore: regStore,
		MemoryStore:   &fullMemoryStore{},
		L0Store:       &fullL0Store{},
	})
	if err != nil { fmt.Fprintf(os.Stderr, "FAIL: %v\n", err); os.Exit(1) }
	defer core.Close(ctx)

	sess, _ := core.NewSession("demo-tenant", "demo-user")
	defer sess.Close()

	// Send 3 messages to accumulate context
	for _, msg := range []string{"Hello!", "What is AI?", "Tell me more"} {
		result, err := sess.Send(ctx, msg, nil)
		if err != nil { fmt.Fprintf(os.Stderr, "FAIL: %v\n", err); os.Exit(1) }
		fmt.Printf("\n=== Response ===\n%s\n", result.Response.Content)
	}
}

func toJSON(v any) []byte { d, _ := json.Marshal(v); return d }
