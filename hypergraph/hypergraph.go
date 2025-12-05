// Package hypergraph provides a generic implementation of hypergraphs.
// A hypergraph is a generalization of a graph where edges can connect any number of vertices.
//
// Concurrency: Hypergraph is NOT safe for concurrent use.
// Callers must synchronize access when using from multiple goroutines.
package hypergraph

import (
	"cmp"
	"fmt"
)

// Hypergraph represents a hypergraph with generic vertex type V.
// V must satisfy cmp.Ordered to enable efficient sorting operations.
type Hypergraph[V cmp.Ordered] struct {
	vertices      map[V]struct{}
	edges         map[string]Edge[V]
	vertexToEdges map[V]map[string]struct{}
}

// Edge represents a hyperedge with an ID and a set of vertices.
type Edge[V cmp.Ordered] struct {
	ID  string
	Set map[V]struct{}
}

// NewHypergraph creates a new empty hypergraph.
func NewHypergraph[V cmp.Ordered]() *Hypergraph[V] {
	return &Hypergraph[V]{
		vertices:      make(map[V]struct{}),
		edges:         make(map[string]Edge[V]),
		vertexToEdges: make(map[V]map[string]struct{}),
	}
}

// AddVertex adds a vertex to the hypergraph.
func (h *Hypergraph[V]) AddVertex(v V) {
	if _, exists := h.vertices[v]; !exists {
		h.vertices[v] = struct{}{}
		h.vertexToEdges[v] = make(map[string]struct{})
	}
}

// RemoveVertex removes a vertex and all incident edges.
func (h *Hypergraph[V]) RemoveVertex(v V) {
	if _, exists := h.vertices[v]; !exists {
		return
	}
	delete(h.vertices, v)
	for edgeID := range h.vertexToEdges[v] {
		delete(h.edges[edgeID].Set, v)
		if len(h.edges[edgeID].Set) == 0 {
			delete(h.edges, edgeID)
		}
	}
	delete(h.vertexToEdges, v)
}

// AddEdge adds a hyperedge with the given ID and members.
func (h *Hypergraph[V]) AddEdge(id string, members []V) error {
	if _, exists := h.edges[id]; exists {
		return ErrDuplicateEdge
	}
	if len(members) == 0 {
		return fmt.Errorf("edge cannot be empty")
	}
	edge := Edge[V]{ID: id, Set: make(map[V]struct{})}
	for _, v := range members {
		if _, exists := h.vertices[v]; !exists {
			h.AddVertex(v)
		}
		edge.Set[v] = struct{}{}
		h.vertexToEdges[v][id] = struct{}{}
	}
	h.edges[id] = edge
	return nil
}

// RemoveEdge removes a hyperedge.
func (h *Hypergraph[V]) RemoveEdge(id string) {
	if edge, exists := h.edges[id]; exists {
		for v := range edge.Set {
			delete(h.vertexToEdges[v], id)
		}
		delete(h.edges, id)
	}
}

// NumVertices returns the number of vertices.
func (h *Hypergraph[V]) NumVertices() int {
	return len(h.vertices)
}

// NumEdges returns the number of edges.
func (h *Hypergraph[V]) NumEdges() int {
	return len(h.edges)
}

// Vertices returns a slice of all vertices.
func (h *Hypergraph[V]) Vertices() []V {
	vs := make([]V, 0, len(h.vertices))
	for v := range h.vertices {
		vs = append(vs, v)
	}
	return vs
}

// Edges returns a slice of all edge IDs.
func (h *Hypergraph[V]) Edges() []string {
	es := make([]string, 0, len(h.edges))
	for id := range h.edges {
		es = append(es, id)
	}
	return es
}

// HasVertex checks if a vertex exists.
func (h *Hypergraph[V]) HasVertex(v V) bool {
	_, exists := h.vertices[v]
	return exists
}

// HasEdge checks if an edge exists.
func (h *Hypergraph[V]) HasEdge(id string) bool {
	_, exists := h.edges[id]
	return exists
}

// VertexDegree returns the degree of a vertex.
func (h *Hypergraph[V]) VertexDegree(v V) int {
	if edges, exists := h.vertexToEdges[v]; exists {
		return len(edges)
	}
	return 0
}

// EdgeSize returns the size of an edge.
func (h *Hypergraph[V]) EdgeSize(id string) (int, bool) {
	if edge, exists := h.edges[id]; exists {
		return len(edge.Set), true
	}
	return 0, false
}
