//go:build cubical

package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
	tyctx "github.com/watchthelight/HypergraphGo/kernel/ctx"
	"github.com/watchthelight/HypergraphGo/kernel/subst"
)

// TestPathTypeFormation verifies that Path A x y : Type type-checks.
func TestPathTypeFormation(t *testing.T) {
	c := NewChecker(nil)

	// Path Type0 x x should have type Type1 (since Type0 : Type1)
	pathType := ast.Path{
		A: ast.Sort{U: 0},
		X: ast.Var{Ix: 0}, // x
		Y: ast.Var{Ix: 0}, // x
	}

	// Under context with x : Type0
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})
	ty, err := c.Synth(ctx, NoSpan(), pathType)
	if err != nil {
		t.Fatalf("Path type formation failed: %v", err)
	}

	// Should synthesize to Type1 (A : Type1, so Path A x y : Type1)
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ast.Sprint(ty))
	}
}

// TestPathPTypeFormation verifies that PathP A x y : Type type-checks.
func TestPathPTypeFormation(t *testing.T) {
	c := NewChecker(nil)

	// PathP (λi. Type0) x x should have type Type1 where x : Type0
	// A is a constant type family: Type0 (which has type Type1)
	// x and y are endpoints of type A[i0/i] = Type0 and A[i1/i] = Type0
	pathPType := ast.PathP{
		A: ast.Sort{U: 0}, // Type family (constant): I → Type0
		X: ast.Var{Ix: 0}, // x : Type0
		Y: ast.Var{Ix: 0}, // x : Type0
	}

	// Context with x : Type0
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})
	ty, err := c.Synth(ctx, NoSpan(), pathPType)
	if err != nil {
		t.Fatalf("PathP type formation failed: %v", err)
	}

	// Should synthesize to Type1 (A : Type1, so PathP A x y : Type1)
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ast.Sprint(ty))
	}
}

// TestPathLamIntro verifies that <i> t synthesizes PathP type.
func TestPathLamIntro(t *testing.T) {
	c := NewChecker(nil)

	// <i> Type0 : PathP (λi. Type1) Type0 Type0
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Sort{U: 0}, // constant body
	}

	ty, err := c.Synth(nil, NoSpan(), plam)
	if err != nil {
		t.Fatalf("PathLam intro failed: %v", err)
	}

	// Should synthesize to PathP with endpoints being Type0
	pathp, ok := ty.(ast.PathP)
	if !ok {
		t.Fatalf("Expected PathP type, got %v", ast.Sprint(ty))
	}

	// Endpoints should be Type0
	if sort, ok := pathp.X.(ast.Sort); !ok || sort.U != 0 {
		t.Errorf("Expected left endpoint Type0, got %v", ast.Sprint(pathp.X))
	}
	if sort, ok := pathp.Y.(ast.Sort); !ok || sort.U != 0 {
		t.Errorf("Expected right endpoint Type0, got %v", ast.Sprint(pathp.Y))
	}
}

// TestPathAppBetaI0 verifies that (<i> t) @ i0 reduces to t[i0/i].
func TestPathAppBetaI0(t *testing.T) {
	// <i> x where x is a term variable
	// When applied to i0, should reduce to x (constant)
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Var{Ix: 0}, // body is just a term variable
	}
	papp := ast.PathApp{
		P: plam,
		R: ast.I0{},
	}

	// Evaluate using cubical NbE
	env := &eval.Env{Bindings: []eval.Value{eval.VGlobal{Name: "x"}}}
	result := eval.EvalCubical(env, eval.EmptyIEnv(), papp)

	// Should reduce to the value of variable 0
	if g, ok := result.(eval.VGlobal); !ok || g.Name != "x" {
		t.Errorf("Expected VGlobal{x}, got %T", result)
	}
}

// TestPathAppBetaI1 verifies that (<i> t) @ i1 reduces to t[i1/i].
func TestPathAppBetaI1(t *testing.T) {
	// <i> x where x is a term variable
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Var{Ix: 0},
	}
	papp := ast.PathApp{
		P: plam,
		R: ast.I1{},
	}

	env := &eval.Env{Bindings: []eval.Value{eval.VGlobal{Name: "x"}}}
	result := eval.EvalCubical(env, eval.EmptyIEnv(), papp)

	// Should reduce to the value of variable 0
	if g, ok := result.(eval.VGlobal); !ok || g.Name != "x" {
		t.Errorf("Expected VGlobal{x}, got %T", result)
	}
}

// TestTransportConstant verifies that transport (λi. A) e --> e when A is constant.
func TestTransportConstant(t *testing.T) {
	// transport (λi. Type0) Type1 should reduce to Type1
	tr := ast.Transport{
		A: ast.Sort{U: 0}, // constant type family
		E: ast.Sort{U: 1},
	}

	result := eval.EvalCubical(nil, nil, tr)

	// Should reduce to Type1 (since A is constant)
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 1 {
		t.Errorf("Expected VSort{1}, got %T %v", result, result)
	}
}

// TestTransportTyping verifies transport A e : A[i1/i].
func TestTransportTyping(t *testing.T) {
	c := NewChecker(nil)

	// transport (λi. Type0) x : Type0 where x : Type0
	// A = Type0 (constant), e = x : A[i0/i] = Type0
	// Result type is A[i1/i] = Type0
	tr := ast.Transport{
		A: ast.Sort{U: 0},
		E: ast.Var{Ix: 0}, // x : Type0
	}

	// Context with x : Type0
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})
	ty, err := c.Synth(ctx, NoSpan(), tr)
	if err != nil {
		t.Fatalf("Transport typing failed: %v", err)
	}

	// Should have type Type0 (A[i1/i] = Type0)
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 0 {
		t.Errorf("Expected Type0, got %v", ast.Sprint(ty))
	}
}

// TestReflAsPath verifies that <i> x : Path A x x type-checks.
func TestReflAsPath(t *testing.T) {
	c := NewChecker(nil)

	// <i> x should check against Path Type0 x x
	// where x is a term variable of type Type0
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Var{Ix: 0}, // x (constant in i)
	}

	expectedTy := ast.Path{
		A: ast.Sort{U: 0},
		X: ast.Var{Ix: 0},
		Y: ast.Var{Ix: 0},
	}

	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})
	err := c.Check(ctx, NoSpan(), plam, expectedTy)
	if err != nil {
		t.Fatalf("Refl as path failed: %v", err)
	}
}

// TestISubst verifies interval substitution works correctly.
func TestISubst(t *testing.T) {
	// ISubst(0, i0, IVar{0}) should give i0
	result := subst.ISubst(0, ast.I0{}, ast.IVar{Ix: 0})
	if _, ok := result.(ast.I0); !ok {
		t.Errorf("Expected I0, got %v", ast.Sprint(result))
	}

	// ISubst(0, i1, IVar{0}) should give i1
	result = subst.ISubst(0, ast.I1{}, ast.IVar{Ix: 0})
	if _, ok := result.(ast.I1); !ok {
		t.Errorf("Expected I1, got %v", ast.Sprint(result))
	}

	// ISubst(0, i0, IVar{1}) should give IVar{0} (shifted down)
	result = subst.ISubst(0, ast.I0{}, ast.IVar{Ix: 1})
	if ivar, ok := result.(ast.IVar); !ok || ivar.Ix != 0 {
		t.Errorf("Expected IVar{0}, got %v", ast.Sprint(result))
	}
}

// TestIShift verifies interval shifting works correctly.
func TestIShift(t *testing.T) {
	// IShift(1, 0, IVar{0}) should give IVar{1}
	result := subst.IShift(1, 0, ast.IVar{Ix: 0})
	if ivar, ok := result.(ast.IVar); !ok || ivar.Ix != 1 {
		t.Errorf("Expected IVar{1}, got %v", ast.Sprint(result))
	}

	// IShift(1, 1, IVar{0}) should give IVar{0} (below cutoff)
	result = subst.IShift(1, 1, ast.IVar{Ix: 0})
	if ivar, ok := result.(ast.IVar); !ok || ivar.Ix != 0 {
		t.Errorf("Expected IVar{0}, got %v", ast.Sprint(result))
	}
}

// TestCubicalPrinting verifies cubical terms print correctly.
func TestCubicalPrinting(t *testing.T) {
	tests := []struct {
		term     ast.Term
		expected string
	}{
		{ast.Interval{}, "I"},
		{ast.I0{}, "i0"},
		{ast.I1{}, "i1"},
		{ast.IVar{Ix: 0}, "i{0}"},
		{ast.Path{A: ast.Sort{U: 0}, X: ast.I0{}, Y: ast.I1{}}, "(Path Type0 i0 i1)"},
		{ast.PathLam{Binder: "i", Body: ast.Var{Ix: 0}}, "(<i> {0})"},
		{ast.PathApp{P: ast.Var{Ix: 0}, R: ast.I0{}}, "({0} @ i0)"},
		{ast.Transport{A: ast.Sort{U: 0}, E: ast.Var{Ix: 0}}, "(transport Type0 {0})"},
	}

	for _, tt := range tests {
		got := ast.Sprint(tt.term)
		if got != tt.expected {
			t.Errorf("Sprint(%T) = %q, want %q", tt.term, got, tt.expected)
		}
	}
}

// TestAlphaEqCubical verifies alpha equality for cubical terms.
func TestAlphaEqCubical(t *testing.T) {
	tests := []struct {
		a, b   ast.Term
		expect bool
	}{
		{ast.I0{}, ast.I0{}, true},
		{ast.I0{}, ast.I1{}, false},
		{ast.IVar{Ix: 0}, ast.IVar{Ix: 0}, true},
		{ast.IVar{Ix: 0}, ast.IVar{Ix: 1}, false},
		{ast.Interval{}, ast.Interval{}, true},
		{ast.Path{A: ast.Sort{U: 0}, X: ast.I0{}, Y: ast.I1{}},
			ast.Path{A: ast.Sort{U: 0}, X: ast.I0{}, Y: ast.I1{}}, true},
		{ast.PathLam{Body: ast.Var{Ix: 0}},
			ast.PathLam{Body: ast.Var{Ix: 0}}, true},
	}

	for _, tt := range tests {
		// Use Sprint comparison as a proxy for alpha equality
		got := ast.Sprint(tt.a) == ast.Sprint(tt.b)
		if got != tt.expect {
			t.Errorf("AlphaEq(%v, %v) = %v, want %v",
				ast.Sprint(tt.a), ast.Sprint(tt.b), got, tt.expect)
		}
	}
}

// Helper to create a test context with given types (topmost binding first).
func makeTestContext(types []ast.Term) *tyctx.Ctx {
	ctx := &tyctx.Ctx{}
	for i := len(types) - 1; i >= 0; i-- {
		ctx.Extend("_", types[i])
	}
	return ctx
}
