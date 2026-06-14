package provider

import (
	"context"
	"testing"
	"github.com/LionAnti/agent-core/types"
)

type mockProviderStore struct { providers []*types.LLMProvider }

func (m *mockProviderStore) ListProviders(ctx context.Context) ([]*types.LLMProvider, error) { return m.providers, nil }
func (m *mockProviderStore) GetProvider(ctx context.Context, id string) (*types.LLMProvider, error) { return nil, nil }
func (m *mockProviderStore) CreateProvider(ctx context.Context, p *types.LLMProvider) error { return nil }
func (m *mockProviderStore) UpdateProvider(ctx context.Context, p *types.LLMProvider) error { return nil }
func (m *mockProviderStore) DeleteProvider(ctx context.Context, id string) error { return nil }

type mockLogger struct{}

func (mockLogger) Debug(msg string, keys ...any) {}
func (mockLogger) Info(msg string, keys ...any)  {}
func (mockLogger) Warn(msg string, keys ...any)  {}
func (mockLogger) Error(msg string, keys ...any) {}

func TestNewManager(t *testing.T) {
	pm := NewManager(&mockProviderStore{providers: []*types.LLMProvider{
		{ID: "p1", Name: "Provider 1", Enabled: true, Weight: 1, Priority: 1, DefaultModel: "m1"},
	}}, mockLogger{})
	sel, err := pm.Select(context.Background(), "")
	if err != nil { t.Fatal(err) }
	if sel.ID != "p1" { t.Fatalf("expected p1, got %s", sel.ID) }
}

func TestSelect_EmptyActive(t *testing.T) {
	pm := NewManager(&mockProviderStore{}, mockLogger{})
	_, err := pm.Select(context.Background(), "")
	if err == nil { t.Fatal("expected error for empty providers") }
}

func TestSelect_PreferredID(t *testing.T) {
	pm := NewManager(&mockProviderStore{providers: []*types.LLMProvider{
		{ID: "p1", Name: "A", Enabled: true, Weight: 1, Priority: 1, DefaultModel: "m"},
		{ID: "p2", Name: "B", Enabled: true, Weight: 1, Priority: 2, DefaultModel: "m"},
	}}, mockLogger{})
	sel, err := pm.Select(context.Background(), "p2")
	if err != nil { t.Fatal(err) }
	if sel.ID != "p2" { t.Fatalf("expected p2, got %s", sel.ID) }
}

func TestSelect_PreferredIDNotFound(t *testing.T) {
	pm := NewManager(&mockProviderStore{providers: []*types.LLMProvider{
		{ID: "p1", Name: "A", Enabled: true, Weight: 1, Priority: 1, DefaultModel: "m"},
	}}, mockLogger{})
	sel, err := pm.Select(context.Background(), "nonexistent")
	if err != nil { t.Fatal(err) }
	if sel == nil { t.Fatal("expected some provider") }
}

func TestAll_ReturnsCopy(t *testing.T) {
	pm := NewManager(&mockProviderStore{providers: []*types.LLMProvider{
		{ID: "p1", Name: "Original", Enabled: true, Weight: 1, Priority: 1, DefaultModel: "m"},
	}}, mockLogger{})
	all := pm.All()
	all[0].Name = "Modified"
	sel, _ := pm.Select(context.Background(), "p1")
	if sel.Name != "Original" { t.Fatal("modifying All() result should not affect manager") }
}

func TestNilSafe(t *testing.T) {
	var pm *Manager
	if pm.All() != nil { t.Fatal("expected nil") }
	_, err := pm.Select(context.Background(), "")
	if err == nil { t.Fatal("expected error") }
	if pm.RefreshProviders(context.Background()) == nil { t.Fatal("expected error") }
}
