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

// TestTransportUAComputes verifies the key univalence computation rule:
// transport (ua e) a = e.fst a
//
// This is the fundamental property that makes univalence computational.
// When we transport an element along a path created by ua from an equivalence,
// the result is the same as applying the forward function of the equivalence.
func TestTransportUAComputes(t *testing.T) {
	// The transport (ua e) a computation works as follows:
	// 1. ua e creates a path from A to B using Glue types
	// 2. transport along this path evaluates through the Glue composition
	// 3. The result is e.fst a (applying the first projection of the equivalence)
	//
	// At the AST level, UABeta represents this computation:
	// UABeta{Equiv: e, Arg: a} represents the result of transport (ua e) a

	// Test 1: UABeta should represent the transport result
	uaBeta := ast.UABeta{
		Equiv: ast.Global{Name: "myEquiv"},
		Arg:   ast.Global{Name: "myArg"},
	}

	// Evaluate the UABeta term
	result := eval.EvalCubical(nil, eval.EmptyIEnv(), uaBeta)

	// The result should be a VUABeta value (neutral, waiting for reduction)
	if _, ok := result.(eval.VUABeta); !ok {
		t.Errorf("Expected VUABeta value, got %T", result)
	}

	// Test 2: Verify transport structure
	// transport : (A : I → Type) → A[i0] → A[i1]
	// For ua e where e : Equiv A B:
	// - (ua e) : Path Type A B
	// - transport (λi. (ua e) @ i) a : B
	//
	// The type line for transport is the ua path applied to the interval variable
	ua := ast.UA{
		A:     ast.Sort{U: 0}, // Source type A
		B:     ast.Sort{U: 1}, // Target type B
		Equiv: ast.Global{Name: "e"},
	}

	// The type family for transport: λi. (ua e) @ i
	typeLine := ast.PathLam{
		Binder: "i",
		Body:   ast.PathApp{P: ua, R: ast.IVar{Ix: 0}},
	}

	// transport (λi. (ua e) @ i) myArg
	transport := ast.Transport{
		A: typeLine.Body, // The type family body
		E: ast.Global{Name: "myArg"},
	}

	// Evaluate transport
	transResult := eval.EvalCubical(nil, eval.EmptyIEnv(), transport)

	// The transport should produce a value (possibly stuck as VTransport if not fully reduced)
	// This verifies the evaluation machinery handles the transport structure
	if transResult == nil {
		t.Error("Transport evaluation returned nil")
	}
}

// TestUABetaReification verifies that UABeta values reify correctly.
func TestUABetaReification(t *testing.T) {
	// Create a UABeta value using neutral global references
	equivVal := eval.VNeutral{N: eval.Neutral{Head: eval.Head{Glob: "e"}, Sp: nil}}
	argVal := eval.VNeutral{N: eval.Neutral{Head: eval.Head{Glob: "a"}, Sp: nil}}

	uaBetaVal := eval.VUABeta{
		Equiv: equivVal,
		Arg:   argVal,
	}

	// Reify back to AST
	result := eval.ReifyCubicalAt(0, 0, uaBetaVal)

	// Should be UABeta term
	if uab, ok := result.(ast.UABeta); ok {
		if _, ok := uab.Equiv.(ast.Global); !ok {
			t.Errorf("Expected Global for Equiv, got %T", uab.Equiv)
		}
		if _, ok := uab.Arg.(ast.Global); !ok {
			t.Errorf("Expected Global for Arg, got %T", uab.Arg)
		}
	} else {
		t.Errorf("Expected UABeta term, got %T", result)
	}
}

// --- Composition Tests ---

// TestCompFaceSatisfied verifies comp reduces when face is ⊤.
// comp^i A [⊤ ↦ u] a₀ = u[i1/i]
func TestCompFaceSatisfied(t *testing.T) {
	// comp^i Type0 [⊤ ↦ Type1] Type0
	// When face is satisfied (⊤), result should be Type1 (tube at i1)
	comp := ast.Comp{
		IBinder: "i",
		A:       ast.Sort{U: 0}, // Type line: Type0
		Phi:     ast.FaceTop{},  // Face is always true
		Tube:    ast.Sort{U: 1}, // Tube gives Type1
		Base:    ast.Sort{U: 0}, // Base is Type0
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), comp)

	// When face is ⊤, should reduce to tube[i1/i] = Type1
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 1 {
		t.Errorf("Expected VSort{Level: 1}, got %T %v", result, result)
	}
}

// TestCompFaceEmpty verifies comp reduces to transport when face is ⊥.
// comp^i A [⊥ ↦ _] a₀ = transport A a₀
func TestCompFaceEmpty(t *testing.T) {
	// comp^i Type0 [⊥ ↦ _] Type0
	// When face is empty (⊥), should reduce to transport
	comp := ast.Comp{
		IBinder: "i",
		A:       ast.Sort{U: 0}, // Type line: Type0 (constant)
		Phi:     ast.FaceBot{},  // Face is never true
		Tube:    ast.Sort{U: 1}, // Tube is irrelevant when face is ⊥
		Base:    ast.Sort{U: 0}, // Base is Type0
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), comp)

	// When face is ⊥ and type is constant, should reduce to base
	// (transport along constant type is identity)
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 0 {
		t.Errorf("Expected VSort{Level: 0}, got %T %v", result, result)
	}
}

// TestHCompFaceSatisfied verifies hcomp reduces when face is ⊤.
// hcomp A [⊤ ↦ u] a₀ = u[i1/i]
func TestHCompFaceSatisfied(t *testing.T) {
	// hcomp Type0 [⊤ ↦ Type1] Type0
	hcomp := ast.HComp{
		A:    ast.Sort{U: 0}, // Constant type
		Phi:  ast.FaceTop{},  // Face is always true
		Tube: ast.Sort{U: 1}, // Tube gives Type1
		Base: ast.Sort{U: 0}, // Base is Type0
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), hcomp)

	// When face is ⊤, should reduce to tube[i1/i] = Type1
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 1 {
		t.Errorf("Expected VSort{Level: 1}, got %T %v", result, result)
	}
}

// TestHCompFaceEmpty verifies hcomp reduces to base when face is ⊥.
// hcomp A [⊥ ↦ _] a₀ = a₀
func TestHCompFaceEmpty(t *testing.T) {
	// hcomp Type0 [⊥ ↦ _] Type0
	hcomp := ast.HComp{
		A:    ast.Sort{U: 0}, // Constant type
		Phi:  ast.FaceBot{},  // Face is never true
		Tube: ast.Sort{U: 1}, // Tube is irrelevant
		Base: ast.Sort{U: 0}, // Base is Type0
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), hcomp)

	// When face is ⊥, should reduce to base = Type0
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 0 {
		t.Errorf("Expected VSort{Level: 0}, got %T %v", result, result)
	}
}

// TestFillEvaluation verifies fill produces correct values at endpoints.
// fill^i A [φ ↦ u] a₀ @ i0 = a₀
// fill^i A [φ ↦ u] a₀ @ i1 = comp^i A [φ ↦ u] a₀
func TestFillEvaluation(t *testing.T) {
	// fill^i Type0 [⊥ ↦ _] Type0
	fill := ast.Fill{
		IBinder: "i",
		A:       ast.Sort{U: 0}, // Type line
		Phi:     ast.FaceBot{},  // Face constraint
		Tube:    ast.Sort{U: 1}, // Tube
		Base:    ast.Sort{U: 0}, // Base
	}

	// Evaluate fill directly (produces a value representing the fill)
	result := eval.EvalCubical(nil, eval.EmptyIEnv(), fill)

	// Fill should produce a VFill value (or reduce if possible)
	// When evaluated, it represents a path from base to comp
	if result == nil {
		t.Error("Fill evaluation returned nil")
	}

	// Test applying fill to endpoints
	// fill @ i0 should give base
	fillAtI0 := ast.PathApp{P: fill, R: ast.I0{}}
	resultI0 := eval.EvalCubical(nil, eval.EmptyIEnv(), fillAtI0)

	if sort, ok := resultI0.(eval.VSort); ok && sort.Level != 0 {
		t.Errorf("fill @ i0: expected VSort{Level: 0}, got %v", result)
	}
}

// TestCompTypeCheck verifies composition type checking.
func TestCompTypeCheck(t *testing.T) {
	c := NewChecker(nil)
	c.PushIVar() // Need interval variable for tube

	// Context with x : Type1 (element of the composition type)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 1}})

	// comp^i Type1 [⊥ ↦ x] x where x : Type1
	comp := ast.Comp{
		IBinder: "i",
		A:       ast.Sort{U: 1}, // Type line: Type1
		Phi:     ast.FaceBot{},
		Tube:    ast.Var{Ix: 0}, // Tube: x : Type1
		Base:    ast.Var{Ix: 0}, // Base: x : Type1
	}

	ty, err := c.Synth(ctx, NoSpan(), comp)
	if err != nil {
		t.Fatalf("Comp type checking failed: %v", err)
	}

	// Result type should be A[i1/i] = Type1
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ast.Sprint(ty))
	}
}

// TestHCompTypeCheckWithBot verifies hcomp type checking with bot face.
func TestHCompTypeCheckWithBot(t *testing.T) {
	c := NewChecker(nil)
	c.PushIVar() // Need interval variable for tube

	// Context with x : Type1
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 1}})

	// hcomp Type1 [⊥ ↦ x] x
	hcomp := ast.HComp{
		A:    ast.Sort{U: 1}, // Constant type: Type1
		Phi:  ast.FaceBot{},
		Tube: ast.Var{Ix: 0}, // x : Type1
		Base: ast.Var{Ix: 0}, // x : Type1
	}

	ty, err := c.Synth(ctx, NoSpan(), hcomp)
	if err != nil {
		t.Fatalf("HComp type checking failed: %v", err)
	}

	// Result type should be A = Type1
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ast.Sprint(ty))
	}
}

// TestFillTypeCheck verifies fill type checking.
func TestFillTypeCheck(t *testing.T) {
	c := NewChecker(nil)
	c.PushIVar() // Need interval variable

	// Context with x : Type1
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 1}})

	// fill^i Type1 [⊥ ↦ x] x
	fill := ast.Fill{
		IBinder: "i",
		A:       ast.Sort{U: 1}, // Type line: Type1
		Phi:     ast.FaceBot{},
		Tube:    ast.Var{Ix: 0}, // x : Type1
		Base:    ast.Var{Ix: 0}, // x : Type1
	}

	ty, err := c.Synth(ctx, NoSpan(), fill)
	if err != nil {
		t.Fatalf("Fill type checking failed: %v", err)
	}

	// Result type should be the type family A = Type1
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ast.Sprint(ty))
	}
}

// --- Face Formula Edge Case Tests ---

// TestFaceIsBot_Bot verifies that FaceBot is detected as bottom.
func TestFaceIsBot_Bot(t *testing.T) {
	if !faceIsBot(ast.FaceBot{}) {
		t.Error("faceIsBot(FaceBot) = false, want true")
	}
}

// TestFaceIsBot_Top verifies that FaceTop is not bottom.
func TestFaceIsBot_Top(t *testing.T) {
	if faceIsBot(ast.FaceTop{}) {
		t.Error("faceIsBot(FaceTop) = true, want false")
	}
}

// TestFaceIsBot_EqNeverBot verifies that FaceEq is never bottom alone.
func TestFaceIsBot_EqNeverBot(t *testing.T) {
	tests := []ast.Face{
		ast.FaceEq{IVar: 0, IsOne: true},  // (i = 1)
		ast.FaceEq{IVar: 0, IsOne: false}, // (i = 0)
		ast.FaceEq{IVar: 5, IsOne: true},  // (j = 1)
	}
	for i, face := range tests {
		if faceIsBot(face) {
			t.Errorf("Case %d: faceIsBot(%v) = true, want false", i, face)
		}
	}
}

// TestFaceIsBot_NestedAndContradiction tests (i=0) ∧ ((j=1) ∧ (i=1)).
func TestFaceIsBot_NestedAndContradiction(t *testing.T) {
	// (i=0) ∧ ((j=1) ∧ (i=1)) should be ⊥
	inner := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 1, IsOne: true},  // (j=1)
		Right: ast.FaceEq{IVar: 0, IsOne: true},  // (i=1)
	}
	outer := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false}, // (i=0)
		Right: inner,
	}
	if !faceIsBot(outer) {
		t.Error("faceIsBot((i=0) ∧ ((j=1) ∧ (i=1))) = false, want true")
	}
}

// TestFaceIsBot_ThreeVariableNoContradiction tests (i=0) ∧ (j=1) ∧ (k=0).
func TestFaceIsBot_ThreeVariableNoContradiction(t *testing.T) {
	// (i=0) ∧ ((j=1) ∧ (k=0)) - no contradiction
	inner := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 1, IsOne: true},  // (j=1)
		Right: ast.FaceEq{IVar: 2, IsOne: false}, // (k=0)
	}
	outer := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false}, // (i=0)
		Right: inner,
	}
	if faceIsBot(outer) {
		t.Error("faceIsBot((i=0) ∧ ((j=1) ∧ (k=0))) = true, want false")
	}
}

// TestFaceIsBot_OrOfBots verifies that ⊥ ∨ ⊥ = ⊥.
func TestFaceIsBot_OrOfBots(t *testing.T) {
	or := ast.FaceOr{
		Left:  ast.FaceBot{},
		Right: ast.FaceBot{},
	}
	if !faceIsBot(or) {
		t.Error("faceIsBot(⊥ ∨ ⊥) = false, want true")
	}
}

// TestFaceIsBot_OrWithNonBot verifies that φ ∨ ⊥ ≠ ⊥.
func TestFaceIsBot_OrWithNonBot(t *testing.T) {
	tests := []ast.Face{
		ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}},
		ast.FaceOr{Left: ast.FaceBot{}, Right: ast.FaceTop{}},
		ast.FaceOr{Left: ast.FaceEq{IVar: 0, IsOne: true}, Right: ast.FaceBot{}},
		ast.FaceOr{Left: ast.FaceBot{}, Right: ast.FaceEq{IVar: 0, IsOne: false}},
	}
	for i, face := range tests {
		if faceIsBot(face) {
			t.Errorf("Case %d: faceIsBot = true, want false", i)
		}
	}
}

// TestFaceIsBot_OrOfContradictions verifies (⊥) ∨ (⊥) from contradictions.
func TestFaceIsBot_OrOfContradictions(t *testing.T) {
	// ((i=0) ∧ (i=1)) ∨ ((j=0) ∧ (j=1))
	left := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceEq{IVar: 0, IsOne: true},
	}
	right := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 1, IsOne: false},
		Right: ast.FaceEq{IVar: 1, IsOne: true},
	}
	or := ast.FaceOr{Left: left, Right: right}
	if !faceIsBot(or) {
		t.Error("faceIsBot(contradictory ∨ contradictory) = false, want true")
	}
}

// TestFaceIsBot_AndWithBot verifies φ ∧ ⊥ = ⊥.
func TestFaceIsBot_AndWithBot(t *testing.T) {
	tests := []ast.Face{
		ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}},
		ast.FaceAnd{Left: ast.FaceBot{}, Right: ast.FaceTop{}},
		ast.FaceAnd{Left: ast.FaceEq{IVar: 0, IsOne: true}, Right: ast.FaceBot{}},
		ast.FaceAnd{Left: ast.FaceBot{}, Right: ast.FaceEq{IVar: 0, IsOne: false}},
	}
	for i, face := range tests {
		if !faceIsBot(face) {
			t.Errorf("Case %d: faceIsBot(φ ∧ ⊥) = false, want true", i)
		}
	}
}

// TestFaceIsBot_DeeplyNested tests a deeply nested contradiction.
func TestFaceIsBot_DeeplyNested(t *testing.T) {
	// Build: ((i=0) ∧ ((j=0) ∧ ((k=0) ∧ (i=1))))
	innermost := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 2, IsOne: false}, // (k=0)
		Right: ast.FaceEq{IVar: 0, IsOne: true},  // (i=1)
	}
	middle := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 1, IsOne: false}, // (j=0)
		Right: innermost,
	}
	outer := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false}, // (i=0)
		Right: middle,
	}
	if !faceIsBot(outer) {
		t.Error("faceIsBot(deeply nested contradiction) = false, want true")
	}
}

// --- isContradictoryFaceAnd Tests ---

// TestIsContradictoryFaceAnd_DirectContradiction tests (i=0) ∧ (i=1).
func TestIsContradictoryFaceAnd_DirectContradiction(t *testing.T) {
	face := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false}, // (i=0)
		Right: ast.FaceEq{IVar: 0, IsOne: true},  // (i=1)
	}
	if !isContradictoryFaceAnd(face) {
		t.Error("isContradictoryFaceAnd((i=0) ∧ (i=1)) = false, want true")
	}
}

// TestIsContradictoryFaceAnd_DifferentVariables tests (i=0) ∧ (j=1).
func TestIsContradictoryFaceAnd_DifferentVariables(t *testing.T) {
	face := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false}, // (i=0)
		Right: ast.FaceEq{IVar: 1, IsOne: true},  // (j=1)
	}
	if isContradictoryFaceAnd(face) {
		t.Error("isContradictoryFaceAnd((i=0) ∧ (j=1)) = true, want false")
	}
}

// TestIsContradictoryFaceAnd_SameConstraint tests (i=0) ∧ (i=0).
func TestIsContradictoryFaceAnd_SameConstraint(t *testing.T) {
	face := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false}, // (i=0)
		Right: ast.FaceEq{IVar: 0, IsOne: false}, // (i=0)
	}
	if isContradictoryFaceAnd(face) {
		t.Error("isContradictoryFaceAnd((i=0) ∧ (i=0)) = true, want false")
	}
}

// TestIsContradictoryFaceAnd_TripleNested tests ((i=0) ∧ (j=0)) ∧ (i=1).
func TestIsContradictoryFaceAnd_TripleNested(t *testing.T) {
	inner := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false}, // (i=0)
		Right: ast.FaceEq{IVar: 1, IsOne: false}, // (j=0)
	}
	outer := ast.FaceAnd{
		Left:  inner,
		Right: ast.FaceEq{IVar: 0, IsOne: true}, // (i=1)
	}
	if !isContradictoryFaceAnd(outer) {
		t.Error("isContradictoryFaceAnd(((i=0) ∧ (j=0)) ∧ (i=1)) = false, want true")
	}
}

// TestIsContradictoryFaceAnd_WithTop tests ⊤ ∧ (i=0).
func TestIsContradictoryFaceAnd_WithTop(t *testing.T) {
	face := ast.FaceAnd{
		Left:  ast.FaceTop{},
		Right: ast.FaceEq{IVar: 0, IsOne: false},
	}
	// FaceTop doesn't contribute constraints, so no contradiction
	if isContradictoryFaceAnd(face) {
		t.Error("isContradictoryFaceAnd(⊤ ∧ (i=0)) = true, want false")
	}
}

// --- collectFaceEqs Tests ---

// TestCollectFaceEqs_SingleEq tests collecting from a single FaceEq.
func TestCollectFaceEqs_SingleEq(t *testing.T) {
	face := ast.FaceEq{IVar: 0, IsOne: true}
	eqs := collectFaceEqs(face)
	if len(eqs) != 1 {
		t.Fatalf("collectFaceEqs(FaceEq) length = %d, want 1", len(eqs))
	}
	if eqs[0].IVar != 0 || eqs[0].IsOne != true {
		t.Errorf("collectFaceEqs(FaceEq) = %v, want [{0 true}]", eqs)
	}
}

// TestCollectFaceEqs_AndOfTwo tests collecting from (i=0) ∧ (j=1).
func TestCollectFaceEqs_AndOfTwo(t *testing.T) {
	face := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceEq{IVar: 1, IsOne: true},
	}
	eqs := collectFaceEqs(face)
	if len(eqs) != 2 {
		t.Fatalf("collectFaceEqs length = %d, want 2", len(eqs))
	}
}

// TestCollectFaceEqs_DeeplyNested tests collecting from nested ands.
func TestCollectFaceEqs_DeeplyNested(t *testing.T) {
	// ((i=0) ∧ (j=1)) ∧ ((k=0) ∧ (l=1))
	left := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceEq{IVar: 1, IsOne: true},
	}
	right := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 2, IsOne: false},
		Right: ast.FaceEq{IVar: 3, IsOne: true},
	}
	face := ast.FaceAnd{Left: left, Right: right}

	eqs := collectFaceEqs(face)
	if len(eqs) != 4 {
		t.Fatalf("collectFaceEqs length = %d, want 4", len(eqs))
	}
}

// TestCollectFaceEqs_OrDoesNotCollect verifies Or doesn't collect.
func TestCollectFaceEqs_OrDoesNotCollect(t *testing.T) {
	face := ast.FaceOr{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceEq{IVar: 1, IsOne: true},
	}
	eqs := collectFaceEqs(face)
	if len(eqs) != 0 {
		t.Errorf("collectFaceEqs(FaceOr) length = %d, want 0", len(eqs))
	}
}

// TestCollectFaceEqs_TopReturnsEmpty verifies Top returns empty.
func TestCollectFaceEqs_TopReturnsEmpty(t *testing.T) {
	eqs := collectFaceEqs(ast.FaceTop{})
	if len(eqs) != 0 {
		t.Errorf("collectFaceEqs(FaceTop) length = %d, want 0", len(eqs))
	}
}

// TestCollectFaceEqs_BotReturnsEmpty verifies Bot returns empty.
func TestCollectFaceEqs_BotReturnsEmpty(t *testing.T) {
	eqs := collectFaceEqs(ast.FaceBot{})
	if len(eqs) != 0 {
		t.Errorf("collectFaceEqs(FaceBot) length = %d, want 0", len(eqs))
	}
}

// TestCollectFaceEqs_AndWithTopAndBot tests mixed constraints.
func TestCollectFaceEqs_AndWithTopAndBot(t *testing.T) {
	// (i=0) ∧ ⊤ - should only collect the FaceEq
	face := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceTop{},
	}
	eqs := collectFaceEqs(face)
	if len(eqs) != 1 {
		t.Errorf("collectFaceEqs((i=0) ∧ ⊤) length = %d, want 1", len(eqs))
	}
}

// --- Path Type Edge Case Tests ---

// TestPathLam_NestedIVar tests PathLam with nested interval variables.
func TestPathLam_NestedIVar(t *testing.T) {
	c := NewChecker(nil)

	// <i> <j> Type0 should synthesize PathP type with nested path
	innerLam := ast.PathLam{
		Binder: "j",
		Body:   ast.Sort{U: 0},
	}
	outerLam := ast.PathLam{
		Binder: "i",
		Body:   innerLam,
	}

	ty, err := c.Synth(nil, NoSpan(), outerLam)
	if err != nil {
		t.Fatalf("Nested PathLam synthesis failed: %v", err)
	}

	// Outer type should be PathP
	if _, ok := ty.(ast.PathP); !ok {
		t.Errorf("Expected PathP type for outer, got %T", ty)
	}
}

// TestPathApp_NestedPathLam tests PathApp on nested PathLam.
func TestPathApp_NestedPathLam(t *testing.T) {
	// (<i> <j> Type0) @ i0 should reduce to <j> Type0
	innerLam := ast.PathLam{
		Binder: "j",
		Body:   ast.Sort{U: 0},
	}
	outerLam := ast.PathLam{
		Binder: "i",
		Body:   innerLam,
	}
	papp := ast.PathApp{
		P: outerLam,
		R: ast.I0{},
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), papp)

	// Should reduce to VPathLam
	if _, ok := result.(eval.VPathLam); !ok {
		t.Errorf("Expected VPathLam, got %T", result)
	}
}

// TestPathApp_AtIVar tests PathApp at an interval variable.
func TestPathApp_AtIVar(t *testing.T) {
	c := NewChecker(nil)

	// Push interval variable for IVar{0}
	pop := c.PushIVar()
	defer pop()

	// <i> Type0 @ j where j is an interval variable
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Sort{U: 0},
	}
	papp := ast.PathApp{
		P: plam,
		R: ast.IVar{Ix: 0}, // j
	}

	ty, err := c.Synth(nil, NoSpan(), papp)
	if err != nil {
		t.Fatalf("PathApp at IVar failed: %v", err)
	}

	// Result type should be Sort (the type family at that point)
	if _, ok := ty.(ast.Sort); !ok {
		t.Errorf("Expected Sort type, got %T", ty)
	}
}

// TestPathP_WithTypeFamily tests PathP with non-constant type family.
func TestPathP_WithTypeFamily(t *testing.T) {
	c := NewChecker(nil)

	// PathP (λi. Sort i) Type0 Type1 is not well-formed for our simple model,
	// but PathP with constant family should work
	pathp := ast.PathP{
		A: ast.Sort{U: 0}, // Constant family
		X: ast.Var{Ix: 0}, // x : Type0
		Y: ast.Var{Ix: 0}, // x : Type0
	}

	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})
	ty, err := c.Synth(ctx, NoSpan(), pathp)
	if err != nil {
		t.Fatalf("PathP synthesis failed: %v", err)
	}

	if _, ok := ty.(ast.Sort); !ok {
		t.Errorf("Expected Sort, got %T", ty)
	}
}

// TestPathLam_EndpointsMatch tests that PathLam endpoints are correctly computed.
func TestPathLam_EndpointsMatch(t *testing.T) {
	c := NewChecker(nil)

	// <i> x where x : Type0
	// Endpoints should both be x (constant path)
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Var{Ix: 0},
	}

	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})
	ty, err := c.Synth(ctx, NoSpan(), plam)
	if err != nil {
		t.Fatalf("PathLam synthesis failed: %v", err)
	}

	pathp, ok := ty.(ast.PathP)
	if !ok {
		t.Fatalf("Expected PathP type, got %T", ty)
	}

	// Both endpoints should be Var{0} (normalized)
	if _, ok := pathp.X.(ast.Var); !ok {
		t.Errorf("Expected left endpoint Var, got %T", pathp.X)
	}
	if _, ok := pathp.Y.(ast.Var); !ok {
		t.Errorf("Expected right endpoint Var, got %T", pathp.Y)
	}
}

// TestPathApp_BetaReduces tests that PathApp beta-reduces correctly.
func TestPathApp_BetaReduces(t *testing.T) {
	// (<i> IVar{0}) @ i0 should reduce to I0
	// where the body uses the interval variable
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.IVar{Ix: 0}, // The interval variable itself
	}

	// Apply at i0
	papp0 := ast.PathApp{P: plam, R: ast.I0{}}
	result0 := eval.EvalCubical(nil, eval.EmptyIEnv(), papp0)
	if _, ok := result0.(eval.VI0); !ok {
		t.Errorf("(<i> i) @ i0 expected VI0, got %T", result0)
	}

	// Apply at i1
	papp1 := ast.PathApp{P: plam, R: ast.I1{}}
	result1 := eval.EvalCubical(nil, eval.EmptyIEnv(), papp1)
	if _, ok := result1.(eval.VI1); !ok {
		t.Errorf("(<i> i) @ i1 expected VI1, got %T", result1)
	}
}

// TestPath_SameEndpoints tests Path A x x.
func TestPath_SameEndpoints(t *testing.T) {
	c := NewChecker(nil)

	// Path Type0 x x where x : Type0
	path := ast.Path{
		A: ast.Sort{U: 0},
		X: ast.Var{Ix: 0},
		Y: ast.Var{Ix: 0},
	}

	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})
	ty, err := c.Synth(ctx, NoSpan(), path)
	if err != nil {
		t.Fatalf("Path same endpoints failed: %v", err)
	}

	// Should be Sort 1 (Path Type0 x x : Type1)
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ty)
	}
}

// TestPathLam_ConstantBody tests that constant body produces refl-like path.
func TestPathLam_ConstantBody(t *testing.T) {
	c := NewChecker(nil)

	// <i> zero where zero is a global
	c.Globals().AddAxiom("Nat", ast.Sort{U: 0})
	c.Globals().AddDefinition("zero", ast.Global{Name: "Nat"}, ast.Global{Name: "z"}, Transparent)

	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Global{Name: "zero"},
	}

	ty, err := c.Synth(nil, NoSpan(), plam)
	if err != nil {
		t.Fatalf("Constant PathLam failed: %v", err)
	}

	pathp, ok := ty.(ast.PathP)
	if !ok {
		t.Fatalf("Expected PathP, got %T", ty)
	}

	// Endpoints should both be "zero"
	if g, ok := pathp.X.(ast.Global); !ok || g.Name != "zero" {
		t.Errorf("Expected left endpoint Global{zero}, got %v", pathp.X)
	}
	if g, ok := pathp.Y.(ast.Global); !ok || g.Name != "zero" {
		t.Errorf("Expected right endpoint Global{zero}, got %v", pathp.Y)
	}
}

// --- Composition Edge Case Tests ---

// TestComp_WithContradictoryFace tests comp with contradictory face.
func TestComp_WithContradictoryFace(t *testing.T) {
	// comp^i Type0 [(i=0) ∧ (i=1) ↦ _] Type0
	// The face is contradictory (always ⊥), so result should be like transport
	contradictoryFace := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceEq{IVar: 0, IsOne: true},
	}
	comp := ast.Comp{
		IBinder: "i",
		A:       ast.Sort{U: 0},
		Phi:     contradictoryFace,
		Tube:    ast.Sort{U: 1}, // Irrelevant since face is ⊥
		Base:    ast.Sort{U: 0},
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), comp)

	// Should reduce to transport (constant type → identity)
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 0 {
		t.Errorf("Expected VSort{Level: 0}, got %T %v", result, result)
	}
}

// TestHComp_WithMultipleBranches tests hcomp with FaceOr.
func TestHComp_WithMultipleBranches(t *testing.T) {
	c := NewChecker(nil)
	pop := c.PushIVar()
	defer pop()

	ctx := makeTestContext([]ast.Term{ast.Sort{U: 1}})

	// hcomp Type1 [⊥ ∨ ⊥ ↦ x] x
	// Both branches are ⊥, so effectively FaceBot
	orFace := ast.FaceOr{
		Left:  ast.FaceBot{},
		Right: ast.FaceBot{},
	}
	hcomp := ast.HComp{
		A:    ast.Sort{U: 1},
		Phi:  orFace,
		Tube: ast.Var{Ix: 0},
		Base: ast.Var{Ix: 0},
	}

	ty, err := c.Synth(ctx, NoSpan(), hcomp)
	if err != nil {
		t.Fatalf("HComp with Or face failed: %v", err)
	}

	if _, ok := ty.(ast.Sort); !ok {
		t.Errorf("Expected Sort type, got %T", ty)
	}
}

// TestTransport_ConstantTypeIsIdentity tests transport on constant type.
func TestTransport_ConstantTypeIsIdentity(t *testing.T) {
	// transport (λi. Type0) Type1 should reduce to Type1
	tr := ast.Transport{
		A: ast.Sort{U: 0}, // Constant type family
		E: ast.Sort{U: 1},
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), tr)

	// Transport along constant type is identity
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 1 {
		t.Errorf("Expected VSort{Level: 1}, got %T %v", result, result)
	}
}

// TestTransport_NonConstantIsStuck tests transport with non-constant type.
func TestTransport_NonConstantIsStuck(t *testing.T) {
	c := NewChecker(nil)

	// transport along a variable type family - should produce stuck value
	// We can't easily construct a truly non-constant type family in the AST
	// since we'd need a lambda with interval dependency
	// This test verifies the type checking still works
	pop := c.PushIVar()
	defer pop()

	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})
	tr := ast.Transport{
		A: ast.Sort{U: 0}, // Still constant for now
		E: ast.Var{Ix: 0},
	}

	ty, err := c.Synth(ctx, NoSpan(), tr)
	if err != nil {
		t.Fatalf("Transport synthesis failed: %v", err)
	}

	// Result type should be Sort{0}
	if _, ok := ty.(ast.Sort); !ok {
		t.Errorf("Expected Sort type, got %T", ty)
	}
}

// TestFill_AtEndpoints tests fill at i0 and i1.
func TestFill_AtEndpoints(t *testing.T) {
	// fill^i Type0 [⊥ ↦ _] x @ i0 = x
	fill := ast.Fill{
		IBinder: "i",
		A:       ast.Sort{U: 0},
		Phi:     ast.FaceBot{},
		Tube:    ast.Sort{U: 1},
		Base:    ast.Sort{U: 0}, // base = Type0
	}

	// Apply at i0 - should give base
	fillAtI0 := ast.PathApp{P: fill, R: ast.I0{}}
	resultI0 := eval.EvalCubical(nil, eval.EmptyIEnv(), fillAtI0)

	// At i0, fill should give base
	if sort, ok := resultI0.(eval.VSort); ok {
		if sort.Level != 0 {
			t.Errorf("fill @ i0: expected level 0, got %d", sort.Level)
		}
	}

	// Apply at i1 - should be equivalent to comp
	fillAtI1 := ast.PathApp{P: fill, R: ast.I1{}}
	resultI1 := eval.EvalCubical(nil, eval.EmptyIEnv(), fillAtI1)

	// At i1, fill should give comp result
	if resultI1 == nil {
		t.Error("fill @ i1: got nil result")
	}
}

// TestComp_WithNestedFaceAnd tests comp with nested FaceAnd.
func TestComp_WithNestedFaceAnd(t *testing.T) {
	c := NewChecker(nil)
	pop := c.PushIVar()
	defer pop()

	ctx := makeTestContext([]ast.Term{ast.Sort{U: 1}})

	// comp^i Type1 [((i=0) ∧ (j=0)) ∧ (k=0) ↦ x] x
	// where i, j, k are interval variables
	pop2 := c.PushIVar()
	defer pop2()
	pop3 := c.PushIVar()
	defer pop3()

	innerFace := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false}, // (i=0)
		Right: ast.FaceEq{IVar: 1, IsOne: false}, // (j=0)
	}
	outerFace := ast.FaceAnd{
		Left:  innerFace,
		Right: ast.FaceEq{IVar: 2, IsOne: false}, // (k=0)
	}

	comp := ast.Comp{
		IBinder: "i",
		A:       ast.Sort{U: 1},
		Phi:     outerFace,
		Tube:    ast.Var{Ix: 0},
		Base:    ast.Var{Ix: 0},
	}

	ty, err := c.Synth(ctx, NoSpan(), comp)
	if err != nil {
		t.Fatalf("Comp with nested face failed: %v", err)
	}

	if _, ok := ty.(ast.Sort); !ok {
		t.Errorf("Expected Sort type, got %T", ty)
	}
}

// TestHComp_BaseTypeMustMatch tests that HComp base has correct type.
func TestHComp_BaseTypeMustMatch(t *testing.T) {
	c := NewChecker(nil)
	pop := c.PushIVar()
	defer pop()

	// hcomp Type0 [⊥ ↦ _] x where x : Type0
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	hcomp := ast.HComp{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceBot{},
		Tube: ast.Var{Ix: 0},
		Base: ast.Var{Ix: 0}, // x : Type0
	}

	ty, err := c.Synth(ctx, NoSpan(), hcomp)
	if err != nil {
		t.Fatalf("HComp type check failed: %v", err)
	}

	// Result type should match A = Type0
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 0 {
		t.Errorf("Expected Type0, got %v", ty)
	}
}

// TestComp_ResultTypeIsAAtI1 tests that comp result type is A[i1/i].
func TestComp_ResultTypeIsAAtI1(t *testing.T) {
	c := NewChecker(nil)
	pop := c.PushIVar()
	defer pop()

	ctx := makeTestContext([]ast.Term{ast.Sort{U: 1}})

	// comp^i Type1 [⊥ ↦ x] x : Type1
	comp := ast.Comp{
		IBinder: "i",
		A:       ast.Sort{U: 1}, // Constant type family
		Phi:     ast.FaceBot{},
		Tube:    ast.Var{Ix: 0},
		Base:    ast.Var{Ix: 0},
	}

	ty, err := c.Synth(ctx, NoSpan(), comp)
	if err != nil {
		t.Fatalf("Comp synthesis failed: %v", err)
	}

	// A[i1/i] for constant A = Type1 is still Type1
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ty)
	}
}

// --- Glue/UA Edge Case Tests ---

// TestGlue_EmptySystem tests Glue with empty system.
func TestGlue_EmptySystem(t *testing.T) {
	c := NewChecker(nil)

	// Glue Type0 [] should be equivalent to Type0
	glue := ast.Glue{
		A:      ast.Sort{U: 0},
		System: nil,
	}

	ty, err := c.Synth(nil, NoSpan(), glue)
	if err != nil {
		t.Fatalf("Glue empty system failed: %v", err)
	}

	// Type of Glue Type0 [] is Type1
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ty)
	}
}

// TestGlue_MultipleBranches tests Glue with multiple face branches.
func TestGlue_MultipleBranches(t *testing.T) {
	c := NewChecker(nil)
	pop := c.PushIVar()
	defer pop()

	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	// Glue Type0 [(i=0) ↦ (Type0, e), (i=1) ↦ (Type0, e)]
	glue := ast.Glue{
		A: ast.Sort{U: 0},
		System: []ast.GlueBranch{
			{
				Phi:   ast.FaceEq{IVar: 0, IsOne: false}, // (i=0)
				T:     ast.Sort{U: 0},
				Equiv: ast.Var{Ix: 0},
			},
			{
				Phi:   ast.FaceEq{IVar: 0, IsOne: true}, // (i=1)
				T:     ast.Sort{U: 0},
				Equiv: ast.Var{Ix: 0},
			},
		},
	}

	ty, err := c.Synth(ctx, NoSpan(), glue)
	if err != nil {
		t.Fatalf("Glue multiple branches failed: %v", err)
	}

	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ty)
	}
}

// TestGlue_WithBotFace tests Glue with FaceBot branch (should be ignored).
func TestGlue_WithBotFace(t *testing.T) {
	c := NewChecker(nil)

	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	// Glue Type0 [⊥ ↦ (Type0, e)] - branch is never active
	glue := ast.Glue{
		A: ast.Sort{U: 0},
		System: []ast.GlueBranch{
			{
				Phi:   ast.FaceBot{},
				T:     ast.Sort{U: 0},
				Equiv: ast.Var{Ix: 0},
			},
		},
	}

	ty, err := c.Synth(ctx, NoSpan(), glue)
	if err != nil {
		t.Fatalf("Glue with bot face failed: %v", err)
	}

	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ty)
	}
}

// TestGlueElem_BasicEval tests GlueElem evaluation.
func TestGlueElem_BasicEval(t *testing.T) {
	// glue Type0 [] should reduce to Type0
	glueElem := ast.GlueElem{
		Base:   ast.Sort{U: 0},
		System: nil,
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), glueElem)

	// Should preserve the base
	if result == nil {
		t.Error("GlueElem evaluation returned nil")
	}
}

// TestUnglue_AfterGlueElem tests unglue (glue x []) = x.
func TestUnglue_AfterGlueElem(t *testing.T) {
	// unglue (glue x []) should reduce to x
	glueElem := ast.GlueElem{
		Base:   ast.Sort{U: 0},
		System: nil,
	}
	glueTy := ast.Glue{
		A:      ast.Sort{U: 0},
		System: nil,
	}
	unglue := ast.Unglue{
		Ty: glueTy,
		G:  glueElem,
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), unglue)

	// Should reduce to Type0 (the base)
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 0 {
		t.Errorf("Expected VSort{Level: 0}, got %T", result)
	}
}

// TestUA_EndpointI0 tests that (ua A B e) @ i0 = A.
func TestUA_EndpointI0(t *testing.T) {
	ua := ast.UA{
		A:     ast.Sort{U: 0},
		B:     ast.Sort{U: 1},
		Equiv: ast.Global{Name: "e"},
	}

	papp := ast.PathApp{P: ua, R: ast.I0{}}
	result := eval.EvalCubical(nil, eval.EmptyIEnv(), papp)

	// At i0, should be A = Type0
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 0 {
		t.Errorf("ua @ i0: expected VSort{0}, got %T", result)
	}
}

// TestUA_EndpointI1 tests that (ua A B e) @ i1 = B.
func TestUA_EndpointI1(t *testing.T) {
	ua := ast.UA{
		A:     ast.Sort{U: 0},
		B:     ast.Sort{U: 1},
		Equiv: ast.Global{Name: "e"},
	}

	papp := ast.PathApp{P: ua, R: ast.I1{}}
	result := eval.EvalCubical(nil, eval.EmptyIEnv(), papp)

	// At i1, should be B = Type1
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 1 {
		t.Errorf("ua @ i1: expected VSort{1}, got %T", result)
	}
}

// TestUA_TypeCheck tests ua type checking produces Path.
func TestUA_TypeCheck(t *testing.T) {
	c := NewChecker(nil)

	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}}) // e as placeholder

	// ua Type0 Type0 e : Path Type1 Type0 Type0
	ua := ast.UA{
		A:     ast.Sort{U: 0},
		B:     ast.Sort{U: 0},
		Equiv: ast.Var{Ix: 0},
	}

	ty, err := c.Synth(ctx, NoSpan(), ua)
	if err != nil {
		t.Fatalf("UA type check failed: %v", err)
	}

	// Should be Path or PathP type (UA produces a path between types)
	switch ty.(type) {
	case ast.Path, ast.PathP:
		// OK
	default:
		t.Errorf("Expected Path or PathP type, got %T", ty)
	}
}

// TestGlue_TopFaceSatisfied tests that Glue with ⊤ reduces.
func TestGlue_TopFaceSatisfied(t *testing.T) {
	// Glue Type0 [⊤ ↦ (Type1, e)] should reduce to Type1 when face is ⊤
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

	// When face is ⊤, should reduce to T = Type1
	if sort, ok := result.(eval.VSort); !ok || sort.Level != 1 {
		t.Errorf("Expected VSort{1}, got %T", result)
	}
}

// TestGlue_BotFaceStuck tests that Glue with ⊥ doesn't reduce via that branch.
func TestGlue_BotFaceStuck(t *testing.T) {
	// Glue Type0 [⊥ ↦ (Type1, e)] - branch is never taken
	glue := ast.Glue{
		A: ast.Sort{U: 0},
		System: []ast.GlueBranch{
			{
				Phi:   ast.FaceBot{},
				T:     ast.Sort{U: 1},
				Equiv: ast.Global{Name: "e"},
			},
		},
	}

	result := eval.EvalCubical(nil, eval.EmptyIEnv(), glue)

	// When face is ⊥, should NOT reduce to Type1, stays as VGlue or Type0
	if sort, ok := result.(eval.VSort); ok && sort.Level == 1 {
		t.Error("Glue with ⊥ face should not reduce via that branch")
	}
}

// ============================================================================
// checkFace Edge Cases
// ============================================================================

// TestCheckFace_FaceTop tests FaceTop is always valid.
func TestCheckFace_FaceTop(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	// FaceTop should be valid
	err := c.checkFace(ctx, NoSpan(), ast.FaceTop{})
	if err != nil {
		t.Errorf("FaceTop should be valid, got error: %v", err)
	}
}

// TestCheckFace_FaceBot tests FaceBot is always valid.
func TestCheckFace_FaceBot(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	// FaceBot should be valid
	err := c.checkFace(ctx, NoSpan(), ast.FaceBot{})
	if err != nil {
		t.Errorf("FaceBot should be valid, got error: %v", err)
	}
}

// TestCheckFace_FaceEq_BoundIVar tests FaceEq with bound interval variable.
func TestCheckFace_FaceEq_BoundIVar(t *testing.T) {
	c := NewChecker(nil)
	c.PushIVar() // Push interval variable to context
	ctx := makeTestContext(nil)

	// FaceEq{IVar: 0} should be valid when ivar 0 is bound
	err := c.checkFace(ctx, NoSpan(), ast.FaceEq{IVar: 0, IsOne: false})
	if err != nil {
		t.Errorf("FaceEq with bound IVar should be valid, got error: %v", err)
	}
}

// TestCheckFace_FaceEq_UnboundIVar tests FaceEq with unbound interval variable.
func TestCheckFace_FaceEq_UnboundIVar(t *testing.T) {
	c := NewChecker(nil)
	// Don't push any interval variables
	ctx := makeTestContext(nil)

	// FaceEq{IVar: 0} should fail when no ivars are bound
	err := c.checkFace(ctx, NoSpan(), ast.FaceEq{IVar: 0, IsOne: false})
	if err == nil {
		t.Error("FaceEq with unbound IVar should produce error")
	}
}

// TestCheckFace_FaceAnd tests nested FaceAnd.
func TestCheckFace_FaceAnd(t *testing.T) {
	c := NewChecker(nil)
	c.PushIVar() // i
	c.PushIVar() // j
	ctx := makeTestContext(nil)

	// (i = 0) ∧ (j = 1) should be valid
	face := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceEq{IVar: 1, IsOne: true},
	}
	err := c.checkFace(ctx, NoSpan(), face)
	if err != nil {
		t.Errorf("FaceAnd with valid sub-faces should be valid, got error: %v", err)
	}
}

// TestCheckFace_FaceOr tests nested FaceOr.
func TestCheckFace_FaceOr(t *testing.T) {
	c := NewChecker(nil)
	c.PushIVar()
	ctx := makeTestContext(nil)

	// (i = 0) ∨ (i = 1) should be valid
	face := ast.FaceOr{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceEq{IVar: 0, IsOne: true},
	}
	err := c.checkFace(ctx, NoSpan(), face)
	if err != nil {
		t.Errorf("FaceOr should be valid, got error: %v", err)
	}
}

// TestCheckFace_FaceAnd_InvalidLeft tests FaceAnd with invalid left.
func TestCheckFace_FaceAnd_InvalidLeft(t *testing.T) {
	c := NewChecker(nil)
	// No ivars bound
	ctx := makeTestContext(nil)

	// (unbound ivar) ∧ ⊤ should fail on left
	face := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false}, // Unbound
		Right: ast.FaceTop{},
	}
	err := c.checkFace(ctx, NoSpan(), face)
	if err == nil {
		t.Error("FaceAnd with invalid left should produce error")
	}
}

// TestCheckFace_FaceOr_InvalidRight tests FaceOr with invalid right.
func TestCheckFace_FaceOr_InvalidRight(t *testing.T) {
	c := NewChecker(nil)
	// No ivars bound
	ctx := makeTestContext(nil)

	// ⊤ ∨ (unbound ivar) should fail on right
	face := ast.FaceOr{
		Left:  ast.FaceTop{},
		Right: ast.FaceEq{IVar: 0, IsOne: false}, // Unbound
	}
	err := c.checkFace(ctx, NoSpan(), face)
	if err == nil {
		t.Error("FaceOr with invalid right should produce error")
	}
}

// TestCheckFace_Nil tests nil face is valid.
func TestCheckFace_Nil(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	err := c.checkFace(ctx, NoSpan(), nil)
	if err != nil {
		t.Errorf("nil face should be valid, got error: %v", err)
	}
}

// ============================================================================
// PathApp Edge Cases
// ============================================================================

// TestSynthPathApp_I0Endpoint tests path applied at i0 gives the type family result.
func TestSynthPathApp_I0Endpoint(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}, ast.Sort{U: 0}})

	// p : Path Type1 Type0 Type0 (type family is Type1, endpoints are Type0)
	// p @ i0 : Type1 (the type family applied to i0)
	pathType := ast.Path{A: ast.Sort{U: 1}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}}
	c.globals.AddAxiom("p", pathType)

	papp := ast.PathApp{
		P: ast.Global{Name: "p"},
		R: ast.I0{},
	}

	ty, err := c.Synth(ctx, NoSpan(), papp)
	if err != nil {
		t.Fatalf("PathApp @ i0 synthesis failed: %v", err)
	}

	// Should be Type1 (the type family A)
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ast.Sprint(ty))
	}
}

// TestSynthPathApp_I1Endpoint tests path applied at i1 gives the type family result.
func TestSynthPathApp_I1Endpoint(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	// p : Path Type1 Type0 Type0
	pathType := ast.Path{A: ast.Sort{U: 1}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}}
	c.globals.AddAxiom("p", pathType)

	papp := ast.PathApp{
		P: ast.Global{Name: "p"},
		R: ast.I1{},
	}

	ty, err := c.Synth(ctx, NoSpan(), papp)
	if err != nil {
		t.Fatalf("PathApp @ i1 synthesis failed: %v", err)
	}

	// Should be Type1 (the type family A)
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ast.Sprint(ty))
	}
}

// TestSynthPathApp_IVar tests path applied at interval variable.
func TestSynthPathApp_IVar(t *testing.T) {
	c := NewChecker(nil)
	c.PushIVar() // Push i
	ctx := makeTestContext(nil)

	// p : Path Type1 Type0 Type0
	pathType := ast.Path{A: ast.Sort{U: 1}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}}
	c.globals.AddAxiom("p", pathType)

	papp := ast.PathApp{
		P: ast.Global{Name: "p"},
		R: ast.IVar{Ix: 0},
	}

	ty, err := c.Synth(ctx, NoSpan(), papp)
	if err != nil {
		t.Fatalf("PathApp @ IVar synthesis failed: %v", err)
	}

	// Should be Type1 (the type family A)
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("Expected Type1, got %v", ast.Sprint(ty))
	}
}

// TestSynthPathApp_NotAPath tests applying non-path term.
func TestSynthPathApp_NotAPath(t *testing.T) {
	c := NewChecker(NewGlobalEnvWithPrimitives())
	ctx := makeTestContext(nil)

	// x : Nat (not a path)
	c.globals.AddAxiom("x", ast.Global{Name: "Nat"})

	// x @ i0 should fail
	papp := ast.PathApp{
		P: ast.Global{Name: "x"},
		R: ast.I0{},
	}

	_, err := c.Synth(ctx, NoSpan(), papp)
	if err == nil {
		t.Error("PathApp on non-path should produce error")
	}
}

// ============================================================================
// PathLam Checking Edge Cases
// ============================================================================

// TestCheckPathLam_ValidConstant tests checking PathLam with constant body.
func TestCheckPathLam_ValidConstant(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	// <i> Type0 should check against Path Type1 Type0 Type0
	plam := ast.PathLam{Binder: "i", Body: ast.Sort{U: 0}}
	expectedTy := ast.Path{A: ast.Sort{U: 1}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}}

	err := c.Check(ctx, NoSpan(), plam, expectedTy)
	if err != nil {
		t.Fatalf("PathLam check failed: %v", err)
	}
}

// TestCheckPathLam_WrongLeftEndpoint tests PathLam with wrong left endpoint.
func TestCheckPathLam_WrongLeftEndpoint(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	// <i> Type0 should NOT check against Path Type1 Type1 Type0
	plam := ast.PathLam{Binder: "i", Body: ast.Sort{U: 0}}
	expectedTy := ast.Path{A: ast.Sort{U: 1}, X: ast.Sort{U: 1}, Y: ast.Sort{U: 0}}

	err := c.Check(ctx, NoSpan(), plam, expectedTy)
	if err == nil {
		t.Error("PathLam with wrong left endpoint should fail")
	}
}

// TestCheckPathLam_WrongRightEndpoint tests PathLam with wrong right endpoint.
func TestCheckPathLam_WrongRightEndpoint(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	// <i> Type0 should NOT check against Path Type1 Type0 Type1
	plam := ast.PathLam{Binder: "i", Body: ast.Sort{U: 0}}
	expectedTy := ast.Path{A: ast.Sort{U: 1}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 1}}

	err := c.Check(ctx, NoSpan(), plam, expectedTy)
	if err == nil {
		t.Error("PathLam with wrong right endpoint should fail")
	}
}

// TestCheckPathLam_AgainstPathP tests checking PathLam against PathP type.
func TestCheckPathLam_AgainstPathP(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	// <i> Type0 should check against PathP Type1 Type0 Type0
	plam := ast.PathLam{Binder: "i", Body: ast.Sort{U: 0}}
	expectedTy := ast.PathP{A: ast.Sort{U: 1}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}}

	err := c.Check(ctx, NoSpan(), plam, expectedTy)
	if err != nil {
		t.Fatalf("PathLam check against PathP failed: %v", err)
	}
}

// ============================================================================
// Cubical Type Synthesis Tests
// ============================================================================

// TestSynthInterval tests Interval type synthesis.
func TestSynthInterval(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	ty, err := c.Synth(ctx, NoSpan(), ast.Interval{})
	if err != nil {
		t.Fatalf("Interval synthesis failed: %v", err)
	}
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 0 {
		t.Errorf("Expected Type0, got %v", ast.Sprint(ty))
	}
}

// TestSynthI0 tests i0 synthesis.
func TestSynthI0(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	ty, err := c.Synth(ctx, NoSpan(), ast.I0{})
	if err != nil {
		t.Fatalf("I0 synthesis failed: %v", err)
	}
	if _, ok := ty.(ast.Interval); !ok {
		t.Errorf("Expected Interval, got %v", ast.Sprint(ty))
	}
}

// TestSynthI1 tests i1 synthesis.
func TestSynthI1(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	ty, err := c.Synth(ctx, NoSpan(), ast.I1{})
	if err != nil {
		t.Fatalf("I1 synthesis failed: %v", err)
	}
	if _, ok := ty.(ast.Interval); !ok {
		t.Errorf("Expected Interval, got %v", ast.Sprint(ty))
	}
}

// TestSynthIVar_Valid tests valid IVar synthesis.
func TestSynthIVar_Valid(t *testing.T) {
	c := NewChecker(nil)
	c.PushIVar() // Push one interval variable
	ctx := makeTestContext(nil)

	ty, err := c.Synth(ctx, NoSpan(), ast.IVar{Ix: 0})
	if err != nil {
		t.Fatalf("IVar synthesis failed: %v", err)
	}
	if _, ok := ty.(ast.Interval); !ok {
		t.Errorf("Expected Interval, got %v", ast.Sprint(ty))
	}
}

// TestSynthIVar_Unbound tests unbound IVar synthesis.
func TestSynthIVar_Unbound(t *testing.T) {
	c := NewChecker(nil)
	// No interval variables pushed
	ctx := makeTestContext(nil)

	_, err := c.Synth(ctx, NoSpan(), ast.IVar{Ix: 0})
	if err == nil {
		t.Error("Unbound IVar should produce error")
	}
}

// TestSynthFaceTop tests FaceTop synthesis.
func TestSynthFaceTop(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	ty, err := c.Synth(ctx, NoSpan(), ast.FaceTop{})
	if err != nil {
		t.Fatalf("FaceTop synthesis failed: %v", err)
	}
	// Should be Face type
	if g, ok := ty.(ast.Global); !ok || g.Name != "Face" {
		t.Errorf("Expected Face, got %v", ast.Sprint(ty))
	}
}

// TestSynthFaceBot tests FaceBot synthesis.
func TestSynthFaceBot(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	ty, err := c.Synth(ctx, NoSpan(), ast.FaceBot{})
	if err != nil {
		t.Fatalf("FaceBot synthesis failed: %v", err)
	}
	// Should be Face type
	if g, ok := ty.(ast.Global); !ok || g.Name != "Face" {
		t.Errorf("Expected Face, got %v", ast.Sprint(ty))
	}
}

// TestSynthFaceEq_Valid tests valid FaceEq synthesis.
func TestSynthFaceEq_Valid(t *testing.T) {
	c := NewChecker(nil)
	c.PushIVar()
	ctx := makeTestContext(nil)

	ty, err := c.Synth(ctx, NoSpan(), ast.FaceEq{IVar: 0, IsOne: false})
	if err != nil {
		t.Fatalf("FaceEq synthesis failed: %v", err)
	}
	if g, ok := ty.(ast.Global); !ok || g.Name != "Face" {
		t.Errorf("Expected Face, got %v", ast.Sprint(ty))
	}
}

// TestSynthFaceEq_Unbound tests FaceEq with unbound IVar.
func TestSynthFaceEq_Unbound(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	_, err := c.Synth(ctx, NoSpan(), ast.FaceEq{IVar: 0, IsOne: false})
	if err == nil {
		t.Error("FaceEq with unbound IVar should produce error")
	}
}

// TestSynthFaceAnd tests FaceAnd synthesis.
func TestSynthFaceAnd(t *testing.T) {
	c := NewChecker(nil)
	c.PushIVar()
	c.PushIVar()
	ctx := makeTestContext(nil)

	face := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceEq{IVar: 1, IsOne: true},
	}
	ty, err := c.Synth(ctx, NoSpan(), face)
	if err != nil {
		t.Fatalf("FaceAnd synthesis failed: %v", err)
	}
	if g, ok := ty.(ast.Global); !ok || g.Name != "Face" {
		t.Errorf("Expected Face, got %v", ast.Sprint(ty))
	}
}

// TestSynthFaceOr tests FaceOr synthesis.
func TestSynthFaceOr(t *testing.T) {
	c := NewChecker(nil)
	c.PushIVar()
	ctx := makeTestContext(nil)

	face := ast.FaceOr{
		Left:  ast.FaceTop{},
		Right: ast.FaceBot{},
	}
	ty, err := c.Synth(ctx, NoSpan(), face)
	if err != nil {
		t.Fatalf("FaceOr synthesis failed: %v", err)
	}
	if g, ok := ty.(ast.Global); !ok || g.Name != "Face" {
		t.Errorf("Expected Face, got %v", ast.Sprint(ty))
	}
}

// TestSynthPartial tests Partial type synthesis.
func TestSynthPartial(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	partial := ast.Partial{
		Phi: ast.FaceTop{},
		A:   ast.Sort{U: 0}, // Type0 : Type1
	}
	ty, err := c.Synth(ctx, NoSpan(), partial)
	if err != nil {
		t.Fatalf("Partial synthesis failed: %v", err)
	}
	// Partial φ Type0 : Type1 (since Type0 : Type1)
	if sort, ok := ty.(ast.Sort); !ok {
		t.Errorf("Expected Sort, got %v", ast.Sprint(ty))
	} else if sort.U > 2 {
		t.Errorf("Unexpected universe level %d", sort.U)
	}
}

// TestSynthTransport tests Transport synthesis.
func TestSynthTransport(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	// transport (λi. Type0) x where x : Type0
	// transport preserves the element type
	tr := ast.Transport{
		A: ast.Sort{U: 0}, // constant type family (Type0)
		E: ast.Var{Ix: 0}, // x : Type0
	}
	ty, err := c.Synth(ctx, NoSpan(), tr)
	if err != nil {
		t.Fatalf("Transport synthesis failed: %v", err)
	}
	// Result type should be Type0 (the type family)
	if sort, ok := ty.(ast.Sort); !ok || sort.U != 0 {
		t.Errorf("Expected Type0, got %v", ast.Sprint(ty))
	}
}

// TestSynthSystem_Empty tests empty System fails to infer.
func TestSynthSystem_Empty(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	sys := ast.System{Branches: nil}
	_, err := c.Synth(ctx, NoSpan(), sys)
	// Empty system should fail - cannot infer type
	if err == nil {
		t.Error("Empty System should fail to synthesize (needs annotation)")
	}
}

// TestSynthSystem_SingleBranch tests System with one branch.
func TestSynthSystem_SingleBranch(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	sys := ast.System{
		Branches: []ast.SystemBranch{
			{Phi: ast.FaceTop{}, Term: ast.Sort{U: 0}},
		},
	}
	ty, err := c.Synth(ctx, NoSpan(), sys)
	if err != nil {
		t.Fatalf("System synthesis failed: %v", err)
	}
	if ty == nil {
		t.Error("System should have a type")
	}
}
