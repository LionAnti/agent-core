package agentcore

import (
    "context"
    "testing"
)

type mockProviderStore struct {
    providers []*LLMProvider
}

func (m *mockProviderStore) ListProviders(ctx context.Context) ([]*LLMProvider, error) { return m.providers, nil }
func (m *mockProviderStore) GetProvider(ctx context.Context, id string) (*LLMProvider, error) { return nil, nil }
func (m *mockProviderStore) CreateProvider(ctx context.Context, p *LLMProvider) error { return nil }
func (m *mockProviderStore) UpdateProvider(ctx context.Context, p *LLMProvider) error { return nil }
func (m *mockProviderStore) DeleteProvider(ctx context.Context, id string) error { return nil }

func TestProviderManager_Select(t *testing.T) {
    store := &mockProviderStore{
        providers: []*LLMProvider{
            {ID: "p1", Name: "Provider 1", Enabled: true, Weight: 1, Priority: 1, DefaultModel: "m1"},
        },
    }
    pm := NewProviderManager(store, NoopLogger{})
    selected, _ := pm.Select(context.Background(), "")
    if selected == nil {
        t.Fatal("expected a provider to be selected")
    }
    if selected.ID != "p1" {
        t.Fatalf("expected p1, got %s", selected.ID)
    }
}

func TestProviderManager_SelectNone(t *testing.T) {
    store := &mockProviderStore{providers: []*LLMProvider{}}
    pm := NewProviderManager(store, NoopLogger{})
    selected, _ := pm.Select(context.Background(), "")
    if selected != nil {
        t.Fatal("expected nil when no providers")
    }
}

func TestProviderManager_All(t *testing.T) {
    store := &mockProviderStore{
        providers: []*LLMProvider{
            {ID: "p1", Enabled: true, Weight: 1, Priority: 1},
            {ID: "p2", Enabled: true, Weight: 2, Priority: 2},
        },
    }
    pm := NewProviderManager(store, NoopLogger{})
    all := pm.All()
    if len(all) != 2 {
        t.Fatalf("expected 2 providers, got %d", len(all))
    }
}
