package hypergraph

import (
	"cmp"
	"slices"
	"testing"
	"time"
)

// ============================================================================
// GreedyHittingSet Tests
// ============================================================================

func TestGreedyHittingSet_SingleEdge(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C"})

	hs := h.GreedyHittingSet()

	// Should have exactly one vertex (any of A, B, C)
	if len(hs) != 1 {
		t.Fatalf("GreedyHittingSet size=%d want 1", len(hs))
	}

	// Verify it hits the edge
	if !isHittingSet(h, hs) {
		t.Fatal("result is not a valid hitting set")
	}
}

func TestGreedyHittingSet_DisjointEdges(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"C", "D"})
	_ = h.AddEdge("E3", []string{"E", "F"})

	hs := h.GreedyHittingSet()

	// Need at least 3 vertices (one per disjoint edge)
	if len(hs) < 3 {
		t.Fatalf("GreedyHittingSet size=%d want >=3", len(hs))
	}

	if !isHittingSet(h, hs) {
		t.Fatal("result is not a valid hitting set")
	}
}

func TestGreedyHittingSet_SharedVertex(t *testing.T) {
	t.Parallel()
	// All edges share vertex X - greedy should pick X (or equivalent)
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"X", "A"})
	_ = h.AddEdge("E2", []string{"X", "B"})
	_ = h.AddEdge("E3", []string{"X", "C"})

	hs := h.GreedyHittingSet()

	// Optimal is size 1 (just X). Greedy should find it.
	if len(hs) != 1 {
		t.Fatalf("GreedyHittingSet size=%d want 1", len(hs))
	}

	if hs[0] != "X" {
		t.Fatalf("Expected X in hitting set, got %v", hs[0])
	}

	if !isHittingSet(h, hs) {
		t.Fatal("result is not a valid hitting set")
	}
}

func TestGreedyHittingSet_Empty(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()

	hs := h.GreedyHittingSet()

	if len(hs) != 0 {
		t.Fatalf("GreedyHittingSet on empty graph=%v want empty", hs)
	}
}

func TestGreedyHittingSet_GreedyChoosesMaxDegree(t *testing.T) {
	t.Parallel()
	// B appears in 3 edges, others in fewer
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})
	_ = h.AddEdge("E3", []string{"B", "D"})
	_ = h.AddEdge("E4", []string{"E", "F"}) // Disjoint

	hs := h.GreedyHittingSet()

	// First pick should be B (degree 3), then one of E,F
	if !isHittingSet(h, hs) {
		t.Fatal("result is not a valid hitting set")
	}

	// Should be size 2: B covers E1,E2,E3, then need one of E,F for E4
	if len(hs) != 2 {
		t.Fatalf("GreedyHittingSet size=%d want 2", len(hs))
	}
}

func TestGreedyHittingSet_Deterministic(t *testing.T) {
	t.Parallel()
	// Create a hypergraph where map iteration order could affect results
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C"})
	_ = h.AddEdge("E2", []string{"C", "D", "E"})
	_ = h.AddEdge("E3", []string{"E", "F", "G"})

	// Run GreedyHittingSet multiple times and verify same result
	first := h.GreedyHittingSet()
	for i := 0; i < 100; i++ {
		result := h.GreedyHittingSet()
		if len(result) != len(first) {
			t.Fatalf("Run %d: different size %d vs %d", i, len(result), len(first))
		}
		for j, v := range result {
			if v != first[j] {
				t.Fatalf("Run %d: different result at index %d: %v vs %v", i, j, v, first[j])
			}
		}
	}
}

// ============================================================================
// EnumerateMinimalTransversals Tests
// ============================================================================

func TestEnumerateMinimalTransversals_SingleEdge(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C"})

	trans, err := h.EnumerateMinimalTransversals(100, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Each single vertex is a minimal transversal
	if len(trans) != 3 {
		t.Fatalf("got %d transversals, want 3", len(trans))
	}

	// Verify each is minimal and valid
	for _, tr := range trans {
		if len(tr) != 1 {
			t.Fatalf("expected minimal transversal of size 1, got %v", tr)
		}
		if !isHittingSet(h, tr) {
			t.Fatalf("transversal %v is not valid", tr)
		}
	}
}

func TestEnumerateMinimalTransversals_DisjointEdges(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"C", "D"})

	trans, err := h.EnumerateMinimalTransversals(100, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2x2 = 4 minimal transversals: {A,C}, {A,D}, {B,C}, {B,D}
	if len(trans) != 4 {
		t.Fatalf("got %d transversals, want 4", len(trans))
	}

	for _, tr := range trans {
		if len(tr) != 2 {
			t.Fatalf("expected transversal of size 2, got %v", tr)
		}
		if !isHittingSet(h, tr) {
			t.Fatalf("transversal %v is not valid", tr)
		}
		if !isMinimal(h, tr) {
			t.Fatalf("transversal %v is not minimal", tr)
		}
	}
}

func TestEnumerateMinimalTransversals_SharedVertex(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"X", "A"})
	_ = h.AddEdge("E2", []string{"X", "B"})

	trans, err := h.EnumerateMinimalTransversals(100, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Minimal transversals: {X}, {A,B}
	if len(trans) != 2 {
		t.Fatalf("got %d transversals, want 2", len(trans))
	}

	for _, tr := range trans {
		if !isHittingSet(h, tr) {
			t.Fatalf("transversal %v is not valid", tr)
		}
		if !isMinimal(h, tr) {
			t.Fatalf("transversal %v is not minimal", tr)
		}
	}
}

func TestEnumerateMinimalTransversals_Empty(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()

	trans, err := h.EnumerateMinimalTransversals(100, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty hypergraph has no edges to cover, so empty set is the only transversal
	if len(trans) != 1 || len(trans[0]) != 0 {
		t.Fatalf("expected single empty transversal, got %v", trans)
	}
}

func TestEnumerateMinimalTransversals_SolutionCutoff(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C", "D", "E"})

	// 5 minimal transversals exist, limit to 3
	trans, err := h.EnumerateMinimalTransversals(3, time.Hour)

	if err != ErrCutoff {
		t.Fatalf("expected ErrCutoff, got %v", err)
	}
	if len(trans) != 3 {
		t.Fatalf("expected 3 transversals, got %d", len(trans))
	}
}

func TestEnumerateMinimalTransversals_NoDuplicates(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})

	trans, err := h.EnumerateMinimalTransversals(100, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, tr := range trans {
		slices.Sort(tr)
		key := ""
		for _, v := range tr {
			key += v + ","
		}
		if seen[key] {
			t.Fatalf("duplicate transversal found: %v", tr)
		}
		seen[key] = true
	}
}

// ============================================================================
// GreedyColoring Tests
// ============================================================================

func TestGreedyColoring_SingleEdge(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C"})

	coloring := h.GreedyColoring()

	// All vertices in same edge must have different colors
	if !isValidColoring(h, coloring) {
		t.Fatal("coloring is invalid: vertices in same edge share color")
	}

	// Need exactly 3 colors for 3-element edge
	colors := make(map[int]bool)
	for _, c := range coloring {
		colors[c] = true
	}
	if len(colors) != 3 {
		t.Fatalf("expected 3 colors, got %d", len(colors))
	}
}

func TestGreedyColoring_DisjointEdges(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"C", "D"})

	coloring := h.GreedyColoring()

	if !isValidColoring(h, coloring) {
		t.Fatal("coloring is invalid")
	}

	// Can reuse colors across disjoint components - need only 2 colors
	colors := make(map[int]bool)
	for _, c := range coloring {
		colors[c] = true
	}
	if len(colors) > 2 {
		t.Fatalf("expected at most 2 colors for disjoint edges, got %d", len(colors))
	}
}

func TestGreedyColoring_SharedVertex(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"X", "A"})
	_ = h.AddEdge("E2", []string{"X", "B"})

	coloring := h.GreedyColoring()

	if !isValidColoring(h, coloring) {
		t.Fatal("coloring is invalid")
	}

	// X must differ from A and from B, but A and B can be same
	if coloring["X"] == coloring["A"] {
		t.Fatal("X and A share color but are in same edge")
	}
	if coloring["X"] == coloring["B"] {
		t.Fatal("X and B share color but are in same edge")
	}
}

func TestGreedyColoring_Empty(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()

	coloring := h.GreedyColoring()

	if len(coloring) != 0 {
		t.Fatalf("expected empty coloring, got %v", coloring)
	}
}

func TestGreedyColoring_VerticesWithNoEdges(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	h.AddVertex("A")
	h.AddVertex("B")

	coloring := h.GreedyColoring()

	// Isolated vertices can all be color 0
	for v, c := range coloring {
		if c != 0 {
			t.Fatalf("isolated vertex %s has color %d, expected 0", v, c)
		}
	}
}

func TestGreedyColoring_TriangleHypergraph(t *testing.T) {
	t.Parallel()
	// Three 2-edges forming a triangle: need exactly 2 colors
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})
	_ = h.AddEdge("E3", []string{"C", "A"})

	coloring := h.GreedyColoring()

	if !isValidColoring(h, coloring) {
		t.Fatal("coloring is invalid")
	}

	// This is 2-colorable (bipartite-like structure)
	colors := make(map[int]bool)
	for _, c := range coloring {
		colors[c] = true
	}
	// Greedy may use 2 or 3 depending on order, but must be valid
	if len(colors) > 3 {
		t.Fatalf("unexpected number of colors: %d", len(colors))
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// isHittingSet checks if the given set intersects every edge
func isHittingSet[V cmp.Ordered](h *Hypergraph[V], set []V) bool {
	setMap := make(map[V]bool)
	for _, v := range set {
		setMap[v] = true
	}

	for _, edgeID := range h.Edges() {
		hit := false
		for v := range h.edges[edgeID].Set {
			if setMap[v] {
				hit = true
				break
			}
		}
		if !hit {
			return false
		}
	}
	return true
}

// isMinimal checks if removing any vertex from the set makes it not a hitting set
func isMinimal[V cmp.Ordered](h *Hypergraph[V], set []V) bool {
	if len(set) == 0 {
		return true
	}
	for i := range set {
		subset := make([]V, 0, len(set)-1)
		subset = append(subset, set[:i]...)
		subset = append(subset, set[i+1:]...)
		if isHittingSet(h, subset) {
			return false // Can remove element i and still hit all edges
		}
	}
	return true
}

// isValidColoring checks that no two vertices in the same edge share a color
func isValidColoring[V cmp.Ordered](h *Hypergraph[V], coloring map[V]int) bool {
	for _, edgeID := range h.Edges() {
		colors := make(map[int]bool)
		for v := range h.edges[edgeID].Set {
			c := coloring[v]
			if colors[c] {
				return false // Duplicate color in same edge
			}
			colors[c] = true
		}
	}
	return true
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkGreedyHittingSet(b *testing.B) {
	h := NewHypergraph[int]()
	// Create 500 vertices with varied edges
	for i := 0; i < 200; i++ {
		members := []int{i % 500, (i * 3 + 1) % 500, (i * 7 + 2) % 500}
		_ = h.AddEdge("E"+algIntToStr(i), members)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.GreedyHittingSet()
	}
}

func BenchmarkEnumerateTransversals(b *testing.B) {
	h := NewHypergraph[int]()
	// Small graph for transversal enumeration (exponential complexity)
	for i := 0; i < 6; i++ {
		_ = h.AddEdge("E"+algIntToStr(i), []int{i * 2, i*2 + 1})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h.EnumerateMinimalTransversals(1000, time.Hour)
	}
}

func BenchmarkGreedyColoring(b *testing.B) {
	h := NewHypergraph[int]()
	// Create 200 vertices with 400 edges
	for i := 0; i < 400; i++ {
		members := []int{i % 200, (i + 1) % 200, (i + 2) % 200}
		_ = h.AddEdge("E"+algIntToStr(i), members)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.GreedyColoring()
	}
}

// Helper for benchmark edge IDs
func algIntToStr(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + algIntToStr(-i)
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}
