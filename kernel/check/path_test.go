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

// --- Face Formula Tests ---

// TestFacePrinting verifies face formulas print correctly.
func TestFacePrinting(t *testing.T) {
	tests := []struct {
		term     ast.Term
		expected string
	}{
		{ast.FaceTop{}, "⊤"},
		{ast.FaceBot{}, "⊥"},
		{ast.FaceEq{IVar: 0, IsOne: false}, "(i{0} = 0)"},
		{ast.FaceEq{IVar: 0, IsOne: true}, "(i{0} = 1)"},
		{ast.FaceAnd{Left: ast.FaceEq{IVar: 0, IsOne: false}, Right: ast.FaceEq{IVar: 1, IsOne: true}},
			"((i{0} = 0) ∧ (i{1} = 1))"},
		{ast.FaceOr{Left: ast.FaceEq{IVar: 0, IsOne: false}, Right: ast.FaceEq{IVar: 0, IsOne: true}},
			"((i{0} = 0) ∨ (i{0} = 1))"},
	}

	for _, tt := range tests {
		got := ast.Sprint(tt.term)
		if got != tt.expected {
			t.Errorf("Sprint(%T) = %q, want %q", tt.term, got, tt.expected)
		}
	}
}

// TestFaceEval verifies face formula evaluation.
func TestFaceEval(t *testing.T) {
	// Test face simplification: (i=0) ∧ (i=1) = ⊥
	faceAnd := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceEq{IVar: 0, IsOne: true},
	}
	result := eval.EvalCubical(nil, eval.EmptyIEnv(), faceAnd)
	if _, ok := result.(eval.VFaceBot); !ok {
		t.Errorf("Expected VFaceBot for contradictory face, got %T", result)
	}

	// Test face simplification: (i=0) ∨ (i=1) = ⊤
	faceOr := ast.FaceOr{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceEq{IVar: 0, IsOne: true},
	}
	result = eval.EvalCubical(nil, eval.EmptyIEnv(), faceOr)
	if _, ok := result.(eval.VFaceTop); !ok {
		t.Errorf("Expected VFaceTop for tautological face, got %T", result)
	}

	// Test ⊤ ∧ φ = φ
	andTop := ast.FaceAnd{
		Left:  ast.FaceTop{},
		Right: ast.FaceEq{IVar: 0, IsOne: true},
	}
	result = eval.EvalCubical(nil, eval.EmptyIEnv(), andTop)
	if eq, ok := result.(eval.VFaceEq); !ok || !eq.IsOne {
		t.Errorf("Expected VFaceEq{IsOne: true} for ⊤ ∧ (i=1), got %T %v", result, result)
	}

	// Test ⊥ ∨ φ = φ
	orBot := ast.FaceOr{
		Left:  ast.FaceBot{},
		Right: ast.FaceEq{IVar: 0, IsOne: false},
	}
	result = eval.EvalCubical(nil, eval.EmptyIEnv(), orBot)
	if eq, ok := result.(eval.VFaceEq); !ok || eq.IsOne {
		t.Errorf("Expected VFaceEq{IsOne: false} for ⊥ ∨ (i=0), got %T %v", result, result)
	}
}

// TestFaceSubst verifies face formula substitution.
func TestFaceSubst(t *testing.T) {
	// Substituting i0 for i in (i=0) should give ⊤
	face := ast.FaceEq{IVar: 0, IsOne: false}
	result := subst.ISubst(0, ast.I0{}, face)
	if _, ok := result.(ast.FaceTop); !ok {
		t.Errorf("Expected FaceTop for (i=0)[i0/i], got %v", ast.Sprint(result))
	}

	// Substituting i1 for i in (i=0) should give ⊥
	result = subst.ISubst(0, ast.I1{}, face)
	if _, ok := result.(ast.FaceBot); !ok {
		t.Errorf("Expected FaceBot for (i=0)[i1/i], got %v", ast.Sprint(result))
	}

	// Substituting i0 for i in (i=1) should give ⊥
	face = ast.FaceEq{IVar: 0, IsOne: true}
	result = subst.ISubst(0, ast.I0{}, face)
	if _, ok := result.(ast.FaceBot); !ok {
		t.Errorf("Expected FaceBot for (i=1)[i0/i], got %v", ast.Sprint(result))
	}

	// Substituting i1 for i in (i=1) should give ⊤
	result = subst.ISubst(0, ast.I1{}, face)
	if _, ok := result.(ast.FaceTop); !ok {
		t.Errorf("Expected FaceTop for (i=1)[i1/i], got %v", ast.Sprint(result))
	}
}

// --- Partial Type Tests ---

// TestPartialPrinting verifies partial types print correctly.
func TestPartialPrinting(t *testing.T) {
	partial := ast.Partial{
		Phi: ast.FaceEq{IVar: 0, IsOne: false},
		A:   ast.Sort{U: 0},
	}
	expected := "(Partial (i{0} = 0) Type0)"
	got := ast.Sprint(partial)
	if got != expected {
		t.Errorf("Sprint(Partial) = %q, want %q", got, expected)
	}

	// System printing
	sys := ast.System{
		Branches: []ast.SystemBranch{
			{Phi: ast.FaceEq{IVar: 0, IsOne: false}, Term: ast.Var{Ix: 0}},
			{Phi: ast.FaceEq{IVar: 0, IsOne: true}, Term: ast.Var{Ix: 1}},
		},
	}
	expected = "[(i{0} = 0) ↦ {0}, (i{0} = 1) ↦ {1}]"
	got = ast.Sprint(sys)
	if got != expected {
		t.Errorf("Sprint(System) = %q, want %q", got, expected)
	}
}

// TestPartialTypeCheck verifies Partial type formation.
func TestPartialTypeCheck(t *testing.T) {
	c := NewChecker(nil)

	// Partial ⊤ Type0 : Type1
	partial := ast.Partial{
		Phi: ast.FaceTop{},
		A:   ast.Sort{U: 0},
	}

	ty, err := c.Synth(nil, NoSpan(), partial)
	if err != nil {
		t.Fatalf("Partial type formation failed: %v", err)
	}

	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ast.Sprint(ty))
	}
}

// TestSystemTypeCheck verifies System type checking.
func TestSystemTypeCheck(t *testing.T) {
	c := NewChecker(nil)

	// Push an interval variable for the face constraints
	pop := c.PushIVar()
	defer pop()

	// [i=0 ↦ Type0, i=1 ↦ Type0] : Partial (i=0 ∨ i=1) Type1
	sys := ast.System{
		Branches: []ast.SystemBranch{
			{Phi: ast.FaceEq{IVar: 0, IsOne: false}, Term: ast.Sort{U: 0}},
			{Phi: ast.FaceEq{IVar: 0, IsOne: true}, Term: ast.Sort{U: 0}},
		},
	}

	ty, err := c.Synth(nil, NoSpan(), sys)
	if err != nil {
		t.Fatalf("System type checking failed: %v", err)
	}

	partial, ok := ty.(ast.Partial)
	if !ok {
		t.Fatalf("Expected Partial type, got %v", ast.Sprint(ty))
	}

	// The type should be Sort{U: 0} (Type0 : Type1)
	if sort, ok := partial.A.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected type Type1, got %v", ast.Sprint(partial.A))
	}
}

// --- Composition Tests ---

// TestCompPrinting verifies composition operations print correctly.
func TestCompPrinting(t *testing.T) {
	tests := []struct {
		term     ast.Term
		expected string
	}{
		{
			ast.Comp{
				IBinder: "i",
				A:       ast.Sort{U: 0},
				Phi:     ast.FaceTop{},
				Tube:    ast.Var{Ix: 0},
				Base:    ast.Var{Ix: 1},
			},
			"(comp^i Type0 [⊤ ↦ {0}] {1})",
		},
		{
			ast.HComp{
				A:    ast.Sort{U: 0},
				Phi:  ast.FaceBot{},
				Tube: ast.Var{Ix: 0},
				Base: ast.Var{Ix: 1},
			},
			"(hcomp Type0 [⊥ ↦ {0}] {1})",
		},
		{
			ast.Fill{
				IBinder: "i",
				A:       ast.Sort{U: 0},
				Phi:     ast.FaceEq{IVar: 0, IsOne: false},
				Tube:    ast.Var{Ix: 0},
				Base:    ast.Var{Ix: 1},
			},
			"(fill^i Type0 [(i{0} = 0) ↦ {0}] {1})",
		},
	}

	for _, tt := range tests {
		got := ast.Sprint(tt.term)
		if got != tt.expected {
			t.Errorf("Sprint(%T) = %q, want %q", tt.term, got, tt.expected)
		}
	}
}

// TestCompEvalFaceSatisfied verifies comp reduces when face is satisfied.
func TestCompEvalFaceSatisfied(t *testing.T) {
	// comp^i Type0 [⊤ ↦ Type1] Type2 should reduce to Type1
	comp := ast.Comp{
		IBinder: "i",
		A:       ast.Sort{U: 0},
		Phi:     ast.FaceTop{},
		Tube:    ast.Sort{U: 1},
		Base:    ast.Sort{U: 2},
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), comp)

	// Should reduce to Type1 (tube at i1)
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 1 {
		t.Errorf("Expected VSort{1} for comp with ⊤ face, got %T %v", result, result)
	}
}

// TestHCompEvalFaceEmpty verifies hcomp reduces to base when face is empty.
func TestHCompEvalFaceEmpty(t *testing.T) {
	// hcomp Type0 [⊥ ↦ _] Type1 should reduce to Type1
	hcomp := ast.HComp{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceBot{},
		Tube: ast.Sort{U: 2},
		Base: ast.Sort{U: 1},
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), hcomp)

	// Should reduce to Type1 (base, since face is empty)
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 1 {
		t.Errorf("Expected VSort{1} for hcomp with ⊥ face, got %T %v", result, result)
	}
}

// TestCompEvalStuck verifies comp stays stuck with unresolved face.
func TestCompEvalStuck(t *testing.T) {
	// comp^i Type0 [(i=0) ↦ Type1] Type2 should stay stuck
	comp := ast.Comp{
		IBinder: "i",
		A:       ast.Sort{U: 0},
		Phi:     ast.FaceEq{IVar: 0, IsOne: false},
		Tube:    ast.Sort{U: 1},
		Base:    ast.Sort{U: 2},
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), comp)

	// Should stay stuck as VComp
	if _, ok := result.(eval.VComp); !ok {
		t.Errorf("Expected VComp for comp with unresolved face, got %T", result)
	}
}

// TestHCompTypeCheck verifies hcomp type synthesis.
func TestHCompTypeCheck(t *testing.T) {
	c := NewChecker(nil)

	// Context with A : Type0
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	// hcomp Type0 [⊤ ↦ A] A : Type0
	// We use Var{0} which refers to A : Type0 in context
	hcomp := ast.HComp{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceTop{},
		Tube: ast.Var{Ix: 0}, // A : Type0 (shifted for interval binder)
		Base: ast.Var{Ix: 0}, // A : Type0
	}

	ty, err := c.Synth(ctx, NoSpan(), hcomp)
	if err != nil {
		t.Fatalf("HComp type checking failed: %v", err)
	}

	// Result should be Type0
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 0 {
		t.Errorf("Expected Type0, got %v", ast.Sprint(ty))
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

// --- Glue Type Tests ---

// TestGluePrinting verifies Glue types print correctly.
func TestGluePrinting(t *testing.T) {
	// Glue Type0 [⊤ ↦ (Type0, idEquiv)]
	glue := ast.Glue{
		A: ast.Sort{U: 0},
		System: []ast.GlueBranch{
			{
				Phi:   ast.FaceTop{},
				T:     ast.Sort{U: 0},
				Equiv: ast.Global{Name: "idEquiv"},
			},
		},
	}

	printed := ast.Sprint(glue)
	if printed != "(Glue Type0 [⊤ ↦ (Type0, idEquiv)])" {
		t.Errorf("Glue printed as: %s", printed)
	}

	// glue [⊤ ↦ x] y
	glueElem := ast.GlueElem{
		System: []ast.GlueElemBranch{
			{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}},
		},
		Base: ast.Var{Ix: 1},
	}

	printed = ast.Sprint(glueElem)
	if printed != "(glue [⊤ ↦ {0}] {1})" {
		t.Errorf("GlueElem printed as: %s", printed)
	}

	// unglue g
	unglue := ast.Unglue{
		Ty: ast.Glue{A: ast.Sort{U: 0}, System: nil},
		G:  ast.Var{Ix: 0},
	}

	printed = ast.Sprint(unglue)
	if printed != "(unglue {0})" {
		t.Errorf("Unglue printed as: %s", printed)
	}
}

// TestGlueEvalFaceSatisfied verifies Glue [⊤ ↦ (T, e)] = T.
func TestGlueEvalFaceSatisfied(t *testing.T) {
	// Glue Type0 [⊤ ↦ (Type1, e)] should evaluate to Type1
	glue := ast.Glue{
		A: ast.Sort{U: 0},
		System: []ast.GlueBranch{
			{
				Phi:   ast.FaceTop{},
				T:     ast.Sort{U: 1},
				Equiv: ast.Global{Name: "e"},
			},
		},
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), glue)

	// Should simplify to Type1
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 1 {
		t.Errorf("Expected VSort{Level: 1}, got %T", result)
	}
}

// TestGlueEvalFaceEmpty verifies Glue A [] = A.
func TestGlueEvalFaceEmpty(t *testing.T) {
	// Glue Type0 [] should evaluate to Type0
	glue := ast.Glue{
		A:      ast.Sort{U: 0},
		System: nil,
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), glue)

	// Should simplify to Type0
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 0 {
		t.Errorf("Expected VSort{Level: 0}, got %T", result)
	}
}

// TestGlueElemEvalFaceSatisfied verifies glue [⊤ ↦ t] a = t.
func TestGlueElemEvalFaceSatisfied(t *testing.T) {
	// glue [⊤ ↦ Type1] Type0 should evaluate to Type1
	glueElem := ast.GlueElem{
		System: []ast.GlueElemBranch{
			{Phi: ast.FaceTop{}, Term: ast.Sort{U: 1}},
		},
		Base: ast.Sort{U: 0},
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), glueElem)

	// Should simplify to Type1
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 1 {
		t.Errorf("Expected VSort{Level: 1}, got %T", result)
	}
}

// TestUnglueEvalGlueElem verifies unglue (glue [φ ↦ t] a) = a.
func TestUnglueEvalGlueElem(t *testing.T) {
	// unglue (glue [] Type0) should evaluate to Type0 (empty system)
	unglue := ast.Unglue{
		Ty: ast.Glue{A: ast.Sort{U: 0}, System: nil},
		G: ast.GlueElem{
			System: nil, // Empty system - glue element doesn't reduce
			Base:   ast.Sort{U: 0},
		},
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), unglue)

	// Should extract the base Type0
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 0 {
		t.Errorf("Expected VSort{Level: 0}, got %T", result)
	}
}

// TestUnglueEvalStuck verifies unglue stays stuck when input is not a GlueElem.
func TestUnglueEvalStuck(t *testing.T) {
	// unglue (glue [⊤ ↦ Type1] Type0)
	// The inner glue elem reduces to Type1 first, then unglue is stuck
	unglue := ast.Unglue{
		Ty: ast.Glue{A: ast.Sort{U: 0}, System: nil},
		G: ast.GlueElem{
			System: []ast.GlueElemBranch{
				{Phi: ast.FaceTop{}, Term: ast.Sort{U: 1}},
			},
			Base: ast.Sort{U: 0},
		},
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), unglue)

	// Inner glue elem reduces to Type1, so unglue sees VSort, not VGlueElem
	// Unglue then stays stuck as VUnglue
	if _, ok := result.(eval.VUnglue); !ok {
		t.Errorf("Expected VUnglue (stuck), got %T", result)
	}
}

// TestGlueTypeCheck verifies Glue type synthesis.
func TestGlueTypeCheck(t *testing.T) {
	c := NewChecker(nil)

	// Glue Type0 [] : Type1 (empty system)
	glue := ast.Glue{
		A:      ast.Sort{U: 0},
		System: nil,
	}

	ty, err := c.Synth(nil, NoSpan(), glue)
	if err != nil {
		t.Fatalf("Glue type checking failed: %v", err)
	}

	// Result should be Type1 (universe of Type0)
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ast.Sprint(ty))
	}
}

// TestGlueTypeCheckWithBranch verifies Glue with a face branch.
func TestGlueTypeCheckWithBranch(t *testing.T) {
	c := NewChecker(nil)

	// Context with e : Type0 (as placeholder for equivalence)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	// Glue Type0 [⊤ ↦ (Type0, e)] where e is a variable
	glue := ast.Glue{
		A: ast.Sort{U: 0},
		System: []ast.GlueBranch{
			{
				Phi:   ast.FaceTop{},
				T:     ast.Sort{U: 0},
				Equiv: ast.Var{Ix: 0}, // Use variable instead of global
			},
		},
	}

	ty, err := c.Synth(ctx, NoSpan(), glue)
	if err != nil {
		t.Fatalf("Glue type checking failed: %v", err)
	}

	// Result should be Type1 (universe of Type0)
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ast.Sprint(ty))
	}
}

// --- Univalence (UA) Tests ---

// TestUAPrinting verifies UA terms print correctly.
func TestUAPrinting(t *testing.T) {
	// ua A B e
	ua := ast.UA{
		A:     ast.Sort{U: 0},
		B:     ast.Sort{U: 0},
		Equiv: ast.Var{Ix: 0},
	}

	printed := ast.Sprint(ua)
	if printed != "(ua Type0 Type0 {0})" {
		t.Errorf("UA printed as: %s", printed)
	}

	// ua-β e a
	uaBeta := ast.UABeta{
		Equiv: ast.Var{Ix: 0},
		Arg:   ast.Var{Ix: 1},
	}

	printed = ast.Sprint(uaBeta)
	if printed != "(ua-β {0} {1})" {
		t.Errorf("UABeta printed as: %s", printed)
	}
}

// TestUAPathApplyI0 verifies (ua e) @ i0 = A.
func TestUAPathApplyI0(t *testing.T) {
	// (ua Type0 Type1 e) @ i0 should evaluate to Type0
	ua := ast.UA{
		A:     ast.Sort{U: 0},
		B:     ast.Sort{U: 1},
		Equiv: ast.Global{Name: "e"},
	}

	papp := ast.PathApp{
		P: ua,
		R: ast.I0{},
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), papp)

	// Should evaluate to Type0 (the A component)
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 0 {
		t.Errorf("Expected VSort{Level: 0}, got %T %v", result, result)
	}
}

// TestUAPathApplyI1 verifies (ua e) @ i1 = B.
func TestUAPathApplyI1(t *testing.T) {
	// (ua Type0 Type1 e) @ i1 should evaluate to Type1
	ua := ast.UA{
		A:     ast.Sort{U: 0},
		B:     ast.Sort{U: 1},
		Equiv: ast.Global{Name: "e"},
	}

	papp := ast.PathApp{
		P: ua,
		R: ast.I1{},
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), papp)

	// Should evaluate to Type1 (the B component)
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 1 {
		t.Errorf("Expected VSort{Level: 1}, got %T %v", result, result)
	}
}

// TestUAPathApplyIntermediate verifies (ua e) @ i gives Glue type.
func TestUAPathApplyIntermediate(t *testing.T) {
	// (ua Type0 Type1 e) @ i where i is a variable should give a Glue type
	ua := ast.UA{
		A:     ast.Sort{U: 0},
		B:     ast.Sort{U: 1},
		Equiv: ast.Global{Name: "e"},
	}

	papp := ast.PathApp{
		P: ua,
		R: ast.IVar{Ix: 0}, // interval variable
	}

	// Create an interval environment with one variable
	ienv := eval.EmptyIEnv().Extend(eval.VIVar{Level: 0})

	result := eval.EvalCubical(nil, ienv, papp)

	// Should stay as VGlue with the (i=0) face mapping to A
	if _, ok := result.(eval.VGlue); !ok {
		t.Errorf("Expected VGlue for intermediate interval, got %T %v", result, result)
	}
}

// TestUATypeCheck verifies UA type synthesis.
func TestUATypeCheck(t *testing.T) {
	c := NewChecker(nil)

	// Context with e : Type0 (as placeholder for equivalence)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	// ua Type0 Type0 e should have type Path Type1 Type0 Type0
	ua := ast.UA{
		A:     ast.Sort{U: 0},
		B:     ast.Sort{U: 0},
		Equiv: ast.Var{Ix: 0},
	}

	ty, err := c.Synth(ctx, NoSpan(), ua)
	if err != nil {
		t.Fatalf("UA type checking failed: %v", err)
	}

	// Result should be Path Type1 Type0 Type0
	path, ok := ty.(ast.Path)
	if !ok {
		t.Fatalf("Expected Path type, got %v", ast.Sprint(ty))
	}

	// The type family should be Type1
	if sort, ok := path.A.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1 as path type family, got %v", ast.Sprint(path.A))
	}

	// Endpoints should be Type0
	if sort, ok := path.X.(ast.Sort); !ok || sort.U != 0 {
		t.Errorf("Expected Type0 as left endpoint, got %v", ast.Sprint(path.X))
	}
	if sort, ok := path.Y.(ast.Sort); !ok || sort.U != 0 {
		t.Errorf("Expected Type0 as right endpoint, got %v", ast.Sprint(path.Y))
	}
}

// TestUABetaTypeCheck verifies UABeta type synthesis.
func TestUABetaTypeCheck(t *testing.T) {
	c := NewChecker(nil)

	// Context with e : Type0 (placeholder for equivalence), a : Type0
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}, ast.Sort{U: 0}})

	// ua-β e a
	uaBeta := ast.UABeta{
		Equiv: ast.Var{Ix: 1}, // e
		Arg:   ast.Var{Ix: 0}, // a
	}

	// Should synthesize without error (the return type is B in Equiv A B)
	_, err := c.Synth(ctx, NoSpan(), uaBeta)
	if err != nil {
		t.Fatalf("UABeta type checking failed: %v", err)
	}
}
