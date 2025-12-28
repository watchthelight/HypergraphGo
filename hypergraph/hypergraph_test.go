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

// ============================================================================
// EdgeMembers Tests
// ============================================================================

func TestEdgeMembers_ExistingEdge(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C"})

	members := h.EdgeMembers("E1")

	if len(members) != 3 {
		t.Fatalf("EdgeMembers(E1)=%d members, want 3", len(members))
	}

	// Check that all expected vertices are present (order not guaranteed)
	memberSet := make(map[string]bool)
	for _, m := range members {
		memberSet[m] = true
	}
	for _, v := range []string{"A", "B", "C"} {
		if !memberSet[v] {
			t.Fatalf("EdgeMembers missing %q", v)
		}
	}
}

func TestEdgeMembers_NonExistentEdge(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})

	members := h.EdgeMembers("E2") // doesn't exist

	if members != nil {
		t.Fatalf("EdgeMembers(E2)=%v want nil for non-existent edge", members)
	}
}

// ============================================================================
// Copy Tests
// ============================================================================

func TestCopy_DeepCopySemantics(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})

	copyH := h.Copy()

	// Verify copy has same structure
	if copyH.NumVertices() != h.NumVertices() {
		t.Fatalf("Copy NumVertices=%d want %d", copyH.NumVertices(), h.NumVertices())
	}
	if copyH.NumEdges() != h.NumEdges() {
		t.Fatalf("Copy NumEdges=%d want %d", copyH.NumEdges(), h.NumEdges())
	}

	// Verify all vertices exist in copy
	for _, v := range h.Vertices() {
		if !copyH.HasVertex(v) {
			t.Fatalf("Copy missing vertex %v", v)
		}
	}

	// Verify all edges exist in copy with same members
	for _, id := range h.Edges() {
		if !copyH.HasEdge(id) {
			t.Fatalf("Copy missing edge %s", id)
		}
		origSize, _ := h.EdgeSize(id)
		copySize, _ := copyH.EdgeSize(id)
		if origSize != copySize {
			t.Fatalf("Copy edge %s size=%d want %d", id, copySize, origSize)
		}
	}
}

func TestCopy_Independence(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})

	copyH := h.Copy()

	// Modify original
	h.AddVertex("Z")
	_ = h.AddEdge("E2", []string{"X", "Y"})

	// Copy should be unaffected
	if copyH.HasVertex("Z") {
		t.Fatal("Copy was affected by adding vertex to original")
	}
	if copyH.HasEdge("E2") {
		t.Fatal("Copy was affected by adding edge to original")
	}
	if copyH.NumVertices() != 2 {
		t.Fatalf("Copy NumVertices=%d want 2", copyH.NumVertices())
	}
	if copyH.NumEdges() != 1 {
		t.Fatalf("Copy NumEdges=%d want 1", copyH.NumEdges())
	}
}

func TestCopy_EmptyHypergraph(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()

	copyH := h.Copy()

	if copyH.NumVertices() != 0 || copyH.NumEdges() != 0 {
		t.Fatalf("Copy of empty=(V=%d,E=%d) want (0,0)", copyH.NumVertices(), copyH.NumEdges())
	}
}

// ============================================================================
// AddEdge with Duplicate Vertices Tests
// ============================================================================

func TestAddEdge_DuplicateVertices(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()

	// Add edge with duplicate vertices in input
	err := h.AddEdge("E1", []string{"A", "A", "B"})
	if err != nil {
		t.Fatalf("AddEdge with duplicates: %v", err)
	}

	// Edge should contain only unique vertices (set semantics)
	size, ok := h.EdgeSize("E1")
	if !ok {
		t.Fatal("Edge E1 not found")
	}
	if size != 2 {
		t.Fatalf("EdgeSize(E1)=%d want 2 (duplicates should be deduplicated)", size)
	}

	// Hypergraph should have only 2 vertices
	if h.NumVertices() != 2 {
		t.Fatalf("NumVertices=%d want 2", h.NumVertices())
	}
}

func TestAddEdge_AllDuplicates(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()

	// Add edge where all vertices are the same
	err := h.AddEdge("E1", []string{"A", "A", "A"})
	if err != nil {
		t.Fatalf("AddEdge: %v", err)
	}

	// Edge should contain only 1 unique vertex
	size, ok := h.EdgeSize("E1")
	if !ok {
		t.Fatal("Edge E1 not found")
	}
	if size != 1 {
		t.Fatalf("EdgeSize(E1)=%d want 1", size)
	}
}

// ============================================================================
// VertexDegree and EdgeSize Edge Cases
// ============================================================================

func TestVertexDegree_NonExistentVertex(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})

	degree := h.VertexDegree("Z") // doesn't exist

	if degree != 0 {
		t.Fatalf("VertexDegree(Z)=%d want 0 for non-existent vertex", degree)
	}
}

func TestEdgeSize_NonExistentEdge(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})

	size, ok := h.EdgeSize("E2") // doesn't exist

	if ok {
		t.Fatal("EdgeSize(E2) should return false for non-existent edge")
	}
	if size != 0 {
		t.Fatalf("EdgeSize(E2)=%d want 0", size)
	}
}

func TestVertexDegree_MultipleEdges(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"A", "C"})
	_ = h.AddEdge("E3", []string{"A", "D"})

	degree := h.VertexDegree("A")

	if degree != 3 {
		t.Fatalf("VertexDegree(A)=%d want 3", degree)
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
