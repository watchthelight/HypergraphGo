package proofstate

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func TestNewProofState(t *testing.T) {
	goalType := ast.Sort{U: 0}
	state := NewProofState(goalType, nil)

	if state.GoalCount() != 1 {
		t.Errorf("expected 1 goal, got %d", state.GoalCount())
	}

	if state.IsComplete() {
		t.Error("expected incomplete proof")
	}

	if state.CurrentGoal() == nil {
		t.Error("expected a current goal")
	}
}

func TestNewProofStateWithHypotheses(t *testing.T) {
	goalType := ast.Sort{U: 0}
	hyps := []Hypothesis{
		{Name: "A", Type: ast.Sort{U: 0}},
		{Name: "x", Type: ast.Var{Ix: 0}},
	}
	state := NewProofState(goalType, hyps)

	goal := state.CurrentGoal()
	if goal == nil {
		t.Fatal("expected a current goal")
	}

	if len(goal.Hypotheses) != 2 {
		t.Errorf("expected 2 hypotheses, got %d", len(goal.Hypotheses))
	}

	if goal.Hypotheses[0].Name != "A" {
		t.Errorf("expected first hypothesis to be A, got %s", goal.Hypotheses[0].Name)
	}
}

func TestProofStateCurrentGoalEmpty(t *testing.T) {
	state := &ProofState{Goals: nil}
	if state.CurrentGoal() != nil {
		t.Error("expected nil current goal for empty state")
	}
}

func TestProofStateFocus(t *testing.T) {
	state := NewProofState(ast.Sort{U: 0}, nil)

	// Add more goals
	id1 := state.AddGoal(ast.Sort{U: 1}, nil)
	id2 := state.AddGoal(ast.Sort{U: 2}, nil)

	// Focus on id2
	if err := state.Focus(id2); err != nil {
		t.Fatalf("Focus failed: %v", err)
	}

	if state.CurrentGoal().ID != id2 {
		t.Errorf("expected focused goal to be %d, got %d", id2, state.CurrentGoal().ID)
	}

	// Focus on id1
	if err := state.Focus(id1); err != nil {
		t.Fatalf("Focus failed: %v", err)
	}

	if state.CurrentGoal().ID != id1 {
		t.Errorf("expected focused goal to be %d, got %d", id1, state.CurrentGoal().ID)
	}

	// Focus on non-existent goal
	if err := state.Focus(999); err == nil {
		t.Error("expected error for non-existent goal")
	}
}

func TestProofStateAddGoal(t *testing.T) {
	state := NewProofState(ast.Sort{U: 0}, nil)
	initialCount := state.GoalCount()

	id := state.AddGoal(ast.Sort{U: 1}, []Hypothesis{{Name: "x", Type: ast.Sort{U: 0}}})

	if state.GoalCount() != initialCount+1 {
		t.Errorf("expected %d goals, got %d", initialCount+1, state.GoalCount())
	}

	goal, ok := state.GetGoal(id)
	if !ok {
		t.Fatal("expected to find added goal")
	}

	if len(goal.Hypotheses) != 1 {
		t.Errorf("expected 1 hypothesis, got %d", len(goal.Hypotheses))
	}
}

func TestProofStateSolveGoal(t *testing.T) {
	state := NewProofState(ast.Sort{U: 0}, nil)
	goalID := state.CurrentGoal().ID

	err := state.SolveGoal(goalID, ast.Sort{U: 0})
	if err != nil {
		t.Fatalf("SolveGoal failed: %v", err)
	}

	if !state.IsComplete() {
		t.Error("expected proof to be complete after solving only goal")
	}

	// Try to solve non-existent goal
	err = state.SolveGoal(999, ast.Sort{U: 0})
	if err == nil {
		t.Error("expected error for non-existent goal")
	}
}

func TestProofStateReplaceGoal(t *testing.T) {
	state := NewProofState(ast.Sort{U: 0}, nil)
	goalID := state.CurrentGoal().ID

	newGoals := []Goal{
		{Type: ast.Sort{U: 1}},
		{Type: ast.Sort{U: 2}},
	}

	err := state.ReplaceGoal(goalID, newGoals)
	if err != nil {
		t.Fatalf("ReplaceGoal failed: %v", err)
	}

	if state.GoalCount() != 2 {
		t.Errorf("expected 2 goals, got %d", state.GoalCount())
	}

	// Try to replace non-existent goal
	err = state.ReplaceGoal(999, newGoals)
	if err == nil {
		t.Error("expected error for non-existent goal")
	}
}

func TestProofStateClone(t *testing.T) {
	state := NewProofState(ast.Sort{U: 0}, []Hypothesis{{Name: "x", Type: ast.Sort{U: 0}}})
	state.AddGoal(ast.Sort{U: 1}, nil)

	clone := state.Clone()

	// Verify clone has same structure
	if clone.GoalCount() != state.GoalCount() {
		t.Errorf("expected %d goals, got %d", state.GoalCount(), clone.GoalCount())
	}

	// Modify original, verify clone unchanged
	state.AddGoal(ast.Sort{U: 2}, nil)
	if clone.GoalCount() == state.GoalCount() {
		t.Error("clone should not be affected by original modifications")
	}
}

func TestProofStateUndoNoHistory(t *testing.T) {
	state := NewProofState(ast.Sort{U: 0}, nil)

	if state.Undo() {
		t.Error("expected Undo to fail with no history")
	}
}

func TestProofStateSaveAndUndo(t *testing.T) {
	state := NewProofState(ast.Sort{U: 0}, nil)
	originalCount := state.GoalCount()

	state.SaveState()
	state.AddGoal(ast.Sort{U: 1}, nil)

	if state.GoalCount() != originalCount+1 {
		t.Errorf("expected %d goals after add, got %d", originalCount+1, state.GoalCount())
	}

	if !state.Undo() {
		t.Error("Undo should succeed with history")
	}

	if state.GoalCount() != originalCount {
		t.Errorf("expected %d goals after undo, got %d", originalCount, state.GoalCount())
	}
}

func TestProofStateExtractTermIncomplete(t *testing.T) {
	state := NewProofState(ast.Sort{U: 0}, nil)

	_, err := state.ExtractTerm()
	if err == nil {
		t.Error("expected error for incomplete proof")
	}
}

func TestProofStateExtractTermComplete(t *testing.T) {
	state := NewProofState(ast.Sort{U: 0}, nil)
	goalID := state.CurrentGoal().ID
	solution := ast.Sort{U: 0}

	if err := state.SolveGoal(goalID, solution); err != nil {
		t.Fatalf("SolveGoal failed: %v", err)
	}

	term, err := state.ExtractTerm()
	if err != nil {
		t.Fatalf("ExtractTerm failed: %v", err)
	}

	if _, ok := term.(ast.Sort); !ok {
		t.Errorf("expected Sort, got %T", term)
	}
}

func TestGoalLookupHypothesis(t *testing.T) {
	goal := Goal{
		Hypotheses: []Hypothesis{
			{Name: "A", Type: ast.Sort{U: 0}},
			{Name: "x", Type: ast.Var{Ix: 0}},
			{Name: "y", Type: ast.Var{Ix: 1}},
		},
	}

	// Find existing hypothesis
	hyp, ix, ok := goal.LookupHypothesis("x")
	if !ok {
		t.Error("expected to find hypothesis x")
	}
	if hyp.Name != "x" {
		t.Errorf("expected name x, got %s", hyp.Name)
	}
	// y is at index 2 (last), x is at index 1, so ix for x should be 1
	if ix != 1 {
		t.Errorf("expected de Bruijn index 1, got %d", ix)
	}

	// Find first hypothesis
	hyp, ix, ok = goal.LookupHypothesis("A")
	if !ok {
		t.Error("expected to find hypothesis A")
	}
	if ix != 2 {
		t.Errorf("expected de Bruijn index 2, got %d", ix)
	}

	// Look up non-existent
	_, _, ok = goal.LookupHypothesis("z")
	if ok {
		t.Error("expected not to find z")
	}
}

func TestGoalAddHypothesis(t *testing.T) {
	goal := Goal{Hypotheses: nil}

	goal.AddHypothesis("x", ast.Sort{U: 0})
	if len(goal.Hypotheses) != 1 {
		t.Errorf("expected 1 hypothesis, got %d", len(goal.Hypotheses))
	}

	goal.AddHypothesis("y", ast.Sort{U: 1})
	if len(goal.Hypotheses) != 2 {
		t.Errorf("expected 2 hypotheses, got %d", len(goal.Hypotheses))
	}
}

func TestProofStateGetGoal(t *testing.T) {
	state := NewProofState(ast.Sort{U: 0}, nil)
	id := state.CurrentGoal().ID

	goal, ok := state.GetGoal(id)
	if !ok {
		t.Error("expected to find goal")
	}
	if goal.ID != id {
		t.Errorf("expected ID %d, got %d", id, goal.ID)
	}

	_, ok = state.GetGoal(999)
	if ok {
		t.Error("expected not to find non-existent goal")
	}
}

func TestProofStateFormatGoal(t *testing.T) {
	state := NewProofState(ast.Sort{U: 0}, []Hypothesis{
		{Name: "A", Type: ast.Sort{U: 0}},
	})

	goal := state.CurrentGoal()
	output := state.FormatGoal(goal)

	if output == "" {
		t.Error("FormatGoal returned empty string")
	}
}

func TestProofStateFormatStateComplete(t *testing.T) {
	state := NewProofState(ast.Sort{U: 0}, nil)
	state.SolveGoal(state.CurrentGoal().ID, ast.Sort{U: 0})

	output := state.FormatState()
	if output != "No more goals.\n" {
		t.Errorf("expected 'No more goals.' for complete state, got %q", output)
	}
}

func TestProofStateFormatStateIncomplete(t *testing.T) {
	state := NewProofState(ast.Sort{U: 0}, nil)
	state.AddGoal(ast.Sort{U: 1}, nil)

	output := state.FormatState()
	if output == "" {
		t.Error("FormatState returned empty string")
	}
}

func TestHypothesisStruct(t *testing.T) {
	hyp := Hypothesis{Name: "x", Type: ast.Sort{U: 0}}
	if hyp.Name != "x" {
		t.Error("unexpected name")
	}
	if _, ok := hyp.Type.(ast.Sort); !ok {
		t.Error("unexpected type")
	}
}
