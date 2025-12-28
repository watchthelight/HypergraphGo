package hypergraph

import (
	"slices"
	"time"
)

// GreedyHittingSet computes a hitting set using a greedy algorithm.
//
// A hitting set is a subset of vertices that intersects every hyperedge.
// The greedy algorithm iteratively selects the vertex with maximum degree
// (number of remaining uncovered edges) until all edges are covered.
//
// Time complexity: O(|V| * |E|) where |V| is vertices and |E| is edges.
// This is a polynomial-time approximation; the optimal hitting set problem is NP-hard.
func (h *Hypergraph[V]) GreedyHittingSet() []V {
	hittingSet := []V{}
	remainingEdges := make(map[string]struct{})
	for e := range h.edges {
		remainingEdges[e] = struct{}{}
	}
	// Sort vertices for deterministic iteration order
	vertices := h.Vertices()
	slices.Sort(vertices)

	for len(remainingEdges) > 0 {
		// Find vertex with max degree in remaining
		var bestV V
		maxDeg := -1
		for _, v := range vertices {
			deg := 0
			for e := range h.vertexToEdges[v] {
				if _, exists := remainingEdges[e]; exists {
					deg++
				}
			}
			if deg > maxDeg {
				maxDeg = deg
				bestV = v
			}
		}
		if maxDeg <= 0 {
			break // no vertex found with positive degree
		}
		hittingSet = append(hittingSet, bestV)
		for e := range h.vertexToEdges[bestV] {
			delete(remainingEdges, e)
		}
	}
	return hittingSet
}

// EnumerateMinimalTransversals enumerates minimal transversals using backtracking.
//
// A transversal (or hitting set) intersects every hyperedge. A minimal transversal
// has no proper subset that is also a transversal. This function enumerates all
// minimal transversals up to the specified limits.
//
// Parameters:
//   - maxSolutions: stop after finding this many transversals
//   - maxTime: stop after this duration
//
// Returns ErrCutoff if either limit is reached before complete enumeration.
// Time complexity: exponential in worst case (NP-hard problem).
func (h *Hypergraph[V]) EnumerateMinimalTransversals(maxSolutions int, maxTime time.Duration) ([][]V, error) {
	start := time.Now()
	var transversals [][]V
	vertices := h.Vertices()
	slices.Sort(vertices)
	var backtrack func(int, []V)
	backtrack = func(index int, current []V) {
		if time.Since(start) > maxTime {
			return
		}
		if len(transversals) >= maxSolutions {
			return
		}
		// Check if current covers all edges
		covered := make(map[string]struct{})
		for _, v := range current {
			for e := range h.vertexToEdges[v] {
				covered[e] = struct{}{}
			}
		}
		if len(covered) == len(h.edges) {
			// Minimal check: no subset works
			minimal := true
			for i := range current {
				sub := make([]V, 0, len(current)-1)
				sub = append(sub, current[:i]...)
				sub = append(sub, current[i+1:]...)
				subCovered := make(map[string]struct{})
				for _, u := range sub {
					for e := range h.vertexToEdges[u] {
						subCovered[e] = struct{}{}
					}
				}
				if len(subCovered) == len(h.edges) {
					minimal = false
					break
				}
			}
			if minimal {
				transversals = append(transversals, append([]V(nil), current...))
			}
			return
		}
		for i := index; i < len(vertices); i++ {
			backtrack(i+1, append(current, vertices[i]))
		}
	}
	backtrack(0, []V{})
	if time.Since(start) > maxTime || len(transversals) >= maxSolutions {
		return transversals, ErrCutoff
	}
	return transversals, nil
}

// GreedyColoring computes a vertex coloring using a greedy algorithm.
//
// A valid coloring assigns colors (integers) to vertices such that no two
// vertices sharing a hyperedge have the same color. The greedy algorithm
// processes vertices in sorted order, assigning the smallest available color.
//
// Returns a map from vertex to color (0-indexed integers).
// Time complexity: O(|V| * |E|).
func (h *Hypergraph[V]) GreedyColoring() map[V]int {
	coloring := make(map[V]int)
	vertices := h.Vertices()
	slices.Sort(vertices)
	for _, v := range vertices {
		used := make(map[int]struct{})
		for e := range h.vertexToEdges[v] {
			for u := range h.edges[e].Set {
				if c, exists := coloring[u]; exists {
					used[c] = struct{}{}
				}
			}
		}
		color := 0
		for {
			if _, exists := used[color]; !exists {
				break
			}
			color++
		}
		coloring[v] = color
	}
	return coloring
}
