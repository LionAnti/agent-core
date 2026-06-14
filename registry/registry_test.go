package registry

import (
	"context"
	"testing"
	"github.com/agent-core/types"
)

type mockStore struct { data map[string][]byte }

func (m *mockStore) Get(ctx context.Context, key string) ([]byte, error) {
	if m.data == nil { return nil, nil }
	if d, ok := m.data[key]; ok { return d, nil }
	return nil, nil
}
func (m *mockStore) Put(ctx context.Context, key string, data []byte) error {
	if m.data == nil { m.data = make(map[string][]byte) }
	m.data[key] = data; return nil
}
func (m *mockStore) Delete(ctx context.Context, key string) error { if m.data != nil { delete(m.data, key) }; return nil }
func (m *mockStore) List(ctx context.Context, prefix string) ([][]byte, error) {
	var r [][]byte
	if m.data != nil {
		for k, v := range m.data {
			if len(k) >= len(prefix) && k[:len(prefix)] == prefix { r = append(r, v) }
		}
	}
	return r, nil
}

type mockLogger struct{}
func (mockLogger) Debug(msg string, keys ...any) {}
func (mockLogger) Info(msg string, keys ...any)  {}
func (mockLogger) Warn(msg string, keys ...any)  {}
func (mockLogger) Error(msg string, keys ...any) {}

func TestRegisterAndGet(t *testing.T) {
	tr := NewToolRegistry(&mockStore{}, mockLogger{})
	tool := &types.ToolSpec{Name: "test-tool", TenantID: "t1", Description: "test", Version: "1.0", Status: types.ToolStatusEnabled, ToolType: types.ToolBuiltin}
	if err := tr.Register(context.Background(), tool); err != nil { t.Fatal(err) }
	got, err := tr.Get(context.Background(), "t1", "test-tool")
	if err != nil { t.Fatal(err) }
	if got.Name != "test-tool" { t.Fatalf("expected test-tool, got %s", got.Name) }
}

func TestUnregister(t *testing.T) {
	tr := NewToolRegistry(&mockStore{}, mockLogger{})
	tool := &types.ToolSpec{Name: "del", TenantID: "t1", Status: types.ToolStatusEnabled}
	tr.Register(context.Background(), tool)
	if err := tr.Unregister(context.Background(), "t1", "del"); err != nil { t.Fatal(err) }
	_, err := tr.Get(context.Background(), "t1", "del")
	if err == nil { t.Fatal("expected error after delete") }
}

func TestList_OnlyEnabled(t *testing.T) {
	tr := NewToolRegistry(&mockStore{}, mockLogger{})
	tr.Register(context.Background(), &types.ToolSpec{Name: "enabled", TenantID: "t1", Status: types.ToolStatusEnabled})
	tr.Register(context.Background(), &types.ToolSpec{Name: "disabled", TenantID: "t1", Status: types.ToolStatusDisabled})
	list, err := tr.List(context.Background(), "t1")
	if err != nil { t.Fatal(err) }
	if len(list) != 1 { t.Fatalf("expected 1 enabled tool, got %d", len(list)) }
}

func TestSearch(t *testing.T) {
	tr := NewToolRegistry(&mockStore{}, mockLogger{})
	tr.Register(context.Background(), &types.ToolSpec{Name: "search-api", TenantID: "t1", Description: "API search tool", Status: types.ToolStatusEnabled})
	tr.Register(context.Background(), &types.ToolSpec{Name: "other", TenantID: "t1", Description: "something else", Status: types.ToolStatusEnabled})
	results, err := tr.Search(context.Background(), "t1", "search")
	if err != nil { t.Fatal(err) }
	if len(results) != 1 { t.Fatalf("expected 1 result, got %d", len(results)) }
}

func TestRuleToolSelector(t *testing.T) {
	sel := NewRuleToolSelector(func(ctx *types.RuleContext) []types.RuleOutput { return nil })
	tools := []types.ToolSpec{
		{Name: "t1", Tags: []string{"search"}},
		{Name: "t2", Tags: []string{"action"}},
	}
	selected, err := sel.Select(context.Background(), &types.IntentClassification{Type: types.IntentTask}, tools)
	if err != nil { t.Fatal(err) }
	if len(selected) == 2 { t.Log("selector returned all tools (fallback)") }
}

func TestCopyTool(t *testing.T) {
	orig := &types.ToolSpec{Name: "orig", Parameters: []types.ToolParameter{{Name: "p1"}}, Tags: []string{"tag1"}}
	copied := copyTool(orig)
	copied.Name = "modified"
	if orig.Name != "orig" { t.Fatal("modifying copy should not affect original") }
}

func TestNilSafe(t *testing.T) {
	var tr *ToolRegistry
	if err := tr.Register(context.Background(), nil); err == nil { t.Fatal("expected error") }
	if _, err := tr.Get(context.Background(), "a", "b"); err == nil { t.Fatal("expected error") }
	if _, err := tr.List(context.Background(), "a"); err == nil { t.Fatal("expected error") }
	var sel *RuleToolSelector
	if s, _ := sel.Select(context.Background(), nil, nil); s != nil { t.Fatal("expected nil") }
}
