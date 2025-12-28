package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/kernel/ctx"
)

// ============================================================================
// Complex Application Chain Tests
// ============================================================================

// TestSynthApp_MultiArg tests multi-argument function application: f a b c
func TestSynthApp_MultiArg(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// Define f : Nat -> Nat -> Nat -> Nat
	checker.globals.AddAxiom("f", ast.Pi{
		Binder: "_", A: ast.Global{Name: "Nat"},
		B: ast.Pi{
			Binder: "_", A: ast.Global{Name: "Nat"},
			B: ast.Pi{
				Binder: "_", A: ast.Global{Name: "Nat"},
				B: ast.Global{Name: "Nat"},
			},
		},
	})

	// f zero zero zero : Nat
	app := ast.MkApps(
		ast.Global{Name: "f"},
		ast.Global{Name: "zero"},
		ast.Global{Name: "zero"},
		ast.Global{Name: "zero"},
	)

	ty, err := checker.Synth(context, NoSpan(), app)
	if err != nil {
		t.Fatalf("Failed to synth f zero zero zero: %v", err)
	}
	if !checker.conv(ty, ast.Global{Name: "Nat"}) {
		t.Errorf("Expected Nat, got %s", ast.Sprint(ty))
	}
}

// TestSynthApp_NestedApplications tests nested applications: (f x) (g y)
func TestSynthApp_NestedApplications(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// Define:
	// f : Nat -> Nat -> Nat
	// g : Nat -> Nat
	checker.globals.AddAxiom("f", ast.Pi{
		Binder: "_", A: ast.Global{Name: "Nat"},
		B: ast.Pi{
			Binder: "_", A: ast.Global{Name: "Nat"},
			B: ast.Global{Name: "Nat"},
		},
	})
	checker.globals.AddAxiom("g", ast.Pi{
		Binder: "_", A: ast.Global{Name: "Nat"},
		B: ast.Global{Name: "Nat"},
	})

	// (f zero) (g zero) : Nat
	app := ast.App{
		T: ast.App{T: ast.Global{Name: "f"}, U: ast.Global{Name: "zero"}},
		U: ast.App{T: ast.Global{Name: "g"}, U: ast.Global{Name: "zero"}},
	}

	ty, err := checker.Synth(context, NoSpan(), app)
	if err != nil {
		t.Fatalf("Failed to synth (f zero) (g zero): %v", err)
	}
	if !checker.conv(ty, ast.Global{Name: "Nat"}) {
		t.Errorf("Expected Nat, got %s", ast.Sprint(ty))
	}
}

// TestSynthApp_TypeArgument tests applications with type arguments
func TestSynthApp_TypeArgument(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// id : (A : Type) -> A -> A
	idType := ast.Pi{
		Binder: "A", A: ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "_", A: ast.Var{Ix: 0},
			B: ast.Var{Ix: 1},
		},
	}
	checker.globals.AddAxiom("id", idType)

	// id Nat zero : Nat
	app := ast.MkApps(
		ast.Global{Name: "id"},
		ast.Global{Name: "Nat"},
		ast.Global{Name: "zero"},
	)

	ty, err := checker.Synth(context, NoSpan(), app)
	if err != nil {
		t.Fatalf("Failed to synth id Nat zero: %v", err)
	}
	if !checker.conv(ty, ast.Global{Name: "Nat"}) {
		t.Errorf("Expected Nat, got %s", ast.Sprint(ty))
	}
}

// ============================================================================
// Identity Type (Id) and J Elimination Tests
// ============================================================================

// TestSynthId tests Id type synthesis
func TestSynthId(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// Id Nat zero zero : Type0
	idTerm := ast.Id{
		A: ast.Global{Name: "Nat"},
		X: ast.Global{Name: "zero"},
		Y: ast.Global{Name: "zero"},
	}

	ty, err := checker.Synth(context, NoSpan(), idTerm)
	if err != nil {
		t.Fatalf("Failed to synth Id Nat zero zero: %v", err)
	}
	if !checker.conv(ty, ast.Sort{U: 0}) {
		t.Errorf("Expected Type0, got %s", ast.Sprint(ty))
	}
}

// TestSynthRefl tests Refl synthesis
func TestSynthRefl(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// refl Nat zero : Id Nat zero zero
	reflTerm := ast.Refl{
		A: ast.Global{Name: "Nat"},
		X: ast.Global{Name: "zero"},
	}

	ty, err := checker.Synth(context, NoSpan(), reflTerm)
	if err != nil {
		t.Fatalf("Failed to synth refl Nat zero: %v", err)
	}

	expectedTy := ast.Id{
		A: ast.Global{Name: "Nat"},
		X: ast.Global{Name: "zero"},
		Y: ast.Global{Name: "zero"},
	}
	if !checker.conv(ty, expectedTy) {
		t.Errorf("Expected Id Nat zero zero, got %s", ast.Sprint(ty))
	}
}

// TestSynthJ tests J elimination synthesis
func TestSynthJ(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// J Nat C d zero zero (refl Nat zero) : C zero (refl Nat zero)
	// where C : (y : Nat) -> Id Nat zero y -> Type0

	// C = λy. λp. Nat (a simple motive that ignores the proof)
	motive := ast.Lam{
		Binder: "y", Ann: ast.Global{Name: "Nat"},
		Body: ast.Lam{
			Binder: "p",
			Ann: ast.Id{
				A: ast.Global{Name: "Nat"},
				X: ast.Global{Name: "zero"},
				Y: ast.Var{Ix: 0}, // y
			},
			Body: ast.Global{Name: "Nat"},
		},
	}

	// d : C zero (refl Nat zero) = Nat, so d can be zero
	baseCase := ast.Global{Name: "zero"}

	jTerm := ast.J{
		A: ast.Global{Name: "Nat"},
		C: motive,
		D: baseCase,
		X: ast.Global{Name: "zero"},
		Y: ast.Global{Name: "zero"},
		P: ast.Refl{A: ast.Global{Name: "Nat"}, X: ast.Global{Name: "zero"}},
	}

	ty, err := checker.Synth(context, NoSpan(), jTerm)
	if err != nil {
		t.Fatalf("Failed to synth J: %v", err)
	}

	// Result should be C zero (refl Nat zero) = Nat
	if !checker.conv(ty, ast.Global{Name: "Nat"}) {
		t.Errorf("Expected Nat, got %s", ast.Sprint(ty))
	}
}

// TestSynthId_Error tests error case for Id with mismatched endpoints
func TestSynthId_Error(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// Id Nat zero true should fail (true is not a Nat)
	idTerm := ast.Id{
		A: ast.Global{Name: "Nat"},
		X: ast.Global{Name: "zero"},
		Y: ast.Global{Name: "true"}, // Bool, not Nat
	}

	_, err := checker.Synth(context, NoSpan(), idTerm)
	if err == nil {
		t.Fatal("Expected error for Id with mismatched endpoints")
	}
}

// ============================================================================
// Universe Level Tests
// ============================================================================

// TestUniverseLevelMax tests max level computation for Pi/Sigma
func TestUniverseLevelMax(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())
	context := emptyCtx()

	// Pi (A : Type1) . Type0 : Type2 (max 1, 0 = 1, +1 = 2)
	pi := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 1},
		B:      ast.Sort{U: 0},
	}

	ty, err := checker.Synth(context, NoSpan(), pi)
	if err != nil {
		t.Fatalf("Failed to synth Pi: %v", err)
	}
	if !checker.conv(ty, ast.Sort{U: 2}) {
		t.Errorf("Expected Type2, got %s", ast.Sprint(ty))
	}

	// Sigma (A : Type2) . Type1 : Type3 (max 2, 1 = 2, +1 = 3)
	sigma := ast.Sigma{
		Binder: "A",
		A:      ast.Sort{U: 2},
		B:      ast.Sort{U: 1},
	}

	ty, err = checker.Synth(context, NoSpan(), sigma)
	if err != nil {
		t.Fatalf("Failed to synth Sigma: %v", err)
	}
	if !checker.conv(ty, ast.Sort{U: 3}) {
		t.Errorf("Expected Type3, got %s", ast.Sprint(ty))
	}
}

// ============================================================================
// Check Mode Tests
// ============================================================================

// TestCheckLam_Unannotated tests checking unannotated lambda against Pi
func TestCheckLam_Unannotated(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// λx. x checked against Nat -> Nat
	lam := ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}}
	piType := ast.Pi{Binder: "_", A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "Nat"}}

	err := checker.Check(context, NoSpan(), lam, piType)
	if err != nil {
		t.Errorf("Check unannotated lambda failed: %v", err)
	}
}

// TestCheckPair tests checking pair against Sigma
func TestCheckPair(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// (zero, true) checked against Σ(_ : Nat). Bool
	pair := ast.Pair{
		Fst: ast.Global{Name: "zero"},
		Snd: ast.Global{Name: "true"},
	}
	sigmaType := ast.Sigma{
		Binder: "_",
		A:      ast.Global{Name: "Nat"},
		B:      ast.Global{Name: "Bool"},
	}

	err := checker.Check(context, NoSpan(), pair, sigmaType)
	if err != nil {
		t.Errorf("Check pair failed: %v", err)
	}
}

// TestCheckPair_Dependent tests checking pair against dependent Sigma
func TestCheckPair_Dependent(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// Add Vec : Nat -> Type
	checker.globals.AddAxiom("Vec", ast.Pi{
		Binder: "n", A: ast.Global{Name: "Nat"},
		B: ast.Sort{U: 0},
	})
	// Add vnil : Vec zero
	checker.globals.AddAxiom("vnil", ast.App{
		T: ast.Global{Name: "Vec"},
		U: ast.Global{Name: "zero"},
	})

	// (zero, vnil) checked against Σ(n : Nat). Vec n
	pair := ast.Pair{
		Fst: ast.Global{Name: "zero"},
		Snd: ast.Global{Name: "vnil"},
	}
	sigmaType := ast.Sigma{
		Binder: "n",
		A:      ast.Global{Name: "Nat"},
		B:      ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 0}},
	}

	err := checker.Check(context, NoSpan(), pair, sigmaType)
	if err != nil {
		t.Errorf("Check dependent pair failed: %v", err)
	}
}

// ============================================================================
// Deep Context Tests
// ============================================================================

// TestDeepContext tests variable lookup in deeply nested context
func TestDeepContext(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// λA. λB. λC. λx. λy. λz. x (returns variable at depth 2 from innermost)
	term := ast.Lam{
		Binder: "A", Ann: ast.Sort{U: 0},
		Body: ast.Lam{
			Binder: "B", Ann: ast.Sort{U: 0},
			Body: ast.Lam{
				Binder: "C", Ann: ast.Sort{U: 0},
				Body: ast.Lam{
					Binder: "x", Ann: ast.Var{Ix: 2}, // A
					Body: ast.Lam{
						Binder: "y", Ann: ast.Var{Ix: 2}, // B
						Body: ast.Lam{
							Binder: "z", Ann: ast.Var{Ix: 2}, // C
							Body: ast.Var{Ix: 2}, // x
						},
					},
				},
			},
		},
	}

	ty, err := checker.Synth(context, NoSpan(), term)
	if err != nil {
		t.Fatalf("Failed to synth deeply nested term: %v", err)
	}

	// Should be a Pi type
	if _, ok := ty.(ast.Pi); !ok {
		t.Errorf("Expected Pi type, got %T", ty)
	}
}

// TestContextCleanup tests that context is properly cleaned up after synthesis
func TestContextCleanup(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := &ctx.Ctx{Tele: nil}

	// Synthesize a term that extends context
	term := ast.Lam{
		Binder: "x", Ann: ast.Sort{U: 0},
		Body: ast.Var{Ix: 0},
	}

	_, err := checker.Synth(context, NoSpan(), term)
	if err != nil {
		t.Fatalf("Synth failed: %v", err)
	}

	// Context should be back to empty
	if context.Len() != 0 {
		t.Errorf("Context not cleaned up, length = %d", context.Len())
	}
}

// ============================================================================
// Ensure* Helper Tests
// ============================================================================

// TestEnsurePi_NotPi tests error when non-Pi is expected to be Pi
func TestEnsurePi_NotPi(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// Apply zero to something - zero is not a function
	app := ast.App{
		T: ast.Global{Name: "zero"},
		U: ast.Global{Name: "zero"},
	}

	_, err := checker.Synth(context, NoSpan(), app)
	if err == nil {
		t.Fatal("Expected error when applying non-function")
	}
	if err.Kind != ErrNotAFunction {
		t.Errorf("Expected ErrNotAFunction, got %v", err.Kind)
	}
}

// TestEnsureSigma_NotSigma tests error when non-Sigma is expected to be Sigma
func TestEnsureSigma_NotSigma(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// fst zero - zero is not a pair
	fst := ast.Fst{P: ast.Global{Name: "zero"}}

	_, err := checker.Synth(context, NoSpan(), fst)
	if err == nil {
		t.Fatal("Expected error when projecting from non-pair")
	}
	if err.Kind != ErrNotAPair {
		t.Errorf("Expected ErrNotAPair, got %v", err.Kind)
	}
}

// TestEnsureSort_NotSort tests error when non-Sort is expected to be Sort
func TestEnsureSort_NotSort(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// Pi (x : zero) . x - zero is not a type
	pi := ast.Pi{
		Binder: "x",
		A:      ast.Global{Name: "zero"},
		B:      ast.Var{Ix: 0},
	}

	_, err := checker.Synth(context, NoSpan(), pi)
	if err == nil {
		t.Fatal("Expected error when domain is not a type")
	}
	if err.Kind != ErrNotAType {
		t.Errorf("Expected ErrNotAType, got %v", err.Kind)
	}
}

// ============================================================================
// Let Expression Tests
// ============================================================================

// TestSynthLet_Annotated tests let with annotation
func TestSynthLet_Annotated(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// let x : Nat = zero in succ x
	letTerm := ast.Let{
		Binder: "x",
		Ann:    ast.Global{Name: "Nat"},
		Val:    ast.Global{Name: "zero"},
		Body:   ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 0}},
	}

	ty, err := checker.Synth(context, NoSpan(), letTerm)
	if err != nil {
		t.Fatalf("Failed to synth let: %v", err)
	}
	if !checker.conv(ty, ast.Global{Name: "Nat"}) {
		t.Errorf("Expected Nat, got %s", ast.Sprint(ty))
	}
}

// TestSynthLet_Unannotated tests let without annotation (synthesizes value type)
func TestSynthLet_Unannotated(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// let x = zero in succ x (no annotation)
	letTerm := ast.Let{
		Binder: "x",
		Val:    ast.Global{Name: "zero"},
		Body:   ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 0}},
	}

	ty, err := checker.Synth(context, NoSpan(), letTerm)
	if err != nil {
		t.Fatalf("Failed to synth unannotated let: %v", err)
	}
	if !checker.conv(ty, ast.Global{Name: "Nat"}) {
		t.Errorf("Expected Nat, got %s", ast.Sprint(ty))
	}
}

// ============================================================================
// Nil Term Tests
// ============================================================================

// TestSynthNil tests that nil term produces error
func TestSynthNil(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())
	context := emptyCtx()

	_, err := checker.Synth(context, NoSpan(), nil)
	if err == nil {
		t.Fatal("Expected error for nil term")
	}
	if err.Kind != ErrCannotInfer {
		t.Errorf("Expected ErrCannotInfer, got %v", err.Kind)
	}
}

// TestCheckNil tests that nil term produces error in check mode
func TestCheckNil(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	err := checker.Check(context, NoSpan(), nil, ast.Global{Name: "Nat"})
	if err == nil {
		t.Fatal("Expected error for nil term in check mode")
	}
}
