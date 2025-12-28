package hypergraph

import (
	"slices"
	"testing"
)

// ============================================================================
// Dual Tests
// ============================================================================

func TestDual_Basic(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})

	dual := h.Dual()

	// Dual vertices = original edges
	if dual.NumVertices() != h.NumEdges() {
		t.Fatalf("dual vertices=%d want %d (original edges)", dual.NumVertices(), h.NumEdges())
	}

	// Dual edges correspond to original vertices with at least one edge membership
	// A is in E1, B is in E1,E2, C is in E2 → 3 edges in dual
	if dual.NumEdges() != 3 {
		t.Fatalf("dual edges=%d want 3", dual.NumEdges())
	}
}

func TestDual_VertexEdgeSwap(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C"})
	_ = h.AddEdge("E2", []string{"C", "D"})

	dual := h.Dual()

	// Original: 2 edges, 4 vertices
	// Dual: vertices are {E1, E2}, edges correspond to {A, B, C, D} (but only those with memberships)

	// Vertices in dual = original edge IDs
	dualVerts := dual.Vertices()
	slices.Sort(dualVerts)
	if len(dualVerts) != 2 || dualVerts[0] != "E1" || dualVerts[1] != "E2" {
		t.Fatalf("dual vertices=%v want [E1 E2]", dualVerts)
	}
}

func TestDual_IncidencePreservation(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})

	dual := h.Dual()

	// If A was in E1 in original, then E1 should be in edge "A" in dual
	// Edge "A" in dual should contain exactly E1
	aEdgeSize, ok := dual.EdgeSize("A")
	if !ok {
		t.Fatal("dual should have edge 'A'")
	}
	if aEdgeSize != 1 {
		t.Fatalf("edge A size=%d want 1", aEdgeSize)
	}

	// Edge "B" should contain E1 and E2 (B was in both)
	bEdgeSize, ok := dual.EdgeSize("B")
	if !ok {
		t.Fatal("dual should have edge 'B'")
	}
	if bEdgeSize != 2 {
		t.Fatalf("edge B size=%d want 2", bEdgeSize)
	}
}

func TestDual_Empty(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()

	dual := h.Dual()

	if dual.NumVertices() != 0 || dual.NumEdges() != 0 {
		t.Fatalf("dual of empty: V=%d E=%d, want 0,0", dual.NumVertices(), dual.NumEdges())
	}
}

func TestDual_SingleVertexEdge(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A"})

	dual := h.Dual()

	// Dual has vertex E1, edge A containing E1
	if dual.NumVertices() != 1 {
		t.Fatalf("dual vertices=%d want 1", dual.NumVertices())
	}
	if dual.NumEdges() != 1 {
		t.Fatalf("dual edges=%d want 1", dual.NumEdges())
	}
}

// ============================================================================
// TwoSection Tests
// ============================================================================

func TestTwoSection_SingleEdge(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C"})

	g := h.TwoSection()

	// 3-element edge → complete graph K3 with 3 edges
	if len(g.Vertices()) != 3 {
		t.Fatalf("2-section vertices=%d want 3", len(g.Vertices()))
	}
	// C(3,2) = 3 edges
	if len(g.Edges()) != 3 {
		t.Fatalf("2-section edges=%d want 3", len(g.Edges()))
	}
}

func TestTwoSection_DisjointEdges(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"C", "D"})

	g := h.TwoSection()

	// 4 vertices, 2 edges (A-B from E1, C-D from E2)
	if len(g.Vertices()) != 4 {
		t.Fatalf("2-section vertices=%d want 4", len(g.Vertices()))
	}
	if len(g.Edges()) != 2 {
		t.Fatalf("2-section edges=%d want 2", len(g.Edges()))
	}

	// No cross-edges between disjoint components
	edges := g.Edges()
	for _, e := range edges {
		// Each edge should connect within same hyperedge
		sameE1 := (e.From == "A" || e.From == "B") && (e.To == "A" || e.To == "B")
		sameE2 := (e.From == "C" || e.From == "D") && (e.To == "C" || e.To == "D")
		if !sameE1 && !sameE2 {
			t.Fatalf("unexpected cross-edge: %v-%v", e.From, e.To)
		}
	}
}

func TestTwoSection_SharedVertex(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"X", "A"})
	_ = h.AddEdge("E2", []string{"X", "B"})

	g := h.TwoSection()

	// 3 vertices: X, A, B
	// 2 edges: X-A, X-B (A and B are NOT connected directly)
	if len(g.Vertices()) != 3 {
		t.Fatalf("2-section vertices=%d want 3", len(g.Vertices()))
	}
	if len(g.Edges()) != 2 {
		t.Fatalf("2-section edges=%d want 2", len(g.Edges()))
	}
}

func TestTwoSection_EdgeDeduplication(t *testing.T) {
	t.Parallel()
	// Two edges both containing A and B should only create one A-B edge
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"A", "B", "C"})

	g := h.TwoSection()

	// A-B appears in both edges but should only be one graph edge
	// Total edges: A-B, A-C, B-C = 3
	if len(g.Edges()) != 3 {
		t.Fatalf("2-section edges=%d want 3 (deduplication)", len(g.Edges()))
	}
}

func TestTwoSection_Empty(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()

	g := h.TwoSection()

	if len(g.Vertices()) != 0 || len(g.Edges()) != 0 {
		t.Fatalf("2-section of empty: V=%d E=%d, want 0,0", len(g.Vertices()), len(g.Edges()))
	}
}

func TestTwoSection_LargeEdge(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C", "D", "E"})

	g := h.TwoSection()

	// 5-element edge → K5 with C(5,2) = 10 edges
	if len(g.Vertices()) != 5 {
		t.Fatalf("2-section vertices=%d want 5", len(g.Vertices()))
	}
	if len(g.Edges()) != 10 {
		t.Fatalf("2-section edges=%d want 10 (C(5,2))", len(g.Edges()))
	}
}

// ============================================================================
// LineGraph Tests
// ============================================================================

func TestLineGraph_DisjointEdges(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"C", "D"})

	g := h.LineGraph()

	// Two disjoint edges → 2 vertices, 0 edges (no shared vertices)
	if len(g.Vertices()) != 2 {
		t.Fatalf("line graph vertices=%d want 2", len(g.Vertices()))
	}
	if len(g.Edges()) != 0 {
		t.Fatalf("line graph edges=%d want 0 (disjoint)", len(g.Edges()))
	}
}

func TestLineGraph_IntersectingEdges(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})

	g := h.LineGraph()

	// E1 and E2 share B → connected in line graph
	if len(g.Vertices()) != 2 {
		t.Fatalf("line graph vertices=%d want 2", len(g.Vertices()))
	}
	if len(g.Edges()) != 1 {
		t.Fatalf("line graph edges=%d want 1 (E1-E2)", len(g.Edges()))
	}
}

func TestLineGraph_Star(t *testing.T) {
	t.Parallel()
	// All edges share vertex X → line graph is complete
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"X", "A"})
	_ = h.AddEdge("E2", []string{"X", "B"})
	_ = h.AddEdge("E3", []string{"X", "C"})

	g := h.LineGraph()

	// 3 vertices (E1, E2, E3), all connected → K3 with 3 edges
	if len(g.Vertices()) != 3 {
		t.Fatalf("line graph vertices=%d want 3", len(g.Vertices()))
	}
	if len(g.Edges()) != 3 {
		t.Fatalf("line graph edges=%d want 3 (complete K3)", len(g.Edges()))
	}
}

func TestLineGraph_Empty(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()

	g := h.LineGraph()

	if len(g.Vertices()) != 0 || len(g.Edges()) != 0 {
		t.Fatalf("line graph of empty: V=%d E=%d, want 0,0", len(g.Vertices()), len(g.Edges()))
	}
}

func TestLineGraph_SingleEdge(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C"})

	g := h.LineGraph()

	// Single edge → single vertex, no edges
	if len(g.Vertices()) != 1 {
		t.Fatalf("line graph vertices=%d want 1", len(g.Vertices()))
	}
	if len(g.Edges()) != 0 {
		t.Fatalf("line graph edges=%d want 0", len(g.Edges()))
	}
}

func TestLineGraph_Chain(t *testing.T) {
	t.Parallel()
	// Chain: E1-E2-E3-E4 (each consecutive pair shares one vertex)
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})
	_ = h.AddEdge("E3", []string{"C", "D"})
	_ = h.AddEdge("E4", []string{"D", "E"})

	g := h.LineGraph()

	// 4 vertices, 3 edges (E1-E2, E2-E3, E3-E4)
	if len(g.Vertices()) != 4 {
		t.Fatalf("line graph vertices=%d want 4", len(g.Vertices()))
	}
	if len(g.Edges()) != 3 {
		t.Fatalf("line graph edges=%d want 3 (chain)", len(g.Edges()))
	}
}

// ============================================================================
// Graph Method Tests
// ============================================================================

func TestGraph_VerticesAndEdges(t *testing.T) {
	t.Parallel()
	g := NewGraph[string]()
	g.vertices["A"] = struct{}{}
	g.vertices["B"] = struct{}{}
	g.edges["A-B"] = struct{ From, To string }{"A", "B"}

	verts := g.Vertices()
	if len(verts) != 2 {
		t.Fatalf("vertices=%d want 2", len(verts))
	}

	edges := g.Edges()
	if len(edges) != 1 {
		t.Fatalf("edges=%d want 1", len(edges))
	}
	if edges[0].From != "A" || edges[0].To != "B" {
		t.Fatalf("edge=%v want A-B", edges[0])
	}
}
