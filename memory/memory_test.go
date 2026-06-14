package memory

import (
	"context"
	"testing"
	"github.com/LionAnti/agent-core/types"
)

type mockMemStore struct {
	l1s []*types.L1Memory
}

func (m *mockMemStore) SaveL1(ctx context.Context, mem *types.L1Memory) error { m.l1s = append(m.l1s, mem); return nil }
func (m *mockMemStore) SearchL1(ctx context.Context, tid, uid, q string, limit int) ([]*types.L1Memory, error) { return nil, nil }
func (m *mockMemStore) GetL1ByID(ctx context.Context, id string) (*types.L1Memory, error) { return nil, nil }
func (m *mockMemStore) UpdateL1Utility(ctx context.Context, id string, s float64) error { return nil }
func (m *mockMemStore) IncrementL1Recall(ctx context.Context, id string) error { return nil }
func (m *mockMemStore) CountActiveL1(ctx context.Context, tid, uid string) (int, error) { return 0, nil }
func (m *mockMemStore) DeleteL1(ctx context.Context, id string) error { return nil }
func (m *mockMemStore) ArchiveL1(ctx context.Context, id string, arch bool) error { return nil }
func (m *mockMemStore) SaveL2(ctx context.Context, s *types.L2Scene) error { return nil }
func (m *mockMemStore) GetL2ByUser(ctx context.Context, tid, uid string) ([]*types.L2Scene, error) { return nil, nil }
func (m *mockMemStore) IncrementL2Heat(ctx context.Context, id string) error { return nil }
func (m *mockMemStore) DeleteL2(ctx context.Context, id string) error { return nil }
func (m *mockMemStore) SaveL3(ctx context.Context, p *types.L3Persona) error { return nil }
func (m *mockMemStore) GetCurrentL3(ctx context.Context, tid, uid string) (*types.L3Persona, error) { return nil, nil }
func (m *mockMemStore) ListL3Versions(ctx context.Context, tid, uid string, l int) ([]*types.L3Persona, error) { return nil, nil }

type mockLogger struct{}
func (mockLogger) Debug(msg string, keys ...any) {}
func (mockLogger) Info(msg string, keys ...any)  {}
func (mockLogger) Warn(msg string, keys ...any)  {}
func (mockLogger) Error(msg string, keys ...any) {}

func TestExtractL1_Success(t *testing.T) {
	store := &mockMemStore{}
	p := NewPipeline(store, mockLogger{})
	recs := []types.L0Record{
		{ID: "1", SessionKey: "t1/u1/s1", Role: "user", Content: "hello"},
		{ID: "2", SessionKey: "t1/u1/s1", Role: "assistant", Content: "hi"},
		{ID: "3", SessionKey: "t1/u1/s1", Role: "user", Content: "help me"},
	}
	if err := p.ExtractL1(context.Background(), recs); err != nil { t.Fatal(err) }
	if len(store.l1s) != 1 { t.Fatalf("expected 1 L1 memory, got %d", len(store.l1s)) }
	if store.l1s[0].TenantID != "t1" { t.Fatalf("expected t1, got %s", store.l1s[0].TenantID) }
	if store.l1s[0].UserID != "u1" { t.Fatalf("expected u1, got %s", store.l1s[0].UserID) }
}

func TestExtractL1_FewerThan3(t *testing.T) {
	store := &mockMemStore{}
	p := NewPipeline(store, mockLogger{})
	recs := []types.L0Record{{ID: "1", SessionKey: "t/u/s", Role: "user", Content: "only one"}}
	if err := p.ExtractL1(context.Background(), recs); err != nil { t.Fatal(err) }
	if len(store.l1s) != 0 { t.Fatal("should not extract L1 with <3 records") }
}

func TestRecall_NilStoreNoPanic(t *testing.T) {
	r := NewRecall(nil, nil, mockLogger{})
	result, err := r.Recall(context.Background(), &types.IntentClassification{Type: types.IntentRecall}, "t", "u")
	if err != nil { t.Fatal(err) }
	if result == nil { t.Fatal("expected empty result, not nil") }
}

func TestRecall_NoRecall(t *testing.T) {
	store := &mockMemStore{}
	r := NewRecall(store, nil, mockLogger{})
	result, err := r.Recall(context.Background(), &types.IntentClassification{Type: types.IntentFactual}, "t", "u")
	if err != nil { t.Fatal(err) }
	if len(result.Memories) != 0 { t.Fatal("should not recall for factual intent") }
}

func TestExtractTenantID(t *testing.T) {
	if id := extractTID("t/u/s"); id != "t" { t.Fatalf("expected t, got %s", id) }
	if id := extractTID(""); id != "" { t.Fatalf("expected empty, got %s", id) }
}

func TestExtractUserID(t *testing.T) {
	if id := extractUID("t/u/s"); id != "u" { t.Fatalf("expected u, got %s", id) }
	if id := extractUID("t"); id != "" { t.Fatalf("expected empty, got %s", id) }
}

func TestNilSafe(t *testing.T) {
	var p *Pipeline
	if err := p.ExtractL1(context.Background(), nil); err != nil { t.Fatal("expected nil, got error") }
	var r *Recall
	result, err := r.Recall(context.Background(), nil, "", "")
	if err != nil { t.Fatal(err) }
	if result == nil { t.Fatal("expected result") }
}
