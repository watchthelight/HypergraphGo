package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/watchthelight/HypergraphGo/hypergraph"
)

// captureStdout captures stdout during execution of f and returns the output.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestCmdInfo tests the info command.
func TestCmdInfo(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing_file_flag",
			args:    []string{},
			wantErr: true,
			errMsg:  "missing required flag: -f FILE",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := cmdInfo(tc.args)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("error = %q, want containing %q", err.Error(), tc.errMsg)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCmdInfo_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := writeTestGraphFile(t, dir, "test.json")

	output := captureStdout(t, func() {
		err := cmdInfo([]string{"-f", path})
		if err != nil {
			t.Fatalf("cmdInfo failed: %v", err)
		}
	})

	if !strings.Contains(output, "Vertices: 3") {
		t.Errorf("output should contain 'Vertices: 3', got: %s", output)
	}
	if !strings.Contains(output, "Edges:    2") {
		t.Errorf("output should contain 'Edges:    2', got: %s", output)
	}
}

func TestCmdInfo_MissingFile(t *testing.T) {
	err := cmdInfo([]string{"-f", "/nonexistent/file.json"})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// TestCmdNew tests the new command.
func TestCmdNew(t *testing.T) {
	t.Run("missing_output_flag", func(t *testing.T) {
		err := cmdNew([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "missing required flag") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("creates_empty_graph", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "new.json")

		err := cmdNew([]string{"-o", path})
		if err != nil {
			t.Fatalf("cmdNew failed: %v", err)
		}

		hg, err := loadGraph(path)
		if err != nil {
			t.Fatalf("failed to load new graph: %v", err)
		}
		if hg.NumVertices() != 0 || hg.NumEdges() != 0 {
			t.Errorf("new graph should be empty")
		}
	})
}

// TestCmdValidate tests the validate command.
func TestCmdValidate(t *testing.T) {
	t.Run("missing_file_flag", func(t *testing.T) {
		err := cmdValidate([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("valid_file", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "valid.json")

		output := captureStdout(t, func() {
			err := cmdValidate([]string{"-f", path})
			if err != nil {
				t.Fatalf("cmdValidate failed: %v", err)
			}
		})

		if !strings.Contains(output, "valid") {
			t.Errorf("output should contain 'valid', got: %s", output)
		}
	})

	t.Run("invalid_file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "invalid.json")
		os.WriteFile(path, []byte("not json"), 0644)

		err := cmdValidate([]string{"-f", path})
		if err == nil {
			t.Fatal("expected error for invalid file")
		}
		if !strings.Contains(err.Error(), "invalid") {
			t.Errorf("error should contain 'invalid', got: %v", err)
		}
	})
}

// TestCmdAddVertex tests the add-vertex command.
func TestCmdAddVertex(t *testing.T) {
	t.Run("missing_flags", func(t *testing.T) {
		err := cmdAddVertex([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "missing required flags") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("add_vertex_in_place", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		err := cmdAddVertex([]string{"-f", path, "-v", "d"})
		if err != nil {
			t.Fatalf("cmdAddVertex failed: %v", err)
		}

		hg, err := loadGraph(path)
		if err != nil {
			t.Fatal(err)
		}
		if !hg.HasVertex("d") {
			t.Error("vertex 'd' should have been added")
		}
		if hg.NumVertices() != 4 {
			t.Errorf("should have 4 vertices, got %d", hg.NumVertices())
		}
	})

	t.Run("add_vertex_to_output", func(t *testing.T) {
		dir := t.TempDir()
		inputPath := writeTestGraphFile(t, dir, "input.json")
		outputPath := filepath.Join(dir, "output.json")

		err := cmdAddVertex([]string{"-f", inputPath, "-v", "newvert", "-o", outputPath})
		if err != nil {
			t.Fatalf("cmdAddVertex failed: %v", err)
		}

		// Check input unchanged
		inputHg, _ := loadGraph(inputPath)
		if inputHg.HasVertex("newvert") {
			t.Error("input should be unchanged")
		}

		// Check output has new vertex
		outputHg, _ := loadGraph(outputPath)
		if !outputHg.HasVertex("newvert") {
			t.Error("output should have new vertex")
		}
	})
}

// TestCmdRemoveVertex tests the remove-vertex command.
func TestCmdRemoveVertex(t *testing.T) {
	t.Run("missing_flags", func(t *testing.T) {
		err := cmdRemoveVertex([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("remove_existing_vertex", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		err := cmdRemoveVertex([]string{"-f", path, "-v", "a"})
		if err != nil {
			t.Fatalf("cmdRemoveVertex failed: %v", err)
		}

		hg, _ := loadGraph(path)
		if hg.HasVertex("a") {
			t.Error("vertex 'a' should have been removed")
		}
	})

	t.Run("remove_nonexistent_vertex", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		err := cmdRemoveVertex([]string{"-f", path, "-v", "nonexistent"})
		if err == nil {
			t.Fatal("expected error for nonexistent vertex")
		}
		if !strings.Contains(err.Error(), "vertex not found") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// TestCmdHasVertex tests the has-vertex command.
func TestCmdHasVertex(t *testing.T) {
	t.Run("missing_flags", func(t *testing.T) {
		err := cmdHasVertex([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("existing_vertex", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		output := captureStdout(t, func() {
			err := cmdHasVertex([]string{"-f", path, "-v", "a"})
			if err != nil {
				t.Fatalf("cmdHasVertex failed: %v", err)
			}
		})

		if !strings.Contains(output, "true") {
			t.Errorf("output should contain 'true', got: %s", output)
		}
	})

	t.Run("nonexistent_vertex", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		output := captureStdout(t, func() {
			err := cmdHasVertex([]string{"-f", path, "-v", "nonexistent"})
			if err != nil {
				t.Fatalf("cmdHasVertex failed: %v", err)
			}
		})

		if !strings.Contains(output, "false") {
			t.Errorf("output should contain 'false', got: %s", output)
		}
	})
}

// TestCmdAddEdge tests the add-edge command.
func TestCmdAddEdge(t *testing.T) {
	t.Run("missing_flags", func(t *testing.T) {
		err := cmdAddEdge([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "missing required flags") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("add_edge_in_place", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		err := cmdAddEdge([]string{"-f", path, "-id", "e3", "-m", "a,c"})
		if err != nil {
			t.Fatalf("cmdAddEdge failed: %v", err)
		}

		hg, _ := loadGraph(path)
		if !hg.HasEdge("e3") {
			t.Error("edge 'e3' should have been added")
		}
		members := hg.EdgeMembers("e3")
		if len(members) != 2 {
			t.Errorf("edge should have 2 members, got %d", len(members))
		}
	})

	t.Run("add_edge_with_whitespace", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		err := cmdAddEdge([]string{"-f", path, "-id", "e3", "-m", " a , c "})
		if err != nil {
			t.Fatalf("cmdAddEdge failed: %v", err)
		}

		hg, _ := loadGraph(path)
		members := hg.EdgeMembers("e3")
		hasA, hasC := false, false
		for _, m := range members {
			if m == "a" {
				hasA = true
			}
			if m == "c" {
				hasC = true
			}
		}
		if !hasA || !hasC {
			t.Error("whitespace should be trimmed from member list")
		}
	})
}

// TestCmdRemoveEdge tests the remove-edge command.
func TestCmdRemoveEdge(t *testing.T) {
	t.Run("missing_flags", func(t *testing.T) {
		err := cmdRemoveEdge([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("remove_existing_edge", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		err := cmdRemoveEdge([]string{"-f", path, "-id", "e1"})
		if err != nil {
			t.Fatalf("cmdRemoveEdge failed: %v", err)
		}

		hg, _ := loadGraph(path)
		if hg.HasEdge("e1") {
			t.Error("edge 'e1' should have been removed")
		}
	})

	t.Run("remove_nonexistent_edge", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		err := cmdRemoveEdge([]string{"-f", path, "-id", "nonexistent"})
		if err == nil {
			t.Fatal("expected error for nonexistent edge")
		}
		if !strings.Contains(err.Error(), "edge not found") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// TestCmdHasEdge tests the has-edge command.
func TestCmdHasEdge(t *testing.T) {
	t.Run("missing_flags", func(t *testing.T) {
		err := cmdHasEdge([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("existing_edge", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		output := captureStdout(t, func() {
			err := cmdHasEdge([]string{"-f", path, "-id", "e1"})
			if err != nil {
				t.Fatalf("cmdHasEdge failed: %v", err)
			}
		})

		if !strings.Contains(output, "true") {
			t.Errorf("output should contain 'true', got: %s", output)
		}
	})

	t.Run("nonexistent_edge", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		output := captureStdout(t, func() {
			err := cmdHasEdge([]string{"-f", path, "-id", "nonexistent"})
			if err != nil {
				t.Fatalf("cmdHasEdge failed: %v", err)
			}
		})

		if !strings.Contains(output, "false") {
			t.Errorf("output should contain 'false', got: %s", output)
		}
	})
}

// TestCmdVertices tests the vertices command.
func TestCmdVertices(t *testing.T) {
	t.Run("missing_file_flag", func(t *testing.T) {
		err := cmdVertices([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("list_vertices", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		output := captureStdout(t, func() {
			err := cmdVertices([]string{"-f", path})
			if err != nil {
				t.Fatalf("cmdVertices failed: %v", err)
			}
		})

		// Vertices should be sorted alphabetically
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 3 {
			t.Errorf("expected 3 lines, got %d: %v", len(lines), lines)
		}
		if lines[0] != "a" || lines[1] != "b" || lines[2] != "c" {
			t.Errorf("vertices not sorted correctly: %v", lines)
		}
	})
}

// TestCmdEdges tests the edges command.
func TestCmdEdges(t *testing.T) {
	t.Run("missing_file_flag", func(t *testing.T) {
		err := cmdEdges([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("list_edges", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		output := captureStdout(t, func() {
			err := cmdEdges([]string{"-f", path})
			if err != nil {
				t.Fatalf("cmdEdges failed: %v", err)
			}
		})

		// Should contain both edges with sorted member lists
		if !strings.Contains(output, "e1:") {
			t.Errorf("output should contain 'e1:', got: %s", output)
		}
		if !strings.Contains(output, "e2:") {
			t.Errorf("output should contain 'e2:', got: %s", output)
		}
	})
}

// TestCmdDegree tests the degree command.
func TestCmdDegree(t *testing.T) {
	t.Run("missing_flags", func(t *testing.T) {
		err := cmdDegree([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get_degree", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		output := captureStdout(t, func() {
			err := cmdDegree([]string{"-f", path, "-v", "b"})
			if err != nil {
				t.Fatalf("cmdDegree failed: %v", err)
			}
		})

		// vertex b is in edges e1 and e2, so degree should be 2
		if !strings.Contains(output, "2") {
			t.Errorf("degree of b should be 2, got: %s", output)
		}
	})

	t.Run("nonexistent_vertex", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		err := cmdDegree([]string{"-f", path, "-v", "nonexistent"})
		if err == nil {
			t.Fatal("expected error for nonexistent vertex")
		}
		if !strings.Contains(err.Error(), "vertex not found") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// TestCmdEdgeSize tests the edge-size command.
func TestCmdEdgeSize(t *testing.T) {
	t.Run("missing_flags", func(t *testing.T) {
		err := cmdEdgeSize([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get_edge_size", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		output := captureStdout(t, func() {
			err := cmdEdgeSize([]string{"-f", path, "-id", "e1"})
			if err != nil {
				t.Fatalf("cmdEdgeSize failed: %v", err)
			}
		})

		// edge e1 has members a and b, so size should be 2
		if !strings.Contains(output, "2") {
			t.Errorf("size of e1 should be 2, got: %s", output)
		}
	})

	t.Run("nonexistent_edge", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "graph.json")

		err := cmdEdgeSize([]string{"-f", path, "-id", "nonexistent"})
		if err == nil {
			t.Fatal("expected error for nonexistent edge")
		}
		if !strings.Contains(err.Error(), "edge not found") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// TestCmdCopy tests the copy command.
func TestCmdCopy(t *testing.T) {
	t.Run("missing_flags", func(t *testing.T) {
		err := cmdCopy([]string{})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "missing required flags") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("copy_graph", func(t *testing.T) {
		dir := t.TempDir()
		inputPath := writeTestGraphFile(t, dir, "input.json")
		outputPath := filepath.Join(dir, "output.json")

		err := cmdCopy([]string{"-f", inputPath, "-o", outputPath})
		if err != nil {
			t.Fatalf("cmdCopy failed: %v", err)
		}

		// Verify copy was made
		outputHg, err := loadGraph(outputPath)
		if err != nil {
			t.Fatalf("failed to load copied graph: %v", err)
		}

		inputHg, _ := loadGraph(inputPath)
		if outputHg.NumVertices() != inputHg.NumVertices() {
			t.Errorf("copy should have same number of vertices")
		}
		if outputHg.NumEdges() != inputHg.NumEdges() {
			t.Errorf("copy should have same number of edges")
		}
	})

	t.Run("copy_is_independent", func(t *testing.T) {
		dir := t.TempDir()
		inputPath := writeTestGraphFile(t, dir, "input.json")
		outputPath := filepath.Join(dir, "output.json")

		cmdCopy([]string{"-f", inputPath, "-o", outputPath})

		// Modify original
		cmdAddVertex([]string{"-f", inputPath, "-v", "new"})

		// Check copy is unchanged
		outputHg, _ := loadGraph(outputPath)
		if outputHg.HasVertex("new") {
			t.Error("copy should be independent of original")
		}
	})
}

// TestCmdHelp tests the help command.
func TestCmdHelp(t *testing.T) {
	t.Run("no_args_shows_usage", func(t *testing.T) {
		output := captureStdout(t, func() {
			err := cmdHelp([]string{})
			if err != nil {
				t.Fatalf("cmdHelp failed: %v", err)
			}
		})

		if !strings.Contains(output, "hg") {
			t.Errorf("should show usage, got: %s", output)
		}
	})

	t.Run("specific_command_help", func(t *testing.T) {
		commands := []string{
			"info", "new", "validate", "add-vertex", "remove-vertex",
			"has-vertex", "add-edge", "remove-edge", "has-edge",
			"vertices", "edges", "degree", "edge-size", "copy",
			"dual", "two-section", "line-graph", "bfs", "dfs",
			"components", "hitting-set", "transversals", "coloring",
			"incidence", "repl",
		}

		for _, cmd := range commands {
			t.Run(cmd, func(t *testing.T) {
				output := captureStdout(t, func() {
					err := cmdHelp([]string{cmd})
					if err != nil {
						t.Fatalf("cmdHelp %s failed: %v", cmd, err)
					}
				})

				if !strings.Contains(output, cmd) {
					t.Errorf("help for %s should contain command name", cmd)
				}
			})
		}
	})

	t.Run("unknown_command", func(t *testing.T) {
		err := cmdHelp([]string{"nonexistent"})
		if err == nil {
			t.Fatal("expected error for unknown command")
		}
		if !strings.Contains(err.Error(), "unknown command") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

// Additional edge case tests

func TestCommands_MissingInputFile(t *testing.T) {
	commands := []struct {
		name string
		fn   func([]string) error
		args []string
	}{
		{"info", cmdInfo, []string{"-f", "/nonexistent/file.json"}},
		{"validate", cmdValidate, []string{"-f", "/nonexistent/file.json"}},
		{"add-vertex", cmdAddVertex, []string{"-f", "/nonexistent/file.json", "-v", "x"}},
		{"remove-vertex", cmdRemoveVertex, []string{"-f", "/nonexistent/file.json", "-v", "x"}},
		{"has-vertex", cmdHasVertex, []string{"-f", "/nonexistent/file.json", "-v", "x"}},
		{"add-edge", cmdAddEdge, []string{"-f", "/nonexistent/file.json", "-id", "e", "-m", "a,b"}},
		{"remove-edge", cmdRemoveEdge, []string{"-f", "/nonexistent/file.json", "-id", "e"}},
		{"has-edge", cmdHasEdge, []string{"-f", "/nonexistent/file.json", "-id", "e"}},
		{"vertices", cmdVertices, []string{"-f", "/nonexistent/file.json"}},
		{"edges", cmdEdges, []string{"-f", "/nonexistent/file.json"}},
		{"degree", cmdDegree, []string{"-f", "/nonexistent/file.json", "-v", "x"}},
		{"edge-size", cmdEdgeSize, []string{"-f", "/nonexistent/file.json", "-id", "e"}},
		{"copy", cmdCopy, []string{"-f", "/nonexistent/file.json", "-o", "/tmp/out.json"}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn(tc.args)
			if err == nil {
				t.Errorf("%s should fail for missing input file", tc.name)
			}
		})
	}
}

func TestCommands_EmptyGraph(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")

	// Create empty graph
	hg := hypergraph.NewHypergraph[string]()
	saveGraph(hg, path)

	t.Run("info_empty", func(t *testing.T) {
		output := captureStdout(t, func() {
			cmdInfo([]string{"-f", path})
		})
		if !strings.Contains(output, "Vertices: 0") {
			t.Errorf("should show 0 vertices, got: %s", output)
		}
		if !strings.Contains(output, "Empty:    true") {
			t.Errorf("should show Empty: true, got: %s", output)
		}
	})

	t.Run("vertices_empty", func(t *testing.T) {
		output := captureStdout(t, func() {
			cmdVertices([]string{"-f", path})
		})
		if strings.TrimSpace(output) != "" {
			t.Errorf("should have no output for empty graph, got: %s", output)
		}
	})

	t.Run("edges_empty", func(t *testing.T) {
		output := captureStdout(t, func() {
			cmdEdges([]string{"-f", path})
		})
		if strings.TrimSpace(output) != "" {
			t.Errorf("should have no output for empty graph, got: %s", output)
		}
	})
}
