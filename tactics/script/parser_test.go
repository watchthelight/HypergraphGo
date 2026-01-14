package script

import (
	"strings"
	"testing"
)

func TestParse_SimpleTheorem(t *testing.T) {
	input := `
-- Simple identity function proof
Theorem id : (Pi A Type (Pi x (Var 0) (Var 1)))
Proof
  intro A
  intro x
  exact (Var 0)
Qed
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(script.Theorems) != 1 {
		t.Fatalf("expected 1 theorem, got %d", len(script.Theorems))
	}

	thm := script.Theorems[0]
	if thm.Name != "id" {
		t.Errorf("expected name 'id', got %q", thm.Name)
	}

	if len(thm.Proof) != 3 {
		t.Errorf("expected 3 tactic commands, got %d", len(thm.Proof))
	}

	expected := []struct {
		name string
		args []string
	}{
		{"intro", []string{"A"}},
		{"intro", []string{"x"}},
		{"exact", []string{"(Var", "0)"}},
	}

	for i, exp := range expected {
		if i >= len(thm.Proof) {
			break
		}
		cmd := thm.Proof[i]
		if cmd.Name != exp.name {
			t.Errorf("command %d: expected name %q, got %q", i, exp.name, cmd.Name)
		}
		if len(cmd.Args) != len(exp.args) {
			t.Errorf("command %d: expected %d args, got %d", i, len(exp.args), len(cmd.Args))
		}
	}
}

func TestParse_MultipleTheorems(t *testing.T) {
	input := `
Theorem id : (Pi A Type (Pi x (Var 0) (Var 1)))
Proof
  intro A
  intro x
  assumption
Qed

Theorem const : (Pi A Type (Pi B Type (Pi a (Var 1) (Pi b (Var 1) (Var 3)))))
Proof
  intro A
  intro B
  intro a
  intro b
  exact (Var 1)
Qed
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(script.Theorems) != 2 {
		t.Fatalf("expected 2 theorems, got %d", len(script.Theorems))
	}

	if script.Theorems[0].Name != "id" {
		t.Errorf("expected first theorem 'id', got %q", script.Theorems[0].Name)
	}
	if script.Theorems[1].Name != "const" {
		t.Errorf("expected second theorem 'const', got %q", script.Theorems[1].Name)
	}
}

func TestParse_EmptyScript(t *testing.T) {
	input := `
-- Just comments
-- Nothing else
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(script.Theorems) != 0 {
		t.Errorf("expected 0 theorems, got %d", len(script.Theorems))
	}
}

func TestParse_NoProofBlock(t *testing.T) {
	input := `
Theorem bad : Type
`
	_, err := ParseString(input)
	if err == nil {
		t.Fatal("expected error for missing Proof block")
	}
	if !strings.Contains(err.Error(), "end of file") {
		t.Errorf("expected 'end of file' error, got: %v", err)
	}
}

func TestParse_NoQed(t *testing.T) {
	input := `
Theorem bad : Type
Proof
  intro x
`
	_, err := ParseString(input)
	if err == nil {
		t.Fatal("expected error for missing Qed")
	}
	if !strings.Contains(err.Error(), "end of file") {
		t.Errorf("expected 'end of file' error, got: %v", err)
	}
}

func TestParse_MissingColon(t *testing.T) {
	input := `
Theorem bad Type
Proof
Qed
`
	_, err := ParseString(input)
	if err == nil {
		t.Fatal("expected error for missing colon")
	}
	if !strings.Contains(err.Error(), ":") {
		t.Errorf("expected colon error, got: %v", err)
	}
}

func TestParse_InvalidType(t *testing.T) {
	input := `
Theorem bad : (invalid syntax here
Proof
Qed
`
	_, err := ParseString(input)
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(err.Error(), "parsing type") {
		t.Errorf("expected parsing type error, got: %v", err)
	}
}

func TestParse_EmptyTheoremName(t *testing.T) {
	input := `
Theorem  : Type
Proof
Qed
`
	_, err := ParseString(input)
	if err == nil {
		t.Fatal("expected error for empty theorem name")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected empty name error, got: %v", err)
	}
}

func TestParse_UnexpectedToken(t *testing.T) {
	input := `
garbage line
`
	_, err := ParseString(input)
	if err == nil {
		t.Fatal("expected error for unexpected token")
	}
	if !strings.Contains(err.Error(), "unexpected") {
		t.Errorf("expected 'unexpected' error, got: %v", err)
	}
}

func TestParse_LineNumbers(t *testing.T) {
	input := `
-- Comment on line 2
Theorem test : Type
Proof
  intro x
  bad_tactic
Qed
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find the bad_tactic command and check its line number
	found := false
	for _, thm := range script.Theorems {
		for _, cmd := range thm.Proof {
			if cmd.Name == "bad_tactic" {
				found = true
				// Line 6 (1-indexed): blank, comment, Theorem, Proof, intro, bad_tactic
				if cmd.Line != 6 {
					t.Errorf("expected line 6, got %d", cmd.Line)
				}
			}
		}
	}
	if !found {
		t.Error("bad_tactic not found in parsed script")
	}
}
