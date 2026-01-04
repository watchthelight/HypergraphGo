package hypergraph

import (
	"cmp"
	"encoding/json"
	"io"
	"slices"
)

// SaveJSON saves the hypergraph to JSON.
// Vertices and edge members are sorted for stable output; JSON map key order is not guaranteed.
func (h *Hypergraph[V]) SaveJSON(w io.Writer) error {
	vertices := h.Vertices()
	slices.Sort(vertices)
	edges := make(map[string][]V)
	for id, edge := range h.edges {
		members := make([]V, 0)
		for v := range edge.Set {
			members = append(members, v)
		}
		slices.Sort(members)
		edges[id] = members
	}
	data := map[string]interface{}{
		"vertices": vertices,
		"edges":    edges,
	}
	return json.NewEncoder(w).Encode(data)
}

// LoadJSON loads the hypergraph from JSON.
func LoadJSON[V cmp.Ordered](r io.Reader) (*Hypergraph[V], error) {
	var data struct {
		Vertices []V            `json:"vertices"`
		Edges    map[string][]V `json:"edges"`
	}
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return nil, err
	}
	h := NewHypergraph[V]()
	for _, v := range data.Vertices {
		h.AddVertex(v)
	}
	for id, members := range data.Edges {
		if err := h.AddEdge(id, members); err != nil {
			return nil, err
		}
	}
	return h, nil
}
