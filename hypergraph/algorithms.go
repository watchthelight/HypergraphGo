package hypergraph

import (
	"fmt"
	"sort"
	"time"
)

// GreedyHittingSet returns a greedy hitting set.
func (h *Hypergraph[V]) GreedyHittingSet() []V {
	hittingSet := []V{}
	remainingEdges := make(map[string]struct{})
	for e := range h.edges {
		remainingEdges[e] = struct{}{}
	}
	for len(remainingEdges) > 0 {
		// Find vertex with max degree in remaining
		var bestV V
		maxDeg := -1
		for v := range h.vertices {
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

// EnumerateMinimalTransversals enumerates minimal transversals with cutoffs.
func (h *Hypergraph[V]) EnumerateMinimalTransversals(maxSolutions int, maxTime time.Duration) ([][]V, error) {
	start := time.Now()
	var transversals [][]V
	vertices := h.Vertices()
	sort.Slice(vertices, func(i, j int) bool {
		return fmt.Sprintf("%v", vertices[i]) < fmt.Sprintf("%v", vertices[j])
	})
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

// GreedyColoring returns a greedy coloring.
func (h *Hypergraph[V]) GreedyColoring() map[V]int {
	coloring := make(map[V]int)
	vertices := h.Vertices()
	sort.Slice(vertices, func(i, j int) bool {
		return fmt.Sprintf("%v", vertices[i]) < fmt.Sprintf("%v", vertices[j])
	})
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
