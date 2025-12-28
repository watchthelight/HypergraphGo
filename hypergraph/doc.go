// Package hypergraph provides a generic implementation of hypergraphs.
//
// A hypergraph generalizes a graph by allowing edges (hyperedges) to connect
// any number of vertices, not just two. This enables modeling complex
// relationships in domains like databases, combinatorics, and constraint
// satisfaction problems.
//
// # Core Types
//
// The main type is [Hypergraph], parameterized by vertex type V:
//
//	type Hypergraph[V cmp.Ordered] struct { ... }
//
// Vertices can be any ordered type (int, string, etc.). Edges are identified
// by string IDs and contain sets of vertices.
//
// # Vertex and Edge Operations
//
// Basic operations for managing vertices and edges:
//
//   - [Hypergraph.AddVertex], [Hypergraph.RemoveVertex] - vertex management
//   - [Hypergraph.AddEdge], [Hypergraph.RemoveEdge] - edge management
//   - [Hypergraph.HasVertex], [Hypergraph.HasEdge] - existence checks
//   - [Hypergraph.Vertices], [Hypergraph.Edges] - enumeration
//   - [Hypergraph.EdgeMembers], [Hypergraph.EdgeSize] - edge queries
//   - [Hypergraph.VertexDegree] - vertex degree (number of incident edges)
//
// # Graph Algorithms
//
// The package includes several algorithms for hypergraph analysis:
//
//   - [Hypergraph.GreedyHittingSet] - approximates minimum hitting set
//   - [Hypergraph.EnumerateMinimalTransversals] - enumerates all minimal transversals
//   - [Hypergraph.GreedyColoring] - computes a vertex coloring
//   - [Hypergraph.ConnectedComponents] - finds connected components
//
// # Transformations
//
// Hypergraphs can be transformed into related structures:
//
//   - [Hypergraph.Dual] - swaps vertices and edges
//   - [Hypergraph.TwoSection] - projects to ordinary graph
//   - [Hypergraph.Primal] - synonym for TwoSection
//
// # Serialization
//
// Hypergraphs can be serialized to and from JSON:
//
//   - [Hypergraph.SaveJSON] - writes to io.Writer
//   - [LoadJSON] - reads from io.Reader
//
// # Thread Safety
//
// Hypergraph is NOT safe for concurrent use. If multiple goroutines access
// a Hypergraph concurrently, and at least one modifies it, external
// synchronization is required.
//
// # Error Handling
//
// Operations that can fail return errors:
//
//   - [ErrDuplicateEdge] - returned by AddEdge if edge ID exists
//   - [ErrCutoff] - returned by EnumerateMinimalTransversals when limits reached
//
// # Example
//
// Basic usage creating a hypergraph and computing a hitting set:
//
//	h := hypergraph.NewHypergraph[string]()
//	h.AddVertex("A")
//	h.AddVertex("B")
//	h.AddVertex("C")
//
//	if err := h.AddEdge("E1", []string{"A", "B"}); err != nil {
//	    log.Fatal(err)
//	}
//	if err := h.AddEdge("E2", []string{"B", "C"}); err != nil {
//	    log.Fatal(err)
//	}
//
//	hittingSet := h.GreedyHittingSet()
//	fmt.Println("Hitting set:", hittingSet) // e.g., [B]
package hypergraph
