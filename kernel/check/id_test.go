package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
	"github.com/watchthelight/HypergraphGo/kernel/ctx"
)

// idTestCtx creates an empty context for identity type tests
func idTestCtx() *ctx.Ctx {
	return &ctx.Ctx{Tele: nil}
}

// TestIdTypeFormation tests that Id A x y : Type when A : Type and x, y : A
func TestIdTypeFormation(t *testing.T) {
	globals := NewGlobalEnvWithPrimitives()
	checker := NewChecker(globals)
	context := idTestCtx()

	// Id Nat zero zero : Type0
	idType := ast.Id{
		A: ast.Global{Name: "Nat"},
		X: ast.Global{Name: "zero"},
		Y: ast.Global{Name: "zero"},
	}

	ty, err := checker.Synth(context, NoSpan(), idType)
	if err != nil {
		t.Fatalf("expected Id Nat zero zero to typecheck, got error: %v", err)
	}

	// Should be Sort{U: 0}
	sort, ok := ty.(ast.Sort)
	if !ok {
		t.Fatalf("expected Sort, got %T", ty)
	}
	if sort.U != 0 {
		t.Errorf("expected Type0, got Type%d", sort.U)
	}
}

// TestIdTypeFormationDifferentEndpoints tests Id with different endpoints
func TestIdTypeFormationDifferentEndpoints(t *testing.T) {
	globals := NewGlobalEnvWithPrimitives()
	checker := NewChecker(globals)
	context := idTestCtx()

	// Id Nat zero (succ zero) : Type0
	succZero := ast.App{T: ast.Global{Name: "succ"}, U: ast.Global{Name: "zero"}}
	idType := ast.Id{
		A: ast.Global{Name: "Nat"},
		X: ast.Global{Name: "zero"},
		Y: succZero,
	}

	ty, err := checker.Synth(context, NoSpan(), idType)
	if err != nil {
		t.Fatalf("expected Id Nat zero (succ zero) to typecheck, got error: %v", err)
	}

	sort, ok := ty.(ast.Sort)
	if !ok {
		t.Fatalf("expected Sort, got %T", ty)
	}
	if sort.U != 0 {
		t.Errorf("expected Type0, got Type%d", sort.U)
	}
}

// TestRefl tests that refl A x : Id A x x
func TestRefl(t *testing.T) {
	globals := NewGlobalEnvWithPrimitives()
	checker := NewChecker(globals)
	context := idTestCtx()

	// refl Nat zero : Id Nat zero zero
	refl := ast.Refl{
		A: ast.Global{Name: "Nat"},
		X: ast.Global{Name: "zero"},
	}

	ty, err := checker.Synth(context, NoSpan(), refl)
	if err != nil {
		t.Fatalf("expected refl Nat zero to typecheck, got error: %v", err)
	}

	// Should be Id Nat zero zero
	id, ok := ty.(ast.Id)
	if !ok {
		t.Fatalf("expected Id type, got %T", ty)
	}
	if !alphaEq(id.A, ast.Global{Name: "Nat"}) {
		t.Errorf("expected Id A = Nat, got %v", ast.Sprint(id.A))
	}
	if !alphaEq(id.X, ast.Global{Name: "zero"}) {
		t.Errorf("expected Id X = zero, got %v", ast.Sprint(id.X))
	}
	if !alphaEq(id.Y, ast.Global{Name: "zero"}) {
		t.Errorf("expected Id Y = zero, got %v", ast.Sprint(id.Y))
	}
}

// TestJComputation tests that J A C d x x (refl A x) --> d
func TestJComputation(t *testing.T) {
	// Build: J Nat (\y. \p. Nat) (succ zero) zero zero (refl Nat zero)
	// The motive C = \y. \p. Nat (constant motive)
	// The base case d = succ zero
	// Should reduce to: succ zero

	nat := ast.Global{Name: "Nat"}
	zero := ast.Global{Name: "zero"}
	succZero := ast.App{T: ast.Global{Name: "succ"}, U: zero}

	// Motive: \y. \p. Nat (ignores y and p, returns Nat)
	motive := ast.Lam{
		Binder: "y",
		Ann:    nat,
		Body: ast.Lam{
			Binder: "p",
			Ann:    ast.Id{A: nat, X: zero, Y: ast.Var{Ix: 0}},
			Body:   nat,
		},
	}

	j := ast.J{
		A: nat,
		C: motive,
		D: succZero,
		X: zero,
		Y: zero,
		P: ast.Refl{A: nat, X: zero},
	}

	// Evaluate using NbE
	result := eval.EvalNBE(j)

	// Should normalize to (succ zero)
	if !alphaEq(result, succZero) {
		t.Errorf("expected J to reduce to succ zero, got %v", ast.Sprint(result))
	}
}

// TestJTyping tests that J typechecks correctly
func TestJTyping(t *testing.T) {
	globals := NewGlobalEnvWithPrimitives()
	checker := NewChecker(globals)
	context := idTestCtx()

	nat := ast.Global{Name: "Nat"}
	zero := ast.Global{Name: "zero"}
	succZero := ast.App{T: ast.Global{Name: "succ"}, U: zero}

	// Motive: \y. \p. Nat
	motive := ast.Lam{
		Binder: "y",
		Ann:    nat,
		Body: ast.Lam{
			Binder: "p",
			Ann:    ast.Id{A: nat, X: zero, Y: ast.Var{Ix: 0}},
			Body:   nat,
		},
	}

	j := ast.J{
		A: nat,
		C: motive,
		D: succZero,
		X: zero,
		Y: zero,
		P: ast.Refl{A: nat, X: zero},
	}

	ty, err := checker.Synth(context, NoSpan(), j)
	if err != nil {
		t.Fatalf("expected J to typecheck, got error: %v", err)
	}

	// Result type should be C y p = Nat
	if !alphaEq(ty, nat) {
		t.Errorf("expected result type Nat, got %v", ast.Sprint(ty))
	}
}

// TestTransport tests the transport function derivable from J
// transport : (A : Type) -> (P : A -> Type) -> (x y : A) -> Id A x y -> P x -> P y
// transport A P x y p px = J A (\z. \q. P z) px x y p
func TestTransport(t *testing.T) {
	globals := NewGlobalEnvWithPrimitives()
	checker := NewChecker(globals)
	context := idTestCtx()

	// For simplicity, we test transport with Nat and a constant predicate
	// P = \n. Nat (constant predicate)
	// transport Nat (\n. Nat) zero zero (refl Nat zero) (succ zero) = succ zero

	nat := ast.Global{Name: "Nat"}
	zero := ast.Global{Name: "zero"}
	succZero := ast.App{T: ast.Global{Name: "succ"}, U: zero}

	// Predicate P : Nat -> Type = \n. Nat
	predP := ast.Lam{
		Binder: "n",
		Ann:    nat,
		Body:   ast.Sort{U: 0}, // returns Type0
	}

	// Motive for J: \z. \q. P z = \z. \q. Nat (since P n = Nat for all n)
	// Note: This is simplified. The actual motive should be (\z. \q. P z)
	// For P = \n. Nat, this becomes \z. \q. Nat
	motive := ast.Lam{
		Binder: "z",
		Ann:    nat,
		Body: ast.Lam{
			Binder: "q",
			Ann:    ast.Id{A: nat, X: zero, Y: ast.Var{Ix: 0}},
			Body:   nat, // P z = Nat (constant)
		},
	}

	// transport implemented as J
	transportExpr := ast.J{
		A: nat,
		C: motive,
		D: succZero, // px : P x = Nat
		X: zero,
		Y: zero,
		P: ast.Refl{A: nat, X: zero},
	}

	// Should typecheck
	ty, err := checker.Synth(context, NoSpan(), transportExpr)
	if err != nil {
		t.Fatalf("expected transport to typecheck, got error: %v", err)
	}

	// Result type should be Nat (= P y)
	if !alphaEq(ty, nat) {
		t.Errorf("expected result type Nat, got %v", ast.Sprint(ty))
	}

	// Should evaluate to succZero
	result := eval.EvalNBE(transportExpr)
	if !alphaEq(result, succZero) {
		t.Errorf("expected transport to evaluate to succ zero, got %v", ast.Sprint(result))
	}

	// Also verify the predicate typechecks
	_, err = checker.Synth(context, NoSpan(), predP)
	if err != nil {
		t.Fatalf("expected predicate to typecheck, got error: %v", err)
	}
}

// TestSymmetry tests that we can prove symmetry: Id A x y -> Id A y x
func TestSymmetry(t *testing.T) {
	globals := NewGlobalEnvWithPrimitives()
	checker := NewChecker(globals)
	context := idTestCtx()

	nat := ast.Global{Name: "Nat"}
	zero := ast.Global{Name: "zero"}

	// sym x y p = J A (\z. \q. Id A z x) (refl A x) x y p
	// Motive: \z. \q. Id A z x (note: x is free, this proves Id A y x from Id A x y)

	// For sym refl: sym A x x (refl A x) should give refl A x
	// Motive: \z. \q. Id Nat z zero
	motive := ast.Lam{
		Binder: "z",
		Ann:    nat,
		Body: ast.Lam{
			Binder: "q",
			Ann:    ast.Id{A: nat, X: zero, Y: ast.Var{Ix: 0}},
			Body:   ast.Id{A: nat, X: ast.Var{Ix: 1}, Y: zero}, // Id Nat z zero
		},
	}

	symExpr := ast.J{
		A: nat,
		C: motive,
		D: ast.Refl{A: nat, X: zero}, // d : C x (refl A x) = Id Nat zero zero
		X: zero,
		Y: zero,
		P: ast.Refl{A: nat, X: zero},
	}

	ty, err := checker.Synth(context, NoSpan(), symExpr)
	if err != nil {
		t.Fatalf("expected symmetry proof to typecheck, got error: %v", err)
	}

	// Result should normalize to Id Nat zero zero
	// The raw type is ((\z.\q. Id Nat z zero) zero (refl Nat zero))
	// which normalizes to Id Nat zero zero
	expectedTy := ast.Id{A: nat, X: zero, Y: zero}
	if !alphaEq(ty, expectedTy) {
		t.Errorf("expected Id Nat zero zero, got %v", ast.Sprint(eval.EvalNBE(ty)))
	}

	// Should evaluate to refl
	result := eval.EvalNBE(symExpr)
	if _, ok := result.(ast.Refl); !ok {
		t.Errorf("expected symmetry on refl to give refl, got %v", ast.Sprint(result))
	}
}

// TestIdTypeMismatch tests that mismatched types are rejected
func TestIdTypeMismatch(t *testing.T) {
	globals := NewGlobalEnvWithPrimitives()
	checker := NewChecker(globals)
	context := idTestCtx()

	// Id Nat true zero should fail (true : Bool, not Nat)
	idType := ast.Id{
		A: ast.Global{Name: "Nat"},
		X: ast.Global{Name: "true"}, // wrong type!
		Y: ast.Global{Name: "zero"},
	}

	_, err := checker.Synth(context, NoSpan(), idType)
	if err == nil {
		t.Error("expected type error for Id Nat true zero")
	}
}

// TestReflTypeMismatch tests that refl with wrong type is rejected
func TestReflTypeMismatch(t *testing.T) {
	globals := NewGlobalEnvWithPrimitives()
	checker := NewChecker(globals)
	context := idTestCtx()

	// refl Nat true should fail (true : Bool, not Nat)
	refl := ast.Refl{
		A: ast.Global{Name: "Nat"},
		X: ast.Global{Name: "true"}, // wrong type!
	}

	_, err := checker.Synth(context, NoSpan(), refl)
	if err == nil {
		t.Error("expected type error for refl Nat true")
	}
}

// alphaEq is a helper for comparing terms (using the internal AlphaEq)
func alphaEq(a, b ast.Term) bool {
	// Use NbE to normalize both terms first
	normA := eval.EvalNBE(a)
	normB := eval.EvalNBE(b)

	// Then compare structurally
	return structuralEq(normA, normB)
}

// structuralEq compares terms structurally
func structuralEq(a, b ast.Term) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch a := a.(type) {
	case ast.Sort:
		if bb, ok := b.(ast.Sort); ok {
			return a.U == bb.U
		}
	case ast.Var:
		if bb, ok := b.(ast.Var); ok {
			return a.Ix == bb.Ix
		}
	case ast.Global:
		if bb, ok := b.(ast.Global); ok {
			return a.Name == bb.Name
		}
	case ast.Pi:
		if bb, ok := b.(ast.Pi); ok {
			return structuralEq(a.A, bb.A) && structuralEq(a.B, bb.B)
		}
	case ast.Lam:
		if bb, ok := b.(ast.Lam); ok {
			return structuralEq(a.Body, bb.Body)
		}
	case ast.App:
		if bb, ok := b.(ast.App); ok {
			return structuralEq(a.T, bb.T) && structuralEq(a.U, bb.U)
		}
	case ast.Sigma:
		if bb, ok := b.(ast.Sigma); ok {
			return structuralEq(a.A, bb.A) && structuralEq(a.B, bb.B)
		}
	case ast.Pair:
		if bb, ok := b.(ast.Pair); ok {
			return structuralEq(a.Fst, bb.Fst) && structuralEq(a.Snd, bb.Snd)
		}
	case ast.Fst:
		if bb, ok := b.(ast.Fst); ok {
			return structuralEq(a.P, bb.P)
		}
	case ast.Snd:
		if bb, ok := b.(ast.Snd); ok {
			return structuralEq(a.P, bb.P)
		}
	case ast.Id:
		if bb, ok := b.(ast.Id); ok {
			return structuralEq(a.A, bb.A) && structuralEq(a.X, bb.X) && structuralEq(a.Y, bb.Y)
		}
	case ast.Refl:
		if bb, ok := b.(ast.Refl); ok {
			return structuralEq(a.A, bb.A) && structuralEq(a.X, bb.X)
		}
	case ast.J:
		if bb, ok := b.(ast.J); ok {
			return structuralEq(a.A, bb.A) && structuralEq(a.C, bb.C) && structuralEq(a.D, bb.D) &&
				structuralEq(a.X, bb.X) && structuralEq(a.Y, bb.Y) && structuralEq(a.P, bb.P)
		}
	}
	return false
}
