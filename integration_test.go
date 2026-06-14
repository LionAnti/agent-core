package agentcore

import (
	"context"
	"testing"
	"github.com/agent-core/types"
)

type slowLLM struct{}

func (s *slowLLM) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	return &types.ChatResponse{Content: "slow response"}, nil
}

func TestConcurrentSessions(t *testing.T) {
	core, _ := New(Config{
		LLMClient:     &slowLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	n := 5
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		go func(id int) {
			s, e := core.NewSession("t1", "u1")
			if e != nil { errs <- e; return }
			_, e = s.Send(context.Background(), "ping", nil)
			errs <- e
		}(i)
	}
	for i := 0; i < n; i++ {
		if e := <-errs; e != nil { t.Error(e) }
	}
}

func TestSessionTiming(t *testing.T) {
	core, _ := New(Config{
		LLMClient:     &slowLLM{},
		ProviderStore: &mockStore{},
		RegistryStore: &mockStore{},
		MemoryStore:   &mockStore{},
	})
	sess, _ := core.NewSession("t1", "u1")
	r, err := sess.Send(context.Background(), "timing test", nil)
	if err != nil { t.Fatal(err) }
	if r.Response == nil { t.Fatal("expected response") }
}
