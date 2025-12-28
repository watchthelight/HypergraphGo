package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/watchthelight/HypergraphGo/hypergraph"
)

// createTestGraph creates a simple test hypergraph with vertices and edges.
func createTestGraph(t *testing.T) *hypergraph.Hypergraph[string] {
	t.Helper()
	hg := hypergraph.NewHypergraph[string]()
	hg.AddVertex("a")
	hg.AddVertex("b")
	hg.AddVertex("c")
	if err := hg.AddEdge("e1", []string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	if err := hg.AddEdge("e2", []string{"b", "c"}); err != nil {
		t.Fatal(err)
	}
	return hg
}

// writeTestGraphFile writes a hypergraph JSON file for testing.
func writeTestGraphFile(t *testing.T, dir string, name string) string {
	t.Helper()
	hg := createTestGraph(t)
	path := filepath.Join(dir, name)
	if err := saveGraph(hg, path); err != nil {
		t.Fatalf("failed to save test graph: %v", err)
	}
	return path
}

func TestLoadGraph_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := writeTestGraphFile(t, dir, "test.json")

	hg, err := loadGraph(path)
	if err != nil {
		t.Fatalf("loadGraph failed: %v", err)
	}

	if hg.NumVertices() != 3 {
		t.Errorf("NumVertices = %d, want 3", hg.NumVertices())
	}
	if hg.NumEdges() != 2 {
		t.Errorf("NumEdges = %d, want 2", hg.NumEdges())
	}
}

func TestLoadGraph_MissingFile(t *testing.T) {
	_, err := loadGraph("/nonexistent/path/graph.json")
	if err == nil {
		t.Fatal("loadGraph should fail for missing file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected os.IsNotExist error, got: %v", err)
	}
}

func TestLoadGraph_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")

	if err := os.WriteFile(path, []byte("not valid json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := loadGraph(path)
	if err == nil {
		t.Fatal("loadGraph should fail for invalid JSON")
	}
}

func TestLoadGraph_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")

	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := loadGraph(path)
	if err == nil {
		t.Fatal("loadGraph should fail for empty file")
	}
}

func TestSaveGraph_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.json")

	hg := createTestGraph(t)
	if err := saveGraph(hg, path); err != nil {
		t.Fatalf("saveGraph failed: %v", err)
	}

	// Verify file was created and can be loaded
	loaded, err := loadGraph(path)
	if err != nil {
		t.Fatalf("failed to load saved graph: %v", err)
	}

	if loaded.NumVertices() != hg.NumVertices() {
		t.Errorf("loaded vertices = %d, want %d", loaded.NumVertices(), hg.NumVertices())
	}
	if loaded.NumEdges() != hg.NumEdges() {
		t.Errorf("loaded edges = %d, want %d", loaded.NumEdges(), hg.NumEdges())
	}
}

func TestSaveGraph_Overwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.json")

	// Write initial graph
	hg1 := hypergraph.NewHypergraph[string]()
	hg1.AddVertex("x")
	if err := saveGraph(hg1, path); err != nil {
		t.Fatal(err)
	}

	// Overwrite with different graph
	hg2 := createTestGraph(t)
	if err := saveGraph(hg2, path); err != nil {
		t.Fatalf("saveGraph overwrite failed: %v", err)
	}

	// Verify overwritten content
	loaded, err := loadGraph(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.NumVertices() != 3 {
		t.Errorf("overwritten graph should have 3 vertices, got %d", loaded.NumVertices())
	}
}

func TestSaveGraph_InvalidPath(t *testing.T) {
	hg := hypergraph.NewHypergraph[string]()
	err := saveGraph(hg, "/nonexistent/dir/file.json")
	if err == nil {
		t.Fatal("saveGraph should fail for invalid path")
	}
}

func TestLoadSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()

	// Create a more complex graph
	hg := hypergraph.NewHypergraph[string]()
	hg.AddVertex("v1")
	hg.AddVertex("v2")
	hg.AddVertex("v3")
	hg.AddVertex("v4")
	if err := hg.AddEdge("edge1", []string{"v1", "v2", "v3"}); err != nil {
		t.Fatal(err)
	}
	if err := hg.AddEdge("edge2", []string{"v2", "v4"}); err != nil {
		t.Fatal(err)
	}
	if err := hg.AddEdge("edge3", []string{"v1", "v4"}); err != nil {
		t.Fatal(err)
	}

	// Save and reload
	path := filepath.Join(dir, "roundtrip.json")
	if err := saveGraph(hg, path); err != nil {
		t.Fatal(err)
	}

	loaded, err := loadGraph(path)
	if err != nil {
		t.Fatal(err)
	}

	// Verify all structure preserved
	if loaded.NumVertices() != 4 {
		t.Errorf("NumVertices = %d, want 4", loaded.NumVertices())
	}
	if loaded.NumEdges() != 3 {
		t.Errorf("NumEdges = %d, want 3", loaded.NumEdges())
	}

	// Verify specific vertices and edges
	for _, v := range []string{"v1", "v2", "v3", "v4"} {
		if !loaded.HasVertex(v) {
			t.Errorf("missing vertex %s", v)
		}
	}
	for _, e := range []string{"edge1", "edge2", "edge3"} {
		if !loaded.HasEdge(e) {
			t.Errorf("missing edge %s", e)
		}
	}

	// Verify edge members
	e1Members := loaded.EdgeMembers("edge1")
	if len(e1Members) != 3 {
		t.Errorf("edge1 should have 3 members, got %d", len(e1Members))
	}
}

func TestLoadGraph_MalformedJSON(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"missing_brace", `{"vertices": ["a", "b"]`},
		{"extra_comma", `{"vertices": ["a",], "edges": {}}`},
		{"wrong_type_vertices", `{"vertices": "not_an_array", "edges": {}}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "malformed.json")
			if err := os.WriteFile(path, []byte(tc.content), 0644); err != nil {
				t.Fatal(err)
			}

			_, err := loadGraph(path)
			if err == nil {
				t.Errorf("loadGraph should fail for %s", tc.name)
			}
		})
	}
}

func TestSaveGraph_EmptyGraph(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")

	hg := hypergraph.NewHypergraph[string]()
	if err := saveGraph(hg, path); err != nil {
		t.Fatalf("saveGraph failed for empty graph: %v", err)
	}

	loaded, err := loadGraph(path)
	if err != nil {
		t.Fatalf("loadGraph failed for empty graph: %v", err)
	}

	if loaded.NumVertices() != 0 {
		t.Errorf("empty graph should have 0 vertices")
	}
	if loaded.NumEdges() != 0 {
		t.Errorf("empty graph should have 0 edges")
	}
}
