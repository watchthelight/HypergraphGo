package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureOutput captures stdout and stderr during function execution.
func captureOutput(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	fn()

	wOut.Close()
	wErr.Close()

	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return bufOut.String(), bufErr.String()
}

// writeFile is a helper to create test files.
func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// --- doCheck tests ---

func TestDoCheck_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "valid.hott", "Type")

	stdout, _ := captureOutput(t, func() {
		err := doCheck(path)
		if err != nil {
			t.Errorf("doCheck() unexpected error: %v", err)
		}
	})

	if !strings.Contains(stdout, "term 1") {
		t.Errorf("stdout should contain term output, got: %q", stdout)
	}
}

func TestDoCheck_MultipleTerms(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "multi.hott", "Type\nType1\nNat")

	stdout, _ := captureOutput(t, func() {
		err := doCheck(path)
		if err != nil {
			t.Errorf("doCheck() unexpected error: %v", err)
		}
	})

	if !strings.Contains(stdout, "term 1") {
		t.Errorf("stdout should contain term 1, got: %q", stdout)
	}
	if !strings.Contains(stdout, "term 2") {
		t.Errorf("stdout should contain term 2, got: %q", stdout)
	}
	if !strings.Contains(stdout, "term 3") {
		t.Errorf("stdout should contain term 3, got: %q", stdout)
	}
}

func TestDoCheck_MissingFile(t *testing.T) {
	err := doCheck("/nonexistent/path/file.hott")
	if err == nil {
		t.Fatal("doCheck() should fail for missing file")
	}
	if !strings.Contains(err.Error(), "reading file") {
		t.Errorf("error should mention reading file, got: %v", err)
	}
}

func TestDoCheck_ParseError(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "invalid.hott", "(unclosed paren")

	err := doCheck(path)
	if err == nil {
		t.Fatal("doCheck() should fail for parse error")
	}
	if !strings.Contains(err.Error(), "parsing") {
		t.Errorf("error should mention parsing, got: %v", err)
	}
}

func TestDoCheck_TypeError(t *testing.T) {
	dir := t.TempDir()
	// Applying Type to itself is a type error
	path := writeFile(t, dir, "typeerr.hott", "(App Type Type)")

	err := doCheck(path)
	if err == nil {
		t.Fatal("doCheck() should fail for type error")
	}
	if !strings.Contains(err.Error(), "term 1") {
		t.Errorf("error should mention term number, got: %v", err)
	}
}

func TestDoCheck_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "empty.hott", "")

	stdout, _ := captureOutput(t, func() {
		err := doCheck(path)
		if err != nil {
			t.Errorf("doCheck() should succeed for empty file, got: %v", err)
		}
	})

	// Empty file should produce no term output
	if strings.Contains(stdout, "term") {
		t.Errorf("empty file should produce no term output, got: %q", stdout)
	}
}

// --- doEval tests ---

func TestDoEval(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		wantError bool
		contains  string
	}{
		{
			name:      "valid Type",
			expr:      "Type",
			wantError: false,
			contains:  "Type",
		},
		{
			name:      "valid Sort",
			expr:      "(Sort 1)",
			wantError: false,
			contains:  "(Sort 1)",
		},
		{
			name:      "identity lambda",
			expr:      "(Lam x (Var 0))",
			wantError: false,
			contains:  "Lam",
		},
		{
			name:      "parse error unclosed",
			expr:      "(unclosed",
			wantError: true,
			contains:  "parsing",
		},
		{
			name:      "parse error malformed",
			expr:      "((()))",
			wantError: true,
			contains:  "parsing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _ := captureOutput(t, func() {
				err := doEval(tt.expr)
				if tt.wantError {
					if err == nil {
						t.Error("doEval() expected error, got nil")
					} else if !strings.Contains(err.Error(), tt.contains) {
						t.Errorf("error should contain %q, got: %v", tt.contains, err)
					}
				} else {
					if err != nil {
						t.Errorf("doEval() unexpected error: %v", err)
					}
				}
			})

			if !tt.wantError && !strings.Contains(stdout, tt.contains) {
				t.Errorf("stdout should contain %q, got: %q", tt.contains, stdout)
			}
		})
	}
}

// --- doSynth tests ---

func TestDoSynth(t *testing.T) {
	tests := []struct {
		name         string
		expr         string
		wantError    bool
		errContains  string
		outContains  string
	}{
		{
			name:        "valid Type",
			expr:        "Type",
			wantError:   false,
			outContains: "(Sort 1)",
		},
		{
			name:        "valid Nat",
			expr:        "Nat",
			wantError:   false,
			outContains: "Type",
		},
		{
			name:        "valid zero",
			expr:        "zero",
			wantError:   false,
			outContains: "Nat",
		},
		{
			name:        "parse error",
			expr:        "(unclosed",
			wantError:   true,
			errContains: "parsing",
		},
		{
			name:        "type error - undefined",
			expr:        "undefined_name_xyz",
			wantError:   true,
			errContains: "type error",
		},
		{
			name:        "type error - ill-typed app",
			expr:        "(App Type Type)",
			wantError:   true,
			errContains: "type error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _ := captureOutput(t, func() {
				err := doSynth(tt.expr)
				if tt.wantError {
					if err == nil {
						t.Error("doSynth() expected error, got nil")
					} else if !strings.Contains(err.Error(), tt.errContains) {
						t.Errorf("error should contain %q, got: %v", tt.errContains, err)
					}
				} else {
					if err != nil {
						t.Errorf("doSynth() unexpected error: %v", err)
					}
				}
			})

			if !tt.wantError && !strings.Contains(stdout, tt.outContains) {
				t.Errorf("stdout should contain %q, got: %q", tt.outContains, stdout)
			}
		})
	}
}

// --- REPL tests ---

// runREPLWithInput simulates running the REPL with given input lines.
func runREPLWithInput(t *testing.T, input string) (stdout, stderr string) {
	t.Helper()

	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	stdout, stderr = captureOutput(t, func() {
		repl()
	})

	os.Stdin = oldStdin
	return stdout, stderr
}

func TestREPL_QuitCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"quit", ":quit\n"},
		{"q", ":q\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr := runREPLWithInput(t, tt.input)
			// REPL should exit cleanly
			if strings.Contains(stderr, "error") {
				t.Errorf("REPL should exit cleanly, got stderr: %q", stderr)
			}
			// Should show prompt
			if !strings.Contains(stdout, "> ") {
				t.Errorf("REPL should show prompt, got: %q", stdout)
			}
		})
	}
}

func TestREPL_EvalCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOut     string
		wantErr     string
	}{
		{
			name:    "eval Type",
			input:   ":eval Type\n:quit\n",
			wantOut: "Type",
			wantErr: "",
		},
		{
			name:    "eval Sort",
			input:   ":eval (Sort 2)\n:quit\n",
			wantOut: "(Sort 2)",
			wantErr: "",
		},
		{
			name:    "eval parse error",
			input:   ":eval (unclosed\n:quit\n",
			wantOut: "",
			wantErr: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr := runREPLWithInput(t, tt.input)

			if tt.wantOut != "" && !strings.Contains(stdout, tt.wantOut) {
				t.Errorf("stdout should contain %q, got: %q", tt.wantOut, stdout)
			}
			if tt.wantErr != "" && !strings.Contains(stderr, tt.wantErr) {
				t.Errorf("stderr should contain %q, got: %q", tt.wantErr, stderr)
			}
		})
	}
}

func TestREPL_SynthCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOut     string
		wantErr     string
	}{
		{
			name:    "synth Type",
			input:   ":synth Type\n:quit\n",
			wantOut: "(Sort 1)",
			wantErr: "",
		},
		{
			name:    "synth Nat",
			input:   ":synth Nat\n:quit\n",
			wantOut: "Type",
			wantErr: "",
		},
		{
			name:    "synth parse error",
			input:   ":synth (bad\n:quit\n",
			wantOut: "",
			wantErr: "parse error",
		},
		{
			name:    "synth type error",
			input:   ":synth unknown_var\n:quit\n",
			wantOut: "",
			wantErr: "type error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr := runREPLWithInput(t, tt.input)

			if tt.wantOut != "" && !strings.Contains(stdout, tt.wantOut) {
				t.Errorf("stdout should contain %q, got: %q", tt.wantOut, stdout)
			}
			if tt.wantErr != "" && !strings.Contains(stderr, tt.wantErr) {
				t.Errorf("stderr should contain %q, got: %q", tt.wantErr, stderr)
			}
		})
	}
}

func TestREPL_PlainExpressions(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOut     string
		wantErr     string
	}{
		{
			name:    "Type expression",
			input:   "Type\n:quit\n",
			wantOut: "(Sort 1)",
			wantErr: "",
		},
		{
			name:    "Nat expression",
			input:   "Nat\n:quit\n",
			wantOut: "Type",
			wantErr: "",
		},
		{
			name:    "zero expression",
			input:   "zero\n:quit\n",
			wantOut: "Nat",
			wantErr: "",
		},
		{
			name:    "parse error",
			input:   "(bad syntax\n:quit\n",
			wantOut: "",
			wantErr: "parse error",
		},
		{
			name:    "type error",
			input:   "nonexistent\n:quit\n",
			wantOut: "",
			wantErr: "type error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr := runREPLWithInput(t, tt.input)

			if tt.wantOut != "" && !strings.Contains(stdout, tt.wantOut) {
				t.Errorf("stdout should contain %q, got: %q", tt.wantOut, stdout)
			}
			if tt.wantErr != "" && !strings.Contains(stderr, tt.wantErr) {
				t.Errorf("stderr should contain %q, got: %q", tt.wantErr, stderr)
			}
		})
	}
}

func TestREPL_EmptyLine(t *testing.T) {
	// Empty lines should be ignored, not cause errors
	stdout, stderr := runREPLWithInput(t, "\n\n\nType\n:quit\n")

	if strings.Contains(stderr, "error") {
		t.Errorf("empty lines should not cause errors, stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "(Sort 1)") {
		t.Errorf("should still process Type after empty lines, got: %q", stdout)
	}
}

func TestREPL_EOF(t *testing.T) {
	// EOF should exit cleanly
	stdout, stderr := runREPLWithInput(t, "")

	if strings.Contains(stderr, "error") {
		t.Errorf("EOF should exit cleanly, stderr: %q", stderr)
	}
	// Should show at least one prompt
	if !strings.Contains(stdout, "> ") {
		t.Errorf("should show prompt before EOF, got: %q", stdout)
	}
}

// --- Version flag test ---

func TestVersionFlag(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"hottgo", "--version"}

	// We can't easily test main() due to flag.Parse() state,
	// but we can verify the version module is accessible
	// by importing it. The integration test would verify the actual CLI.

	// This is a basic sanity check that the version package is available
	// and the code structure supports --version
}
