package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/watchthelight/HypergraphGo/hypergraph"
)

// TestCmdDual tests the dual command.
func TestCmdDual(t *testing.T) {
	t.Run("missing_flags", func(t *testing.T) {
		err := cmdDual([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "missing required flags") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing_output_flag", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "input.json")

		err := cmdDual([]string{"-f", path})
		if err == nil {
			t.Fatal("expected error for missing -o flag")
		}
	})

	t.Run("compute_dual", func(t *testing.T) {
		dir := t.TempDir()
		inputPath := writeTestGraphFile(t, dir, "input.json")
		outputPath := filepath.Join(dir, "dual.json")

		err := cmdDual([]string{"-f", inputPath, "-o", outputPath})
		if err != nil {
			t.Fatalf("cmdDual failed: %v", err)
		}

		// Load the dual graph
		dual, err := loadGraph(outputPath)
		if err != nil {
			t.Fatalf("failed to load dual: %v", err)
		}

		// In the dual, original edges become vertices
		// Original graph has edges e1 and e2, so dual has vertices e1 and e2
		if dual.NumVertices() != 2 {
			t.Errorf("dual should have 2 vertices (original edge count), got %d", dual.NumVertices())
		}
	})

	t.Run("dual_of_dual", func(t *testing.T) {
		dir := t.TempDir()
		inputPath := writeTestGraphFile(t, dir, "input.json")
		dualPath := filepath.Join(dir, "dual.json")
		doubleDualPath := filepath.Join(dir, "double_dual.json")

		// Compute dual
		cmdDual([]string{"-f", inputPath, "-o", dualPath})
		// Compute dual of dual
		cmdDual([]string{"-f", dualPath, "-o", doubleDualPath})

		original, _ := loadGraph(inputPath)
		doubleDual, _ := loadGraph(doubleDualPath)

		// Dual of dual should have same vertex count as original
		// (structure may differ but cardinalities should match for simple graphs)
		if doubleDual.NumVertices() != original.NumVertices() {
			t.Errorf("dual of dual should have same vertex count as original: %d vs %d",
				doubleDual.NumVertices(), original.NumVertices())
		}
	})

	t.Run("missing_input_file", func(t *testing.T) {
		dir := t.TempDir()
		outputPath := filepath.Join(dir, "output.json")

		err := cmdDual([]string{"-f", "/nonexistent/file.json", "-o", outputPath})
		if err == nil {
			t.Fatal("expected error for missing input file")
		}
	})
}

// TestCmdTwoSection tests the two-section command.
func TestCmdTwoSection(t *testing.T) {
	t.Run("missing_flags", func(t *testing.T) {
		err := cmdTwoSection([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "missing required flags") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("compute_two_section", func(t *testing.T) {
		dir := t.TempDir()
		inputPath := writeTestGraphFile(t, dir, "input.json")
		outputPath := filepath.Join(dir, "two_section.json")

		err := cmdTwoSection([]string{"-f", inputPath, "-o", outputPath})
		if err != nil {
			t.Fatalf("cmdTwoSection failed: %v", err)
		}

		// Read and verify output structure
		data, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("failed to read output: %v", err)
		}

		var result graphJSON
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("failed to parse output JSON: %v", err)
		}

		// Vertices should be same as original
		if len(result.Vertices) != 3 {
			t.Errorf("two-section should have 3 vertices, got %d", len(result.Vertices))
		}

		// Edges should be pairs of vertices that share a hyperedge
		// Original: e1={a,b}, e2={b,c}
		// Two-section edges: (a,b) from e1, (b,c) from e2
		if len(result.Edges) < 2 {
			t.Errorf("two-section should have at least 2 edges, got %d", len(result.Edges))
		}
	})

	t.Run("missing_input_file", func(t *testing.T) {
		dir := t.TempDir()
		outputPath := filepath.Join(dir, "output.json")

		err := cmdTwoSection([]string{"-f", "/nonexistent/file.json", "-o", outputPath})
		if err == nil {
			t.Fatal("expected error for missing input file")
		}
	})

	t.Run("empty_graph", func(t *testing.T) {
		dir := t.TempDir()
		inputPath := filepath.Join(dir, "empty.json")
		outputPath := filepath.Join(dir, "two_section.json")

		// Create empty graph
		hg := hypergraph.NewHypergraph[string]()
		saveGraph(hg, inputPath)

		err := cmdTwoSection([]string{"-f", inputPath, "-o", outputPath})
		if err != nil {
			t.Fatalf("cmdTwoSection failed on empty graph: %v", err)
		}

		data, _ := os.ReadFile(outputPath)
		var result graphJSON
		json.Unmarshal(data, &result)

		if len(result.Vertices) != 0 {
			t.Error("empty graph should produce empty two-section")
		}
	})
}

// TestCmdLineGraph tests the line-graph command.
func TestCmdLineGraph(t *testing.T) {
	t.Run("missing_flags", func(t *testing.T) {
		err := cmdLineGraph([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "missing required flags") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("compute_line_graph", func(t *testing.T) {
		dir := t.TempDir()
		inputPath := writeTestGraphFile(t, dir, "input.json")
		outputPath := filepath.Join(dir, "line_graph.json")

		err := cmdLineGraph([]string{"-f", inputPath, "-o", outputPath})
		if err != nil {
			t.Fatalf("cmdLineGraph failed: %v", err)
		}

		data, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("failed to read output: %v", err)
		}

		var result graphJSON
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("failed to parse output JSON: %v", err)
		}

		// Vertices should be the edge IDs from original graph
		if len(result.Vertices) != 2 {
			t.Errorf("line graph should have 2 vertices (original edge count), got %d", len(result.Vertices))
		}

		// e1 and e2 share vertex b, so there should be an edge between them
		if len(result.Edges) < 1 {
			t.Errorf("line graph should have at least 1 edge, got %d", len(result.Edges))
		}
	})

	t.Run("missing_input_file", func(t *testing.T) {
		dir := t.TempDir()
		outputPath := filepath.Join(dir, "output.json")

		err := cmdLineGraph([]string{"-f", "/nonexistent/file.json", "-o", outputPath})
		if err == nil {
			t.Fatal("expected error for missing input file")
		}
	})

	t.Run("disjoint_edges", func(t *testing.T) {
		dir := t.TempDir()
		inputPath := filepath.Join(dir, "disjoint.json")
		outputPath := filepath.Join(dir, "line_graph.json")

		// Create graph with disjoint edges
		hg := hypergraph.NewHypergraph[string]()
		hg.AddVertex("a")
		hg.AddVertex("b")
		hg.AddVertex("c")
		hg.AddVertex("d")
		hg.AddEdge("e1", []string{"a", "b"})
		hg.AddEdge("e2", []string{"c", "d"}) // Disjoint from e1
		saveGraph(hg, inputPath)

		err := cmdLineGraph([]string{"-f", inputPath, "-o", outputPath})
		if err != nil {
			t.Fatalf("cmdLineGraph failed: %v", err)
		}

		data, _ := os.ReadFile(outputPath)
		var result graphJSON
		json.Unmarshal(data, &result)

		// Disjoint edges should produce isolated vertices in line graph
		if len(result.Vertices) != 2 {
			t.Errorf("expected 2 vertices, got %d", len(result.Vertices))
		}
		if len(result.Edges) != 0 {
			t.Errorf("disjoint edges should produce no edges in line graph, got %d", len(result.Edges))
		}
	})
}

// TestSaveSimpleGraphString tests the helper function for saving simple graphs.
func TestSaveSimpleGraphString(t *testing.T) {
	t.Run("basic_graph", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "simple.json")

		vertices := []string{"a", "b", "c"}
		edges := []struct{ From, To string }{
			{"a", "b"},
			{"b", "c"},
		}

		err := saveSimpleGraphString(vertices, edges, path)
		if err != nil {
			t.Fatalf("saveSimpleGraphString failed: %v", err)
		}

		data, _ := os.ReadFile(path)
		var result graphJSON
		json.Unmarshal(data, &result)

		if len(result.Vertices) != 3 {
			t.Errorf("expected 3 vertices, got %d", len(result.Vertices))
		}
		if len(result.Edges) != 2 {
			t.Errorf("expected 2 edges, got %d", len(result.Edges))
		}
	})

	t.Run("sorted_output", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "sorted.json")

		// Provide unsorted input
		vertices := []string{"c", "a", "b"}
		edges := []struct{ From, To string }{
			{"c", "a"},
			{"b", "a"},
		}

		saveSimpleGraphString(vertices, edges, path)

		data, _ := os.ReadFile(path)
		var result graphJSON
		json.Unmarshal(data, &result)

		// Vertices should be sorted
		if result.Vertices[0] != "a" || result.Vertices[1] != "b" || result.Vertices[2] != "c" {
			t.Errorf("vertices should be sorted: %v", result.Vertices)
		}

		// Edges should be sorted (each pair sorted, then pairs sorted)
		for _, edge := range result.Edges {
			if len(edge) == 2 && edge[0] > edge[1] {
				t.Errorf("edge vertices should be sorted: %v", edge)
			}
		}
	})

	t.Run("empty_graph", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.json")

		err := saveSimpleGraphString(nil, nil, path)
		if err != nil {
			t.Fatalf("saveSimpleGraphString failed for empty: %v", err)
		}

		data, _ := os.ReadFile(path)
		var result graphJSON
		json.Unmarshal(data, &result)

		if len(result.Vertices) != 0 {
			t.Error("empty graph should have no vertices")
		}
		if len(result.Edges) != 0 {
			t.Error("empty graph should have no edges")
		}
	})

	t.Run("invalid_path", func(t *testing.T) {
		err := saveSimpleGraphString(nil, nil, "/nonexistent/dir/file.json")
		if err == nil {
			t.Fatal("expected error for invalid path")
		}
	})
}

// TestTransforms_LargerGraph tests transforms on a larger graph.
func TestTransforms_LargerGraph(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "large.json")

	// Create a larger hypergraph
	hg := hypergraph.NewHypergraph[string]()
	vertices := []string{"v1", "v2", "v3", "v4", "v5"}
	for _, v := range vertices {
		hg.AddVertex(v)
	}
	hg.AddEdge("e1", []string{"v1", "v2", "v3"})
	hg.AddEdge("e2", []string{"v2", "v3", "v4"})
	hg.AddEdge("e3", []string{"v4", "v5"})
	hg.AddEdge("e4", []string{"v1", "v5"})
	saveGraph(hg, inputPath)

	t.Run("dual", func(t *testing.T) {
		outputPath := filepath.Join(dir, "large_dual.json")
		err := cmdDual([]string{"-f", inputPath, "-o", outputPath})
		if err != nil {
			t.Fatalf("cmdDual failed: %v", err)
		}

		dual, _ := loadGraph(outputPath)
		if dual.NumVertices() != 4 { // 4 original edges
			t.Errorf("dual should have 4 vertices, got %d", dual.NumVertices())
		}
	})

	t.Run("two_section", func(t *testing.T) {
		outputPath := filepath.Join(dir, "large_two_section.json")
		err := cmdTwoSection([]string{"-f", inputPath, "-o", outputPath})
		if err != nil {
			t.Fatalf("cmdTwoSection failed: %v", err)
		}

		data, _ := os.ReadFile(outputPath)
		var result graphJSON
		json.Unmarshal(data, &result)

		if len(result.Vertices) != 5 { // Same as original vertices
			t.Errorf("two-section should have 5 vertices, got %d", len(result.Vertices))
		}
		// Should have edges for all pairs within each hyperedge
		if len(result.Edges) < 4 {
			t.Errorf("two-section should have multiple edges, got %d", len(result.Edges))
		}
	})

	t.Run("line_graph", func(t *testing.T) {
		outputPath := filepath.Join(dir, "large_line_graph.json")
		err := cmdLineGraph([]string{"-f", inputPath, "-o", outputPath})
		if err != nil {
			t.Fatalf("cmdLineGraph failed: %v", err)
		}

		data, _ := os.ReadFile(outputPath)
		var result graphJSON
		json.Unmarshal(data, &result)

		if len(result.Vertices) != 4 { // 4 original edges
			t.Errorf("line graph should have 4 vertices, got %d", len(result.Vertices))
		}
	})
}

// TestTransforms_SingletonEdges tests transforms with singleton (1-element) edges.
func TestTransforms_SingletonEdges(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "singleton.json")

	hg := hypergraph.NewHypergraph[string]()
	hg.AddVertex("a")
	hg.AddVertex("b")
	hg.AddEdge("e1", []string{"a"}) // Singleton edge
	hg.AddEdge("e2", []string{"b"}) // Another singleton
	saveGraph(hg, inputPath)

	t.Run("two_section_singleton", func(t *testing.T) {
		outputPath := filepath.Join(dir, "singleton_two_section.json")
		err := cmdTwoSection([]string{"-f", inputPath, "-o", outputPath})
		if err != nil {
			t.Fatalf("cmdTwoSection failed: %v", err)
		}

		data, _ := os.ReadFile(outputPath)
		var result graphJSON
		json.Unmarshal(data, &result)

		// Singleton edges produce no 2-section edges (need pairs)
		if len(result.Edges) != 0 {
			t.Errorf("singleton edges should produce no two-section edges, got %d", len(result.Edges))
		}
	})

	t.Run("line_graph_singleton", func(t *testing.T) {
		outputPath := filepath.Join(dir, "singleton_line_graph.json")
		err := cmdLineGraph([]string{"-f", inputPath, "-o", outputPath})
		if err != nil {
			t.Fatalf("cmdLineGraph failed: %v", err)
		}

		data, _ := os.ReadFile(outputPath)
		var result graphJSON
		json.Unmarshal(data, &result)

		// Disjoint singleton edges don't share vertices
		if len(result.Edges) != 0 {
			t.Errorf("disjoint singletons should produce no line graph edges, got %d", len(result.Edges))
		}
	})
}
