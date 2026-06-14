package main

import (
	"context"
	"fmt"
	"os"

	"github.com/LionAnti/agent-core"
	"github.com/LionAnti/agent-core/types"
)

// ─── Implement LLMClient ───

type myLLM struct{}

func (m *myLLM) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	return &types.ChatResponse{
		Content: "Hello from the LLM! I'm ready to help.",
		Usage:   types.Usage{PromptTokens: len(req.Messages) * 10, CompletionTokens: 15, TotalTokens: len(req.Messages)*10 + 15},
	}, nil
}

// ─── Implement ProviderStore ───

type myProviderStore struct{}

func (m *myProviderStore) ListProviders(ctx context.Context) ([]*types.LLMProvider, error) {
	return []*types.LLMProvider{
		{ID: "default", Name: "Default Provider", DefaultModel: "gpt-4", Enabled: true, Weight: 1, Priority: 1, Models: []string{"gpt-4", "gpt-3.5"}},
	}, nil
}
func (m *myProviderStore) GetProvider(ctx context.Context, id string) (*types.LLMProvider, error) { return nil, nil }
func (m *myProviderStore) CreateProvider(ctx context.Context, p *types.LLMProvider) error { return nil }
func (m *myProviderStore) UpdateProvider(ctx context.Context, p *types.LLMProvider) error { return nil }
func (m *myProviderStore) DeleteProvider(ctx context.Context, id string) error { return nil }

// ─── Implement RegistryStore ───

type myRegistryStore struct{ data map[string][]byte }

func (m *myRegistryStore) Get(ctx context.Context, key string) ([]byte, error) {
	if m.data == nil { return nil, nil }
	v, _ := m.data[key]; return v, nil
}
func (m *myRegistryStore) Put(ctx context.Context, key string, data []byte) error {
	if m.data == nil { m.data = make(map[string][]byte) }
	m.data[key] = data; return nil
}
func (m *myRegistryStore) Delete(ctx context.Context, key string) error { return nil }
func (m *myRegistryStore) List(ctx context.Context, prefix string) ([][]byte, error) { return nil, nil }

// ─── Implement MemoryStore ───

type myMemoryStore struct{}

func (m *myMemoryStore) SaveL1(ctx context.Context, mem *types.L1Memory) error { return nil }
func (m *myMemoryStore) SearchL1(ctx context.Context, tid, uid, q string, limit int) ([]*types.L1Memory, error) { return nil, nil }
func (m *myMemoryStore) GetL1ByID(ctx context.Context, id string) (*types.L1Memory, error) { return nil, nil }
func (m *myMemoryStore) UpdateL1Utility(ctx context.Context, id string, s float64) error { return nil }
func (m *myMemoryStore) IncrementL1Recall(ctx context.Context, id string) error { return nil }
func (m *myMemoryStore) CountActiveL1(ctx context.Context, tid, uid string) (int, error) { return 0, nil }
func (m *myMemoryStore) DeleteL1(ctx context.Context, id string) error { return nil }
func (m *myMemoryStore) ArchiveL1(ctx context.Context, id string, a bool) error { return nil }
func (m *myMemoryStore) SaveL2(ctx context.Context, s *types.L2Scene) error { return nil }
func (m *myMemoryStore) GetL2ByUser(ctx context.Context, tid, uid string) ([]*types.L2Scene, error) { return nil, nil }
func (m *myMemoryStore) IncrementL2Heat(ctx context.Context, id string) error { return nil }
func (m *myMemoryStore) DeleteL2(ctx context.Context, id string) error { return nil }
func (m *myMemoryStore) SaveL3(ctx context.Context, p *types.L3Persona) error { return nil }
func (m *myMemoryStore) GetCurrentL3(ctx context.Context, tid, uid string) (*types.L3Persona, error) { return nil, nil }
func (m *myMemoryStore) ListL3Versions(ctx context.Context, tid, uid string, l int) ([]*types.L3Persona, error) { return nil, nil }

func main() {
	ctx := context.Background()

	// Build the core
	core, err := agentcore.New(agentcore.Config{
		LLMClient:     &myLLM{},
		ProviderStore: &myProviderStore{},
		RegistryStore: &myRegistryStore{},
		MemoryStore:   &myMemoryStore{},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
	defer core.Close(ctx)

	fmt.Println("agent-core initialized successfully")
	fmt.Println()

	// Create a session and send a message
	sess, err := core.NewSession("demo-tenant", "demo-user",
		agentcore.WithModel("gpt-4"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: NewSession: %v\n", err)
		os.Exit(1)
	}
	defer sess.Close()

	result, err := sess.Send(ctx, "Hello, what can you do?", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Send: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("LLM: %s\n", result.Response.Content)
	fmt.Printf("Stats: %d messages, %d total tokens\n",
		result.Stats.TotalMessages, result.Stats.TotalTokens)
}
