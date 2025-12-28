package hypergraph

import (
	"bytes"
	"strings"
	"testing"
)

// ============================================================================
// SaveJSON/LoadJSON Round-Trip Tests
// ============================================================================

func TestSaveLoadJSON_EmptyGraph(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()

	var buf bytes.Buffer
	if err := h.SaveJSON(&buf); err != nil {
		t.Fatalf("SaveJSON error: %v", err)
	}

	loaded, err := LoadJSON[string](&buf)
	if err != nil {
		t.Fatalf("LoadJSON error: %v", err)
	}

	if loaded.NumVertices() != 0 {
		t.Fatalf("Loaded vertices=%d, want 0", loaded.NumVertices())
	}
	if loaded.NumEdges() != 0 {
		t.Fatalf("Loaded edges=%d, want 0", loaded.NumEdges())
	}
}

func TestSaveLoadJSON_SingleVertex(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	h.AddVertex("A")

	var buf bytes.Buffer
	if err := h.SaveJSON(&buf); err != nil {
		t.Fatalf("SaveJSON error: %v", err)
	}

	loaded, err := LoadJSON[string](&buf)
	if err != nil {
		t.Fatalf("LoadJSON error: %v", err)
	}

	if loaded.NumVertices() != 1 {
		t.Fatalf("Loaded vertices=%d, want 1", loaded.NumVertices())
	}
	if !loaded.HasVertex("A") {
		t.Fatal("Loaded graph missing vertex A")
	}
}

func TestSaveLoadJSON_VerticesAndEdges(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C"})
	_ = h.AddEdge("E2", []string{"C", "D"})
	h.AddVertex("X") // Isolated vertex

	var buf bytes.Buffer
	if err := h.SaveJSON(&buf); err != nil {
		t.Fatalf("SaveJSON error: %v", err)
	}

	loaded, err := LoadJSON[string](&buf)
	if err != nil {
		t.Fatalf("LoadJSON error: %v", err)
	}

	if loaded.NumVertices() != 5 {
		t.Fatalf("Loaded vertices=%d, want 5", loaded.NumVertices())
	}
	if loaded.NumEdges() != 2 {
		t.Fatalf("Loaded edges=%d, want 2", loaded.NumEdges())
	}

	// Check specific vertices
	for _, v := range []string{"A", "B", "C", "D", "X"} {
		if !loaded.HasVertex(v) {
			t.Fatalf("Loaded graph missing vertex %s", v)
		}
	}

	// Check edges
	if !loaded.HasEdge("E1") || !loaded.HasEdge("E2") {
		t.Fatal("Loaded graph missing expected edges")
	}
}

func TestSaveLoadJSON_IntVertices(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[int]()
	_ = h.AddEdge("E1", []int{1, 2, 3})
	_ = h.AddEdge("E2", []int{3, 4, 5})

	var buf bytes.Buffer
	if err := h.SaveJSON(&buf); err != nil {
		t.Fatalf("SaveJSON error: %v", err)
	}

	loaded, err := LoadJSON[int](&buf)
	if err != nil {
		t.Fatalf("LoadJSON error: %v", err)
	}

	if loaded.NumVertices() != 5 {
		t.Fatalf("Loaded vertices=%d, want 5", loaded.NumVertices())
	}
	for i := 1; i <= 5; i++ {
		if !loaded.HasVertex(i) {
			t.Fatalf("Loaded graph missing vertex %d", i)
		}
	}
}

func TestSaveJSON_DeterministicOutput(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	// Add vertices and edges in different order
	_ = h.AddEdge("E2", []string{"Z", "Y", "X"})
	_ = h.AddEdge("E1", []string{"C", "B", "A"})

	var buf1, buf2 bytes.Buffer
	if err := h.SaveJSON(&buf1); err != nil {
		t.Fatalf("SaveJSON 1 error: %v", err)
	}
	if err := h.SaveJSON(&buf2); err != nil {
		t.Fatalf("SaveJSON 2 error: %v", err)
	}

	// Output should be identical
	if buf1.String() != buf2.String() {
		t.Fatalf("SaveJSON not deterministic:\n%s\nvs\n%s", buf1.String(), buf2.String())
	}

	// Vertices should be sorted
	json := buf1.String()
	if !strings.Contains(json, `"vertices":["A","B","C","X","Y","Z"]`) {
		t.Fatalf("Vertices not sorted in JSON: %s", json)
	}
}

func TestSaveLoadJSON_LargeGraph(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[int]()

	// Create 100 vertices with 50 edges
	for i := 0; i < 50; i++ {
		members := []int{i * 2, i*2 + 1, (i*2 + 2) % 100}
		_ = h.AddEdge("E"+string(rune('A'+i%26))+string(rune('0'+i/26)), members)
	}

	var buf bytes.Buffer
	if err := h.SaveJSON(&buf); err != nil {
		t.Fatalf("SaveJSON error: %v", err)
	}

	loaded, err := LoadJSON[int](&buf)
	if err != nil {
		t.Fatalf("LoadJSON error: %v", err)
	}

	if loaded.NumVertices() != h.NumVertices() {
		t.Fatalf("Loaded vertices=%d, want %d", loaded.NumVertices(), h.NumVertices())
	}
	if loaded.NumEdges() != h.NumEdges() {
		t.Fatalf("Loaded edges=%d, want %d", loaded.NumEdges(), h.NumEdges())
	}
}

func TestSaveLoadJSON_PreservesEdgeMembers(t *testing.T) {
	t.Parallel()
	h := NewHypergraph[string]()
	_ = h.AddEdge("E1", []string{"A", "B", "C", "D", "E"})

	var buf bytes.Buffer
	if err := h.SaveJSON(&buf); err != nil {
		t.Fatalf("SaveJSON error: %v", err)
	}

	loaded, err := LoadJSON[string](&buf)
	if err != nil {
		t.Fatalf("LoadJSON error: %v", err)
	}

	// Verify edge size matches
	size, ok := loaded.EdgeSize("E1")
	if !ok {
		t.Fatal("Edge E1 not found")
	}
	if size != 5 {
		t.Fatalf("Edge E1 size=%d, want 5", size)
	}
}
