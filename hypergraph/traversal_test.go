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
