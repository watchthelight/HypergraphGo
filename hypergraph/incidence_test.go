package hypergraph

import (
	"testing"
)

// ============================================================================
// IncidenceMatrix Tests
// ============================================================================

func TestIncidenceMatrix_Basic(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})

	vIdx, eIdx, coo := h.IncidenceMatrix()

	// 3 vertices, 2 edges
	if len(vIdx) != 3 {
		t.Fatalf("vertex index size=%d want 3", len(vIdx))
	}
	if len(eIdx) != 2 {
		t.Fatalf("edge index size=%d want 2", len(eIdx))
	}

	// Total memberships: E1 has 2, E2 has 2 → 4 non-zeros
	if len(coo.Rows) != 4 || len(coo.Cols) != 4 {
		t.Fatalf("COO size: rows=%d cols=%d want 4,4", len(coo.Rows), len(coo.Cols))
	}
}

func TestIncidenceMatrix_IndexStability(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})

	// Call twice and verify indices are identical
	vIdx1, eIdx1, coo1 := h.IncidenceMatrix()
	vIdx2, eIdx2, coo2 := h.IncidenceMatrix()

	// Vertex indices should match
	for v, idx := range vIdx1 {
		if vIdx2[v] != idx {
			t.Fatalf("vertex index instability: %s got %d then %d", v, idx, vIdx2[v])
		}
	}

	// Edge indices should match
	for e, idx := range eIdx1 {
		if eIdx2[e] != idx {
			t.Fatalf("edge index instability: %s got %d then %d", e, idx, eIdx2[e])
		}
	}

	// COO should match
	if len(coo1.Rows) != len(coo2.Rows) {
		t.Fatalf("COO size instability")
	}
}

func TestIncidenceMatrix_SortedIndices(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E2", []string{"C", "A"})
	_ = h.AddEdge("E1", []string{"B", "D"})

	vIdx, eIdx, _ := h.IncidenceMatrix()

	// Vertices should be sorted: A=0, B=1, C=2, D=3
	if vIdx["A"] != 0 || vIdx["B"] != 1 || vIdx["C"] != 2 || vIdx["D"] != 3 {
		t.Fatalf("vertex indices not sorted: %v", vIdx)
	}

	// Edges should be sorted: E1=0, E2=1
	if eIdx["E1"] != 0 || eIdx["E2"] != 1 {
		t.Fatalf("edge indices not sorted: %v", eIdx)
	}
}

func TestIncidenceMatrix_IndexBounds(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C"})
	_ = h.AddEdge("E2", []string{"C", "D"})

	vIdx, eIdx, coo := h.IncidenceMatrix()

	numV := len(vIdx)
	numE := len(eIdx)

	for i, row := range coo.Rows {
		if row < 0 || row >= numV {
			t.Fatalf("row index %d out of bounds [0,%d) at position %d", row, numV, i)
		}
	}

	for i, col := range coo.Cols {
		if col < 0 || col >= numE {
			t.Fatalf("col index %d out of bounds [0,%d) at position %d", col, numE, i)
		}
	}
}

func TestIncidenceMatrix_Reconstruction(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})

	vIdx, eIdx, coo := h.IncidenceMatrix()

	// Invert the index maps
	vName := make(map[int]string)
	for v, idx := range vIdx {
		vName[idx] = v
	}
	eName := make(map[int]string)
	for e, idx := range eIdx {
		eName[idx] = e
	}

	// Reconstruct edge memberships from COO
	reconstructed := make(map[string]map[string]bool)
	for i := range coo.Rows {
		v := vName[coo.Rows[i]]
		e := eName[coo.Cols[i]]
		if reconstructed[e] == nil {
			reconstructed[e] = make(map[string]bool)
		}
		reconstructed[e][v] = true
	}

	// Verify E1 contains A, B
	if !reconstructed["E1"]["A"] || !reconstructed["E1"]["B"] {
		t.Fatalf("E1 reconstruction failed: %v", reconstructed["E1"])
	}
	if len(reconstructed["E1"]) != 2 {
		t.Fatalf("E1 size=%d want 2", len(reconstructed["E1"]))
	}

	// Verify E2 contains B, C
	if !reconstructed["E2"]["B"] || !reconstructed["E2"]["C"] {
		t.Fatalf("E2 reconstruction failed: %v", reconstructed["E2"])
	}
	if len(reconstructed["E2"]) != 2 {
		t.Fatalf("E2 size=%d want 2", len(reconstructed["E2"]))
	}
}

func TestIncidenceMatrix_RowSumsVertexDegrees(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})
	_ = h.AddEdge("E3", []string{"B", "D"})

	vIdx, _, coo := h.IncidenceMatrix()

	// Count row occurrences
	rowCounts := make(map[int]int)
	for _, row := range coo.Rows {
		rowCounts[row]++
	}

	// Verify row counts match vertex degrees
	// A: degree 1, B: degree 3, C: degree 1, D: degree 1
	expectedDegrees := map[string]int{"A": 1, "B": 3, "C": 1, "D": 1}
	for v, expectedDeg := range expectedDegrees {
		actualDeg := rowCounts[vIdx[v]]
		if actualDeg != expectedDeg {
			t.Fatalf("vertex %s: row count=%d, expected degree=%d", v, actualDeg, expectedDeg)
		}
	}
}

func TestIncidenceMatrix_ColSumsEdgeSizes(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C"})
	_ = h.AddEdge("E2", []string{"D", "E"})

	_, eIdx, coo := h.IncidenceMatrix()

	// Count column occurrences
	colCounts := make(map[int]int)
	for _, col := range coo.Cols {
		colCounts[col]++
	}

	// Verify column counts match edge sizes
	// E1: size 3, E2: size 2
	if colCounts[eIdx["E1"]] != 3 {
		t.Fatalf("E1 col count=%d want 3", colCounts[eIdx["E1"]])
	}
	if colCounts[eIdx["E2"]] != 2 {
		t.Fatalf("E2 col count=%d want 2", colCounts[eIdx["E2"]])
	}
}

func TestIncidenceMatrix_Empty(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()

	vIdx, eIdx, coo := h.IncidenceMatrix()

	if len(vIdx) != 0 {
		t.Fatalf("empty hypergraph vertex index=%d want 0", len(vIdx))
	}
	if len(eIdx) != 0 {
		t.Fatalf("empty hypergraph edge index=%d want 0", len(eIdx))
	}
	if len(coo.Rows) != 0 || len(coo.Cols) != 0 {
		t.Fatalf("empty hypergraph COO: rows=%d cols=%d want 0,0", len(coo.Rows), len(coo.Cols))
	}
}

func TestIncidenceMatrix_SingleElement(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A"})

	vIdx, eIdx, coo := h.IncidenceMatrix()

	if len(vIdx) != 1 || vIdx["A"] != 0 {
		t.Fatalf("vertex index: %v", vIdx)
	}
	if len(eIdx) != 1 || eIdx["E1"] != 0 {
		t.Fatalf("edge index: %v", eIdx)
	}
	if len(coo.Rows) != 1 || len(coo.Cols) != 1 {
		t.Fatalf("COO size: rows=%d cols=%d want 1,1", len(coo.Rows), len(coo.Cols))
	}
	if coo.Rows[0] != 0 || coo.Cols[0] != 0 {
		t.Fatalf("COO entry: (%d,%d) want (0,0)", coo.Rows[0], coo.Cols[0])
	}
}

func TestIncidenceMatrix_NoDuplicates(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B"})
	_ = h.AddEdge("E2", []string{"B", "C"})

	_, _, coo := h.IncidenceMatrix()

	// Check for duplicate (row,col) pairs
	seen := make(map[[2]int]bool)
	for i := range coo.Rows {
		pair := [2]int{coo.Rows[i], coo.Cols[i]}
		if seen[pair] {
			t.Fatalf("duplicate COO entry: (%d,%d)", pair[0], pair[1])
		}
		seen[pair] = true
	}
}
