package hypergraph

import (
	"bytes"
	"slices"
	"testing"
)

// TestAddRemoveVertexInvariant tests that adding and removing a vertex
// returns the graph to its original state.
func TestAddRemoveVertexInvariant(t *testing.T) {
	h := NewHypergraph[int]()
	h.AddVertex(1)
	h.AddVertex(2)
	h.AddVertex(3)

	originalCount := h.NumVertices()
	originalVertices := h.Vertices()
	slices.Sort(originalVertices)

	// Add a new vertex
	h.AddVertex(4)
	if h.NumVertices() != originalCount+1 {
		t.Errorf("After add: expected %d vertices, got %d", originalCount+1, h.NumVertices())
	}
	if !h.HasVertex(4) {
		t.Error("Added vertex not found")
	}

	// Remove it
	h.RemoveVertex(4)
	if h.NumVertices() != originalCount {
		t.Errorf("After remove: expected %d vertices, got %d", originalCount, h.NumVertices())
	}
	if h.HasVertex(4) {
		t.Error("Removed vertex still found")
	}

	// Verify original vertices unchanged
	finalVertices := h.Vertices()
	slices.Sort(finalVertices)
	if !slices.Equal(originalVertices, finalVertices) {
		t.Errorf("Vertices changed:\n  Original: %v\n  Final:    %v", originalVertices, finalVertices)
	}
}

// TestAddRemoveEdgeInvariant tests that adding and removing an edge
// returns the graph to its original state.
func TestAddRemoveEdgeInvariant(t *testing.T) {
	h := NewHypergraph[int]()
	h.AddVertex(1)
	h.AddVertex(2)
	h.AddVertex(3)
	h.AddEdge("e1", []int{1, 2})

	originalEdgeCount := h.NumEdges()
	originalEdges := h.Edges()
	slices.Sort(originalEdges)

	// Add a new edge
	err := h.AddEdge("e2", []int{2, 3})
	if err != nil {
		t.Fatalf("Failed to add edge: %v", err)
	}
	if h.NumEdges() != originalEdgeCount+1 {
		t.Errorf("After add: expected %d edges, got %d", originalEdgeCount+1, h.NumEdges())
	}
	if !h.HasEdge("e2") {
		t.Error("Added edge not found")
	}

	// Remove it
	h.RemoveEdge("e2")
	if h.NumEdges() != originalEdgeCount {
		t.Errorf("After remove: expected %d edges, got %d", originalEdgeCount, h.NumEdges())
	}
	if h.HasEdge("e2") {
		t.Error("Removed edge still found")
	}

	// Verify original edges unchanged
	finalEdges := h.Edges()
	slices.Sort(finalEdges)
	if !slices.Equal(originalEdges, finalEdges) {
		t.Errorf("Edges changed:\n  Original: %v\n  Final:    %v", originalEdges, finalEdges)
	}
}

// TestCopyIndependence tests that Copy creates an independent copy.
func TestCopyIndependence(t *testing.T) {
	h1 := NewHypergraph[int]()
	h1.AddVertex(1)
	h1.AddVertex(2)
	h1.AddEdge("e", []int{1, 2})

	h2 := h1.Copy()

	// Modify h1
	h1.AddVertex(3)
	h1.AddEdge("e2", []int{1, 3})
	h1.RemoveVertex(2)

	// h2 should be unchanged
	if h2.NumVertices() != 2 {
		t.Errorf("Copy affected: expected 2 vertices, got %d", h2.NumVertices())
	}
	if !h2.HasVertex(2) {
		t.Error("Copy lost vertex 2")
	}
	if h2.HasVertex(3) {
		t.Error("Copy incorrectly has vertex 3")
	}
	if h2.NumEdges() != 1 {
		t.Errorf("Copy affected: expected 1 edge, got %d", h2.NumEdges())
	}
	if h2.HasEdge("e2") {
		t.Error("Copy incorrectly has edge e2")
	}
}

// TestJSONRoundTripPreservesStructure tests that JSON serialization
// preserves graph structure.
func TestJSONRoundTripPreservesStructure(t *testing.T) {
	h1 := NewHypergraph[string]()
	h1.AddVertex("a")
	h1.AddVertex("b")
	h1.AddVertex("c")
	h1.AddEdge("e1", []string{"a", "b"})
	h1.AddEdge("e2", []string{"b", "c"})
	h1.AddEdge("hyper", []string{"a", "b", "c"})

	// Save to JSON
	var buf bytes.Buffer
	if err := h1.SaveJSON(&buf); err != nil {
		t.Fatalf("SaveJSON failed: %v", err)
	}

	// Load from JSON
	h2, err := LoadJSON[string](&buf)
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}

	// Verify structure preserved
	if h1.NumVertices() != h2.NumVertices() {
		t.Errorf("NumVertices mismatch: %d vs %d", h1.NumVertices(), h2.NumVertices())
	}
	if h1.NumEdges() != h2.NumEdges() {
		t.Errorf("NumEdges mismatch: %d vs %d", h1.NumEdges(), h2.NumEdges())
	}

	for _, v := range h1.Vertices() {
		if !h2.HasVertex(v) {
			t.Errorf("Vertex %v lost in round-trip", v)
		}
		if h1.VertexDegree(v) != h2.VertexDegree(v) {
			t.Errorf("Degree of %v changed: %d vs %d",
				v, h1.VertexDegree(v), h2.VertexDegree(v))
		}
	}

	for _, e := range h1.Edges() {
		if !h2.HasEdge(e) {
			t.Errorf("Edge %v lost in round-trip", e)
		}
		size1, _ := h1.EdgeSize(e)
		size2, _ := h2.EdgeSize(e)
		if size1 != size2 {
			t.Errorf("Size of edge %v changed: %d vs %d",
				e, size1, size2)
		}
		members1 := h1.EdgeMembers(e)
		members2 := h2.EdgeMembers(e)
		slices.Sort(members1)
		slices.Sort(members2)
		if !slices.Equal(members1, members2) {
			t.Errorf("Members of edge %v changed:\n  Before: %v\n  After:  %v",
				e, members1, members2)
		}
	}
}

// TestEdgeMembershipInvariant tests that edge membership is consistent.
func TestEdgeMembershipInvariant(t *testing.T) {
	h := NewHypergraph[int]()
	h.AddVertex(1)
	h.AddVertex(2)
	h.AddVertex(3)
	h.AddEdge("e", []int{1, 2, 3})

	// All edge members should be vertices
	for _, v := range h.EdgeMembers("e") {
		if !h.HasVertex(v) {
			t.Errorf("Edge member %v is not a vertex", v)
		}
	}

	// Edge size should equal number of members
	size, _ := h.EdgeSize("e")
	if size != len(h.EdgeMembers("e")) {
		t.Errorf("EdgeSize(%d) != len(EdgeMembers(%d))", size, len(h.EdgeMembers("e")))
	}

	// Removing a vertex should remove it from edges or remove the edge
	h.RemoveVertex(2)
	if h.HasEdge("e") {
		// If edge still exists, it shouldn't contain 2
		for _, v := range h.EdgeMembers("e") {
			if v == 2 {
				t.Error("Removed vertex still in edge")
			}
		}
	}
}

// TestDegreeInvariant tests that vertex degree is consistent.
func TestDegreeInvariant(t *testing.T) {
	h := NewHypergraph[int]()
	h.AddVertex(1)
	h.AddVertex(2)
	h.AddVertex(3)
	h.AddEdge("e1", []int{1, 2})
	h.AddEdge("e2", []int{2, 3})
	h.AddEdge("e3", []int{1, 2, 3})

	// Degree should equal number of edges containing the vertex
	for _, v := range h.Vertices() {
		degree := h.VertexDegree(v)
		count := 0
		for _, e := range h.Edges() {
			for _, m := range h.EdgeMembers(e) {
				if m == v {
					count++
					break
				}
			}
		}
		if degree != count {
			t.Errorf("Degree of %v (%d) != count of containing edges (%d)", v, degree, count)
		}
	}
}

// TestIsEmptyInvariant tests that IsEmpty is consistent.
func TestIsEmptyInvariant(t *testing.T) {
	h := NewHypergraph[int]()

	// New graph should be empty
	if !h.IsEmpty() {
		t.Error("New graph should be empty")
	}
	if h.NumVertices() != 0 {
		t.Error("Empty graph should have 0 vertices")
	}
	if h.NumEdges() != 0 {
		t.Error("Empty graph should have 0 edges")
	}

	// Adding a vertex makes it non-empty
	h.AddVertex(1)
	if h.IsEmpty() {
		t.Error("Graph with vertex should not be empty")
	}

	// Removing all vertices makes it empty again
	h.RemoveVertex(1)
	if !h.IsEmpty() {
		t.Error("Graph with no vertices should be empty")
	}
}

// TestAddEdgeErrorCases tests error handling for AddEdge.
func TestAddEdgeErrorCases(t *testing.T) {
	h := NewHypergraph[int]()
	h.AddVertex(1)
	h.AddVertex(2)

	// Adding empty edge should fail
	err := h.AddEdge("empty", []int{})
	if err == nil {
		t.Error("AddEdge with empty members should fail")
	}

	// Adding edge with non-existent vertex auto-adds the vertex
	err = h.AddEdge("e", []int{1, 3})
	if err != nil {
		t.Errorf("AddEdge with non-existent vertex should auto-add it: %v", err)
	}
	if !h.HasVertex(3) {
		t.Error("Non-existent vertex should be auto-added")
	}

	// Adding edge with existing ID should fail
	err = h.AddEdge("e", []int{1, 2})
	if err == nil {
		t.Error("AddEdge with duplicate ID should fail")
	}
}

// TestVertexDegreeNonExistent tests VertexDegree for non-existent vertex.
func TestVertexDegreeNonExistent(t *testing.T) {
	h := NewHypergraph[int]()
	h.AddVertex(1)

	// Should return 0 for non-existent vertex
	degree := h.VertexDegree(999)
	if degree != 0 {
		t.Errorf("Degree of non-existent vertex should be 0, got %d", degree)
	}
}

// TestEdgeSizeNonExistent tests EdgeSize for non-existent edge.
func TestEdgeSizeNonExistent(t *testing.T) {
	h := NewHypergraph[int]()
	h.AddVertex(1)

	// Should return 0, false for non-existent edge
	size, ok := h.EdgeSize("nonexistent")
	if ok || size != 0 {
		t.Errorf("Size of non-existent edge should be 0, false; got %d, %v", size, ok)
	}
}

// TestEdgeMembersNonExistent tests EdgeMembers for non-existent edge.
func TestEdgeMembersNonExistent(t *testing.T) {
	h := NewHypergraph[int]()
	h.AddVertex(1)

	// Should return nil for non-existent edge
	members := h.EdgeMembers("nonexistent")
	if members != nil {
		t.Errorf("Members of non-existent edge should be nil, got %v", members)
	}
}

// TestRemoveVertexCascade tests that removing a vertex updates associated edges.
func TestRemoveVertexCascade(t *testing.T) {
	h := NewHypergraph[int]()
	h.AddVertex(1)
	h.AddVertex(2)
	h.AddVertex(3)
	h.AddEdge("e1", []int{1, 2})
	h.AddEdge("e2", []int{2, 3})

	// Remove vertex 2 - should remove it from edges, edges become single-vertex
	h.RemoveVertex(2)

	// Vertex 2 should be gone
	if h.HasVertex(2) {
		t.Error("Vertex 2 should be removed")
	}

	// Edges should remain but with reduced membership
	if !h.HasEdge("e1") {
		t.Error("Edge e1 should still exist with reduced membership")
	}
	if !h.HasEdge("e2") {
		t.Error("Edge e2 should still exist with reduced membership")
	}

	// Each edge should now have only 1 member
	size1, _ := h.EdgeSize("e1")
	size2, _ := h.EdgeSize("e2")
	if size1 != 1 {
		t.Errorf("Edge e1 should have 1 member after removing vertex 2, got %d", size1)
	}
	if size2 != 1 {
		t.Errorf("Edge e2 should have 1 member after removing vertex 2, got %d", size2)
	}

	// Test that removing the last vertex from an edge removes the edge
	h.RemoveVertex(1) // e1 becomes empty, should be deleted
	if h.HasEdge("e1") {
		t.Error("Edge e1 should be removed when it becomes empty")
	}
}

// TestDuplicateVertexAdd tests adding the same vertex twice.
func TestDuplicateVertexAdd(t *testing.T) {
	h := NewHypergraph[int]()
	h.AddVertex(1)
	countBefore := h.NumVertices()

	h.AddVertex(1) // Add again
	countAfter := h.NumVertices()

	if countBefore != countAfter {
		t.Errorf("Adding duplicate vertex changed count: %d -> %d", countBefore, countAfter)
	}
}

// TestEmptyEdgeRejected tests that empty edges are rejected.
func TestEmptyEdgeRejected(t *testing.T) {
	h := NewHypergraph[int]()

	// Empty edges should be rejected
	err := h.AddEdge("empty", []int{})
	if err == nil {
		t.Error("Empty edge should be rejected")
	}

	if h.HasEdge("empty") {
		t.Error("Empty edge should not exist")
	}
}

// TestSelfLoop tests edge with single vertex (self-loop).
func TestSelfLoop(t *testing.T) {
	h := NewHypergraph[int]()
	h.AddVertex(1)

	err := h.AddEdge("loop", []int{1})
	if err != nil {
		t.Fatalf("Failed to add self-loop: %v", err)
	}

	if !h.HasEdge("loop") {
		t.Error("Self-loop edge should exist")
	}
	size, ok := h.EdgeSize("loop")
	if !ok || size != 1 {
		t.Errorf("Self-loop edge should have size 1, true; got %d, %v", size, ok)
	}
	if h.VertexDegree(1) != 1 {
		t.Errorf("Vertex in self-loop should have degree 1, got %d", h.VertexDegree(1))
	}
}
