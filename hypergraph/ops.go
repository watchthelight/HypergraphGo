package hypergraph

// IsEmpty checks if the hypergraph has no vertices or edges.
func (h *Hypergraph[V]) IsEmpty() bool {
	return len(h.vertices) == 0 && len(h.edges) == 0
}

// Copy returns a deep copy of the hypergraph.
func (h *Hypergraph[V]) Copy() *Hypergraph[V] {
	copy := NewHypergraph[V]()
	for v := range h.vertices {
		copy.AddVertex(v)
	}
	for id, edge := range h.edges {
		members := make([]V, 0)
		for v := range edge.Set {
			members = append(members, v)
		}
		copy.AddEdge(id, members)
	}
	return copy
}
