package hypergraph

import (
	"testing"
)

// ============================================================================
// BFS Tests
// ============================================================================

func TestBFS_EmptyGraph(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()

	result := h.BFS("A")

	if result != nil {
		t.Fatalf("BFS on empty graph: got %v, want nil", result)
	}
}

func TestBFS_SingleVertex(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	h.AddVertex("A")

	result := h.BFS("A")

	if len(result) != 1 || result[0] != "A" {
		t.Fatalf("BFS single vertex: got %v, want [A]", result)
	}
}

func TestBFS_NonExistentStart(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C"})

	result := h.BFS("Z")

	if result != nil {
		t.Fatalf("BFS non-existent start: got %v, want nil", result)
	}
}

func TestBFS_LinearChain(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})
	_ = h.AddEdge("E3", []string{"C", "D"})

	result := h.BFS("A")

	if len(result) != 4 {
		t.Fatalf("BFS linear chain: got %d vertices, want 4", len(result))
	}

	// All vertices should be reachable
	seen := make(map[string]bool)
	for _, v := range result {
		seen[v] = true
	}
	for _, v := range []string{"A", "B", "C", "D"} {
		if !seen[v] {
			t.Fatalf("BFS linear chain: missing vertex %s", v)
		}
	}
}

func TestBFS_Cycle(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})
	_ = h.AddEdge("E3", []string{"C", "A"}) // Cycle back to A

	result := h.BFS("A")

	if len(result) != 3 {
		t.Fatalf("BFS cycle: got %d vertices, want 3", len(result))
	}

	// First vertex should be start
	if result[0] != "A" {
		t.Fatalf("BFS cycle: first vertex=%s, want A", result[0])
	}
}

func TestBFS_DisconnectedComponent(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"C", "D"}) // Disconnected from A,B

	result := h.BFS("A")

	// Should only reach A and B
	if len(result) != 2 {
		t.Fatalf("BFS disconnected: got %d vertices, want 2", len(result))
	}

	seen := make(map[string]bool)
	for _, v := range result {
		seen[v] = true
	}
	if !seen["A"] || !seen["B"] {
		t.Fatalf("BFS disconnected: expected A and B, got %v", result)
	}
	if seen["C"] || seen["D"] {
		t.Fatalf("BFS disconnected: should not reach C or D, got %v", result)
	}
}

func TestBFS_LargeHyperedge(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C", "D", "E", "F"})

	result := h.BFS("A")

	if len(result) != 6 {
		t.Fatalf("BFS large hyperedge: got %d vertices, want 6", len(result))
	}
}

func TestBFS_NoDuplicates(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	// Multiple paths to D: A-B-D and A-C-D
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"A", "C"})
	_ = h.AddEdge("E3", []string{"B", "D"})
	_ = h.AddEdge("E4", []string{"C", "D"})

	result := h.BFS("A")

	// Check for duplicates
	seen := make(map[string]int)
	for _, v := range result {
		seen[v]++
		if seen[v] > 1 {
			t.Fatalf("BFS duplicates: vertex %s appeared %d times", v, seen[v])
		}
	}

	if len(result) != 4 {
		t.Fatalf("BFS no duplicates: got %d vertices, want 4", len(result))
	}
}

func TestBFS_VisitOrder(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[int]()
	// Create a simple tree: 0 -> 1,2 -> 3,4,5,6
	_ = h.AddEdge("E1", []int{0, 1})
	_ = h.AddEdge("E2", []int{0, 2})
	_ = h.AddEdge("E3", []int{1, 3})
	_ = h.AddEdge("E4", []int{1, 4})
	_ = h.AddEdge("E5", []int{2, 5})
	_ = h.AddEdge("E6", []int{2, 6})

	result := h.BFS(0)

	// First vertex should be 0
	if result[0] != 0 {
		t.Fatalf("BFS order: first vertex=%d, want 0", result[0])
	}

	// In BFS, 1 and 2 should appear before 3,4,5,6
	indexOf := func(v int) int {
		for i, x := range result {
			if x == v {
				return i
			}
		}
		return -1
	}

	for _, level1 := range []int{1, 2} {
		for _, level2 := range []int{3, 4, 5, 6} {
			if indexOf(level1) > indexOf(level2) {
				t.Fatalf("BFS order: %d should appear before %d", level1, level2)
			}
		}
	}
}

// ============================================================================
// DFS Tests
// ============================================================================

func TestDFS_EmptyGraph(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()

	result := h.DFS("A")

	if result != nil {
		t.Fatalf("DFS on empty graph: got %v, want nil", result)
	}
}

func TestDFS_SingleVertex(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	h.AddVertex("A")

	result := h.DFS("A")

	if len(result) != 1 || result[0] != "A" {
		t.Fatalf("DFS single vertex: got %v, want [A]", result)
	}
}

func TestDFS_NonExistentStart(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C"})

	result := h.DFS("Z")

	if result != nil {
		t.Fatalf("DFS non-existent start: got %v, want nil", result)
	}
}

func TestDFS_LinearChain(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})
	_ = h.AddEdge("E3", []string{"C", "D"})

	result := h.DFS("A")

	if len(result) != 4 {
		t.Fatalf("DFS linear chain: got %d vertices, want 4", len(result))
	}

	// All vertices should be reachable
	seen := make(map[string]bool)
	for _, v := range result {
		seen[v] = true
	}
	for _, v := range []string{"A", "B", "C", "D"} {
		if !seen[v] {
			t.Fatalf("DFS linear chain: missing vertex %s", v)
		}
	}
}

func TestDFS_Cycle(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})
	_ = h.AddEdge("E3", []string{"C", "A"}) // Cycle back to A

	result := h.DFS("A")

	if len(result) != 3 {
		t.Fatalf("DFS cycle: got %d vertices, want 3", len(result))
	}

	// First vertex should be start
	if result[0] != "A" {
		t.Fatalf("DFS cycle: first vertex=%s, want A", result[0])
	}
}

func TestDFS_DisconnectedComponent(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"C", "D"}) // Disconnected from A,B

	result := h.DFS("A")

	// Should only reach A and B
	if len(result) != 2 {
		t.Fatalf("DFS disconnected: got %d vertices, want 2", len(result))
	}

	seen := make(map[string]bool)
	for _, v := range result {
		seen[v] = true
	}
	if !seen["A"] || !seen["B"] {
		t.Fatalf("DFS disconnected: expected A and B, got %v", result)
	}
	if seen["C"] || seen["D"] {
		t.Fatalf("DFS disconnected: should not reach C or D, got %v", result)
	}
}

func TestDFS_NoDuplicates(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	// Multiple paths to D: A-B-D and A-C-D
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"A", "C"})
	_ = h.AddEdge("E3", []string{"B", "D"})
	_ = h.AddEdge("E4", []string{"C", "D"})

	result := h.DFS("A")

	// Check for duplicates
	seen := make(map[string]int)
	for _, v := range result {
		seen[v]++
		if seen[v] > 1 {
			t.Fatalf("DFS duplicates: vertex %s appeared %d times", v, seen[v])
		}
	}

	if len(result) != 4 {
		t.Fatalf("DFS no duplicates: got %d vertices, want 4", len(result))
	}
}

func TestDFS_DepthFirst(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[int]()
	// Create a simple tree: 0 -> 1,2 -> 1 has 3,4
	_ = h.AddEdge("E1", []int{0, 1})
	_ = h.AddEdge("E2", []int{0, 2})
	_ = h.AddEdge("E3", []int{1, 3})
	_ = h.AddEdge("E4", []int{1, 4})

	result := h.DFS(0)

	// First vertex should be 0
	if result[0] != 0 {
		t.Fatalf("DFS order: first vertex=%d, want 0", result[0])
	}

	// All vertices should be reached
	if len(result) != 5 {
		t.Fatalf("DFS depth first: got %d vertices, want 5", len(result))
	}
}

// ============================================================================
// ConnectedComponents Tests
// ============================================================================

func TestConnectedComponents_EmptyGraph(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()

	components := h.ConnectedComponents()

	if len(components) != 0 {
		t.Fatalf("ConnectedComponents empty graph: got %d components, want 0", len(components))
	}
}

func TestConnectedComponents_SingleVertex(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	h.AddVertex("A")

	components := h.ConnectedComponents()

	if len(components) != 1 {
		t.Fatalf("ConnectedComponents single vertex: got %d components, want 1", len(components))
	}
	if len(components[0]) != 1 || components[0][0] != "A" {
		t.Fatalf("ConnectedComponents single vertex: got %v, want [[A]]", components)
	}
}

func TestConnectedComponents_SingleComponent(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})
	_ = h.AddEdge("E3", []string{"C", "D"})

	components := h.ConnectedComponents()

	if len(components) != 1 {
		t.Fatalf("ConnectedComponents single: got %d components, want 1", len(components))
	}
	if len(components[0]) != 4 {
		t.Fatalf("ConnectedComponents single: got %d vertices, want 4", len(components[0]))
	}
}

func TestConnectedComponents_TwoComponents(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"C", "D"})

	components := h.ConnectedComponents()

	if len(components) != 2 {
		t.Fatalf("ConnectedComponents two: got %d components, want 2", len(components))
	}

	// Each component should have 2 vertices
	for i, comp := range components {
		if len(comp) != 2 {
			t.Fatalf("ConnectedComponents two: component %d has %d vertices, want 2", i, len(comp))
		}
	}
}

func TestConnectedComponents_ThreeComponents(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"C", "D"})
	_ = h.AddEdge("E3", []string{"E", "F", "G"})

	components := h.ConnectedComponents()

	if len(components) != 3 {
		t.Fatalf("ConnectedComponents three: got %d components, want 3", len(components))
	}

	// Count total vertices
	total := 0
	for _, comp := range components {
		total += len(comp)
	}
	if total != 7 {
		t.Fatalf("ConnectedComponents three: total vertices=%d, want 7", total)
	}
}

func TestConnectedComponents_IsolatedVertices(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	h.AddVertex("A")
	h.AddVertex("B")
	h.AddVertex("C")

	components := h.ConnectedComponents()

	// Each isolated vertex is its own component
	if len(components) != 3 {
		t.Fatalf("ConnectedComponents isolated: got %d components, want 3", len(components))
	}

	for i, comp := range components {
		if len(comp) != 1 {
			t.Fatalf("ConnectedComponents isolated: component %d has %d vertices, want 1", i, len(comp))
		}
	}
}

func TestConnectedComponents_MixedIsolatedAndConnected(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C"})
	h.AddVertex("X") // Isolated
	h.AddVertex("Y") // Isolated

	components := h.ConnectedComponents()

	if len(components) != 3 {
		t.Fatalf("ConnectedComponents mixed: got %d components, want 3", len(components))
	}

	// Find the large component
	foundLarge := false
	for _, comp := range components {
		if len(comp) == 3 {
			foundLarge = true
		}
	}
	if !foundLarge {
		t.Fatal("ConnectedComponents mixed: did not find component of size 3")
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkBFS_Large(b *testing.B) {
	h := NewHypergraph[int]()
	// Create 2000 vertices with 4000 edges, each with 5 members
	N, E, k := 2000, 4000, 5
	for i := 0; i < E; i++ {
		members := make([]int, 0, k)
		start := (i * 3) % (N - k)
		for j := 0; j < k; j++ {
			members = append(members, start+j)
		}
		_ = h.AddEdge("E"+intToStr(i), members)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.BFS(0)
	}
}

func BenchmarkDFS_Large(b *testing.B) {
	h := NewHypergraph[int]()
	// Same graph as BFS benchmark
	N, E, k := 2000, 4000, 5
	for i := 0; i < E; i++ {
		members := make([]int, 0, k)
		start := (i * 3) % (N - k)
		for j := 0; j < k; j++ {
			members = append(members, start+j)
		}
		_ = h.AddEdge("E"+intToStr(i), members)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.DFS(0)
	}
}

func BenchmarkConnectedComponents(b *testing.B) {
	h := NewHypergraph[int]()
	// Create 500 vertices with 5 disconnected components
	// Each component has ~100 vertices
	for comp := 0; comp < 5; comp++ {
		base := comp * 100
		for i := 0; i < 50; i++ {
			members := []int{base + i*2, base + i*2 + 1}
			_ = h.AddEdge("E"+intToStr(comp)+"_"+intToStr(i), members)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.ConnectedComponents()
	}
}

// Helper for benchmark edge IDs
func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + intToStr(-i)
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}
