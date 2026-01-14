package hypergraph

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// FuzzLoadJSON fuzzes the JSON hypergraph loader.
// It verifies that:
// 1. The loader never panics on arbitrary JSON input
// 2. Loaded graphs can be serialized back to JSON
func FuzzLoadJSON(f *testing.F) {
	// Seed corpus with valid JSON hypergraph representations
	validSeeds := []string{
		// Empty graph
		`{"vertices":[],"edges":{}}`,

		// Simple graph
		`{"vertices":["a","b","c"],"edges":{"e1":["a","b"]}}`,

		// Single vertex, no edges
		`{"vertices":["x"],"edges":{}}`,

		// Multiple edges
		`{"vertices":["1","2","3","4"],"edges":{"e1":["1","2"],"e2":["2","3"],"e3":["3","4"]}}`,

		// Hyperedge (more than 2 vertices)
		`{"vertices":["a","b","c","d"],"edges":{"hyper":["a","b","c","d"]}}`,

		// Self-loop (edge with single vertex)
		`{"vertices":["v"],"edges":{"loop":["v"]}}`,

		// Unicode vertex names
		`{"vertices":["α","β","γ"],"edges":{"e":["α","β"]}}`,

		// Long vertex names
		`{"vertices":["verylongvertexname123456789"],"edges":{}}`,

		// Many vertices and edges
		`{"vertices":["v1","v2","v3","v4","v5"],"edges":{"e1":["v1","v2"],"e2":["v2","v3"],"e3":["v3","v4"],"e4":["v4","v5"],"e5":["v1","v5"]}}`,

		// Whitespace variations
		`{
			"vertices": ["a", "b"],
			"edges": {
				"e": ["a", "b"]
			}
		}`,

		// Minimal whitespace
		`{"vertices":["a"],"edges":{}}`,

		// Empty edge (hyperedge with no vertices)
		`{"vertices":["a"],"edges":{"empty":[]}}`,
	}

	for _, seed := range validSeeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		r := strings.NewReader(input)
		h, err := LoadJSON[string](r)

		if err == nil && h != nil {
			// Verify graph invariants
			if h.NumVertices() < 0 {
				t.Error("NumVertices should never be negative")
			}
			if h.NumEdges() < 0 {
				t.Error("NumEdges should never be negative")
			}

			// Try to save the graph back to JSON
			var buf bytes.Buffer
			if err := h.SaveJSON(&buf); err != nil {
				// SaveJSON should succeed for any valid hypergraph
				t.Errorf("SaveJSON failed on valid graph: %v", err)
			}

			// Verify all vertices exist
			for _, v := range h.Vertices() {
				if !h.HasVertex(v) {
					t.Errorf("HasVertex(%v) = false, but vertex in Vertices()", v)
				}
			}

			// Verify all edges exist
			for _, e := range h.Edges() {
				if !h.HasEdge(e) {
					t.Errorf("HasEdge(%v) = false, but edge in Edges()", e)
				}
			}
		}
	})
}

// FuzzLoadJSONMalformed specifically tests malformed JSON inputs.
func FuzzLoadJSONMalformed(f *testing.F) {
	malformed := []string{
		// Not JSON at all
		"",
		"not json",
		"{",
		"}",
		"{{}}",
		"[]",
		"null",
		"true",
		"false",
		"123",
		`"string"`,

		// Missing fields
		`{}`,
		`{"vertices":[]}`,
		`{"edges":{}}`,

		// Wrong types
		`{"vertices":"not an array","edges":{}}`,
		`{"vertices":[],"edges":[]}`,
		`{"vertices":[],"edges":{"e":"not an array"}}`,
		`{"vertices":[1,2,3],"edges":{}}`, // numbers instead of strings
		`{"vertices":null,"edges":null}`,

		// Invalid JSON values
		`{"vertices":[undefined],"edges":{}}`,
		`{"vertices":[NaN],"edges":{}}`,
		`{"vertices":[Infinity],"edges":{}}`,

		// Duplicate keys (behavior depends on decoder)
		`{"vertices":["a"],"vertices":["b"],"edges":{}}`,

		// Deeply nested
		`{"vertices":[[[[["a"]]]]],"edges":{}}`,

		// Very large numbers
		`{"vertices":["a"],"edges":{"e":["` + strings.Repeat("a", 10000) + `"]}}`,

		// Null bytes
		`{"vertices":["\u0000"],"edges":{}}`,

		// Control characters
		`{"vertices":["a\nb"],"edges":{}}`,

		// Invalid UTF-8 (raw bytes - might be interpreted as Latin-1)
		"{\"vertices\":[\"\xff\xfe\"],\"edges\":{}}",

		// Extra fields (should be ignored)
		`{"vertices":[],"edges":{},"extra":"field"}`,

		// Numeric edge IDs
		`{"vertices":["a"],"edges":{123:["a"]}}`,

		// Array as edge ID
		`{"vertices":["a"],"edges":{["e"]:["a"]}}`,

		// Edge referencing non-existent vertex
		`{"vertices":["a"],"edges":{"e":["b"]}}`,
	}

	for _, m := range malformed {
		f.Add(m)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		r := strings.NewReader(input)
		h, _ := LoadJSON[string](r)

		// If loaded successfully, verify basic operations don't panic
		if h != nil {
			_ = h.NumVertices()
			_ = h.NumEdges()
			_ = h.Vertices()
			_ = h.Edges()
			_ = h.IsEmpty()
			_ = h.Copy()
		}
	})
}

// FuzzJSONRoundTrip tests that valid hypergraphs survive JSON round-trips.
func FuzzJSONRoundTrip(f *testing.F) {
	// Seed with valid graphs via operations
	seeds := [][]byte{
		[]byte(`{"vertices":[],"edges":{}}`),
		[]byte(`{"vertices":["a","b","c"],"edges":{"e":["a","b"]}}`),
		[]byte(`{"vertices":["1","2","3"],"edges":{"e1":["1"],"e2":["2","3"]}}`),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input []byte) {
		// Try to load
		r := bytes.NewReader(input)
		h1, err := LoadJSON[string](r)
		if err != nil {
			return // Invalid input, skip
		}

		// Save to JSON
		var buf bytes.Buffer
		if err := h1.SaveJSON(&buf); err != nil {
			t.Fatalf("SaveJSON failed: %v", err)
		}

		// Load again
		h2, err := LoadJSON[string](&buf)
		if err != nil {
			t.Fatalf("Round-trip LoadJSON failed: %v", err)
		}

		// Verify structural equality
		if h1.NumVertices() != h2.NumVertices() {
			t.Errorf("NumVertices mismatch: %d vs %d", h1.NumVertices(), h2.NumVertices())
		}
		if h1.NumEdges() != h2.NumEdges() {
			t.Errorf("NumEdges mismatch: %d vs %d", h1.NumEdges(), h2.NumEdges())
		}

		// All vertices in h1 should be in h2
		for _, v := range h1.Vertices() {
			if !h2.HasVertex(v) {
				t.Errorf("Vertex %v lost in round-trip", v)
			}
		}

		// All edges in h1 should be in h2
		for _, e := range h1.Edges() {
			if !h2.HasEdge(e) {
				t.Errorf("Edge %v lost in round-trip", e)
			}
		}
	})
}

// FuzzHypergraphOperations tests random sequences of operations.
func FuzzHypergraphOperations(f *testing.F) {
	// Seeds: byte slices encoding operation sequences
	f.Add([]byte{0, 1, 2, 3, 4, 5})
	f.Add([]byte{1, 1, 1, 1, 1})
	f.Add([]byte{0, 0, 2, 2, 0, 0})

	f.Fuzz(func(t *testing.T, ops []byte) {
		h := NewHypergraph[int]()
		nextVertex := 0
		nextEdge := 0

		for _, op := range ops {
			switch op % 6 {
			case 0: // AddVertex
				h.AddVertex(nextVertex)
				nextVertex++
			case 1: // RemoveVertex (if any exist)
				vertices := h.Vertices()
				if len(vertices) > 0 {
					h.RemoveVertex(vertices[0])
				}
			case 2: // AddEdge (if we have vertices)
				vertices := h.Vertices()
				if len(vertices) >= 2 {
					edgeID := "e" + string(rune('0'+nextEdge%10))
					_ = h.AddEdge(edgeID, []int{vertices[0], vertices[1]})
					nextEdge++
				}
			case 3: // RemoveEdge (if any exist)
				edges := h.Edges()
				if len(edges) > 0 {
					h.RemoveEdge(edges[0])
				}
			case 4: // Copy
				h = h.Copy()
			case 5: // Query operations
				_ = h.NumVertices()
				_ = h.NumEdges()
				_ = h.IsEmpty()
				for _, v := range h.Vertices() {
					_ = h.VertexDegree(v)
				}
				for _, e := range h.Edges() {
					_, _ = h.EdgeSize(e)
					_ = h.EdgeMembers(e)
				}
			}
		}

		// Final invariant checks
		if h.NumVertices() < 0 {
			t.Error("NumVertices negative after operations")
		}
		if h.NumEdges() < 0 {
			t.Error("NumEdges negative after operations")
		}

		// All edges should only contain existing vertices
		for _, e := range h.Edges() {
			for _, v := range h.EdgeMembers(e) {
				if !h.HasVertex(v) {
					t.Errorf("Edge %v contains non-existent vertex %v", e, v)
				}
			}
		}
	})
}

// FuzzJSONSpecialCharacters tests JSON with special characters in strings.
func FuzzJSONSpecialCharacters(f *testing.F) {
	special := []string{
		// JSON escape sequences
		`{"vertices":["a\"b"],"edges":{}}`,
		`{"vertices":["a\\b"],"edges":{}}`,
		`{"vertices":["a\/b"],"edges":{}}`,
		`{"vertices":["a\bb"],"edges":{}}`,
		`{"vertices":["a\fb"],"edges":{}}`,
		`{"vertices":["a\nb"],"edges":{}}`,
		`{"vertices":["a\rb"],"edges":{}}`,
		`{"vertices":["a\tb"],"edges":{}}`,

		// Unicode escapes
		`{"vertices":["\u0000"],"edges":{}}`,
		`{"vertices":["\u001f"],"edges":{}}`,
		`{"vertices":["\u00e9"],"edges":{}}`, // é
		`{"vertices":["\u4e2d"],"edges":{}}`, // 中
		`{"vertices":["\ud83d\ude00"],"edges":{}}`, // 😀 (surrogate pair)

		// Mixed
		`{"vertices":["hello\nworld","tab\there"],"edges":{"e":["hello\nworld","tab\there"]}}`,
	}

	for _, s := range special {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		r := strings.NewReader(input)
		h, err := LoadJSON[string](r)

		if err == nil && h != nil {
			// Verify we can marshal/unmarshal vertex names
			for _, v := range h.Vertices() {
				// Should be valid JSON string
				data, err := json.Marshal(v)
				if err != nil {
					continue // Some strings might not be valid
				}
				var decoded string
				if err := json.Unmarshal(data, &decoded); err != nil {
					continue
				}
				if decoded != v {
					t.Errorf("JSON round-trip failed for vertex: %q -> %q", v, decoded)
				}
			}
		}
	})
}
