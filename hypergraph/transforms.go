package hypergraph

import "fmt"

// Graph represents a simple undirected graph.
type Graph[V comparable] struct {
	vertices map[V]struct{}
	edges    map[string]struct{ From, To V }
}

// NewGraph creates a new graph.
func NewGraph[V comparable]() *Graph[V] {
	return &Graph[V]{
		vertices: make(map[V]struct{}),
		edges:    make(map[string]struct{ From, To V }),
	}
}

// Dual returns the dual hypergraph where vertices become edges and vice versa.
func (h *Hypergraph[V]) Dual() *Hypergraph[string] {
	dual := NewHypergraph[string]()
	for _, e := range h.Edges() {
		dual.AddVertex(e)
	}
	for v := range h.vertices {
		members := make([]string, 0)
		for e := range h.vertexToEdges[v] {
			members = append(members, e)
		}
		if len(members) > 0 {
			dual.AddEdge(fmt.Sprintf("%v", v), members)
		}
	}
	return dual
}

// TwoSection returns the 2-section graph where vertices are connected if they share an edge.
func (h *Hypergraph[V]) TwoSection() *Graph[V] {
	g := NewGraph[V]()
	for v := range h.vertices {
		g.vertices[v] = struct{}{}
	}
	for _, e := range h.edges {
		vs := make([]V, 0)
		for v := range e.Set {
			vs = append(vs, v)
		}
		for i := 0; i < len(vs); i++ {
			for j := i + 1; j < len(vs); j++ {
				from, to := vs[i], vs[j]
				if from > to {
					from, to = to, from
				}
				edgeID := fmt.Sprintf("%v-%v", from, to)
				if _, exists := g.edges[edgeID]; !exists {
					g.edges[edgeID] = struct{ From, To V }{from, to}
				}
			}
		}
	}
	return g
}

// LineGraph returns the line graph where vertices are edges, connected if they share a vertex.
func (h *Hypergraph[V]) LineGraph() *Graph[string] {
	g := NewGraph[string]()
	for _, e := range h.Edges() {
		g.vertices[e] = struct{}{}
	}
	for _, e1 := range h.Edges() {
		for _, e2 := range h.Edges() {
			if e1 >= e2 {
				continue
			}
			shared := false
			for v := range h.edges[e1].Set {
				if _, exists := h.edges[e2].Set[v]; exists {
					shared = true
					break
				}
			}
			if shared {
				from, to := e1, e2
				if from > to {
					from, to = to, from
				}
				edgeID := fmt.Sprintf("%s-%s", from, to)
				if _, exists := g.edges[edgeID]; !exists {
					g.edges[edgeID] = struct{ From, To string }{from, to}
				}
			}
		}
	}
	return g
}
