package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
)

// TestUnitType verifies the Unit type is correctly defined.
func TestUnitType(t *testing.T) {
	checker := NewCheckerWithStdlib()

	// Unit : Type₀
	unitTy, err := checker.Synth(nil, NoSpan(), ast.Global{Name: "Unit"})
	if err != nil {
		t.Fatalf("Unit type synthesis failed: %v", err)
	}
	if _, ok := unitTy.(ast.Sort); !ok {
		t.Errorf("Unit should have type Sort, got %T", unitTy)
	}

	// tt : Unit
	ttTy, err := checker.Synth(nil, NoSpan(), ast.Global{Name: "tt"})
	if err != nil {
		t.Fatalf("tt synthesis failed: %v", err)
	}
	ttGlobal, ok := ttTy.(ast.Global)
	if !ok || ttGlobal.Name != "Unit" {
		t.Errorf("tt should have type Unit, got %v", ttTy)
	}
}

// TestUnitElimType verifies the unitElim eliminator has the correct type.
func TestUnitElimType(t *testing.T) {
	checker := NewCheckerWithStdlib()

	// unitElim should exist and have a Pi type
	elimTy, err := checker.Synth(nil, NoSpan(), ast.Global{Name: "unitElim"})
	if err != nil {
		t.Fatalf("unitElim synthesis failed: %v", err)
	}

	if _, ok := elimTy.(ast.Pi); !ok {
		t.Errorf("unitElim should have Pi type, got %T", elimTy)
	}
}

// TestUnitElimComputation verifies the computation rule:
// unitElim P p tt → p
func TestUnitElimComputation(t *testing.T) {
	// Build: unitElim P p tt
	// where P : Unit → Type, p : P tt
	// Result should be: p

	// We use a simple test: unitElim (λ_. Nat) zero tt should reduce to zero
	term := ast.MkApps(
		ast.Global{Name: "unitElim"},
		// P : Unit → Type = λ_. Nat
		ast.Lam{Binder: "_", Body: ast.Global{Name: "Nat"}},
		// p : P tt = zero
		ast.Global{Name: "zero"},
		// u : Unit = tt
		ast.Global{Name: "tt"},
	)

	result := eval.EvalNBE(term)

	// Should reduce to zero
	if g, ok := result.(ast.Global); !ok || g.Name != "zero" {
		t.Errorf("unitElim P zero tt should reduce to zero, got %v", result)
	}
}

// TestEmptyType verifies the Empty type is correctly defined.
func TestEmptyType(t *testing.T) {
	checker := NewCheckerWithStdlib()

	// Empty : Type₀
	emptyTy, err := checker.Synth(nil, NoSpan(), ast.Global{Name: "Empty"})
	if err != nil {
		t.Fatalf("Empty type synthesis failed: %v", err)
	}
	if _, ok := emptyTy.(ast.Sort); !ok {
		t.Errorf("Empty should have type Sort, got %T", emptyTy)
	}
}

// TestEmptyElimType verifies the emptyElim eliminator has the correct type.
func TestEmptyElimType(t *testing.T) {
	checker := NewCheckerWithStdlib()

	// emptyElim should exist and have a Pi type
	elimTy, err := checker.Synth(nil, NoSpan(), ast.Global{Name: "emptyElim"})
	if err != nil {
		t.Fatalf("emptyElim synthesis failed: %v", err)
	}

	if _, ok := elimTy.(ast.Pi); !ok {
		t.Errorf("emptyElim should have Pi type, got %T", elimTy)
	}
}

// TestEmptyElimStuck verifies that emptyElim is stuck (doesn't reduce)
// when applied to a neutral term.
func TestEmptyElimStuck(t *testing.T) {
	// Build: emptyElim P e where e is a variable (neutral)
	// The eliminator should remain stuck since we can't pattern match on e

	// Use Var{0} as a stand-in for a hypothetical Empty value
	// In practice, we can't construct such a value, but the eliminator
	// should still be well-typed and stuck.

	term := ast.MkApps(
		ast.Global{Name: "emptyElim"},
		// P : Empty → Type = λ_. Nat
		ast.Lam{Binder: "_", Body: ast.Global{Name: "Nat"}},
		// e : Empty = Var{0} (hypothetical)
		ast.Var{Ix: 0},
	)

	result := eval.EvalNBE(term)

	// Should NOT reduce to a simple value - should be stuck as an application
	if _, ok := result.(ast.Global); ok {
		t.Errorf("emptyElim should be stuck on variable, got %v", result)
	}
}

// TestNewCheckerWithStdlib verifies the convenience constructor works.
func TestNewCheckerWithStdlib(t *testing.T) {
	checker := NewCheckerWithStdlib()
	if checker == nil {
		t.Fatal("NewCheckerWithStdlib returned nil")
	}

	// Should have both primitives and stdlib
	globals := checker.Globals()

	// Primitives
	if globals.LookupType("Nat") == nil {
		t.Error("missing Nat primitive")
	}
	if globals.LookupType("Bool") == nil {
		t.Error("missing Bool primitive")
	}

	// Stdlib
	if globals.LookupType("Unit") == nil {
		t.Error("missing Unit type")
	}
	if globals.LookupType("Empty") == nil {
		t.Error("missing Empty type")
	}
	if globals.LookupType("Sum") == nil {
		t.Error("missing Sum type")
	}
}

// TestSumType verifies the Sum type is correctly defined.
func TestSumType(t *testing.T) {
	checker := NewCheckerWithStdlib()

	// Sum : Type → Type → Type
	sumTy, err := checker.Synth(nil, NoSpan(), ast.Global{Name: "Sum"})
	if err != nil {
		t.Fatalf("Sum type synthesis failed: %v", err)
	}
	if _, ok := sumTy.(ast.Pi); !ok {
		t.Errorf("Sum should have Pi type, got %T", sumTy)
	}

	// inl : (A : Type) → (B : Type) → A → Sum A B
	inlTy, err := checker.Synth(nil, NoSpan(), ast.Global{Name: "inl"})
	if err != nil {
		t.Fatalf("inl synthesis failed: %v", err)
	}
	if _, ok := inlTy.(ast.Pi); !ok {
		t.Errorf("inl should have Pi type, got %T", inlTy)
	}

	// inr : (A : Type) → (B : Type) → B → Sum A B
	inrTy, err := checker.Synth(nil, NoSpan(), ast.Global{Name: "inr"})
	if err != nil {
		t.Fatalf("inr synthesis failed: %v", err)
	}
	if _, ok := inrTy.(ast.Pi); !ok {
		t.Errorf("inr should have Pi type, got %T", inrTy)
	}
}

// TestSumElimType verifies the sumElim eliminator has the correct type.
func TestSumElimType(t *testing.T) {
	checker := NewCheckerWithStdlib()

	// sumElim should exist and have a Pi type
	elimTy, err := checker.Synth(nil, NoSpan(), ast.Global{Name: "sumElim"})
	if err != nil {
		t.Fatalf("sumElim synthesis failed: %v", err)
	}

	if _, ok := elimTy.(ast.Pi); !ok {
		t.Errorf("sumElim should have Pi type, got %T", elimTy)
	}
}

// TestSumElimComputationInl verifies the computation rule:
// sumElim A B P f g (inl A B a) → f a
func TestSumElimComputationInl(t *testing.T) {
	// Build: sumElim Nat Bool (λ_. Nat) (λa. a) (λb. zero) (inl Nat Bool zero)
	// Result should be: zero (applying f = id to zero)

	nat := ast.Global{Name: "Nat"}
	bool_ := ast.Global{Name: "Bool"}
	zero := ast.Global{Name: "zero"}

	// Build (inl Nat Bool zero)
	inlTerm := ast.MkApps(
		ast.Global{Name: "inl"},
		nat,   // A
		bool_, // B
		zero,  // a
	)

	// Build sumElim Nat Bool P f g (inl Nat Bool zero)
	term := ast.MkApps(
		ast.Global{Name: "sumElim"},
		nat,   // A
		bool_, // B
		// P : Sum Nat Bool → Type = λ_. Nat
		ast.Lam{Binder: "_", Body: nat},
		// f : (a : Nat) → P (inl Nat Bool a) = λa. a
		ast.Lam{Binder: "a", Body: ast.Var{Ix: 0}},
		// g : (b : Bool) → P (inr Nat Bool b) = λb. zero
		ast.Lam{Binder: "b", Body: zero},
		// s : Sum Nat Bool = inl Nat Bool zero
		inlTerm,
	)

	result := eval.EvalNBE(term)

	// Should reduce to zero
	if g, ok := result.(ast.Global); !ok || g.Name != "zero" {
		t.Errorf("sumElim ... (inl Nat Bool zero) should reduce to zero, got %v", result)
	}
}

// TestSumElimComputationInr verifies the computation rule:
// sumElim A B P f g (inr A B b) → g b
func TestSumElimComputationInr(t *testing.T) {
	nat := ast.Global{Name: "Nat"}
	bool_ := ast.Global{Name: "Bool"}
	zero := ast.Global{Name: "zero"}
	true_ := ast.Global{Name: "true"}

	// Build (inr Nat Bool true)
	inrTerm := ast.MkApps(
		ast.Global{Name: "inr"},
		nat,   // A
		bool_, // B
		true_, // b
	)

	// Build sumElim Nat Bool P f g (inr Nat Bool true)
	term := ast.MkApps(
		ast.Global{Name: "sumElim"},
		nat,   // A
		bool_, // B
		// P : Sum Nat Bool → Type = λ_. Nat
		ast.Lam{Binder: "_", Body: nat},
		// f : (a : Nat) → P (inl Nat Bool a) = λa. a
		ast.Lam{Binder: "a", Body: ast.Var{Ix: 0}},
		// g : (b : Bool) → P (inr Nat Bool b) = λb. zero
		ast.Lam{Binder: "b", Body: zero},
		// s : Sum Nat Bool = inr Nat Bool true
		inrTerm,
	)

	result := eval.EvalNBE(term)

	// Should reduce to zero (g true = zero)
	if g, ok := result.(ast.Global); !ok || g.Name != "zero" {
		t.Errorf("sumElim ... (inr Nat Bool true) should reduce to zero, got %v", result)
	}
}
