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

func TestExecute_Definition(t *testing.T) {
	input := `
Definition id : (Pi A Type (Pi x (Var 0) (Var 1))) := (Lam A (Lam x (Var 0)))
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}

	item := result.Items[0]
	if item.Kind != ItemDefinition {
		t.Errorf("expected ItemDefinition, got %v", item.Kind)
	}
	if !item.Success {
		t.Errorf("definition should succeed: %v", item.Error)
	}

	// Check that definition was added to global env
	globals := checker.Globals()
	if !globals.Has("id") {
		t.Error("definition 'id' should be added to global environment")
	}
}

func TestExecute_Axiom(t *testing.T) {
	input := `
Axiom funext : (Pi A Type (Pi B Type Type))
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}

	item := result.Items[0]
	if item.Kind != ItemAxiom {
		t.Errorf("expected ItemAxiom, got %v", item.Kind)
	}
	if !item.Success {
		t.Errorf("axiom should succeed: %v", item.Error)
	}

	// Check that axiom was added to global env
	globals := checker.Globals()
	if !globals.Has("funext") {
		t.Error("axiom 'funext' should be added to global environment")
	}
}

func TestExecute_DefinitionAndTheorem(t *testing.T) {
	// Test that a theorem can reference a previously defined definition
	// The theorem proves that myid has the expected type
	input := `
Definition myid : (Pi A Type (Pi x (Var 0) (Var 1))) := (Lam A (Lam x (Var 0)))

Theorem myid_has_type : (Sigma f (Pi A Type (Pi x (Var 0) (Var 1))) Unit)
Proof
  split
  exact myid
  constructor
Qed
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}

	// Definition should succeed
	if !result.Items[0].Success {
		t.Errorf("definition should succeed: %v", result.Items[0].Error)
	}

	// Theorem should succeed (referencing the definition)
	if !result.Items[1].Success {
		t.Errorf("theorem should succeed: %v", result.Items[1].Error)
	}
}

func TestExecute_TheoremUsesEarlierTheorem(t *testing.T) {
	// Test that a later theorem can reference an earlier theorem
	input := `
Theorem refl_zero : (Id Nat zero zero)
Proof
  reflexivity
Qed

Theorem refl_zero_exists : (Sigma p (Id Nat zero zero) Unit)
Proof
  split
  exact refl_zero
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
		t.Fatalf("expected 2 theorems, got %d", len(result.Theorems))
	}

	// Both theorems should succeed
	for i, thm := range result.Theorems {
		if !thm.Success {
			t.Errorf("theorem %d (%s) should succeed: %v", i, thm.Name, thm.Error)
		}
	}

	// Check that both theorems were added to global env
	globals := checker.Globals()
	if !globals.Has("refl_zero") {
		t.Error("theorem 'refl_zero' should be added to global environment")
	}
	if !globals.Has("refl_zero_exists") {
		t.Error("theorem 'refl_zero_exists' should be added to global environment")
	}
}

func TestExecute_DefinitionBadType(t *testing.T) {
	input := `
Definition bad : (App x y) := Type
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}

	if result.Items[0].Success {
		t.Error("definition with invalid type should fail")
	}
}

func TestExecute_DefinitionBodyMismatch(t *testing.T) {
	input := `
Definition bad : Nat := true
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}

	if result.Items[0].Success {
		t.Error("definition with mismatched body type should fail")
	}
}

func TestExecute_AxiomBadType(t *testing.T) {
	input := `
Axiom bad : (App x y)
`
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	checker := check.NewCheckerWithStdlib()
	result := Execute(script, checker)

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}

	if result.Items[0].Success {
		t.Error("axiom with invalid type should fail")
	}
}
