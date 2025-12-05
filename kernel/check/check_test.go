package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/kernel/ctx"
)

// Helper to create an empty context
func emptyCtx() *ctx.Ctx {
	return &ctx.Ctx{Tele: nil}
}

// TestIdentityFunction tests the critical success criterion:
// id : Π(A:Type). A → A
// id = λA. λx. x
func TestIdentityFunction(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// id = λA. λx. x
	// With annotations: λ(A : Type0). λ(x : A). x
	idTerm := ast.Lam{
		Binder: "A",
		Ann:    ast.Sort{U: 0}, // A : Type0
		Body: ast.Lam{
			Binder: "x",
			Ann:    ast.Var{Ix: 0}, // x : A (A is Var{0} under this binder)
			Body:   ast.Var{Ix: 0}, // x (x is Var{0} under both binders)
		},
	}

	// Expected type: Π(A:Type0). A → A
	// = Π(A:Type0). Π(_:A). A
	idType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "_",
			A:      ast.Var{Ix: 0}, // A
			B:      ast.Var{Ix: 1}, // A (shifted by 1)
		},
	}

	// Synthesize type
	inferredType, err := checker.Synth(context, NoSpan(), idTerm)
	if err != nil {
		t.Fatalf("Failed to synthesize identity function type: %v", err)
	}

	// Check it matches expected
	if !checker.conv(inferredType, idType) {
		t.Errorf("Type mismatch:\n  expected: %s\n  got: %s",
			ast.Sprint(idType), ast.Sprint(inferredType))
	}

	// Also test checking mode
	if checkErr := checker.Check(context, NoSpan(), idTerm, idType); checkErr != nil {
		t.Errorf("Check failed: %v", checkErr)
	}
}

// TestIdentityFunctionUnannotated tests checking an unannotated lambda.
func TestIdentityFunctionUnannotated(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// Unannotated: λA. λx. x
	idTerm := ast.Lam{
		Binder: "A",
		Body: ast.Lam{
			Binder: "x",
			Body:   ast.Var{Ix: 0}, // x
		},
	}

	// Type: Π(A:Type0). A → A
	idType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "_",
			A:      ast.Var{Ix: 0},
			B:      ast.Var{Ix: 1},
		},
	}

	// Checking mode should succeed
	if err := checker.Check(context, NoSpan(), idTerm, idType); err != nil {
		t.Errorf("Check failed for unannotated identity: %v", err)
	}
}

// TestCompositionFunction tests compose : Π(A B C : Type). (B → C) → (A → B) → A → C
func TestCompositionFunction(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	type0 := ast.Sort{U: 0}

	// compose = λA. λB. λC. λf. λg. λx. f (g x)
	composeTerm := ast.Lam{
		Binder: "A", Ann: type0,
		Body: ast.Lam{
			Binder: "B", Ann: type0,
			Body: ast.Lam{
				Binder: "C", Ann: type0,
				Body: ast.Lam{
					Binder: "f",
					Ann: ast.Pi{Binder: "_", A: ast.Var{Ix: 1}, B: ast.Var{Ix: 1}}, // B → C
					Body: ast.Lam{
						Binder: "g",
						Ann: ast.Pi{Binder: "_", A: ast.Var{Ix: 3}, B: ast.Var{Ix: 3}}, // A → B
						Body: ast.Lam{
							Binder: "x",
							Ann:    ast.Var{Ix: 4}, // A
							Body: ast.App{
								T: ast.Var{Ix: 2}, // f
								U: ast.App{
									T: ast.Var{Ix: 1}, // g
									U: ast.Var{Ix: 0}, // x
								},
							},
						},
					},
				},
			},
		},
	}

	// Just check it type checks
	_, err := checker.Synth(context, NoSpan(), composeTerm)
	if err != nil {
		t.Fatalf("Failed to synthesize composition function: %v", err)
	}
}

// TestNatPrimitives tests that Nat, zero, succ type check correctly.
func TestNatPrimitives(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// Nat : Type0
	natTy, err := checker.Synth(context, NoSpan(), ast.Global{Name: "Nat"})
	if err != nil {
		t.Fatalf("Failed to synth Nat: %v", err)
	}
	if !checker.conv(natTy, ast.Sort{U: 0}) {
		t.Errorf("Nat should have type Type0, got %s", ast.Sprint(natTy))
	}

	// zero : Nat
	zeroTy, err := checker.Synth(context, NoSpan(), ast.Global{Name: "zero"})
	if err != nil {
		t.Fatalf("Failed to synth zero: %v", err)
	}
	if !checker.conv(zeroTy, ast.Global{Name: "Nat"}) {
		t.Errorf("zero should have type Nat, got %s", ast.Sprint(zeroTy))
	}

	// succ zero : Nat
	succZero := ast.App{T: ast.Global{Name: "succ"}, U: ast.Global{Name: "zero"}}
	succZeroTy, err := checker.Synth(context, NoSpan(), succZero)
	if err != nil {
		t.Fatalf("Failed to synth succ zero: %v", err)
	}
	if !checker.conv(succZeroTy, ast.Global{Name: "Nat"}) {
		t.Errorf("succ zero should have type Nat, got %s", ast.Sprint(succZeroTy))
	}
}

// TestBoolPrimitives tests that Bool, true, false type check correctly.
func TestBoolPrimitives(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// Bool : Type0
	boolTy, err := checker.Synth(context, NoSpan(), ast.Global{Name: "Bool"})
	if err != nil {
		t.Fatalf("Failed to synth Bool: %v", err)
	}
	if !checker.conv(boolTy, ast.Sort{U: 0}) {
		t.Errorf("Bool should have type Type0, got %s", ast.Sprint(boolTy))
	}

	// true : Bool
	trueTy, err := checker.Synth(context, NoSpan(), ast.Global{Name: "true"})
	if err != nil {
		t.Fatalf("Failed to synth true: %v", err)
	}
	if !checker.conv(trueTy, ast.Global{Name: "Bool"}) {
		t.Errorf("true should have type Bool, got %s", ast.Sprint(trueTy))
	}

	// false : Bool
	falseTy, err := checker.Synth(context, NoSpan(), ast.Global{Name: "false"})
	if err != nil {
		t.Fatalf("Failed to synth false: %v", err)
	}
	if !checker.conv(falseTy, ast.Global{Name: "Bool"}) {
		t.Errorf("false should have type Bool, got %s", ast.Sprint(falseTy))
	}
}

// TestSortHierarchy tests Type0 : Type1 : Type2 : ...
func TestSortHierarchy(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())
	context := emptyCtx()

	tests := []struct {
		level    ast.Level
		expected ast.Level
	}{
		{0, 1},
		{1, 2},
		{5, 6},
	}

	for _, tt := range tests {
		ty, err := checker.Synth(context, NoSpan(), ast.Sort{U: tt.level})
		if err != nil {
			t.Errorf("Failed to synth Type%d: %v", tt.level, err)
			continue
		}
		expectedTy := ast.Sort{U: tt.expected}
		if !checker.conv(ty, expectedTy) {
			t.Errorf("Type%d : Type%d, got %s", tt.level, tt.expected, ast.Sprint(ty))
		}
	}
}

// TestPiTypeFormation tests that Pi types are well-formed.
func TestPiTypeFormation(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())
	context := emptyCtx()

	// Π(x : Type0). Type0 : Type1
	piType := ast.Pi{
		Binder: "x",
		A:      ast.Sort{U: 0},
		B:      ast.Sort{U: 0},
	}

	ty, err := checker.Synth(context, NoSpan(), piType)
	if err != nil {
		t.Fatalf("Failed to synth Pi type: %v", err)
	}
	if !checker.conv(ty, ast.Sort{U: 1}) {
		t.Errorf("Π(x:Type0).Type0 should be Type1, got %s", ast.Sprint(ty))
	}
}

// TestSigmaTypeFormation tests that Sigma types are well-formed.
func TestSigmaTypeFormation(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())
	context := emptyCtx()

	// Σ(x : Type0). Type0 : Type1
	sigmaType := ast.Sigma{
		Binder: "x",
		A:      ast.Sort{U: 0},
		B:      ast.Sort{U: 0},
	}

	ty, err := checker.Synth(context, NoSpan(), sigmaType)
	if err != nil {
		t.Fatalf("Failed to synth Sigma type: %v", err)
	}
	if !checker.conv(ty, ast.Sort{U: 1}) {
		t.Errorf("Σ(x:Type0).Type0 should be Type1, got %s", ast.Sprint(ty))
	}
}

// TestPairTyping tests pair checking against Sigma types.
func TestPairTyping(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// (zero, zero) : Σ(_ : Nat). Nat
	pair := ast.Pair{
		Fst: ast.Global{Name: "zero"},
		Snd: ast.Global{Name: "zero"},
	}
	sigmaType := ast.Sigma{
		Binder: "_",
		A:      ast.Global{Name: "Nat"},
		B:      ast.Global{Name: "Nat"},
	}

	if err := checker.Check(context, NoSpan(), pair, sigmaType); err != nil {
		t.Errorf("Pair check failed: %v", err)
	}
}

// TestProjections tests fst and snd typing.
func TestProjections(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// Add a pair to context: p : Σ(_ : Nat). Nat
	sigmaType := ast.Sigma{
		Binder: "_",
		A:      ast.Global{Name: "Nat"},
		B:      ast.Global{Name: "Nat"},
	}
	context.Extend("p", sigmaType)

	// fst p : Nat
	fstTy, err := checker.Synth(context, NoSpan(), ast.Fst{P: ast.Var{Ix: 0}})
	if err != nil {
		t.Fatalf("Failed to synth fst: %v", err)
	}
	if !checker.conv(fstTy, ast.Global{Name: "Nat"}) {
		t.Errorf("fst p should have type Nat, got %s", ast.Sprint(fstTy))
	}

	// snd p : Nat
	sndTy, err := checker.Synth(context, NoSpan(), ast.Snd{P: ast.Var{Ix: 0}})
	if err != nil {
		t.Fatalf("Failed to synth snd: %v", err)
	}
	if !checker.conv(sndTy, ast.Global{Name: "Nat"}) {
		t.Errorf("snd p should have type Nat, got %s", ast.Sprint(sndTy))
	}
}

// TestLetExpression tests let bindings.
func TestLetExpression(t *testing.T) {
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
		t.Errorf("let expression should have type Nat, got %s", ast.Sprint(ty))
	}
}

// TestUnboundVariableError tests error for unbound variables.
func TestUnboundVariableError(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())
	context := emptyCtx()

	// Var{0} in empty context
	_, err := checker.Synth(context, NewSpan("test.hott", 1, 1, 1, 2), ast.Var{Ix: 0})
	if err == nil {
		t.Fatal("Expected error for unbound variable")
	}
	if err.Kind != ErrUnboundVariable {
		t.Errorf("Expected ErrUnboundVariable, got %v", err.Kind)
	}
	if err.Span.File != "test.hott" {
		t.Errorf("Expected span file 'test.hott', got %s", err.Span.File)
	}
}

// TestTypeMismatchError tests error for type mismatches.
func TestTypeMismatchError(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// Check true against Nat
	span := NewSpan("test.hott", 5, 10, 5, 14)
	err := checker.Check(context, span, ast.Global{Name: "true"}, ast.Global{Name: "Nat"})
	if err == nil {
		t.Fatal("Expected error for type mismatch")
	}
	if err.Kind != ErrTypeMismatch {
		t.Errorf("Expected ErrTypeMismatch, got %v", err.Kind)
	}
	// Check details
	if details, ok := err.Details.(TypeMismatchDetails); ok {
		if !checker.conv(details.Expected, ast.Global{Name: "Nat"}) {
			t.Error("Details should show Nat as expected type")
		}
	} else {
		t.Error("Expected TypeMismatchDetails")
	}
}

// TestNotAFunctionError tests error when applying a non-function.
func TestNotAFunctionError(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// zero zero (applying zero to zero)
	app := ast.App{
		T: ast.Global{Name: "zero"},
		U: ast.Global{Name: "zero"},
	}

	_, err := checker.Synth(context, NewSpan("test.hott", 3, 1, 3, 10), app)
	if err == nil {
		t.Fatal("Expected error for non-function application")
	}
	if err.Kind != ErrNotAFunction {
		t.Errorf("Expected ErrNotAFunction, got %v", err.Kind)
	}
}

// TestUnknownGlobalError tests error for unknown globals.
func TestUnknownGlobalError(t *testing.T) {
	checker := NewChecker(NewGlobalEnv()) // No primitives
	context := emptyCtx()

	_, err := checker.Synth(context, NewSpan("test.hott", 1, 1, 1, 5), ast.Global{Name: "foo"})
	if err == nil {
		t.Fatal("Expected error for unknown global")
	}
	if err.Kind != ErrUnknownGlobal {
		t.Errorf("Expected ErrUnknownGlobal, got %v", err.Kind)
	}
	if details, ok := err.Details.(UnknownGlobalDetails); ok {
		if details.Name != "foo" {
			t.Errorf("Expected name 'foo', got %s", details.Name)
		}
	}
}

// TestCannotInferError tests error when synthesis is not possible.
func TestCannotInferError(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())
	context := emptyCtx()

	// Unannotated lambda cannot be synthesized
	lam := ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}}

	_, err := checker.Synth(context, NoSpan(), lam)
	if err == nil {
		t.Fatal("Expected error for unannotated lambda synthesis")
	}
	if err.Kind != ErrCannotInfer {
		t.Errorf("Expected ErrCannotInfer, got %v", err.Kind)
	}
}

// TestSpanFormatting tests span string formatting.
func TestSpanFormatting(t *testing.T) {
	tests := []struct {
		span     Span
		expected string
	}{
		{NoSpan(), "<no location>"},
		{NewSpan("", 1, 5, 1, 5), "1:5"},
		{NewSpan("", 1, 5, 1, 10), "1:5-10"},
		{NewSpan("", 1, 5, 3, 10), "1:5-3:10"},
		{NewSpan("foo.hott", 1, 5, 1, 10), "foo.hott:1:5-10"},
	}

	for _, tt := range tests {
		got := tt.span.String()
		if got != tt.expected {
			t.Errorf("Span.String() = %q, want %q", got, tt.expected)
		}
	}
}

// TestDependentPair tests dependent pairs (where second type depends on first).
func TestDependentPair(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// Add an axiom: Vec : Nat → Type0
	checker.globals.AddAxiom("Vec", ast.Pi{
		Binder: "n",
		A:      ast.Global{Name: "Nat"},
		B:      ast.Sort{U: 0},
	})

	// Type: Σ(n : Nat). Vec n
	sigmaType := ast.Sigma{
		Binder: "n",
		A:      ast.Global{Name: "Nat"},
		B:      ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 0}},
	}

	// Check it's a valid type
	level, err := checker.CheckIsType(context, NoSpan(), sigmaType)
	if err != nil {
		t.Fatalf("Failed to check Sigma type: %v", err)
	}
	if level != 0 {
		t.Errorf("Expected level 0, got %d", level)
	}
}

// TestContextManagement tests proper context extension and dropping.
func TestContextManagement(t *testing.T) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// λ(A : Type). λ(x : A). λ(y : A). x
	term := ast.Lam{
		Binder: "A", Ann: ast.Sort{U: 0},
		Body: ast.Lam{
			Binder: "x", Ann: ast.Var{Ix: 0},
			Body: ast.Lam{
				Binder: "y", Ann: ast.Var{Ix: 1},
				Body:   ast.Var{Ix: 1}, // x, not y
			},
		},
	}

	ty, err := checker.Synth(context, NoSpan(), term)
	if err != nil {
		t.Fatalf("Failed to synth: %v", err)
	}

	// Context should be empty after synthesis
	if context.Len() != 0 {
		t.Errorf("Context should be empty after synthesis, got length %d", context.Len())
	}

	// Type should be: Π(A:Type). Π(x:A). Π(y:A). A
	if _, ok := ty.(ast.Pi); !ok {
		t.Errorf("Expected Pi type, got %T", ty)
	}
}

// BenchmarkIdentityTypeCheck benchmarks type checking the identity function.
func BenchmarkIdentityTypeCheck(b *testing.B) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	idTerm := ast.Lam{
		Binder: "A", Ann: ast.Sort{U: 0},
		Body: ast.Lam{
			Binder: "x", Ann: ast.Var{Ix: 0},
			Body:   ast.Var{Ix: 0},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = checker.Synth(context, NoSpan(), idTerm)
	}
}

// BenchmarkDeepNesting benchmarks type checking deeply nested terms.
func BenchmarkDeepNesting(b *testing.B) {
	checker := NewChecker(NewGlobalEnvWithPrimitives())
	context := emptyCtx()

	// Build: succ (succ (succ ... zero))
	var term ast.Term = ast.Global{Name: "zero"}
	for i := 0; i < 20; i++ {
		term = ast.App{T: ast.Global{Name: "succ"}, U: term}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = checker.Synth(context, NoSpan(), term)
	}
}
