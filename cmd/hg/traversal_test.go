package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/watchthelight/HypergraphGo/hypergraph"
)

// TestCmdBFS tests the bfs command.
func TestCmdBFS(t *testing.T) {
	t.Run("missing_flags", func(t *testing.T) {
		err := cmdBFS([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "missing required flags") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing_start_flag", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "test.json")

		err := cmdBFS([]string{"-f", path})
		if err == nil {
			t.Fatal("expected error for missing -start flag")
		}
	})

	t.Run("bfs_from_vertex", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "test.json")

		output := captureStdout(t, func() {
			err := cmdBFS([]string{"-f", path, "-start", "a"})
			if err != nil {
				t.Fatalf("cmdBFS failed: %v", err)
			}
		})

		// BFS from 'a' should include 'a' first
		parts := strings.Fields(output)
		if len(parts) == 0 {
			t.Fatal("BFS should produce output")
		}
		if parts[0] != "a" {
			t.Errorf("BFS should start with 'a', got: %s", parts[0])
		}

		// Should reach all vertices in connected component
		// a connects to b via e1, b connects to c via e2
		if len(parts) != 3 {
			t.Errorf("BFS should visit all 3 vertices, got %d: %v", len(parts), parts)
		}
	})

	t.Run("bfs_nonexistent_vertex", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "test.json")

		err := cmdBFS([]string{"-f", path, "-start", "nonexistent"})
		if err == nil {
			t.Fatal("expected error for nonexistent vertex")
		}
		if !strings.Contains(err.Error(), "vertex not found") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing_input_file", func(t *testing.T) {
		err := cmdBFS([]string{"-f", "/nonexistent/file.json", "-start", "a"})
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

// TestCmdDFS tests the dfs command.
func TestCmdDFS(t *testing.T) {
	t.Run("missing_flags", func(t *testing.T) {
		err := cmdDFS([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "missing required flags") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing_start_flag", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "test.json")

		err := cmdDFS([]string{"-f", path})
		if err == nil {
			t.Fatal("expected error for missing -start flag")
		}
	})

	t.Run("dfs_from_vertex", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "test.json")

		output := captureStdout(t, func() {
			err := cmdDFS([]string{"-f", path, "-start", "a"})
			if err != nil {
				t.Fatalf("cmdDFS failed: %v", err)
			}
		})

		// DFS from 'a' should include 'a' first
		parts := strings.Fields(output)
		if len(parts) == 0 {
			t.Fatal("DFS should produce output")
		}
		if parts[0] != "a" {
			t.Errorf("DFS should start with 'a', got: %s", parts[0])
		}

		// Should reach all vertices in connected component
		if len(parts) != 3 {
			t.Errorf("DFS should visit all 3 vertices, got %d: %v", len(parts), parts)
		}
	})

	t.Run("dfs_nonexistent_vertex", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "test.json")

		err := cmdDFS([]string{"-f", path, "-start", "nonexistent"})
		if err == nil {
			t.Fatal("expected error for nonexistent vertex")
		}
		if !strings.Contains(err.Error(), "vertex not found") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing_input_file", func(t *testing.T) {
		err := cmdDFS([]string{"-f", "/nonexistent/file.json", "-start", "a"})
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

// TestCmdComponents tests the components command.
func TestCmdComponents(t *testing.T) {
	t.Run("missing_file_flag", func(t *testing.T) {
		err := cmdComponents([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "missing required flag") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("single_component", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "test.json")

		output := captureStdout(t, func() {
			err := cmdComponents([]string{"-f", path})
			if err != nil {
				t.Fatalf("cmdComponents failed: %v", err)
			}
		})

		// Test graph is connected, so should have 1 component
		if !strings.Contains(output, "Component 1:") {
			t.Error("should show Component 1")
		}

		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 1 {
			t.Errorf("connected graph should have 1 component, got %d lines", len(lines))
		}

		// Component should contain all vertices
		if !strings.Contains(output, "a") || !strings.Contains(output, "b") || !strings.Contains(output, "c") {
			t.Error("component should contain all vertices")
		}
	})

	t.Run("multiple_components", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "disconnected.json")

		hg := hypergraph.NewHypergraph[string]()
		hg.AddVertex("a")
		hg.AddVertex("b")
		hg.AddVertex("c")
		hg.AddVertex("d")
		hg.AddEdge("e1", []string{"a", "b"}) // Component 1
		hg.AddEdge("e2", []string{"c", "d"}) // Component 2
		saveGraph(hg, path)

		output := captureStdout(t, func() {
			err := cmdComponents([]string{"-f", path})
			if err != nil {
				t.Fatalf("cmdComponents failed: %v", err)
			}
		})

		if !strings.Contains(output, "Component 1:") {
			t.Error("should show Component 1")
		}
		if !strings.Contains(output, "Component 2:") {
			t.Error("should show Component 2")
		}

		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 2 {
			t.Errorf("disconnected graph should have 2 components, got %d lines", len(lines))
		}
	})

	t.Run("missing_input_file", func(t *testing.T) {
		err := cmdComponents([]string{"-f", "/nonexistent/file.json"})
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("empty_graph", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.json")

		hg := hypergraph.NewHypergraph[string]()
		saveGraph(hg, path)

		output := captureStdout(t, func() {
			err := cmdComponents([]string{"-f", path})
			if err != nil {
				t.Fatalf("cmdComponents failed: %v", err)
			}
		})

		// Empty graph should produce no components
		if strings.TrimSpace(output) != "" {
			t.Errorf("empty graph should have no components, got: %s", output)
		}
	})

	t.Run("isolated_vertices", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "isolated.json")

		hg := hypergraph.NewHypergraph[string]()
		hg.AddVertex("a")
		hg.AddVertex("b")
		hg.AddVertex("c")
		// No edges - each vertex is its own component
		saveGraph(hg, path)

		output := captureStdout(t, func() {
			err := cmdComponents([]string{"-f", path})
			if err != nil {
				t.Fatalf("cmdComponents failed: %v", err)
			}
		})

		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 3 {
			t.Errorf("3 isolated vertices should have 3 components, got %d", len(lines))
		}
	})
}

// TestTraversal_LargerGraph tests traversal on a larger graph.
func TestTraversal_LargerGraph(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "large.json")

	// Create a larger connected hypergraph
	hg := hypergraph.NewHypergraph[string]()
	vertices := []string{"v1", "v2", "v3", "v4", "v5"}
	for _, v := range vertices {
		hg.AddVertex(v)
	}
	hg.AddEdge("e1", []string{"v1", "v2", "v3"})
	hg.AddEdge("e2", []string{"v3", "v4"})
	hg.AddEdge("e3", []string{"v4", "v5"})
	saveGraph(hg, path)

	t.Run("bfs_visits_all", func(t *testing.T) {
		output := captureStdout(t, func() {
			cmdBFS([]string{"-f", path, "-start", "v1"})
		})

		parts := strings.Fields(output)
		if len(parts) != 5 {
			t.Errorf("BFS should visit all 5 vertices, got %d: %v", len(parts), parts)
		}
	})

	t.Run("dfs_visits_all", func(t *testing.T) {
		output := captureStdout(t, func() {
			cmdDFS([]string{"-f", path, "-start", "v1"})
		})

		parts := strings.Fields(output)
		if len(parts) != 5 {
			t.Errorf("DFS should visit all 5 vertices, got %d: %v", len(parts), parts)
		}
	})

	t.Run("components_single", func(t *testing.T) {
		output := captureStdout(t, func() {
			cmdComponents([]string{"-f", path})
		})

		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 1 {
			t.Errorf("connected graph should have 1 component, got %d", len(lines))
		}
	})
}

// TestTraversal_ComplexComponents tests components on complex graph structure.
func TestTraversal_ComplexComponents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "complex.json")

	// Create a graph with 3 distinct components:
	// Component 1: a, b (connected via e1)
	// Component 2: c, d, e (connected via e2, e3)
	// Component 3: f (isolated)
	hg := hypergraph.NewHypergraph[string]()
	hg.AddVertex("a")
	hg.AddVertex("b")
	hg.AddVertex("c")
	hg.AddVertex("d")
	hg.AddVertex("e")
	hg.AddVertex("f")
	hg.AddEdge("e1", []string{"a", "b"})
	hg.AddEdge("e2", []string{"c", "d"})
	hg.AddEdge("e3", []string{"d", "e"})
	saveGraph(hg, path)

	t.Run("three_components", func(t *testing.T) {
		output := captureStdout(t, func() {
			cmdComponents([]string{"-f", path})
		})

		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 3 {
			t.Errorf("should have 3 components, got %d:\n%s", len(lines), output)
		}
	})

	t.Run("bfs_from_isolated", func(t *testing.T) {
		output := captureStdout(t, func() {
			cmdBFS([]string{"-f", path, "-start", "f"})
		})

		parts := strings.Fields(output)
		if len(parts) != 1 {
			t.Errorf("BFS from isolated vertex should only visit that vertex, got: %v", parts)
		}
		if parts[0] != "f" {
			t.Errorf("BFS should return 'f', got: %s", parts[0])
		}
	})

	t.Run("dfs_from_component", func(t *testing.T) {
		output := captureStdout(t, func() {
			cmdDFS([]string{"-f", path, "-start", "c"})
		})

		parts := strings.Fields(output)
		if len(parts) != 3 {
			t.Errorf("DFS from 'c' should visit c, d, e (3 vertices), got: %v", parts)
		}
	})
}

// TestTraversal_HyperedgeConnectivity tests that hyperedge connectivity is respected.
func TestTraversal_HyperedgeConnectivity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hyperedge.json")

	// Create a hyperedge with 4 vertices - all should be reachable from any one
	hg := hypergraph.NewHypergraph[string]()
	hg.AddVertex("a")
	hg.AddVertex("b")
	hg.AddVertex("c")
	hg.AddVertex("d")
	hg.AddEdge("big_edge", []string{"a", "b", "c", "d"})
	saveGraph(hg, path)

	for _, start := range []string{"a", "b", "c", "d"} {
		t.Run("bfs_from_"+start, func(t *testing.T) {
			output := captureStdout(t, func() {
				cmdBFS([]string{"-f", path, "-start", start})
			})

			parts := strings.Fields(output)
			if len(parts) != 4 {
				t.Errorf("BFS from '%s' should visit all 4 vertices via hyperedge, got: %v", start, parts)
			}
			if parts[0] != start {
				t.Errorf("BFS should start with '%s', got: %s", start, parts[0])
			}
		})

		t.Run("dfs_from_"+start, func(t *testing.T) {
			output := captureStdout(t, func() {
				cmdDFS([]string{"-f", path, "-start", start})
			})

			parts := strings.Fields(output)
			if len(parts) != 4 {
				t.Errorf("DFS from '%s' should visit all 4 vertices via hyperedge, got: %v", start, parts)
			}
		})
	}
}

// TestTraversal_StartVertexConsistency verifies traversal always starts with the given vertex.
func TestTraversal_StartVertexConsistency(t *testing.T) {
	dir := t.TempDir()
	path := writeTestGraphFile(t, dir, "test.json")

	vertices := []string{"a", "b", "c"}

	for _, v := range vertices {
		t.Run("bfs_starts_with_"+v, func(t *testing.T) {
			output := captureStdout(t, func() {
				cmdBFS([]string{"-f", path, "-start", v})
			})

			parts := strings.Fields(output)
			if len(parts) == 0 || parts[0] != v {
				t.Errorf("BFS should start with '%s', got: %s", v, output)
			}
		})

		t.Run("dfs_starts_with_"+v, func(t *testing.T) {
			output := captureStdout(t, func() {
				cmdDFS([]string{"-f", path, "-start", v})
			})

			parts := strings.Fields(output)
			if len(parts) == 0 || parts[0] != v {
				t.Errorf("DFS should start with '%s', got: %s", v, output)
			}
		})
	}
}

// TestTraversal_ChainGraph tests traversal on a chain graph (linear structure).
func TestTraversal_ChainGraph(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "chain.json")

	// a -- b -- c -- d -- e (linear chain)
	hg := hypergraph.NewHypergraph[string]()
	hg.AddVertex("a")
	hg.AddVertex("b")
	hg.AddVertex("c")
	hg.AddVertex("d")
	hg.AddVertex("e")
	hg.AddEdge("ab", []string{"a", "b"})
	hg.AddEdge("bc", []string{"b", "c"})
	hg.AddEdge("cd", []string{"c", "d"})
	hg.AddEdge("de", []string{"d", "e"})
	saveGraph(hg, path)

	t.Run("bfs_from_end", func(t *testing.T) {
		output := captureStdout(t, func() {
			cmdBFS([]string{"-f", path, "-start", "a"})
		})

		parts := strings.Fields(output)
		if len(parts) != 5 {
			t.Errorf("BFS should visit all 5 vertices, got %d", len(parts))
		}
	})

	t.Run("dfs_from_middle", func(t *testing.T) {
		output := captureStdout(t, func() {
			cmdDFS([]string{"-f", path, "-start", "c"})
		})

		parts := strings.Fields(output)
		if len(parts) != 5 {
			t.Errorf("DFS should visit all 5 vertices, got %d", len(parts))
		}
		if parts[0] != "c" {
			t.Errorf("DFS should start with 'c', got: %s", parts[0])
		}
	})
}
