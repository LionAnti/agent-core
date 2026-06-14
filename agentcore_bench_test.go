package agentcore

import (
	"context"
	"testing"
	"github.com/LionAnti/agent-core/types"
)

type benchLLM struct{}
func (benchLLM) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	return &types.ChatResponse{Content: "benchmark response"}, nil
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		core, err := New(Config{
			LLMClient:     benchLLM{},
			ProviderStore: &mockStore{},
			RegistryStore: &mockStore{},
			MemoryStore:   &mockStore{},
		})
		if err != nil { b.Fatal(err) }
		_ = core
	}
}

func BenchmarkSessionSend(b *testing.B) {
	core, _ := New(Config{
		LLMClient:     benchLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sess, _ := core.NewSession("b", "u")
		_, err := sess.Send(context.Background(), "benchmark", nil)
		if err != nil { b.Fatal(err) }
	}
}

func BenchmarkSessionSendWithHistory(b *testing.B) {
	core, _ := New(Config{
		LLMClient:     benchLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	history := make([]types.Message, 50)
	for i := range history {
		history[i] = types.Message{Role: "user", Content: "previous message "}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sess, _ := core.NewSession("b", "u")
		_, err := sess.Send(context.Background(), "benchmark", history)
		if err != nil { b.Fatal(err) }
	}
}
