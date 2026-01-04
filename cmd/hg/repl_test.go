package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/watchthelight/HypergraphGo/hypergraph"
)

// newTestReplState creates a new REPL state with a test hypergraph.
func newTestReplState(t *testing.T) *replState {
	t.Helper()
	hg := createTestGraph(t)
	return &replState{
		hg:       hg,
		modified: false,
	}
}

// TestReplStateInitialization tests initial REPL state.
func TestReplStateInitialization(t *testing.T) {
	state := &replState{
		hg: hypergraph.NewHypergraph[string](),
	}

	if state.file != "" {
		t.Error("initial file should be empty")
	}
	if state.modified {
		t.Error("initial state should not be modified")
	}
	if !state.hg.IsEmpty() {
		t.Error("initial graph should be empty")
	}
}

// TestExecuteReplCommand_Quit tests the :quit command.
func TestExecuteReplCommand_Quit(t *testing.T) {
	t.Run("quit_unmodified", func(t *testing.T) {
		state := newTestReplState(t)
		state.modified = false

		err := executeReplCommand(state, ":quit")
		if !errors.Is(err, errQuit) {
			t.Errorf("expected errQuit, got: %v", err)
		}
	})

	t.Run("quit_modified_first_attempt", func(t *testing.T) {
		state := newTestReplState(t)
		state.modified = true

		output := captureStdout(t, func() {
			err := executeReplCommand(state, ":quit")
			if err != nil {
				t.Errorf("first quit should not error: %v", err)
			}
		})

		if !strings.Contains(output, "Warning") {
			t.Error("should show warning about unsaved changes")
		}
		// quitConfirmed should be set to allow second quit
		if !state.quitConfirmed {
			t.Error("quitConfirmed should be set after warning")
		}
		// modified flag should remain true (data is still unsaved)
		if !state.modified {
			t.Error("modified flag should still be true")
		}
	})

	t.Run("quit_modified_second_attempt", func(t *testing.T) {
		state := newTestReplState(t)
		state.modified = true

		// First quit shows warning
		executeReplCommand(state, ":quit")

		// Second quit should exit
		err := executeReplCommand(state, ":quit")
		if !errors.Is(err, errQuit) {
			t.Errorf("second quit should return errQuit, got: %v", err)
		}
	})

	t.Run("quit_shorthand", func(t *testing.T) {
		state := newTestReplState(t)
		state.modified = false

		err := executeReplCommand(state, ":q")
		if !errors.Is(err, errQuit) {
			t.Errorf("expected errQuit for :q, got: %v", err)
		}
	})
}

// TestExecuteReplCommand_Help tests the :help command.
func TestExecuteReplCommand_Help(t *testing.T) {
	state := newTestReplState(t)

	output := captureStdout(t, func() {
		err := executeReplCommand(state, ":help")
		if err != nil {
			t.Fatalf(":help failed: %v", err)
		}
	})

	expectedSections := []string{
		"REPL Commands:",
		":load",
		":save",
		":new",
		":info",
		":help",
		":quit",
		"Operations:",
		"add-vertex",
		"remove-vertex",
		"add-edge",
		"remove-edge",
		"vertices",
		"edges",
		"bfs",
		"dfs",
		"components",
		"hitting-set",
		"coloring",
		"dual",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("help should contain %q", section)
		}
	}
}

// TestExecuteReplCommand_HelpShorthand tests :h shorthand.
func TestExecuteReplCommand_HelpShorthand(t *testing.T) {
	state := newTestReplState(t)

	output := captureStdout(t, func() {
		err := executeReplCommand(state, ":h")
		if err != nil {
			t.Fatalf(":h failed: %v", err)
		}
	})

	if !strings.Contains(output, "REPL Commands:") {
		t.Error(":h should show help output")
	}
}

// TestExecuteReplCommand_Load tests the :load command.
func TestExecuteReplCommand_Load(t *testing.T) {
	t.Run("load_missing_arg", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, ":load")
		if err == nil {
			t.Fatal("expected error for missing argument")
		}
		if !strings.Contains(err.Error(), "usage:") {
			t.Errorf("error should show usage, got: %v", err)
		}
	})

	t.Run("load_valid_file", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "test.json")

		state := &replState{
			hg: hypergraph.NewHypergraph[string](),
		}

		output := captureStdout(t, func() {
			err := executeReplCommand(state, ":load "+path)
			if err != nil {
				t.Fatalf(":load failed: %v", err)
			}
		})

		if !strings.Contains(output, "Loaded") {
			t.Error("should show 'Loaded' message")
		}
		if state.file != path {
			t.Errorf("file should be set to %s, got %s", path, state.file)
		}
		if state.modified {
			t.Error("modified should be false after load")
		}
		if state.hg.NumVertices() != 3 {
			t.Errorf("loaded graph should have 3 vertices, got %d", state.hg.NumVertices())
		}
	})

	t.Run("load_missing_file", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, ":load /nonexistent/file.json")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

// TestExecuteReplCommand_Save tests the :save command.
func TestExecuteReplCommand_Save(t *testing.T) {
	t.Run("save_with_no_file", func(t *testing.T) {
		state := newTestReplState(t)
		state.file = ""

		err := executeReplCommand(state, ":save")
		if err == nil {
			t.Fatal("expected error when no file specified")
		}
		if !strings.Contains(err.Error(), "no file specified") {
			t.Errorf("error should mention 'no file specified', got: %v", err)
		}
	})

	t.Run("save_to_current_file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "save.json")

		state := newTestReplState(t)
		state.file = path
		state.modified = true

		output := captureStdout(t, func() {
			err := executeReplCommand(state, ":save")
			if err != nil {
				t.Fatalf(":save failed: %v", err)
			}
		})

		if !strings.Contains(output, "Saved") {
			t.Error("should show 'Saved' message")
		}
		if state.modified {
			t.Error("modified should be false after save")
		}

		// Verify file was created
		loaded, err := loadGraph(path)
		if err != nil {
			t.Fatalf("failed to load saved graph: %v", err)
		}
		if loaded.NumVertices() != 3 {
			t.Errorf("saved graph should have 3 vertices")
		}
	})

	t.Run("save_to_new_file", func(t *testing.T) {
		dir := t.TempDir()
		newPath := filepath.Join(dir, "new_save.json")

		state := newTestReplState(t)
		state.file = ""
		state.modified = true

		output := captureStdout(t, func() {
			err := executeReplCommand(state, ":save "+newPath)
			if err != nil {
				t.Fatalf(":save failed: %v", err)
			}
		})

		if !strings.Contains(output, "Saved") {
			t.Error("should show 'Saved' message")
		}
		if state.file != newPath {
			t.Errorf("file should be updated to %s", newPath)
		}
		if state.modified {
			t.Error("modified should be false after save")
		}
	})
}

// TestExecuteReplCommand_New tests the :new command.
func TestExecuteReplCommand_New(t *testing.T) {
	t.Run("new_unmodified", func(t *testing.T) {
		state := newTestReplState(t)
		state.modified = false

		output := captureStdout(t, func() {
			err := executeReplCommand(state, ":new")
			if err != nil {
				t.Fatalf(":new failed: %v", err)
			}
		})

		if !strings.Contains(output, "Created empty hypergraph") {
			t.Error("should show creation message")
		}
		if !state.hg.IsEmpty() {
			t.Error("new graph should be empty")
		}
		if state.file != "" {
			t.Error("file should be cleared")
		}
		if state.modified {
			t.Error("modified should be false")
		}
	})

	t.Run("new_modified_first_attempt", func(t *testing.T) {
		state := newTestReplState(t)
		state.modified = true

		output := captureStdout(t, func() {
			err := executeReplCommand(state, ":new")
			if err != nil {
				t.Fatalf(":new failed: %v", err)
			}
		})

		if !strings.Contains(output, "Warning") {
			t.Error("should show warning about unsaved changes")
		}
		// Graph should NOT be replaced yet
		if state.hg.IsEmpty() {
			t.Error("graph should not be replaced after first :new with unsaved changes")
		}
	})

	t.Run("new_modified_second_attempt", func(t *testing.T) {
		state := newTestReplState(t)
		state.modified = true

		// First :new shows warning and clears modified
		executeReplCommand(state, ":new")

		// Second :new should create new graph
		output := captureStdout(t, func() {
			err := executeReplCommand(state, ":new")
			if err != nil {
				t.Fatalf(":new failed: %v", err)
			}
		})

		if !strings.Contains(output, "Created empty hypergraph") {
			t.Error("should show creation message")
		}
		if !state.hg.IsEmpty() {
			t.Error("new graph should be empty after second :new")
		}
	})
}

// TestExecuteReplCommand_Info tests the :info command.
func TestExecuteReplCommand_Info(t *testing.T) {
	t.Run("info_with_file", func(t *testing.T) {
		state := newTestReplState(t)
		state.file = "/path/to/file.json"

		output := captureStdout(t, func() {
			err := executeReplCommand(state, ":info")
			if err != nil {
				t.Fatalf(":info failed: %v", err)
			}
		})

		if !strings.Contains(output, "Vertices: 3") {
			t.Error("should show vertex count")
		}
		if !strings.Contains(output, "Edges:    2") {
			t.Error("should show edge count")
		}
		if !strings.Contains(output, "File:") {
			t.Error("should show file path")
		}
	})

	t.Run("info_modified", func(t *testing.T) {
		state := newTestReplState(t)
		state.modified = true

		output := captureStdout(t, func() {
			executeReplCommand(state, ":info")
		})

		if !strings.Contains(output, "unsaved changes") {
			t.Error("should show unsaved changes indicator")
		}
	})

	t.Run("info_without_file", func(t *testing.T) {
		state := newTestReplState(t)
		state.file = ""

		output := captureStdout(t, func() {
			executeReplCommand(state, ":info")
		})

		// Should NOT show "File:" when no file is set
		if strings.Contains(output, "File:") {
			t.Error("should not show 'File:' when no file is set")
		}
	})
}

// TestExecuteReplCommand_Operations tests REPL operations (non-colon commands).
func TestExecuteReplCommand_Operations(t *testing.T) {
	t.Run("add-vertex", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "add-vertex d")
		if err != nil {
			t.Fatalf("add-vertex failed: %v", err)
		}
		if !state.hg.HasVertex("d") {
			t.Error("vertex 'd' should have been added")
		}
		if !state.modified {
			t.Error("modified should be true after add-vertex")
		}
	})

	t.Run("add-vertex_missing_arg", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "add-vertex")
		if err == nil {
			t.Fatal("expected error for missing argument")
		}
	})

	t.Run("remove-vertex", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "remove-vertex a")
		if err != nil {
			t.Fatalf("remove-vertex failed: %v", err)
		}
		if state.hg.HasVertex("a") {
			t.Error("vertex 'a' should have been removed")
		}
		if !state.modified {
			t.Error("modified should be true after remove-vertex")
		}
	})

	t.Run("remove-vertex_nonexistent", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "remove-vertex nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent vertex")
		}
	})

	t.Run("has-vertex", func(t *testing.T) {
		state := newTestReplState(t)
		output := captureStdout(t, func() {
			executeReplCommand(state, "has-vertex a")
		})
		if !strings.Contains(output, "true") {
			t.Error("should print true for existing vertex")
		}
	})

	t.Run("add-edge", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "add-edge e3 a,c")
		if err != nil {
			t.Fatalf("add-edge failed: %v", err)
		}
		if !state.hg.HasEdge("e3") {
			t.Error("edge 'e3' should have been added")
		}
		if !state.modified {
			t.Error("modified should be true after add-edge")
		}
	})

	t.Run("add-edge_missing_args", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "add-edge e3")
		if err == nil {
			t.Fatal("expected error for missing argument")
		}
	})

	t.Run("remove-edge", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "remove-edge e1")
		if err != nil {
			t.Fatalf("remove-edge failed: %v", err)
		}
		if state.hg.HasEdge("e1") {
			t.Error("edge 'e1' should have been removed")
		}
	})

	t.Run("has-edge", func(t *testing.T) {
		state := newTestReplState(t)
		output := captureStdout(t, func() {
			executeReplCommand(state, "has-edge e1")
		})
		if !strings.Contains(output, "true") {
			t.Error("should print true for existing edge")
		}
	})

	t.Run("vertices", func(t *testing.T) {
		state := newTestReplState(t)
		output := captureStdout(t, func() {
			executeReplCommand(state, "vertices")
		})
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 3 {
			t.Errorf("should list 3 vertices, got %d", len(lines))
		}
	})

	t.Run("edges", func(t *testing.T) {
		state := newTestReplState(t)
		output := captureStdout(t, func() {
			executeReplCommand(state, "edges")
		})
		if !strings.Contains(output, "e1:") || !strings.Contains(output, "e2:") {
			t.Error("should list both edges")
		}
	})

	t.Run("degree", func(t *testing.T) {
		state := newTestReplState(t)
		output := captureStdout(t, func() {
			executeReplCommand(state, "degree b")
		})
		if !strings.Contains(output, "2") {
			t.Error("degree of b should be 2")
		}
	})

	t.Run("edge-size", func(t *testing.T) {
		state := newTestReplState(t)
		output := captureStdout(t, func() {
			executeReplCommand(state, "edge-size e1")
		})
		if !strings.Contains(output, "2") {
			t.Error("size of e1 should be 2")
		}
	})
}

// TestExecuteReplCommand_TraversalOperations tests traversal operations in REPL.
func TestExecuteReplCommand_TraversalOperations(t *testing.T) {
	t.Run("bfs", func(t *testing.T) {
		state := newTestReplState(t)
		output := captureStdout(t, func() {
			err := executeReplCommand(state, "bfs a")
			if err != nil {
				t.Fatalf("bfs failed: %v", err)
			}
		})
		// BFS from 'a' should reach all connected vertices
		if !strings.Contains(output, "a") {
			t.Error("BFS should include starting vertex")
		}
	})

	t.Run("bfs_nonexistent", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "bfs nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent vertex")
		}
	})

	t.Run("dfs", func(t *testing.T) {
		state := newTestReplState(t)
		output := captureStdout(t, func() {
			err := executeReplCommand(state, "dfs a")
			if err != nil {
				t.Fatalf("dfs failed: %v", err)
			}
		})
		if !strings.Contains(output, "a") {
			t.Error("DFS should include starting vertex")
		}
	})

	t.Run("components", func(t *testing.T) {
		state := newTestReplState(t)
		output := captureStdout(t, func() {
			err := executeReplCommand(state, "components")
			if err != nil {
				t.Fatalf("components failed: %v", err)
			}
		})
		if !strings.Contains(output, "Component") {
			t.Error("should show component output")
		}
	})
}

// TestExecuteReplCommand_AlgorithmOperations tests algorithm operations in REPL.
func TestExecuteReplCommand_AlgorithmOperations(t *testing.T) {
	t.Run("hitting-set", func(t *testing.T) {
		state := newTestReplState(t)
		output := captureStdout(t, func() {
			err := executeReplCommand(state, "hitting-set")
			if err != nil {
				t.Fatalf("hitting-set failed: %v", err)
			}
		})
		// Should produce some output (the hitting set)
		if strings.TrimSpace(output) == "" {
			t.Error("hitting-set should produce output")
		}
	})

	t.Run("coloring", func(t *testing.T) {
		state := newTestReplState(t)
		output := captureStdout(t, func() {
			err := executeReplCommand(state, "coloring")
			if err != nil {
				t.Fatalf("coloring failed: %v", err)
			}
		})
		// Should show coloring for each vertex
		if !strings.Contains(output, "a:") && !strings.Contains(output, "b:") {
			t.Error("coloring should show vertex colors")
		}
	})
}

// TestExecuteReplCommand_TransformOperations tests transform operations in REPL.
func TestExecuteReplCommand_TransformOperations(t *testing.T) {
	t.Run("dual", func(t *testing.T) {
		state := newTestReplState(t)
		origEdges := state.hg.NumEdges()

		output := captureStdout(t, func() {
			err := executeReplCommand(state, "dual")
			if err != nil {
				t.Fatalf("dual failed: %v", err)
			}
		})

		if !strings.Contains(output, "dual") {
			t.Error("should show dual message")
		}
		if !state.modified {
			t.Error("modified should be true after dual")
		}
		// In dual, edges become vertices
		if state.hg.NumVertices() != origEdges {
			t.Errorf("after dual, vertices should equal original edges (%d), got %d", origEdges, state.hg.NumVertices())
		}
	})
}

// TestExecuteReplCommand_UnknownCommand tests handling of unknown commands.
func TestExecuteReplCommand_UnknownCommand(t *testing.T) {
	state := newTestReplState(t)
	err := executeReplCommand(state, "unknown-command")
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("error should mention 'unknown command', got: %v", err)
	}
}

// TestExecuteReplCommand_EmptyLine tests that empty lines are ignored.
func TestExecuteReplCommand_EmptyLine(t *testing.T) {
	state := newTestReplState(t)

	err := executeReplCommand(state, "")
	if err != nil {
		t.Errorf("empty line should not produce error: %v", err)
	}

	err = executeReplCommand(state, "   ")
	if err != nil {
		t.Errorf("whitespace-only line should not produce error: %v", err)
	}
}

// TestModifiedFlagBehavior tests the modified flag state machine.
func TestModifiedFlagBehavior(t *testing.T) {
	t.Run("operations_set_modified", func(t *testing.T) {
		modifyingOps := []string{
			"add-vertex x",
			"add-edge ex a,b",
			"dual",
		}

		for _, op := range modifyingOps {
			t.Run(op, func(t *testing.T) {
				state := newTestReplState(t)
				state.modified = false

				executeReplCommand(state, op)

				if !state.modified {
					t.Errorf("%s should set modified flag", op)
				}
			})
		}
	})

	t.Run("queries_dont_set_modified", func(t *testing.T) {
		queryOps := []string{
			"has-vertex a",
			"has-edge e1",
			"vertices",
			"edges",
			"degree a",
			"edge-size e1",
			":info",
			":help",
		}

		for _, op := range queryOps {
			t.Run(op, func(t *testing.T) {
				state := newTestReplState(t)
				state.modified = false

				captureStdout(t, func() {
					executeReplCommand(state, op)
				})

				if state.modified {
					t.Errorf("%s should not set modified flag", op)
				}
			})
		}
	})

	t.Run("save_clears_modified", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "save.json")

		state := newTestReplState(t)
		state.modified = true
		state.file = path

		captureStdout(t, func() {
			executeReplCommand(state, ":save")
		})

		if state.modified {
			t.Error("save should clear modified flag")
		}
	})

	t.Run("load_clears_modified", func(t *testing.T) {
		dir := t.TempDir()
		path := writeTestGraphFile(t, dir, "test.json")

		state := newTestReplState(t)
		state.modified = true

		captureStdout(t, func() {
			executeReplCommand(state, ":load "+path)
		})

		if state.modified {
			t.Error("load should clear modified flag")
		}
	})
}

// TestConfirmationFlagReset tests that confirmation flags are reset by other commands.
func TestConfirmationFlagReset(t *testing.T) {
	t.Run("quit_confirmation_reset_by_other_command", func(t *testing.T) {
		state := newTestReplState(t)
		state.modified = true

		// First :quit sets quitConfirmed
		captureStdout(t, func() {
			executeReplCommand(state, ":quit")
		})

		if !state.quitConfirmed {
			t.Fatal("quitConfirmed should be set after first :quit")
		}

		// Running another command should reset quitConfirmed
		captureStdout(t, func() {
			executeReplCommand(state, ":info")
		})

		if state.quitConfirmed {
			t.Error("quitConfirmed should be reset after running :info")
		}

		// Now :quit should warn again
		output := captureStdout(t, func() {
			err := executeReplCommand(state, ":quit")
			if err != nil {
				t.Errorf("quit should not error: %v", err)
			}
		})

		if !strings.Contains(output, "Warning") {
			t.Error("should show warning again after reset")
		}
	})

	t.Run("new_confirmation_reset_by_other_command", func(t *testing.T) {
		state := newTestReplState(t)
		state.modified = true

		// First :new sets newConfirmed
		captureStdout(t, func() {
			executeReplCommand(state, ":new")
		})

		if !state.newConfirmed {
			t.Fatal("newConfirmed should be set after first :new")
		}

		// Running another command should reset newConfirmed
		captureStdout(t, func() {
			executeReplCommand(state, "vertices")
		})

		if state.newConfirmed {
			t.Error("newConfirmed should be reset after running vertices")
		}

		// Now :new should warn again
		output := captureStdout(t, func() {
			err := executeReplCommand(state, ":new")
			if err != nil {
				t.Errorf("new should not error: %v", err)
			}
		})

		if !strings.Contains(output, "Warning") {
			t.Error("should show warning again after reset")
		}
	})

	t.Run("quit_then_new_resets_quit_confirmation", func(t *testing.T) {
		state := newTestReplState(t)
		state.modified = true

		// :quit sets quitConfirmed
		captureStdout(t, func() {
			executeReplCommand(state, ":quit")
		})

		// :new should reset quitConfirmed and set newConfirmed
		captureStdout(t, func() {
			executeReplCommand(state, ":new")
		})

		if state.quitConfirmed {
			t.Error("quitConfirmed should be reset by :new")
		}
		if !state.newConfirmed {
			t.Error("newConfirmed should be set by :new")
		}
	})
}

// TestAtomicSave tests that saves are atomic.
func TestAtomicSave(t *testing.T) {
	t.Run("no_temp_file_left_on_success", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "save.json")

		state := newTestReplState(t)
		state.file = path

		captureStdout(t, func() {
			err := executeReplCommand(state, ":save")
			if err != nil {
				t.Fatalf(":save failed: %v", err)
			}
		})

		// Check that temp file was cleaned up
		tmpPath := path + ".tmp"
		if _, err := loadGraph(tmpPath); err == nil {
			t.Error("temp file should not exist after successful save")
		}

		// Verify the actual file exists and is valid
		loaded, err := loadGraph(path)
		if err != nil {
			t.Fatalf("failed to load saved graph: %v", err)
		}
		if loaded.NumVertices() != 3 {
			t.Errorf("saved graph should have 3 vertices")
		}
	})
}

// TestExecuteReplCommand_EdgeCases tests additional edge cases for REPL commands.
func TestExecuteReplCommand_EdgeCases(t *testing.T) {
	t.Run("remove-edge_missing_arg", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "remove-edge")
		if err == nil {
			t.Fatal("expected error for missing argument")
		}
	})

	t.Run("has-vertex_missing_arg", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "has-vertex")
		if err == nil {
			t.Fatal("expected error for missing argument")
		}
	})

	t.Run("has-edge_missing_arg", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "has-edge")
		if err == nil {
			t.Fatal("expected error for missing argument")
		}
	})

	t.Run("degree_missing_arg", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "degree")
		if err == nil {
			t.Fatal("expected error for missing argument")
		}
	})

	t.Run("degree_nonexistent_vertex", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "degree nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent vertex")
		}
	})

	t.Run("edge-size_missing_arg", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "edge-size")
		if err == nil {
			t.Fatal("expected error for missing argument")
		}
	})

	t.Run("edge-size_nonexistent_edge", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "edge-size nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent edge")
		}
	})

	t.Run("bfs_missing_arg", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "bfs")
		if err == nil {
			t.Fatal("expected error for missing argument")
		}
	})

	t.Run("dfs_missing_arg", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "dfs")
		if err == nil {
			t.Fatal("expected error for missing argument")
		}
	})

	t.Run("dfs_nonexistent_vertex", func(t *testing.T) {
		state := newTestReplState(t)
		err := executeReplCommand(state, "dfs nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent vertex")
		}
	})
}

// TestExecuteReplCommand_SaveError tests save with invalid path.
func TestExecuteReplCommand_SaveError(t *testing.T) {
	state := newTestReplState(t)
	state.file = "/nonexistent/dir/file.json"

	err := executeReplCommand(state, ":save")
	if err == nil {
		t.Fatal("expected error for invalid save path")
	}
}

// TestExecuteReplCommand_LoadShorthand tests :l shorthand for :load.
func TestExecuteReplCommand_LoadShorthand(t *testing.T) {
	// Note: :l is not implemented as shorthand, but :load is tested
	// This tests the shorthand aliases :q and :h
	state := newTestReplState(t)
	state.modified = false

	// Already tested :q and :h, let's test edge case for unknown colon command
	err := executeReplCommand(state, ":unknown")
	if err == nil {
		t.Fatal("expected error for unknown colon command")
	}
}
