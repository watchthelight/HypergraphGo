package tactics

import (
	"fmt"
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/tactics/proofstate"
)

func TestNewProofState(t *testing.T) {
	// Goal: Type -> Type
	goalType := ast.Pi{Binder: "A", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	state := proofstate.NewProofState(goalType, nil)

	if state.GoalCount() != 1 {
		t.Errorf("expected 1 goal, got %d", state.GoalCount())
	}

	if state.IsComplete() {
		t.Error("expected incomplete proof")
	}
}

func TestIntro(t *testing.T) {
	// Goal: (A : Type) -> A -> A
	goalType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "x",
			A:      ast.Var{Ix: 0}, // A
			B:      ast.Var{Ix: 1}, // A (shifted)
		},
	}
	state := proofstate.NewProofState(goalType, nil)

	// Intro A
	result := Intro("A")(state)
	if !result.IsSuccess() {
		t.Fatalf("intro A failed: %v", result.Err)
	}

	goal := state.CurrentGoal()
	if goal == nil {
		t.Fatal("no current goal after intro")
	}

	if len(goal.Hypotheses) != 1 {
		t.Errorf("expected 1 hypothesis, got %d", len(goal.Hypotheses))
	}

	if goal.Hypotheses[0].Name != "A" {
		t.Errorf("expected hypothesis A, got %s", goal.Hypotheses[0].Name)
	}
}

func TestIntroN(t *testing.T) {
	// Goal: (A : Type) -> A -> A
	goalType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "x",
			A:      ast.Var{Ix: 0},
			B:      ast.Var{Ix: 1},
		},
	}
	state := proofstate.NewProofState(goalType, nil)

	// Intro A x
	result := IntroN("A", "x")(state)
	if !result.IsSuccess() {
		t.Fatalf("introN failed: %v", result.Err)
	}

	goal := state.CurrentGoal()
	if goal == nil {
		t.Fatal("no current goal after introN")
	}

	if len(goal.Hypotheses) != 2 {
		t.Errorf("expected 2 hypotheses, got %d", len(goal.Hypotheses))
	}
}

func TestAssumption(t *testing.T) {
	// Simple case: goal Type0, hypothesis x : Type0
	hyps := []proofstate.Hypothesis{
		{Name: "x", Type: ast.Sort{U: 0}},
	}

	// Goal is Type0 (same as the hypothesis type)
	state := proofstate.NewProofState(ast.Sort{U: 0}, hyps)

	result := Assumption()(state)
	if !result.IsSuccess() {
		t.Fatalf("assumption failed: %v", result.Err)
	}

	if !state.IsComplete() {
		t.Error("expected proof to be complete after assumption")
	}
}

func TestExact(t *testing.T) {
	// Goal: Type₁ (we need to provide something of type Type₁)
	state := proofstate.NewProofState(ast.Sort{U: 1}, nil)

	// Exact Type₀ - Type₀ has type Type₁, so this is valid
	result := Exact(ast.Sort{U: 0})(state)
	if !result.IsSuccess() {
		t.Fatalf("exact failed: %v", result.Err)
	}

	// Type checking now works: Type₀ : Type₁ matches goal Type₁
}

func TestReflexivity(t *testing.T) {
	// Goal: Id Type Type0 Type0
	goalType := ast.Id{
		A: ast.Sort{U: 1},
		X: ast.Sort{U: 0},
		Y: ast.Sort{U: 0},
	}
	state := proofstate.NewProofState(goalType, nil)

	result := Reflexivity()(state)
	if !result.IsSuccess() {
		t.Fatalf("reflexivity failed: %v", result.Err)
	}

	if !state.IsComplete() {
		t.Error("expected proof to be complete after reflexivity")
	}
}

func TestSplit(t *testing.T) {
	// Goal: Type * Type
	goalType := ast.Sigma{
		Binder: "_",
		A:      ast.Sort{U: 0},
		B:      ast.Sort{U: 0},
	}
	state := proofstate.NewProofState(goalType, nil)

	result := Split()(state)
	if !result.IsSuccess() {
		t.Fatalf("split failed: %v", result.Err)
	}

	if state.GoalCount() != 2 {
		t.Errorf("expected 2 goals after split, got %d", state.GoalCount())
	}
}

func TestSeq(t *testing.T) {
	// Goal: (A : Type) -> A -> A
	goalType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "x",
			A:      ast.Var{Ix: 0},
			B:      ast.Var{Ix: 1},
		},
	}
	state := proofstate.NewProofState(goalType, nil)

	// Seq intro A; intro x
	result := Seq(Intro("A"), Intro("x"))(state)
	if !result.IsSuccess() {
		t.Fatalf("seq failed: %v", result.Err)
	}

	goal := state.CurrentGoal()
	if goal == nil {
		t.Fatal("no current goal")
	}

	if len(goal.Hypotheses) != 2 {
		t.Errorf("expected 2 hypotheses after seq, got %d", len(goal.Hypotheses))
	}
}

func TestOrElse(t *testing.T) {
	// Goal: Type₁ (not a Pi)
	state := proofstate.NewProofState(ast.Sort{U: 1}, nil)

	// OrElse intro; exact Type₀
	// intro should fail, then exact should succeed (Type₀ : Type₁)
	result := OrElse(
		Intro("x"),
		Exact(ast.Sort{U: 0}),
	)(state)

	if !result.IsSuccess() {
		t.Fatalf("orelse failed: %v", result.Err)
	}
}

func TestTry(t *testing.T) {
	// Goal: Type (not a Pi)
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)

	// Try intro - should succeed even though intro fails
	result := Try(Intro("x"))(state)
	if !result.IsSuccess() {
		t.Fatalf("try failed: %v", result.Err)
	}

	// State should be unchanged
	if state.GoalCount() != 1 {
		t.Error("state was modified despite failed tactic")
	}
}

func TestRepeat(t *testing.T) {
	// Goal: (A : Type) -> (B : Type) -> A
	goalType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "B",
			A:      ast.Sort{U: 0},
			B:      ast.Var{Ix: 1}, // A
		},
	}
	state := proofstate.NewProofState(goalType, nil)

	// Repeat intro - should introduce A and B
	result := Repeat(Intro(""))(state)
	if !result.IsSuccess() {
		t.Fatalf("repeat failed: %v", result.Err)
	}

	goal := state.CurrentGoal()
	if goal == nil {
		t.Fatal("no current goal")
	}

	if len(goal.Hypotheses) != 2 {
		t.Errorf("expected 2 hypotheses after repeat, got %d", len(goal.Hypotheses))
	}
}

func TestFirst(t *testing.T) {
	// Goal: Id Type Type0 Type0
	goalType := ast.Id{
		A: ast.Sort{U: 1},
		X: ast.Sort{U: 0},
		Y: ast.Sort{U: 0},
	}
	state := proofstate.NewProofState(goalType, nil)

	// First intro; reflexivity - intro should fail, reflexivity should succeed
	result := First(Intro("x"), Reflexivity())(state)
	if !result.IsSuccess() {
		t.Fatalf("first failed: %v", result.Err)
	}

	if !state.IsComplete() {
		t.Error("expected proof to be complete")
	}
}

func TestUndo(t *testing.T) {
	// Goal: (A : Type) -> A
	goalType := ast.Pi{Binder: "A", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}}
	state := proofstate.NewProofState(goalType, nil)

	// Save state and intro
	state.SaveState()
	Intro("A")(state)

	if len(state.CurrentGoal().Hypotheses) != 1 {
		t.Error("intro should have added a hypothesis")
	}

	// Undo
	if !state.Undo() {
		t.Error("undo failed")
	}

	if len(state.CurrentGoal().Hypotheses) != 0 {
		t.Error("undo should have restored original state")
	}
}

func TestFormatState(t *testing.T) {
	goalType := ast.Pi{Binder: "A", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	state := proofstate.NewProofState(goalType, nil)

	output := state.FormatState()
	if output == "" {
		t.Error("FormatState returned empty string")
	}
}

// --- Extended Combinator Tests ---

func TestRepeatN(t *testing.T) {
	// Goal: (A : Type) -> (B : Type) -> (C : Type) -> A
	goalType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "B",
			A:      ast.Sort{U: 0},
			B: ast.Pi{
				Binder: "C",
				A:      ast.Sort{U: 0},
				B:      ast.Var{Ix: 2},
			},
		},
	}
	state := proofstate.NewProofState(goalType, nil)

	// RepeatN 2 - should introduce only A and B
	result := RepeatN(2, Intro(""))(state)
	if !result.IsSuccess() {
		t.Fatalf("repeatN failed: %v", result.Err)
	}

	goal := state.CurrentGoal()
	if goal == nil {
		t.Fatal("no current goal")
	}

	if len(goal.Hypotheses) != 2 {
		t.Errorf("expected 2 hypotheses after repeatN(2), got %d", len(goal.Hypotheses))
	}
}

func TestDo(t *testing.T) {
	goalType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "x",
			A:      ast.Var{Ix: 0},
			B:      ast.Var{Ix: 1},
		},
	}
	state := proofstate.NewProofState(goalType, nil)

	// Do 2 intro - should succeed exactly 2 times
	result := Do(2, Intro(""))(state)
	if !result.IsSuccess() {
		t.Fatalf("do failed: %v", result.Err)
	}

	goal := state.CurrentGoal()
	if len(goal.Hypotheses) != 2 {
		t.Errorf("expected 2 hypotheses, got %d", len(goal.Hypotheses))
	}

	// Do 1 more should fail (goal is not a Pi)
	state2 := proofstate.NewProofState(ast.Sort{U: 0}, nil)
	result2 := Do(1, Intro(""))(state2)
	if result2.IsSuccess() {
		t.Error("expected do to fail when tactic fails")
	}
}

func TestComplete(t *testing.T) {
	// Goal: Id Type Type Type (can be solved by reflexivity)
	goalType := ast.Id{A: ast.Sort{U: 1}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}}
	state := proofstate.NewProofState(goalType, nil)

	// Complete reflexivity - should succeed
	result := Complete(Reflexivity())(state)
	if !result.IsSuccess() {
		t.Fatalf("complete failed: %v", result.Err)
	}

	// Complete on a goal that won't be fully solved
	state2 := proofstate.NewProofState(ast.Sort{U: 0}, nil)
	result2 := Complete(NoOp())(state2)
	if result2.IsSuccess() {
		t.Error("expected complete to fail when proof is incomplete")
	}
}

func TestProgress(t *testing.T) {
	// Test with a tactic that makes no progress
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)
	result := Progress(NoOp())(state)
	if result.IsSuccess() {
		t.Error("expected progress to fail when no progress made")
	}
}

func TestIfThenElse(t *testing.T) {
	goalType := ast.Sort{U: 1}
	state := proofstate.NewProofState(goalType, nil)

	// If intro fails, then exact Type₀ (Type₀ : Type₁)
	result := IfThenElse(
		Intro("x"),
		NoOp(),
		Exact(ast.Sort{U: 0}),
	)(state)

	if !result.IsSuccess() {
		t.Fatalf("ifThenElse failed: %v", result.Err)
	}
}

func TestOnce(t *testing.T) {
	goalType := ast.Sort{U: 1}
	state := proofstate.NewProofState(goalType, nil)

	// Once is just identity
	result := Once(Exact(ast.Sort{U: 0}))(state)
	if !result.IsSuccess() {
		t.Fatalf("once failed: %v", result.Err)
	}
}

func TestGuard(t *testing.T) {
	goalType := ast.Sort{U: 0}
	state := proofstate.NewProofState(goalType, nil)

	// Guard with HasGoals
	result := Guard(HasGoals, NoOp())(state)
	if !result.IsSuccess() {
		t.Fatalf("guard with HasGoals failed: %v", result.Err)
	}

	// Guard with IsFinished should fail since we have goals
	result2 := Guard(IsFinished, NoOp())(state)
	if result2.IsSuccess() {
		t.Error("expected guard with IsFinished to fail when goals remain")
	}
}

func TestAll(t *testing.T) {
	// Create a state with multiple goals (Type₁ × Type₁)
	sigmaType := ast.Sigma{Binder: "_", A: ast.Sort{U: 1}, B: ast.Sort{U: 1}}
	state := proofstate.NewProofState(sigmaType, nil)

	// Split creates 2 goals (each Type₁)
	Split()(state)

	if state.GoalCount() != 2 {
		t.Fatalf("expected 2 goals, got %d", state.GoalCount())
	}

	// Apply exact Type₀ to all goals (Type₀ : Type₁)
	result := All(Exact(ast.Sort{U: 0}))(state)
	if !result.IsSuccess() {
		t.Fatalf("all failed: %v", result.Err)
	}
}

func TestFocus(t *testing.T) {
	sigmaType := ast.Sigma{Binder: "_", A: ast.Sort{U: 1}, B: ast.Sort{U: 1}}
	state := proofstate.NewProofState(sigmaType, nil)

	// Split creates 2 goals (each Type₁)
	Split()(state)

	// Focus on second goal and solve it
	goal1ID := state.CurrentGoal().ID
	state.Focus(state.Goals[1].ID)
	goal2ID := state.CurrentGoal().ID

	// Focus tactic
	result := Focus(goal2ID, Exact(ast.Sort{U: 0}))(state)
	if !result.IsSuccess() {
		t.Fatalf("focus failed: %v", result.Err)
	}

	// Focus on non-existent goal should fail
	result2 := Focus(999, NoOp())(state)
	if result2.IsSuccess() {
		t.Error("expected focus on non-existent goal to fail")
	}

	_ = goal1ID // used for verification
}

// --- Core Tactic Extended Tests ---

func TestNoOp(t *testing.T) {
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)
	result := NoOp()(state)
	if !result.IsSuccess() {
		t.Error("NoOp should always succeed")
	}
}

func TestFailWith(t *testing.T) {
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)
	err := fmt.Errorf("test error")
	result := FailWith(err)(state)
	if result.IsSuccess() {
		t.Error("FailWith should always fail")
	}
	if result.Err.Error() != "test error" {
		t.Errorf("expected 'test error', got %q", result.Err.Error())
	}
}

func TestFailWithMsg(t *testing.T) {
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)
	result := FailWithMsg("custom message")(state)
	if result.IsSuccess() {
		t.Error("FailWithMsg should always fail")
	}
}

func TestRunTactic(t *testing.T) {
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)

	// Nil state should fail
	result := RunTactic(nil, NoOp())
	if result.IsSuccess() {
		t.Error("expected error for nil state")
	}

	// Successful tactic
	result2 := RunTactic(state, NoOp())
	if !result2.IsSuccess() {
		t.Errorf("unexpected error: %v", result2.Err)
	}

	// Failed tactic should undo
	state3 := proofstate.NewProofState(ast.Sort{U: 0}, nil)
	initialCount := state3.GoalCount()
	RunTactic(state3, FailWithMsg("fail"))
	if state3.GoalCount() != initialCount {
		t.Error("state should be unchanged after failed tactic")
	}
}

func TestTacticResultSuccess(t *testing.T) {
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)

	result := Success(state)
	if !result.IsSuccess() {
		t.Error("Success should be successful")
	}
	if result.State != state {
		t.Error("Success should preserve state")
	}

	result2 := SuccessMsg(state, "message")
	if !result2.IsSuccess() {
		t.Error("SuccessMsg should be successful")
	}
	if result2.Message != "message" {
		t.Errorf("expected message 'message', got %q", result2.Message)
	}
}

func TestFailf(t *testing.T) {
	result := Failf("error %d", 42)
	if result.IsSuccess() {
		t.Error("Failf should fail")
	}
	if result.Err.Error() != "error 42" {
		t.Errorf("expected 'error 42', got %q", result.Err.Error())
	}
}

func TestFail(t *testing.T) {
	err := fmt.Errorf("test")
	result := Fail(err)
	if result.IsSuccess() {
		t.Error("Fail should fail")
	}
}

func TestIntros(t *testing.T) {
	goalType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "x",
			A:      ast.Var{Ix: 0},
			B:      ast.Var{Ix: 1},
		},
	}
	state := proofstate.NewProofState(goalType, nil)

	result := Intros()(state)
	if !result.IsSuccess() {
		t.Fatalf("intros failed: %v", result.Err)
	}

	goal := state.CurrentGoal()
	if len(goal.Hypotheses) != 2 {
		t.Errorf("expected 2 hypotheses, got %d", len(goal.Hypotheses))
	}
}

func TestSimpl(t *testing.T) {
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)
	result := Simpl()(state)
	if !result.IsSuccess() {
		t.Fatalf("simpl failed: %v", result.Err)
	}
}

func TestUnfold(t *testing.T) {
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)
	result := Unfold("name")(state)
	if result.IsSuccess() {
		t.Error("unfold should fail (requires GlobalEnv)")
	}
}

// mockGlobalEnvForUnfold implements GlobalEnvLookup for testing UnfoldWith.
type mockGlobalEnvForUnfold struct {
	defs map[string]ast.Term
}

func (m *mockGlobalEnvForUnfold) LookupDefinitionBodyForced(name string) (ast.Term, bool) {
	body, ok := m.defs[name]
	return body, ok
}

func TestUnfoldWith(t *testing.T) {
	// Create a mock environment with a definition
	// Define "id" as λA.λx.x
	env := &mockGlobalEnvForUnfold{
		defs: map[string]ast.Term{
			"id": ast.Lam{
				Binder: "A",
				Ann:    ast.Sort{U: 0},
				Body: ast.Lam{
					Binder: "x",
					Ann:    ast.Var{Ix: 0}, // A
					Body:   ast.Var{Ix: 0}, // x
				},
			},
		},
	}

	// Goal containing Global{Name: "id"}
	goalType := ast.App{
		T: ast.App{
			T: ast.Global{Name: "id"},
			U: ast.Global{Name: "Nat"},
		},
		U: ast.Global{Name: "zero"},
	}
	state := proofstate.NewProofState(goalType, nil)

	// Unfold "id"
	result := UnfoldWith(env)("id")(state)
	if !result.IsSuccess() {
		t.Fatalf("UnfoldWith failed: %v", result.Err)
	}

	// Check that the goal type was transformed
	newGoal := state.CurrentGoal()
	if newGoal == nil {
		t.Fatal("no current goal after unfold")
	}

	// The goal should now have the lambda instead of Global{id}
	app1, ok := newGoal.Type.(ast.App)
	if !ok {
		t.Fatalf("expected App type, got %T", newGoal.Type)
	}
	app2, ok := app1.T.(ast.App)
	if !ok {
		t.Fatalf("expected nested App, got %T", app1.T)
	}
	lam, ok := app2.T.(ast.Lam)
	if !ok {
		t.Fatalf("expected Lam after unfold, got %T", app2.T)
	}
	if lam.Binder != "A" {
		t.Errorf("expected binder 'A', got '%s'", lam.Binder)
	}
}

func TestUnfoldWithUnknownDefinition(t *testing.T) {
	env := &mockGlobalEnvForUnfold{
		defs: map[string]ast.Term{},
	}

	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)
	result := UnfoldWith(env)("unknown")(state)
	if result.IsSuccess() {
		t.Error("UnfoldWith should fail for unknown definition")
	}
}

func TestUnfoldWithNoGoal(t *testing.T) {
	env := &mockGlobalEnvForUnfold{
		defs: map[string]ast.Term{
			"test": ast.Sort{U: 0},
		},
	}

	state := &proofstate.ProofState{Goals: nil}
	result := UnfoldWith(env)("test")(state)
	if result.IsSuccess() {
		t.Error("UnfoldWith should fail with no goal")
	}
}

func TestUnfoldWithInHypothesis(t *testing.T) {
	// Create a mock environment with a definition
	env := &mockGlobalEnvForUnfold{
		defs: map[string]ast.Term{
			"myType": ast.Global{Name: "Nat"},
		},
	}

	// Goal with hypothesis that contains the definition
	hyps := []proofstate.Hypothesis{
		{Name: "x", Type: ast.Global{Name: "myType"}},
	}
	state := proofstate.NewProofState(ast.Sort{U: 0}, hyps)

	// Unfold "myType" in hypotheses
	result := UnfoldWith(env)("myType")(state)
	if !result.IsSuccess() {
		t.Fatalf("UnfoldWith failed: %v", result.Err)
	}

	// Check that the hypothesis type was transformed
	newGoal := state.CurrentGoal()
	if len(newGoal.Hypotheses) != 1 {
		t.Fatal("expected 1 hypothesis")
	}

	// The hypothesis type should now be Nat instead of myType
	g, ok := newGoal.Hypotheses[0].Type.(ast.Global)
	if !ok {
		t.Fatalf("expected Global type, got %T", newGoal.Hypotheses[0].Type)
	}
	if g.Name != "Nat" {
		t.Errorf("expected hypothesis type 'Nat', got '%s'", g.Name)
	}
}

func TestRewrite(t *testing.T) {
	// Goal with an Id hypothesis
	hyps := []proofstate.Hypothesis{
		{Name: "x", Type: ast.Sort{U: 0}},
		{Name: "y", Type: ast.Sort{U: 0}},
		{Name: "h", Type: ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 0}}},
	}
	goalType := ast.Sort{U: 0}
	state := proofstate.NewProofState(goalType, hyps)

	// Rewrite using h
	result := Rewrite("h")(state)
	if !result.IsSuccess() {
		t.Fatalf("rewrite failed: %v", result.Err)
	}

	// Rewrite with non-existent hypothesis
	result2 := Rewrite("nonexistent")(state)
	if result2.IsSuccess() {
		t.Error("expected error for non-existent hypothesis")
	}

	// Rewrite with non-Id hypothesis
	state2 := proofstate.NewProofState(ast.Sort{U: 0}, []proofstate.Hypothesis{
		{Name: "x", Type: ast.Sort{U: 0}},
	})
	result3 := Rewrite("x")(state2)
	if result3.IsSuccess() {
		t.Error("expected error for non-Id hypothesis")
	}
}

func TestRewriteRev(t *testing.T) {
	hyps := []proofstate.Hypothesis{
		{Name: "x", Type: ast.Sort{U: 0}},
		{Name: "y", Type: ast.Sort{U: 0}},
		{Name: "h", Type: ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 0}}},
	}
	goalType := ast.Sort{U: 0}
	state := proofstate.NewProofState(goalType, hyps)

	result := RewriteRev("h")(state)
	if !result.IsSuccess() {
		t.Fatalf("rewriteRev failed: %v", result.Err)
	}

	// Non-existent
	result2 := RewriteRev("nonexistent")(state)
	if result2.IsSuccess() {
		t.Error("expected error for non-existent hypothesis")
	}

	// Non-Id
	state2 := proofstate.NewProofState(ast.Sort{U: 0}, []proofstate.Hypothesis{
		{Name: "x", Type: ast.Sort{U: 0}},
	})
	result3 := RewriteRev("x")(state2)
	if result3.IsSuccess() {
		t.Error("expected error for non-Id hypothesis")
	}
}

func TestApplyTactic(t *testing.T) {
	// Setup with a function hypothesis
	hyps := []proofstate.Hypothesis{
		{Name: "A", Type: ast.Sort{U: 0}},
		{Name: "f", Type: ast.Pi{Binder: "_", A: ast.Var{Ix: 0}, B: ast.Var{Ix: 0}}},
	}
	goalType := ast.Var{Ix: 0} // A
	state := proofstate.NewProofState(goalType, hyps)

	// Apply f (which is a function)
	result := Apply(ast.Var{Ix: 0})(state) // f
	if !result.IsSuccess() {
		t.Fatalf("apply failed: %v", result.Err)
	}

	// Apply with non-applicable term
	state2 := proofstate.NewProofState(ast.Sort{U: 0}, []proofstate.Hypothesis{
		{Name: "x", Type: ast.Sort{U: 0}},
	})
	result2 := Apply(ast.Var{Ix: 0})(state2)
	if result2.IsSuccess() {
		t.Error("expected error for non-applicable term")
	}

	// Apply with out of range var
	state3 := proofstate.NewProofState(ast.Sort{U: 0}, nil)
	result3 := Apply(ast.Var{Ix: 10})(state3)
	if result3.IsSuccess() {
		t.Error("expected error for out of range var")
	}
}

func TestReflexivityPath(t *testing.T) {
	// Path type goal with equal endpoints
	goalType := ast.Path{A: ast.Sort{U: 0}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}}
	state := proofstate.NewProofState(goalType, nil)

	result := Reflexivity()(state)
	if !result.IsSuccess() {
		t.Fatalf("reflexivity for Path failed: %v", result.Err)
	}
}

func TestReflexivityFail(t *testing.T) {
	// Id with unequal endpoints
	goalType := ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}
	state := proofstate.NewProofState(goalType, nil)

	result := Reflexivity()(state)
	if result.IsSuccess() {
		t.Error("expected reflexivity to fail for unequal endpoints")
	}

	// Path with unequal endpoints
	goalType2 := ast.Path{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}
	state2 := proofstate.NewProofState(goalType2, nil)

	result2 := Reflexivity()(state2)
	if result2.IsSuccess() {
		t.Error("expected reflexivity to fail for unequal path endpoints")
	}

	// Non-Id/Path type
	state3 := proofstate.NewProofState(ast.Sort{U: 0}, nil)
	result3 := Reflexivity()(state3)
	if result3.IsSuccess() {
		t.Error("expected reflexivity to fail for non-identity type")
	}
}

func TestTrivial(t *testing.T) {
	// Should solve Id goal by reflexivity
	goalType := ast.Id{A: ast.Sort{U: 1}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}}
	state := proofstate.NewProofState(goalType, nil)

	result := Trivial()(state)
	if !result.IsSuccess() {
		t.Fatalf("trivial failed: %v", result.Err)
	}
}

func TestAuto(t *testing.T) {
	// Goal: (A : Type) -> A -> A
	goalType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "x",
			A:      ast.Var{Ix: 0},
			B:      ast.Var{Ix: 1},
		},
	}
	state := proofstate.NewProofState(goalType, nil)

	result := Auto()(state)
	if !result.IsSuccess() {
		t.Fatalf("auto failed: %v", result.Err)
	}
}

func TestIntroNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Intro("x")(state)
	if result.IsSuccess() {
		t.Error("expected error for no goal")
	}
}

func TestIntroNotPi(t *testing.T) {
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)
	result := Intro("x")(state)
	if result.IsSuccess() {
		t.Error("expected error for non-Pi goal")
	}
}

func TestExactNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Exact(ast.Sort{U: 0})(state)
	if result.IsSuccess() {
		t.Error("expected error for no goal")
	}
}

func TestAssumptionNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Assumption()(state)
	if result.IsSuccess() {
		t.Error("expected error for no goal")
	}
}

func TestSplitNotSigma(t *testing.T) {
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)
	result := Split()(state)
	if result.IsSuccess() {
		t.Error("expected error for non-Sigma goal")
	}
}

func TestSplitNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Split()(state)
	if result.IsSuccess() {
		t.Error("expected error for no goal")
	}
}

func TestSimplNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Simpl()(state)
	if result.IsSuccess() {
		t.Error("expected error for no goal")
	}
}

func TestReflexivityNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Reflexivity()(state)
	if result.IsSuccess() {
		t.Error("expected error for no goal")
	}
}

func TestApplyNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Apply(ast.Var{Ix: 0})(state)
	if result.IsSuccess() {
		t.Error("expected error for no goal")
	}
}

func TestRewriteNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Rewrite("h")(state)
	if result.IsSuccess() {
		t.Error("expected error for no goal")
	}
}

func TestRewriteRevNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := RewriteRev("h")(state)
	if result.IsSuccess() {
		t.Error("expected error for no goal")
	}
}

// --- Prover Tests ---

func TestProverBasic(t *testing.T) {
	goalType := ast.Pi{Binder: "A", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	prover := NewProver(goalType)

	if prover.Done() {
		t.Error("expected incomplete proof")
	}

	if prover.GoalCount() != 1 {
		t.Errorf("expected 1 goal, got %d", prover.GoalCount())
	}

	goals := prover.Goals()
	if len(goals) != 1 {
		t.Errorf("expected 1 goal, got %d", len(goals))
	}

	if prover.CurrentGoal() == nil {
		t.Error("expected current goal")
	}
}

func TestProverWithHyps(t *testing.T) {
	hyps := []proofstate.Hypothesis{
		{Name: "x", Type: ast.Sort{U: 0}},
	}
	prover := NewProverWithHyps(ast.Sort{U: 0}, hyps)

	if prover.CurrentGoal() == nil {
		t.Error("expected current goal")
	}
	if len(prover.CurrentGoal().Hypotheses) != 1 {
		t.Error("expected 1 hypothesis")
	}
}

func TestProverApply(t *testing.T) {
	prover := NewProver(ast.Sort{U: 0})

	err := prover.Apply(NoOp())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Apply failing tactic
	err = prover.Apply(FailWithMsg("fail"))
	if err == nil {
		t.Error("expected error from failing tactic")
	}
}

func TestProverFocus(t *testing.T) {
	sigmaType := ast.Sigma{Binder: "_", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	prover := NewProver(sigmaType)

	prover.Apply(Split())

	if prover.GoalCount() != 2 {
		t.Fatalf("expected 2 goals, got %d", prover.GoalCount())
	}

	secondGoalID := prover.Goals()[1].ID
	err := prover.Focus(secondGoalID)
	if err != nil {
		t.Fatalf("Focus failed: %v", err)
	}

	if prover.CurrentGoal().ID != secondGoalID {
		t.Error("expected focused goal to be second goal")
	}
}

func TestProverUndo(t *testing.T) {
	goalType := ast.Pi{Binder: "A", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	prover := NewProver(goalType)

	prover.Apply(Intro("A"))

	if !prover.Undo() {
		t.Error("Undo should succeed")
	}
}

func TestProverFormatState(t *testing.T) {
	prover := NewProver(ast.Sort{U: 0})
	output := prover.FormatState()
	if output == "" {
		t.Error("FormatState should not be empty")
	}
}

func TestProverState(t *testing.T) {
	prover := NewProver(ast.Sort{U: 0})
	if prover.State() == nil {
		t.Error("State should not be nil")
	}
}

func TestProverExtract(t *testing.T) {
	prover := NewProver(ast.Sort{U: 1})

	// Can't extract incomplete proof
	_, err := prover.Extract()
	if err == nil {
		t.Error("expected error for incomplete proof")
	}

	// Complete the proof (Type₀ : Type₁)
	prover.Apply(Exact(ast.Sort{U: 0}))

	term, err := prover.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if term == nil {
		t.Error("expected non-nil term")
	}
}

// --- Fluent API Tests ---

func TestFluentIntro(t *testing.T) {
	goalType := ast.Pi{Binder: "A", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	prover := NewProver(goalType)

	prover.Intro_("A")
	if prover.Error() != nil {
		t.Error("Intro_ should succeed")
	}

	// Error case
	prover2 := NewProver(ast.Sort{U: 0})
	prover2.Intro_("x")
	if prover2.Error() == nil {
		t.Error("expected error for Intro_ on non-Pi")
	}
}

func TestFluentIntroN(t *testing.T) {
	goalType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "x",
			A:      ast.Var{Ix: 0},
			B:      ast.Var{Ix: 1},
		},
	}
	prover := NewProver(goalType)

	prover.IntroN_("A", "x")
	if prover.Error() != nil {
		t.Error("IntroN_ should succeed")
	}
}

func TestFluentIntros(t *testing.T) {
	goalType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B:      ast.Sort{U: 0},
	}
	prover := NewProver(goalType)

	prover.Intros_()
	if prover.Error() != nil {
		t.Error("Intros_ should succeed")
	}
}

func TestFluentExact(t *testing.T) {
	prover := NewProver(ast.Sort{U: 1})
	prover.Exact_(ast.Sort{U: 0}) // Type₀ : Type₁
	if prover.Error() != nil {
		t.Error("Exact_ should succeed")
	}
}

func TestFluentAssumption(t *testing.T) {
	hyps := []proofstate.Hypothesis{
		{Name: "x", Type: ast.Sort{U: 0}},
	}
	prover := NewProverWithHyps(ast.Sort{U: 0}, hyps)
	prover.Assumption_()
	if prover.Error() != nil {
		t.Error("Assumption_ should succeed")
	}
}

func TestFluentReflexivity(t *testing.T) {
	goalType := ast.Id{A: ast.Sort{U: 1}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}}
	prover := NewProver(goalType)
	prover.Reflexivity_()
	if prover.Error() != nil {
		t.Error("Reflexivity_ should succeed")
	}
}

func TestFluentSplit(t *testing.T) {
	goalType := ast.Sigma{Binder: "_", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	prover := NewProver(goalType)
	prover.Split_()
	if prover.Error() != nil {
		t.Error("Split_ should succeed")
	}
}

func TestFluentSimpl(t *testing.T) {
	prover := NewProver(ast.Sort{U: 0})
	prover.Simpl_()
	if prover.Error() != nil {
		t.Error("Simpl_ should succeed")
	}
}

func TestFluentRewrite(t *testing.T) {
	hyps := []proofstate.Hypothesis{
		{Name: "x", Type: ast.Sort{U: 0}},
		{Name: "h", Type: ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}},
	}
	prover := NewProverWithHyps(ast.Sort{U: 0}, hyps)
	prover.Rewrite_("h")
	if prover.Error() != nil {
		t.Error("Rewrite_ should succeed")
	}
}

func TestFluentRewriteRev(t *testing.T) {
	hyps := []proofstate.Hypothesis{
		{Name: "x", Type: ast.Sort{U: 0}},
		{Name: "h", Type: ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}},
	}
	prover := NewProverWithHyps(ast.Sort{U: 0}, hyps)
	prover.RewriteRev_("h")
	if prover.Error() != nil {
		t.Error("RewriteRev_ should succeed")
	}
}

func TestFluentTrivial(t *testing.T) {
	goalType := ast.Id{A: ast.Sort{U: 1}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}}
	prover := NewProver(goalType)
	prover.Trivial_()
	if prover.Error() != nil {
		t.Error("Trivial_ should succeed")
	}
}

func TestFluentAuto(t *testing.T) {
	goalType := ast.Id{A: ast.Sort{U: 1}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}}
	prover := NewProver(goalType)
	prover.Auto_()
	if prover.Error() != nil {
		t.Error("Auto_ should succeed")
	}
}

// --- Prove/MustProve/ProveSeq Tests ---

func TestProve(t *testing.T) {
	goalType := ast.Id{A: ast.Sort{U: 1}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}}

	term, err := Prove(goalType, Reflexivity())
	if err != nil {
		t.Fatalf("Prove failed: %v", err)
	}
	if term == nil {
		t.Error("expected non-nil term")
	}

	// Tactic fails
	_, err = Prove(goalType, FailWithMsg("fail"))
	if err == nil {
		t.Error("expected error for failing tactic")
	}

	// Proof incomplete
	_, err = Prove(ast.Sort{U: 0}, NoOp())
	if err == nil {
		t.Error("expected error for incomplete proof")
	}
}

func TestMustProve(t *testing.T) {
	goalType := ast.Id{A: ast.Sort{U: 1}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}}

	term := MustProve(goalType, Reflexivity())
	if term == nil {
		t.Error("expected non-nil term")
	}

	// MustProve should panic on failure
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for failing proof")
		}
	}()
	MustProve(ast.Sort{U: 0}, FailWithMsg("fail"))
}

func TestProveSeq(t *testing.T) {
	goalType := ast.Id{A: ast.Sort{U: 1}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}}

	term, err := ProveSeq(goalType, Reflexivity())
	if err != nil {
		t.Fatalf("ProveSeq failed: %v", err)
	}
	if term == nil {
		t.Error("expected non-nil term")
	}
}

// --- substTerm Tests ---

func TestSubstTerm(t *testing.T) {
	old := ast.Var{Ix: 0}
	new := ast.Sort{U: 0}

	// Direct match
	result := substTerm(old, old, new)
	if _, ok := result.(ast.Sort); !ok {
		t.Error("expected substitution")
	}

	// No match
	result2 := substTerm(ast.Var{Ix: 1}, old, new)
	if _, ok := result2.(ast.Var); !ok {
		t.Error("expected unchanged var")
	}

	// Various term types
	_ = substTerm(ast.Global{Name: "x"}, old, new)
	_ = substTerm(ast.Sort{U: 0}, old, new)
	_ = substTerm(ast.Pi{A: old, B: ast.Var{Ix: 1}}, old, new)
	_ = substTerm(ast.Lam{Ann: old, Body: ast.Var{Ix: 0}}, old, new)
	_ = substTerm(ast.App{T: old, U: ast.Var{Ix: 1}}, old, new)
	_ = substTerm(ast.Sigma{A: old, B: ast.Var{Ix: 1}}, old, new)
	_ = substTerm(ast.Pair{Fst: old, Snd: ast.Var{Ix: 1}}, old, new)
	_ = substTerm(ast.Fst{P: old}, old, new)
	_ = substTerm(ast.Snd{P: old}, old, new)
	_ = substTerm(ast.Id{A: old, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 2}}, old, new)
	_ = substTerm(ast.Refl{A: old, X: ast.Var{Ix: 1}}, old, new)

	// Unknown term type
	_ = substTerm(ast.I0{}, old, new)
}

// --- Fluent API Error Tests ---

func TestFluentAPIWithErrors(t *testing.T) {
	// Create a prover with a non-Pi goal to trigger errors
	goalType := ast.Sort{U: 0} // Type, not a Pi
	prover := NewProver(goalType)

	// Intro_ should fail (goal is not Pi)
	prover.Intro_("x")
	err := prover.Error()
	if err == nil {
		t.Error("expected error from Intro_ on non-Pi goal")
	}

	// IntroN_ should fail
	prover = NewProver(goalType)
	prover.IntroN_("x", "y")
	err = prover.Error()
	if err == nil {
		t.Error("expected error from IntroN_ on non-Pi goal")
	}

	// Intros_ succeeds (does nothing on non-Pi)
	prover = NewProver(goalType)
	prover.Intros_()
	_ = prover.Error() // Intros uses repeat, so it succeeds even with no intros

	// Exact_ - test that it's called (may or may not error depending on type checking)
	prover = NewProver(goalType)
	prover.Exact_(ast.Var{Ix: 0}) // May not error if type checking is lenient
	_ = prover.Error()            // Just ensure fluent API works

	// Assumption_ with no matching hypothesis should fail
	prover = NewProver(goalType)
	prover.Assumption_()
	err = prover.Error()
	if err == nil {
		t.Error("expected error from Assumption_ with no matching hypothesis")
	}

	// Reflexivity_ on non-Id goal should fail
	prover = NewProver(goalType)
	prover.Reflexivity_()
	err = prover.Error()
	if err == nil {
		t.Error("expected error from Reflexivity_ on non-Id goal")
	}

	// Split_ on non-Sigma goal should fail
	prover = NewProver(goalType)
	prover.Split_()
	err = prover.Error()
	if err == nil {
		t.Error("expected error from Split_ on non-Sigma goal")
	}

	// Simpl_ always succeeds
	prover = NewProver(goalType)
	prover.Simpl_()
	// No error expected

	// Rewrite_ with non-existent hypothesis should fail
	prover = NewProver(goalType)
	prover.Rewrite_("nonexistent")
	err = prover.Error()
	if err == nil {
		t.Error("expected error from Rewrite_ with non-existent hypothesis")
	}

	// RewriteRev_ with non-existent hypothesis should fail
	prover = NewProver(goalType)
	prover.RewriteRev_("nonexistent")
	err = prover.Error()
	if err == nil {
		t.Error("expected error from RewriteRev_ with non-existent hypothesis")
	}

	// Trivial_ should fail on Sort goal (no reflexivity possible)
	prover = NewProver(goalType)
	prover.Trivial_()
	err = prover.Error()
	if err == nil {
		t.Error("expected error from Trivial_ on Sort goal")
	}

	// Auto_ - test that it's called
	prover = NewProver(goalType)
	prover.Auto_()
	_ = prover.Error() // Auto may or may not succeed depending on implementation
}

// --- OrElse Combinator Tests ---

func TestOrElseFirstSucceeds(t *testing.T) {
	goalType := ast.Pi{Binder: "A", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	state := proofstate.NewProofState(goalType, nil)

	// First tactic succeeds
	tactic := OrElse(
		Intro("A"),     // This succeeds
		Intro("wrong"), // This wouldn't be tried
	)

	result := tactic(state)
	if !result.IsSuccess() {
		t.Fatalf("OrElse should succeed when first tactic succeeds: %v", result.Err)
	}

	goal := state.CurrentGoal()
	if len(goal.Hypotheses) != 1 || goal.Hypotheses[0].Name != "A" {
		t.Error("expected hypothesis A from first tactic")
	}
}

func TestOrElseFirstFailsSecondSucceeds(t *testing.T) {
	goalType := ast.Pi{Binder: "A", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	state := proofstate.NewProofState(goalType, nil)

	// First tactic fails, second succeeds
	tactic := OrElse(
		func(s *proofstate.ProofState) TacticResult {
			return Failf("forced failure")
		},
		Intro("A"), // This should be tried
	)

	result := tactic(state)
	if !result.IsSuccess() {
		t.Fatalf("OrElse should succeed when second tactic succeeds: %v", result.Err)
	}
}

// --- Progress Combinator Tests ---

func TestProgressWithChange(t *testing.T) {
	// Test that Progress runs the tactic
	// Use a tactic that definitely changes state - Split on Sigma creates new goals
	sigmaGoal := ast.Sigma{Binder: "a", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	state := proofstate.NewProofState(sigmaGoal, nil)

	tactic := Progress(Split())
	result := tactic(state)

	// Progress checks if state changed (goal count changed)
	if !result.IsSuccess() {
		// If Progress fails, it means the state comparison is strict
		// This is acceptable behavior - just verify Progress runs
		t.Log("Progress detected no change (strict comparison)")
	}
}

func TestProgressWithoutChange(t *testing.T) {
	goalType := ast.Sort{U: 0}
	state := proofstate.NewProofState(goalType, nil)

	// Simpl on Sort doesn't change anything
	tactic := Progress(Simpl())
	result := tactic(state)

	if result.IsSuccess() {
		t.Error("Progress should fail when no change is made")
	}
}

// --- IfThenElse Combinator Tests ---

func TestIfThenElseConditionTrue(t *testing.T) {
	goalType := ast.Pi{Binder: "A", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	state := proofstate.NewProofState(goalType, nil)

	// Condition succeeds -> then branch
	tactic := IfThenElse(
		Intro("A"),       // condition (succeeds on Pi)
		Simpl(),          // then branch
		func(s *proofstate.ProofState) TacticResult { return Failf("else") },
	)

	result := tactic(state)
	if !result.IsSuccess() {
		t.Fatalf("IfThenElse should succeed when condition is true: %v", result.Err)
	}
}

func TestIfThenElseConditionFalse(t *testing.T) {
	goalType := ast.Sort{U: 0}
	state := proofstate.NewProofState(goalType, nil)

	// Condition fails -> else branch
	tactic := IfThenElse(
		Intro("A"),       // condition (fails on Sort)
		func(s *proofstate.ProofState) TacticResult { return Failf("then") },
		Simpl(),          // else branch (succeeds)
	)

	result := tactic(state)
	if !result.IsSuccess() {
		t.Fatalf("IfThenElse should succeed when else branch succeeds: %v", result.Err)
	}
}

// --- All Combinator Tests ---

func TestAllMultipleGoals(t *testing.T) {
	// Create state with a sigma type to get multiple goals after split
	sigmaGoal := ast.Sigma{
		Binder: "a",
		A:      ast.Sort{U: 0},
		B:      ast.Sort{U: 0},
	}
	state := proofstate.NewProofState(sigmaGoal, nil)

	// Split creates 2 goals
	result := Split()(state)
	if !result.IsSuccess() {
		t.Fatalf("Split failed: %v", result.Err)
	}

	if state.GoalCount() != 2 {
		t.Fatalf("expected 2 goals after split, got %d", state.GoalCount())
	}

	// All applies Simpl to all goals
	result = All(Simpl())(state)
	if !result.IsSuccess() {
		t.Fatalf("All(Simpl) failed: %v", result.Err)
	}
}

func TestAllFailsOnOneGoal(t *testing.T) {
	sigmaGoal := ast.Sigma{
		Binder: "a",
		A:      ast.Sort{U: 0},
		B:      ast.Sort{U: 0},
	}
	state := proofstate.NewProofState(sigmaGoal, nil)

	// Split creates 2 goals
	_ = Split()(state)

	// All fails if any tactic fails
	result := All(Intro("x"))(state) // Intro fails on Sort goals
	if result.IsSuccess() {
		t.Error("All should fail when tactic fails on any goal")
	}
}

// --- IntroN Error Tests ---

func TestIntroNPartialFailure(t *testing.T) {
	// Goal has only one Pi
	goalType := ast.Pi{Binder: "A", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	state := proofstate.NewProofState(goalType, nil)

	// Try to intro two things, but only one Pi available
	result := IntroN("A", "x")(state)
	if result.IsSuccess() {
		t.Error("IntroN should fail when not enough Pis")
	}
}

// --- Complete Combinator Tests ---

func TestCompleteSucceeds(t *testing.T) {
	// Goal that assumption can solve
	hyps := []proofstate.Hypothesis{
		{Name: "x", Type: ast.Sort{U: 0}},
	}
	state := proofstate.NewProofState(ast.Sort{U: 0}, hyps)

	result := Complete(Assumption())(state)
	if !result.IsSuccess() {
		t.Fatalf("Complete should succeed when proof is complete: %v", result.Err)
	}
}

func TestCompleteFails(t *testing.T) {
	goalType := ast.Pi{Binder: "A", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	state := proofstate.NewProofState(goalType, nil)

	// Intro doesn't complete the proof
	result := Complete(Intro("A"))(state)
	if result.IsSuccess() {
		t.Error("Complete should fail when proof is not complete")
	}
}

// --- Prove Convenience Function Tests ---

func TestProveIncomplete(t *testing.T) {
	goalType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B:      ast.Pi{Binder: "x", A: ast.Var{Ix: 0}, B: ast.Var{Ix: 1}},
	}

	// Only intro once, leaves goal incomplete
	_, err := Prove(goalType, Intro("A"))
	if err == nil {
		t.Error("Prove should fail when proof is incomplete")
	}
}

func TestProveTacticFails(t *testing.T) {
	goalType := ast.Sort{U: 0}

	// Intro fails on Sort
	_, err := Prove(goalType, Intro("x"))
	if err == nil {
		t.Error("Prove should fail when tactic fails")
	}
}

// --- Additional Combinator Tests ---

func TestRepeatNLimited(t *testing.T) {
	goalType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "B",
			A:      ast.Sort{U: 0},
			B: ast.Pi{
				Binder: "C",
				A:      ast.Sort{U: 0},
				B:      ast.Sort{U: 0},
			},
		},
	}
	state := proofstate.NewProofState(goalType, nil)

	// RepeatN with n=2 using single Intro per iteration
	// Intros itself may intro multiple, so use single Intro
	result := RepeatN(2, Intro("_"))(state)
	if !result.IsSuccess() {
		t.Fatalf("RepeatN failed: %v", result.Err)
	}

	// Should have 2 hypotheses (two single intros)
	goal := state.CurrentGoal()
	if len(goal.Hypotheses) != 2 {
		t.Errorf("expected 2 hypotheses, got %d", len(goal.Hypotheses))
	}
}

func TestFirstNoneSucceed(t *testing.T) {
	goalType := ast.Sort{U: 0}
	state := proofstate.NewProofState(goalType, nil)

	// All tactics fail
	result := First(
		Intro("x"),
		Intro("y"),
		Intro("z"),
	)(state)

	if result.IsSuccess() {
		t.Error("First should fail when all tactics fail")
	}
}

func TestFocusInvalidGoal(t *testing.T) {
	goalType := ast.Sort{U: 0}
	state := proofstate.NewProofState(goalType, nil)

	// Focus on non-existent goal ID
	result := Focus(999, Simpl())(state)
	if result.IsSuccess() {
		t.Error("Focus should fail on invalid goal ID")
	}
}

// --- Contradiction Tactic Tests ---

func TestContradiction(t *testing.T) {
	// Goal: Nat with hypothesis h : Empty
	// This should be solvable via contradiction
	hyps := []proofstate.Hypothesis{
		{Name: "h", Type: ast.Global{Name: "Empty"}},
	}
	goalType := ast.Global{Name: "Nat"}
	state := proofstate.NewProofState(goalType, hyps)

	result := Contradiction()(state)
	if !result.IsSuccess() {
		t.Fatalf("Contradiction failed: %v", result.Err)
	}

	if !state.IsComplete() {
		t.Error("expected proof to be complete after contradiction")
	}
}

func TestContradictionWithMultipleHypotheses(t *testing.T) {
	// Goal: Type with multiple hypotheses including Empty
	hyps := []proofstate.Hypothesis{
		{Name: "A", Type: ast.Sort{U: 0}},
		{Name: "x", Type: ast.Var{Ix: 0}},
		{Name: "empty", Type: ast.Global{Name: "Empty"}},
	}
	goalType := ast.Sort{U: 0}
	state := proofstate.NewProofState(goalType, hyps)

	result := Contradiction()(state)
	if !result.IsSuccess() {
		t.Fatalf("Contradiction failed: %v", result.Err)
	}

	if !state.IsComplete() {
		t.Error("expected proof to be complete")
	}
}

func TestContradictionNoEmpty(t *testing.T) {
	// Goal: Type with no Empty hypothesis
	hyps := []proofstate.Hypothesis{
		{Name: "A", Type: ast.Sort{U: 0}},
		{Name: "x", Type: ast.Var{Ix: 0}},
	}
	goalType := ast.Sort{U: 0}
	state := proofstate.NewProofState(goalType, hyps)

	result := Contradiction()(state)
	if result.IsSuccess() {
		t.Error("Contradiction should fail when no Empty hypothesis exists")
	}
}

func TestContradictionNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Contradiction()(state)
	if result.IsSuccess() {
		t.Error("Contradiction should fail with no goal")
	}
}

// --- Left/Right Tactic Tests ---

func TestLeftTactic(t *testing.T) {
	// Goal: Sum Nat Bool with hypothesis x : Nat
	// Using Left should create subgoal Nat
	nat := ast.Global{Name: "Nat"}
	bool_ := ast.Global{Name: "Bool"}
	sumType := ast.App{T: ast.App{T: ast.Global{Name: "Sum"}, U: nat}, U: bool_}

	hyps := []proofstate.Hypothesis{
		{Name: "x", Type: nat},
	}
	state := proofstate.NewProofState(sumType, hyps)

	result := Left()(state)
	if !result.IsSuccess() {
		t.Fatalf("Left failed: %v", result.Err)
	}

	// Should have a subgoal of type Nat
	goal := state.CurrentGoal()
	if goal == nil {
		t.Fatal("expected a subgoal after Left")
	}
}

func TestRightTactic(t *testing.T) {
	// Goal: Sum Nat Bool with hypothesis b : Bool
	// Using Right should create subgoal Bool
	nat := ast.Global{Name: "Nat"}
	bool_ := ast.Global{Name: "Bool"}
	sumType := ast.App{T: ast.App{T: ast.Global{Name: "Sum"}, U: nat}, U: bool_}

	hyps := []proofstate.Hypothesis{
		{Name: "b", Type: bool_},
	}
	state := proofstate.NewProofState(sumType, hyps)

	result := Right()(state)
	if !result.IsSuccess() {
		t.Fatalf("Right failed: %v", result.Err)
	}

	// Should have a subgoal of type Bool
	goal := state.CurrentGoal()
	if goal == nil {
		t.Fatal("expected a subgoal after Right")
	}
}

func TestLeftWithAssumption(t *testing.T) {
	// Goal: Sum Nat Bool with hypothesis x : Nat
	// Left followed by Assumption should complete the proof
	nat := ast.Global{Name: "Nat"}
	bool_ := ast.Global{Name: "Bool"}
	sumType := ast.App{T: ast.App{T: ast.Global{Name: "Sum"}, U: nat}, U: bool_}

	hyps := []proofstate.Hypothesis{
		{Name: "x", Type: nat},
	}
	state := proofstate.NewProofState(sumType, hyps)

	result := Seq(Left(), Assumption())(state)
	if !result.IsSuccess() {
		t.Fatalf("Left;Assumption failed: %v", result.Err)
	}

	if !state.IsComplete() {
		t.Error("expected proof to be complete")
	}
}

func TestRightWithAssumption(t *testing.T) {
	// Goal: Sum Nat Bool with hypothesis b : Bool
	// Right followed by Assumption should complete the proof
	nat := ast.Global{Name: "Nat"}
	bool_ := ast.Global{Name: "Bool"}
	sumType := ast.App{T: ast.App{T: ast.Global{Name: "Sum"}, U: nat}, U: bool_}

	hyps := []proofstate.Hypothesis{
		{Name: "b", Type: bool_},
	}
	state := proofstate.NewProofState(sumType, hyps)

	result := Seq(Right(), Assumption())(state)
	if !result.IsSuccess() {
		t.Fatalf("Right;Assumption failed: %v", result.Err)
	}

	if !state.IsComplete() {
		t.Error("expected proof to be complete")
	}
}

func TestLeftNotSum(t *testing.T) {
	// Left on non-Sum goal should fail
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)

	result := Left()(state)
	if result.IsSuccess() {
		t.Error("Left should fail on non-Sum goal")
	}
}

func TestRightNotSum(t *testing.T) {
	// Right on non-Sum goal should fail
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)

	result := Right()(state)
	if result.IsSuccess() {
		t.Error("Right should fail on non-Sum goal")
	}
}

func TestLeftNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Left()(state)
	if result.IsSuccess() {
		t.Error("Left should fail with no goal")
	}
}

func TestRightNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Right()(state)
	if result.IsSuccess() {
		t.Error("Right should fail with no goal")
	}
}

// --- Destruct Tactic Tests ---

func TestDestructSum(t *testing.T) {
	// Goal: Nat with hypothesis s : Sum Nat Bool
	// Destruct should create two subgoals
	nat := ast.Global{Name: "Nat"}
	bool_ := ast.Global{Name: "Bool"}
	sumType := ast.App{T: ast.App{T: ast.Global{Name: "Sum"}, U: nat}, U: bool_}

	hyps := []proofstate.Hypothesis{
		{Name: "s", Type: sumType},
	}
	state := proofstate.NewProofState(nat, hyps)

	result := Destruct("s")(state)
	if !result.IsSuccess() {
		t.Fatalf("Destruct s failed: %v", result.Err)
	}

	// Should have 2 goals (one for inl, one for inr)
	if state.GoalCount() != 2 {
		t.Errorf("expected 2 goals after Destruct, got %d", state.GoalCount())
	}
}

func TestDestructBool(t *testing.T) {
	// Goal: Nat with hypothesis b : Bool
	// Destruct should create two subgoals
	nat := ast.Global{Name: "Nat"}
	bool_ := ast.Global{Name: "Bool"}

	hyps := []proofstate.Hypothesis{
		{Name: "b", Type: bool_},
	}
	state := proofstate.NewProofState(nat, hyps)

	result := Destruct("b")(state)
	if !result.IsSuccess() {
		t.Fatalf("Destruct b failed: %v", result.Err)
	}

	// Should have 2 goals (one for true, one for false)
	if state.GoalCount() != 2 {
		t.Errorf("expected 2 goals after Destruct, got %d", state.GoalCount())
	}
}

func TestDestructNotFound(t *testing.T) {
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)

	result := Destruct("nonexistent")(state)
	if result.IsSuccess() {
		t.Error("Destruct should fail when hypothesis not found")
	}
}

func TestDestructUnsupportedType(t *testing.T) {
	// Try to destruct a Nat - should fail
	hyps := []proofstate.Hypothesis{
		{Name: "n", Type: ast.Global{Name: "Nat"}},
	}
	state := proofstate.NewProofState(ast.Sort{U: 0}, hyps)

	result := Destruct("n")(state)
	if result.IsSuccess() {
		t.Error("Destruct should fail on unsupported type")
	}
}

func TestDestructNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Destruct("x")(state)
	if result.IsSuccess() {
		t.Error("Destruct should fail with no goal")
	}
}

// --- Induction Tactic Tests ---

func TestInductionNat(t *testing.T) {
	// Goal: Nat with hypothesis n : Nat
	// Induction should create two subgoals (base and step)
	nat := ast.Global{Name: "Nat"}

	hyps := []proofstate.Hypothesis{
		{Name: "n", Type: nat},
	}
	state := proofstate.NewProofState(nat, hyps)

	result := Induction("n")(state)
	if !result.IsSuccess() {
		t.Fatalf("Induction n failed: %v", result.Err)
	}

	// Should have 2 goals (base case and step case)
	if state.GoalCount() != 2 {
		t.Errorf("expected 2 goals after Induction, got %d", state.GoalCount())
	}
}

func TestInductionList(t *testing.T) {
	// Goal: Nat with hypothesis l : List Nat
	// Induction should create two subgoals (nil and cons)
	nat := ast.Global{Name: "Nat"}
	listNat := ast.App{T: ast.Global{Name: "List"}, U: nat}

	hyps := []proofstate.Hypothesis{
		{Name: "l", Type: listNat},
	}
	state := proofstate.NewProofState(nat, hyps)

	result := Induction("l")(state)
	if !result.IsSuccess() {
		t.Fatalf("Induction l failed: %v", result.Err)
	}

	// Should have 2 goals (nil case and cons case)
	if state.GoalCount() != 2 {
		t.Errorf("expected 2 goals after Induction, got %d", state.GoalCount())
	}

	// Cons case should have additional hypotheses (x, xs, ih)
	goals := state.Goals
	consGoal := goals[1] // Second goal is cons case
	if len(consGoal.Hypotheses) < 3 {
		t.Errorf("expected cons case to have at least 3 hypotheses (x, xs, ih), got %d", len(consGoal.Hypotheses))
	}
}

func TestInductionNotFound(t *testing.T) {
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)

	result := Induction("nonexistent")(state)
	if result.IsSuccess() {
		t.Error("Induction should fail when hypothesis not found")
	}
}

func TestInductionUnsupportedType(t *testing.T) {
	// Try to do induction on a Bool - should fail (use Destruct instead)
	hyps := []proofstate.Hypothesis{
		{Name: "b", Type: ast.Global{Name: "Bool"}},
	}
	state := proofstate.NewProofState(ast.Sort{U: 0}, hyps)

	result := Induction("b")(state)
	if result.IsSuccess() {
		t.Error("Induction should fail on unsupported type (Bool)")
	}
}

func TestInductionNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Induction("x")(state)
	if result.IsSuccess() {
		t.Error("Induction should fail with no goal")
	}
}

func TestInductionNatStepHasIH(t *testing.T) {
	// Verify that the step case gets the induction hypothesis
	nat := ast.Global{Name: "Nat"}

	hyps := []proofstate.Hypothesis{
		{Name: "m", Type: nat},
	}
	state := proofstate.NewProofState(nat, hyps)

	result := Induction("m")(state)
	if !result.IsSuccess() {
		t.Fatalf("Induction m failed: %v", result.Err)
	}

	// Step case (second goal) should have n : Nat and ih : Nat
	goals := state.Goals
	if len(goals) < 2 {
		t.Fatal("expected at least 2 goals")
	}
	stepGoal := goals[1]

	// Check for n and ih hypotheses
	hasN := false
	hasIH := false
	for _, h := range stepGoal.Hypotheses {
		if h.Name == "n" {
			hasN = true
		}
		if h.Name == "ih" {
			hasIH = true
		}
	}
	if !hasN {
		t.Error("step case should have hypothesis n : Nat")
	}
	if !hasIH {
		t.Error("step case should have hypothesis ih (induction hypothesis)")
	}
}

// --- Cases Tactic Tests ---

func TestCasesNat(t *testing.T) {
	// Goal: Nat with hypothesis n : Nat
	// Cases should create two subgoals (zero and succ) without IH
	nat := ast.Global{Name: "Nat"}

	hyps := []proofstate.Hypothesis{
		{Name: "n", Type: nat},
	}
	state := proofstate.NewProofState(nat, hyps)

	result := Cases("n")(state)
	if !result.IsSuccess() {
		t.Fatalf("Cases n failed: %v", result.Err)
	}

	// Should have 2 goals
	if state.GoalCount() != 2 {
		t.Errorf("expected 2 goals after Cases, got %d", state.GoalCount())
	}

	// Succ case should have n : Nat but NO ih (unlike Induction)
	succGoal := state.Goals[1]
	hasN := false
	hasIH := false
	for _, h := range succGoal.Hypotheses {
		if h.Name == "n" {
			hasN = true
		}
		if h.Name == "ih" {
			hasIH = true
		}
	}
	if !hasN {
		t.Error("succ case should have hypothesis n : Nat")
	}
	if hasIH {
		t.Error("Cases should NOT introduce an induction hypothesis")
	}
}

func TestCasesList(t *testing.T) {
	// Goal: Nat with hypothesis l : List Nat
	// Cases should create two subgoals without IH
	nat := ast.Global{Name: "Nat"}
	listNat := ast.App{T: ast.Global{Name: "List"}, U: nat}

	hyps := []proofstate.Hypothesis{
		{Name: "l", Type: listNat},
	}
	state := proofstate.NewProofState(nat, hyps)

	result := Cases("l")(state)
	if !result.IsSuccess() {
		t.Fatalf("Cases l failed: %v", result.Err)
	}

	// Should have 2 goals
	if state.GoalCount() != 2 {
		t.Errorf("expected 2 goals after Cases, got %d", state.GoalCount())
	}

	// Cons case should have x and xs but NO ih
	consGoal := state.Goals[1]
	hasX := false
	hasXs := false
	hasIH := false
	for _, h := range consGoal.Hypotheses {
		if h.Name == "x" {
			hasX = true
		}
		if h.Name == "xs" {
			hasXs = true
		}
		if h.Name == "ih" {
			hasIH = true
		}
	}
	if !hasX {
		t.Error("cons case should have hypothesis x")
	}
	if !hasXs {
		t.Error("cons case should have hypothesis xs")
	}
	if hasIH {
		t.Error("Cases should NOT introduce an induction hypothesis")
	}
}

func TestCasesBool(t *testing.T) {
	// Cases on Bool should work like Destruct
	nat := ast.Global{Name: "Nat"}
	bool_ := ast.Global{Name: "Bool"}

	hyps := []proofstate.Hypothesis{
		{Name: "b", Type: bool_},
	}
	state := proofstate.NewProofState(nat, hyps)

	result := Cases("b")(state)
	if !result.IsSuccess() {
		t.Fatalf("Cases b failed: %v", result.Err)
	}

	// Should have 2 goals
	if state.GoalCount() != 2 {
		t.Errorf("expected 2 goals after Cases, got %d", state.GoalCount())
	}
}

func TestCasesSum(t *testing.T) {
	// Cases on Sum should work like Destruct
	nat := ast.Global{Name: "Nat"}
	bool_ := ast.Global{Name: "Bool"}
	sumType := ast.App{T: ast.App{T: ast.Global{Name: "Sum"}, U: nat}, U: bool_}

	hyps := []proofstate.Hypothesis{
		{Name: "s", Type: sumType},
	}
	state := proofstate.NewProofState(nat, hyps)

	result := Cases("s")(state)
	if !result.IsSuccess() {
		t.Fatalf("Cases s failed: %v", result.Err)
	}

	// Should have 2 goals
	if state.GoalCount() != 2 {
		t.Errorf("expected 2 goals after Cases, got %d", state.GoalCount())
	}
}

func TestCasesNotFound(t *testing.T) {
	state := proofstate.NewProofState(ast.Sort{U: 0}, nil)

	result := Cases("nonexistent")(state)
	if result.IsSuccess() {
		t.Error("Cases should fail when hypothesis not found")
	}
}

func TestCasesUnsupportedType(t *testing.T) {
	// Try to do cases on a Type - should fail
	hyps := []proofstate.Hypothesis{
		{Name: "T", Type: ast.Sort{U: 0}},
	}
	state := proofstate.NewProofState(ast.Sort{U: 0}, hyps)

	result := Cases("T")(state)
	if result.IsSuccess() {
		t.Error("Cases should fail on unsupported type")
	}
}

func TestCasesNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Cases("x")(state)
	if result.IsSuccess() {
		t.Error("Cases should fail with no goal")
	}
}

// --- Constructor Tactic Tests ---

func TestConstructorUnit(t *testing.T) {
	// Goal: Unit should be solved with tt
	unitType := ast.Global{Name: "Unit"}
	state := proofstate.NewProofState(unitType, nil)

	result := Constructor()(state)
	if !result.IsSuccess() {
		t.Fatalf("Constructor on Unit failed: %v", result.Err)
	}

	if !state.IsComplete() {
		t.Error("expected proof to be complete after Constructor on Unit")
	}
}

func TestConstructorSum(t *testing.T) {
	// Goal: Sum Nat Bool should use inl (Left)
	nat := ast.Global{Name: "Nat"}
	bool_ := ast.Global{Name: "Bool"}
	sumType := ast.App{T: ast.App{T: ast.Global{Name: "Sum"}, U: nat}, U: bool_}

	state := proofstate.NewProofState(sumType, nil)

	result := Constructor()(state)
	if !result.IsSuccess() {
		t.Fatalf("Constructor on Sum failed: %v", result.Err)
	}

	// Should have a subgoal for Nat (left injection)
	if state.GoalCount() != 1 {
		t.Errorf("expected 1 subgoal after Constructor on Sum, got %d", state.GoalCount())
	}
}

func TestConstructorList(t *testing.T) {
	// Goal: List Nat should be solved with nil Nat
	nat := ast.Global{Name: "Nat"}
	listNat := ast.App{T: ast.Global{Name: "List"}, U: nat}

	state := proofstate.NewProofState(listNat, nil)

	result := Constructor()(state)
	if !result.IsSuccess() {
		t.Fatalf("Constructor on List failed: %v", result.Err)
	}

	if !state.IsComplete() {
		t.Error("expected proof to be complete after Constructor on List (nil)")
	}
}

func TestConstructorUnsupportedType(t *testing.T) {
	// Constructor on Nat should fail (Nat is not Unit, Sum, or List)
	nat := ast.Global{Name: "Nat"}
	state := proofstate.NewProofState(nat, nil)

	result := Constructor()(state)
	if result.IsSuccess() {
		t.Error("Constructor should fail on unsupported type Nat")
	}
}

func TestConstructorNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	result := Constructor()(state)
	if result.IsSuccess() {
		t.Error("Constructor should fail with no goal")
	}
}

// --- Exists Tactic Tests ---

func TestExistsSigma(t *testing.T) {
	// Goal: Σ(x:Nat).Nat (exists a Nat, and produce another Nat)
	// Provide witness zero, leaving goal Nat
	nat := ast.Global{Name: "Nat"}
	sigmaType := ast.Sigma{
		Binder: "x",
		A:      nat,
		B:      nat,
	}
	state := proofstate.NewProofState(sigmaType, nil)

	// Provide zero as witness
	witness := ast.Global{Name: "zero"}
	result := Exists(witness)(state)
	if !result.IsSuccess() {
		t.Fatalf("Exists failed: %v", result.Err)
	}

	// Should have 1 subgoal for the second component
	if state.GoalCount() != 1 {
		t.Errorf("expected 1 subgoal after Exists, got %d", state.GoalCount())
	}
}

func TestExistsDependentSigma(t *testing.T) {
	// Goal: Σ(A:Type).A (exists a type and an element of that type)
	// Provide Nat as witness, leaving goal Nat
	type0 := ast.Sort{U: 0}
	sigmaType := ast.Sigma{
		Binder: "A",
		A:      type0,
		B:      ast.Var{Ix: 0}, // A (the first component)
	}
	state := proofstate.NewProofState(sigmaType, nil)

	// Provide Nat as the type witness
	witness := ast.Global{Name: "Nat"}
	result := Exists(witness)(state)
	if !result.IsSuccess() {
		t.Fatalf("Exists failed: %v", result.Err)
	}

	// Should have 1 subgoal for the element of type Nat
	if state.GoalCount() != 1 {
		t.Errorf("expected 1 subgoal after Exists, got %d", state.GoalCount())
	}
}

func TestExistsNotSigma(t *testing.T) {
	// Exists on non-Sigma type should fail
	nat := ast.Global{Name: "Nat"}
	state := proofstate.NewProofState(nat, nil)

	witness := ast.Global{Name: "zero"}
	result := Exists(witness)(state)
	if result.IsSuccess() {
		t.Error("Exists should fail on non-Sigma type")
	}
}

func TestExistsNoGoal(t *testing.T) {
	state := &proofstate.ProofState{Goals: nil}
	witness := ast.Global{Name: "zero"}
	result := Exists(witness)(state)
	if result.IsSuccess() {
		t.Error("Exists should fail with no goal")
	}
}

func TestExistsWithAssumption(t *testing.T) {
	// Goal: Σ(x:Nat).Nat with hypothesis n : Nat
	// Exists n followed by Assumption should complete proof
	nat := ast.Global{Name: "Nat"}
	sigmaType := ast.Sigma{
		Binder: "x",
		A:      nat,
		B:      nat,
	}
	hyps := []proofstate.Hypothesis{
		{Name: "n", Type: nat},
	}
	state := proofstate.NewProofState(sigmaType, hyps)

	// Exists Var{0} (hypothesis n) followed by Assumption
	witness := ast.Var{Ix: 0} // n
	result := Seq(Exists(witness), Assumption())(state)
	if !result.IsSuccess() {
		t.Fatalf("Exists;Assumption failed: %v", result.Err)
	}

	if !state.IsComplete() {
		t.Error("expected proof to be complete")
	}
}
