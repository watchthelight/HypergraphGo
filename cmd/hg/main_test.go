package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureOutput captures both stdout and stderr during execution.
func captureOutput(t *testing.T, f func()) (stdout, stderr string) {
	t.Helper()

	// Save original
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create pipes
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	// Run function
	f()

	// Restore and read
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	return bufOut.String(), bufErr.String()
}

// TestPrintUsage tests the usage output.
func TestPrintUsage(t *testing.T) {
	stdout, _ := captureOutput(t, func() {
		printUsage()
	})

	expectedSections := []string{
		"hg - Hypergraph CLI",
		"Usage:",
		"Core:",
		"info",
		"add-vertex",
		"remove-vertex",
		"has-vertex",
		"add-edge",
		"remove-edge",
		"has-edge",
		"vertices",
		"edges",
		"degree",
		"edge-size",
		"copy",
		"Transforms:",
		"dual",
		"two-section",
		"line-graph",
		"Traversal:",
		"bfs",
		"dfs",
		"components",
		"Algorithms:",
		"hitting-set",
		"transversals",
		"coloring",
		"I/O:",
		"new",
		"incidence",
		"validate",
		"Meta:",
		"help",
		"repl",
		"--version",
	}

	for _, section := range expectedSections {
		if !strings.Contains(stdout, section) {
			t.Errorf("usage should contain %q", section)
		}
	}
}

// TestSubcommandRouting tests that subcommands are correctly routed.
func TestSubcommandRouting(t *testing.T) {
	// We can't easily test the main() function directly since it uses os.Exit,
	// but we can verify the command mapping works through individual command tests.
	// This test documents the expected command structure.

	commands := []struct {
		name        string
		description string
	}{
		// Core operations
		{"info", "Display hypergraph info"},
		{"add-vertex", "Add a vertex"},
		{"remove-vertex", "Remove a vertex"},
		{"has-vertex", "Check vertex existence"},
		{"add-edge", "Add a hyperedge"},
		{"remove-edge", "Remove a hyperedge"},
		{"has-edge", "Check edge existence"},
		{"vertices", "List all vertices"},
		{"edges", "List all edges"},
		{"degree", "Get vertex degree"},
		{"edge-size", "Get edge size"},
		{"copy", "Copy hypergraph"},

		// Transforms
		{"dual", "Compute dual hypergraph"},
		{"two-section", "Compute 2-section graph"},
		{"line-graph", "Compute line graph"},

		// Traversal
		{"bfs", "Breadth-first search"},
		{"dfs", "Depth-first search"},
		{"components", "Connected components"},

		// Algorithms
		{"hitting-set", "Greedy hitting set"},
		{"transversals", "Minimal transversals"},
		{"coloring", "Greedy coloring"},

		// I/O
		{"new", "Create empty hypergraph"},
		{"incidence", "Print incidence matrix"},
		{"validate", "Validate JSON file"},

		// Meta
		{"help", "Show command help"},
		{"repl", "Interactive mode"},
	}

	stdout, _ := captureOutput(t, func() {
		printUsage()
	})

	for _, cmd := range commands {
		t.Run(cmd.name, func(t *testing.T) {
			if !strings.Contains(stdout, cmd.name) {
				t.Errorf("usage should list command: %s", cmd.name)
			}
		})
	}
}

// TestCommandHelpEntries verifies all commands have help entries.
func TestCommandHelpEntries(t *testing.T) {
	expectedCommands := []string{
		"info", "new", "validate", "add-vertex", "remove-vertex",
		"has-vertex", "add-edge", "remove-edge", "has-edge",
		"vertices", "edges", "degree", "edge-size", "copy",
		"dual", "two-section", "line-graph",
		"bfs", "dfs", "components",
		"hitting-set", "transversals", "coloring", "incidence",
		"repl",
	}

	for _, cmd := range expectedCommands {
		t.Run(cmd, func(t *testing.T) {
			if _, ok := commandHelp[cmd]; !ok {
				t.Errorf("command %q should have a help entry in commandHelp map", cmd)
			}
		})
	}
}

// TestCommandHelpContent verifies help entries have required content.
func TestCommandHelpContent(t *testing.T) {
	for cmd, help := range commandHelp {
		t.Run(cmd, func(t *testing.T) {
			// Each help entry should contain:
			// 1. The command name
			if !strings.Contains(help, cmd) {
				t.Errorf("help for %q should contain command name", cmd)
			}

			// 2. "Usage:" section
			if !strings.Contains(help, "Usage:") {
				t.Errorf("help for %q should contain 'Usage:' section", cmd)
			}

			// 3. For most commands, "Flags:" section (exceptions: help, repl may not have required flags)
			if cmd != "repl" && !strings.Contains(help, "hg "+cmd) {
				t.Errorf("help for %q should show usage pattern", cmd)
			}
		})
	}
}

// TestUnknownCommand tests that unknown commands produce appropriate error message.
func TestUnknownCommand(t *testing.T) {
	// Since main() calls os.Exit, we test the error path through cmdHelp
	err := cmdHelp([]string{"nonexistent-command"})
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("error should mention 'unknown command', got: %v", err)
	}
}

// TestHelpCommandCategories verifies the categorization of commands in help.
func TestHelpCommandCategories(t *testing.T) {
	stdout, _ := captureOutput(t, func() {
		printUsage()
	})

	categories := []struct {
		name     string
		commands []string
	}{
		{"Core:", []string{"info", "add-vertex", "remove-vertex", "has-vertex",
			"add-edge", "remove-edge", "has-edge", "vertices", "edges",
			"degree", "edge-size", "copy"}},
		{"Transforms:", []string{"dual", "two-section", "line-graph"}},
		{"Traversal:", []string{"bfs", "dfs", "components"}},
		{"Algorithms:", []string{"hitting-set", "transversals", "coloring"}},
		{"I/O:", []string{"new", "incidence", "validate"}},
		{"Meta:", []string{"help", "repl"}},
	}

	for _, cat := range categories {
		t.Run(cat.name, func(t *testing.T) {
			if !strings.Contains(stdout, cat.name) {
				t.Errorf("usage should contain category: %s", cat.name)
			}
			for _, cmd := range cat.commands {
				if !strings.Contains(stdout, cmd) {
					t.Errorf("category %s should contain command: %s", cat.name, cmd)
				}
			}
		})
	}
}

// TestGlobalFlags tests that global flags are documented.
func TestGlobalFlags(t *testing.T) {
	stdout, _ := captureOutput(t, func() {
		printUsage()
	})

	if !strings.Contains(stdout, "Global Flags:") {
		t.Error("usage should contain 'Global Flags:' section")
	}
	if !strings.Contains(stdout, "--version") {
		t.Error("usage should document --version flag")
	}
}

// TestHelpFooter tests the help footer message.
func TestHelpFooter(t *testing.T) {
	stdout, _ := captureOutput(t, func() {
		printUsage()
	})

	if !strings.Contains(stdout, "hg help <command>") {
		t.Error("usage should mention 'hg help <command>' for more information")
	}
}

// TestMultipleHelpRequests tests repeated calls to help work correctly.
func TestMultipleHelpRequests(t *testing.T) {
	commands := []string{"info", "new", "validate"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			for i := 0; i < 3; i++ {
				err := cmdHelp([]string{cmd})
				if err != nil {
					t.Errorf("iteration %d: cmdHelp failed: %v", i, err)
				}
			}
		})
	}
}

// TestHelpWithExtraArgs tests help command ignores extra arguments.
func TestHelpWithExtraArgs(t *testing.T) {
	// help should use first arg and ignore others
	stdout, _ := captureOutput(t, func() {
		err := cmdHelp([]string{"info", "extra", "args"})
		if err != nil {
			t.Fatalf("cmdHelp failed: %v", err)
		}
	})

	if !strings.Contains(stdout, "hg info") {
		t.Error("help should show info command help")
	}
}
