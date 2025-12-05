package hypergraph

import (
	"testing"
)

// TestEmptyHypergraph tests operations on an empty hypergraph.
func TestEmptyHypergraph(t *testing.T) {
	h := NewHypergraph[string]()

	if h.NumVertices() != 0 {
		t.Errorf("expected 0 vertices, got %d", h.NumVertices())
	}
	if h.NumEdges() != 0 {
		t.Errorf("expected 0 edges, got %d", h.NumEdges())
	}
	if !h.IsEmpty() {
		t.Error("expected IsEmpty() to be true")
	}
	if len(h.Vertices()) != 0 {
		t.Error("expected empty vertices slice")
	}
	if len(h.Edges()) != 0 {
		t.Error("expected empty edges slice")
	}

	// Algorithms should handle empty graphs gracefully
	hittingSet := h.GreedyHittingSet()
	if len(hittingSet) != 0 {
		t.Errorf("expected empty hitting set, got %v", hittingSet)
	}

	coloring := h.GreedyColoring()
	if len(coloring) != 0 {
		t.Errorf("expected empty coloring, got %v", coloring)
	}

	components := h.ConnectedComponents()
	if len(components) != 0 {
		t.Errorf("expected no connected components, got %v", components)
	}
}

// TestRemoveVertex tests vertex removal with cascade effects.
func TestRemoveVertex(t *testing.T) {
	h := NewHypergraph[string]()
	h.AddEdge("E1", []string{"A", "B", "C"})
	h.AddEdge("E2", []string{"B", "C", "D"})
	h.AddEdge("E3", []string{"D", "E"})

	// Remove B - should affect E1 and E2
	h.RemoveVertex("B")

	if h.HasVertex("B") {
		t.Error("vertex B should be removed")
	}
	if h.NumVertices() != 4 {
		t.Errorf("expected 4 vertices after removal, got %d", h.NumVertices())
	}

	// E1 should still exist but without B
	size, ok := h.EdgeSize("E1")
	if !ok {
		t.Error("E1 should still exist")
	}
	if size != 2 {
		t.Errorf("E1 should have 2 members, got %d", size)
	}

	// Remove vertex not in graph - should be no-op
	h.RemoveVertex("Z")
	if h.NumVertices() != 4 {
		t.Error("removing non-existent vertex should not change count")
	}
}

// TestRemoveEdge tests edge removal.
func TestRemoveEdge(t *testing.T) {
	h := NewHypergraph[string]()
	h.AddEdge("E1", []string{"A", "B"})
	h.AddEdge("E2", []string{"B", "C"})

	h.RemoveEdge("E1")

	if h.HasEdge("E1") {
		t.Error("E1 should be removed")
	}
	if h.NumEdges() != 1 {
		t.Errorf("expected 1 edge, got %d", h.NumEdges())
	}

	// Vertices should still exist
	if !h.HasVertex("A") || !h.HasVertex("B") {
		t.Error("vertices should remain after edge removal")
	}

	// Degree of A should now be 0
	if h.VertexDegree("A") != 0 {
		t.Errorf("expected degree 0 for A, got %d", h.VertexDegree("A"))
	}

	// Remove non-existent edge - should be no-op
	h.RemoveEdge("E99")
	if h.NumEdges() != 1 {
		t.Error("removing non-existent edge should not change count")
	}
}

// TestBFSInvalidStart tests BFS with non-existent start vertex.
func TestBFSInvalidStart(t *testing.T) {
	h := NewHypergraph[string]()
	h.AddEdge("E1", []string{"A", "B", "C"})

	result := h.BFS("Z")
	if result != nil {
		t.Errorf("BFS with invalid start should return nil, got %v", result)
	}

	result = h.DFS("Z")
	if result != nil {
		t.Errorf("DFS with invalid start should return nil, got %v", result)
	}
}

// TestIsolatedVertex tests behavior with isolated vertices.
func TestIsolatedVertex(t *testing.T) {
	h := NewHypergraph[string]()
	h.AddVertex("A")
	h.AddVertex("B")
	h.AddEdge("E1", []string{"C", "D"})

	if h.NumVertices() != 4 {
		t.Errorf("expected 4 vertices, got %d", h.NumVertices())
	}

	// Degree of isolated vertices should be 0
	if h.VertexDegree("A") != 0 {
		t.Errorf("expected degree 0 for isolated A, got %d", h.VertexDegree("A"))
	}

	// BFS from isolated vertex should return just that vertex
	result := h.BFS("A")
	if len(result) != 1 || result[0] != "A" {
		t.Errorf("BFS from isolated vertex should return [A], got %v", result)
	}

	// Connected components should include isolated vertices
	components := h.ConnectedComponents()
	if len(components) != 3 {
		t.Errorf("expected 3 components (A, B, {C,D}), got %d", len(components))
	}
}

// TestDisconnectedComponents tests BFS on multi-component graphs.
func TestDisconnectedComponents(t *testing.T) {
	h := NewHypergraph[string]()
	h.AddEdge("E1", []string{"A", "B"})
	h.AddEdge("E2", []string{"B", "C"})
	h.AddEdge("E3", []string{"X", "Y"})
	h.AddEdge("E4", []string{"Y", "Z"})

	// BFS from A should not reach X, Y, Z
	reachable := h.BFS("A")
	reachableSet := make(map[string]bool)
	for _, v := range reachable {
		reachableSet[v] = true
	}

	if reachableSet["X"] || reachableSet["Y"] || reachableSet["Z"] {
		t.Error("BFS from A should not reach X, Y, Z")
	}
	if !reachableSet["A"] || !reachableSet["B"] || !reachableSet["C"] {
		t.Error("BFS from A should reach A, B, C")
	}

	// ConnectedComponents should find 2 components
	components := h.ConnectedComponents()
	if len(components) != 2 {
		t.Errorf("expected 2 connected components, got %d", len(components))
	}
}

// TestDFSTraversal tests DFS correctness.
func TestDFSTraversal(t *testing.T) {
	h := NewHypergraph[string]()
	h.AddEdge("E1", []string{"A", "B"})
	h.AddEdge("E2", []string{"B", "C"})
	h.AddEdge("E3", []string{"C", "D"})

	result := h.DFS("A")

	// All vertices should be reachable
	if len(result) != 4 {
		t.Errorf("DFS should visit 4 vertices, got %d", len(result))
	}

	// First vertex should be start
	if result[0] != "A" {
		t.Errorf("DFS should start with A, got %s", result[0])
	}

	// Check all vertices are unique
	seen := make(map[string]bool)
	for _, v := range result {
		if seen[v] {
			t.Errorf("DFS visited %s twice", v)
		}
		seen[v] = true
	}
}

// TestLargeEdges tests edges with many vertices.
func TestLargeEdges(t *testing.T) {
	h := NewHypergraph[int]()

	// Create edge with 100 vertices
	members := make([]int, 100)
	for i := 0; i < 100; i++ {
		members[i] = i
	}
	err := h.AddEdge("large", members)
	if err != nil {
		t.Fatalf("failed to add large edge: %v", err)
	}

	size, ok := h.EdgeSize("large")
	if !ok || size != 100 {
		t.Errorf("expected edge size 100, got %d", size)
	}

	// All vertices should be connected via single edge
	components := h.ConnectedComponents()
	if len(components) != 1 {
		t.Errorf("expected 1 component, got %d", len(components))
	}
}

// BenchmarkAddEdge benchmarks edge insertion performance.
func BenchmarkAddEdge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		h := NewHypergraph[int]()
		for j := 0; j < 100; j++ {
			h.AddEdge(string(rune('A'+j)), []int{j, j + 1, j + 2})
		}
	}
}

// BenchmarkRemoveVertex benchmarks vertex removal performance.
func BenchmarkRemoveVertex(b *testing.B) {
	// Setup
	h := NewHypergraph[int]()
	for j := 0; j < 100; j++ {
		h.AddEdge(string(rune('A'+j)), []int{j, j + 1, j + 2})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hCopy := h.Copy()
		for j := 0; j < 50; j++ {
			hCopy.RemoveVertex(j)
		}
	}
}
