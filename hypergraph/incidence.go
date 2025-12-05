package hypergraph

import (
	"slices"
	"sort"
)

// COO represents a sparse matrix in coordinate format.
type COO struct {
	Rows []int
	Cols []int
}

// IncidenceMatrix returns the incidence matrix in COO format with stable vertex and edge indices.
func (h *Hypergraph[V]) IncidenceMatrix() (vertexIndex map[V]int, edgeIndex map[string]int, coo COO) {
	// Create stable vertex index
	vertices := h.Vertices()
	slices.Sort(vertices)
	vertexIndex = make(map[V]int)
	for i, v := range vertices {
		vertexIndex[v] = i
	}

	// Create stable edge index
	edges := h.Edges()
	sort.Strings(edges)
	edgeIndex = make(map[string]int)
	for i, e := range edges {
		edgeIndex[e] = i
	}

	// Build COO
	var rows, cols []int
	for _, e := range edges {
		col := edgeIndex[e]
		for v := range h.edges[e].Set {
			row := vertexIndex[v]
			rows = append(rows, row)
			cols = append(cols, col)
		}
	}
	coo = COO{Rows: rows, Cols: cols}
	return
}
