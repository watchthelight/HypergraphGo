package script

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/kernel/check"
)

func TestExecute_SimpleTheorem(t *testing.T) {
	input := `
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

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Theorems) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Theorems))
	}

	thm := result.Theorems[0]
	if !thm.Success {
		t.Errorf("theorem should succeed: %v", thm.Error)
	}
	if thm.ProofTerm == nil {
		t.Error("proof term should not be nil")
	}
}

func TestExecute_WithAssumption(t *testing.T) {
	input := `
Theorem id2 : (Pi A Type (Pi x (Var 0) (Var 1)))
Proof
  intro A
  intro x
  assumption
Qed
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Theorems) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Theorems))
	}

	thm := result.Theorems[0]
	if !thm.Success {
		t.Errorf("theorem should succeed: %v", thm.Error)
	}
}

func TestExecute_Reflexivity(t *testing.T) {
	input := `
Theorem refl_test : (Id Nat zero zero)
Proof
  reflexivity
Qed
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Theorems) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Theorems))
	}

	thm := result.Theorems[0]
	if !thm.Success {
		t.Errorf("theorem should succeed: %v", thm.Error)
	}
}

func TestExecute_Unit(t *testing.T) {
	input := `
Theorem unit_proof : Unit
Proof
  constructor
Qed
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Theorems) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Theorems))
	}

	thm := result.Theorems[0]
	if !thm.Success {
		t.Errorf("theorem should succeed: %v", thm.Error)
	}
}

func TestExecute_Sum(t *testing.T) {
	input := `
Theorem sum_left : (Pi A Type (Pi a (Var 0) (App (App Sum (Var 1)) Nat)))
Proof
  intro A
  intro a
  left
  assumption
Qed
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Theorems) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Theorems))
	}

	thm := result.Theorems[0]
	if !thm.Success {
		t.Errorf("theorem should succeed: %v", thm.Error)
	}
}

func TestExecute_MultipleTheorems(t *testing.T) {
	input := `
Theorem id : (Pi A Type (Pi x (Var 0) (Var 1)))
Proof
  intro A
  intro x
  assumption
Qed

Theorem unit : Unit
Proof
  constructor
Qed
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Theorems) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Theorems))
	}

	for i, thm := range result.Theorems {
		if !thm.Success {
			t.Errorf("theorem %d (%s) should succeed: %v", i, thm.Name, thm.Error)
		}
	}
}

func TestExecute_FailingTheorem(t *testing.T) {
	input := `
Theorem bad : (Pi A Type A)
Proof
  intro A
  assumption
Qed
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Theorems) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Theorems))
	}

	thm := result.Theorems[0]
	if thm.Success {
		t.Error("theorem should fail (unprovable)")
	}
	if thm.Error == nil {
		t.Error("expected error for failing theorem")
	}
}

func TestExecute_UnknownTactic(t *testing.T) {
	input := `
Theorem bad : Type
Proof
  unknown_tactic
Qed
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Theorems) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Theorems))
	}

	thm := result.Theorems[0]
	if thm.Success {
		t.Error("theorem should fail (unknown tactic)")
	}
}

func TestExecute_IncompleteProof(t *testing.T) {
	input := `
Theorem incomplete : (Pi A Type (Pi B Type (Var 1)))
Proof
  intro A
Qed
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Theorems) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Theorems))
	}

	thm := result.Theorems[0]
	if thm.Success {
		t.Error("theorem should fail (incomplete proof)")
	}
	if thm.Error == nil {
		t.Error("expected error for incomplete proof")
	}
}

func TestExecute_InvalidGoalType(t *testing.T) {
	input := `
Theorem bad_type : (App x y)
Proof
Qed
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Theorems) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Theorems))
	}

	thm := result.Theorems[0]
	if thm.Success {
		t.Error("theorem should fail (invalid goal type)")
	}
}
