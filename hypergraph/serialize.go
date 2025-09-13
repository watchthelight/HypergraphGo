package hypergraph

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// SaveJSON saves the hypergraph to JSON.
// Vertices and edge members are sorted for stable output; JSON map key order is not guaranteed.
func (h *Hypergraph[V]) SaveJSON(w io.Writer) error {
	data := map[string]interface{}{
		"vertices": h.Vertices(),
		"edges":    make(map[string][]V),
	}
	sort.Slice(data["vertices"].([]V), func(i, j int) bool {
		return fmt.Sprintf("%v", data["vertices"].([]V)[i]) < fmt.Sprintf("%v", data["vertices"].([]V)[j])
	})
	for id, edge := range h.edges {
		members := make([]V, 0)
		for v := range edge.Set {
			members = append(members, v)
		}
		sort.Slice(members, func(i, j int) bool {
			return fmt.Sprintf("%v", members[i]) < fmt.Sprintf("%v", members[j])
		})
		data["edges"].(map[string][]V)[id] = members
	}
	return json.NewEncoder(w).Encode(data)
}

// LoadJSON loads the hypergraph from JSON.
func LoadJSON[V comparable](r io.Reader) (*Hypergraph[V], error) {
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
