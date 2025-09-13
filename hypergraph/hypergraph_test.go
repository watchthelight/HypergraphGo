package hypergraph

import (
	"bytes"
	"testing"
)

func TestHypergraphBasic(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	if err := h.AddEdge("E1", []string{"A", "B"}); err != nil {
		t.Fatalf("AddEdge E1: %v", err)
	}
	if err := h.AddEdge("E2", []string{"B", "C"}); err != nil {
		t.Fatalf("AddEdge E2: %v", err)
	}

	if got, want := h.NumVertices(), 3; got != want {
		t.Fatalf("NumVertices=%d want %d", got, want)
	}
	if got, want := h.NumEdges(), 2; got != want {
		t.Fatalf("NumEdges=%d want %d", got, want)
	}
	if got, want := h.VertexDegree("B"), 2; got != want {
		t.Fatalf("VertexDegree(B)=%d want %d", got, want)
	}
	if got, ok := h.EdgeSize("E1"); !ok || got != 2 {
		t.Fatalf("EdgeSize(E1)=(%d,%v) want (2,true)", got, ok)
	}

	// BFS reachability (order not guaranteed)
	bfs := h.BFS("A")
	if bfs == nil {
		t.Fatalf("BFS returned nil for existing start")
	}
	seen := map[string]bool{}
	for _, v := range bfs {
		seen[v] = true
	}
	for _, v := range []string{"A", "B", "C"} {
		if !seen[v] {
			t.Fatalf("BFS missing %q", v)
		}
	}
}

func TestAddEdgeErrors(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	if err := h.AddEdge("E1", nil); err == nil {
		t.Fatalf("expected error for empty edge members")
	}
	if err := h.AddEdge("E1", []string{"A"}); err != nil {
		t.Fatalf("AddEdge E1: %v", err)
	}
	if err := h.AddEdge("E1", []string{"B"}); err == nil {
		t.Fatalf("expected ErrDuplicateEdge on duplicate ID")
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})

	var buf bytes.Buffer
	if err := h.SaveJSON(&buf); err != nil {
		t.Fatalf("SaveJSON: %v", err)
	}
	h2, err := LoadJSON[string](&buf)
	if err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}
	if h2.NumVertices() != h.NumVertices() || h2.NumEdges() != h.NumEdges() {
		t.Fatalf("roundtrip sizes differ: got (V=%d,E=%d) want (V=%d,E=%d)", h2.NumVertices(), h2.NumEdges(), h.NumVertices(), h.NumEdges())
	}
	// Verify each edge membership matches as sets
	for _, id := range h.Edges() {
		size1, _ := h.EdgeSize(id)
		size2, ok := h2.EdgeSize(id)
		if !ok || size1 != size2 {
			t.Fatalf("edge %s size mismatch: %d vs %d", id, size1, size2)
		}
	}
}

func TestTransformsAndIncidence(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})

	// Two-section should connect A-B and B-C (2 edges)
	g := h.TwoSection()
	if got, want := len(g.Vertices()), 3; got != want {
		t.Fatalf("two-section vertices=%d want %d", got, want)
	}
	if got, want := len(g.Edges()), 2; got != want {
		t.Fatalf("two-section edges=%d want %d", got, want)
	}

	// Incidence (COO) count equals total memberships
	_, _, coo := h.IncidenceMatrix()
	total := 0
	for _, id := range h.Edges() {
		sz, _ := h.EdgeSize(id)
		total += sz
	}
	if got := len(coo.Rows); got != total || len(coo.Cols) != total {
		t.Fatalf("incidence nnz mismatch: rows=%d cols=%d total=%d", len(coo.Rows), len(coo.Cols), total)
	}
}

// Benchmark BFS over a deterministic sliding-window hypergraph.
func BenchmarkBFS(b *testing.B) {
	h := NewHypergraph[int]()
	// Create N vertices and E edges where each edge connects a window of k vertices.
	N, E, k := 2000, 4000, 5
	for i := 0; i < E; i++ {
		members := make([]int, 0, k)
		start := (i * 3) % (N - k) // deterministic stride
		for j := 0; j < k; j++ {
			members = append(members, start+j)
		}
		_ = h.AddEdge("E"+fmtInt(i), members)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.BFS(0)
	}
}

func fmtInt(i int) string {
	// small helper avoiding fmt import in tests
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	n := len(buf)
	for i > 0 {
		n--
		buf[n] = digits[i%10]
		i /= 10
	}
	return string(buf[n:])
}
