package subst

import (
	"reflect"
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// ============================================================================
// ISubst Tests - Interval Substitution
// ============================================================================

// --- IVar Substitution ---

func TestISubst_IVar_Match(t *testing.T) {
	t.Parallel()
	// IVar{1} at j=1 should be replaced with I0
	term := ast.IVar{Ix: 1}
	result := ISubst(1, ast.I0{}, term)

	if _, ok := result.(ast.I0); !ok {
		t.Errorf("expected I0, got %T", result)
	}
}

func TestISubst_IVar_Above(t *testing.T) {
	t.Parallel()
	// IVar{3} at j=1 should become IVar{2} (decremented)
	term := ast.IVar{Ix: 3}
	result := ISubst(1, ast.I0{}, term)

	ivar, ok := result.(ast.IVar)
	if !ok {
		t.Fatalf("expected IVar, got %T", result)
	}
	if ivar.Ix != 2 {
		t.Errorf("expected IVar{2}, got IVar{%d}", ivar.Ix)
	}
}

func TestISubst_IVar_Below(t *testing.T) {
	t.Parallel()
	// IVar{0} at j=1 should be unchanged
	term := ast.IVar{Ix: 0}
	result := ISubst(1, ast.I0{}, term)

	ivar, ok := result.(ast.IVar)
	if !ok {
		t.Fatalf("expected IVar, got %T", result)
	}
	if ivar.Ix != 0 {
		t.Errorf("expected IVar{0}, got IVar{%d}", ivar.Ix)
	}
}

func TestISubst_Constants(t *testing.T) {
	t.Parallel()
	// I0, I1, Interval are unchanged
	tests := []ast.Term{ast.I0{}, ast.I1{}, ast.Interval{}}
	for _, tm := range tests {
		result := ISubst(0, ast.IVar{Ix: 5}, tm)
		if !reflect.DeepEqual(result, tm) {
			t.Errorf("expected %T to be unchanged, got %v", tm, result)
		}
	}
}

func TestISubst_Nil(t *testing.T) {
	t.Parallel()
	result := ISubst(0, ast.I0{}, nil)
	if result != nil {
		t.Errorf("ISubst(nil) should return nil, got %v", result)
	}
}

// --- Path (non-binding) ---

func TestISubst_Path(t *testing.T) {
	t.Parallel()
	// Path doesn't bind interval vars, all fields substituted at same j
	term := ast.Path{
		A: ast.IVar{Ix: 0},
		X: ast.IVar{Ix: 1},
		Y: ast.IVar{Ix: 2},
	}
	result := ISubst(1, ast.I1{}, term)

	path, ok := result.(ast.Path)
	if !ok {
		t.Fatalf("expected Path, got %T", result)
	}

	// A: IVar{0} below j=1, unchanged
	if a, ok := path.A.(ast.IVar); !ok || a.Ix != 0 {
		t.Errorf("Path.A: expected IVar{0}, got %v", path.A)
	}
	// X: IVar{1} at j=1, becomes I1
	if _, ok := path.X.(ast.I1); !ok {
		t.Errorf("Path.X: expected I1, got %v", path.X)
	}
	// Y: IVar{2} above j=1, decremented to IVar{1}
	if y, ok := path.Y.(ast.IVar); !ok || y.Ix != 1 {
		t.Errorf("Path.Y: expected IVar{1}, got %v", path.Y)
	}
}

// --- PathApp (non-binding) ---

func TestISubst_PathApp(t *testing.T) {
	t.Parallel()
	term := ast.PathApp{
		P: ast.IVar{Ix: 0},
		R: ast.IVar{Ix: 1},
	}
	result := ISubst(0, ast.I0{}, term)

	papp, ok := result.(ast.PathApp)
	if !ok {
		t.Fatalf("expected PathApp, got %T", result)
	}

	// P: IVar{0} at j=0, becomes I0
	if _, ok := papp.P.(ast.I0); !ok {
		t.Errorf("PathApp.P: expected I0, got %v", papp.P)
	}
	// R: IVar{1} above j=0, decremented to IVar{0}
	if r, ok := papp.R.(ast.IVar); !ok || r.Ix != 0 {
		t.Errorf("PathApp.R: expected IVar{0}, got %v", papp.R)
	}
}

// --- Glue (non-binding) ---

func TestISubst_Glue(t *testing.T) {
	t.Parallel()
	term := ast.Glue{
		A: ast.IVar{Ix: 0},
		System: []ast.GlueBranch{
			{
				Phi:   ast.FaceEq{IVar: 0, IsOne: false},
				T:     ast.IVar{Ix: 1},
				Equiv: ast.IVar{Ix: 2},
			},
		},
	}
	result := ISubst(0, ast.I1{}, term)

	glue, ok := result.(ast.Glue)
	if !ok {
		t.Fatalf("expected Glue, got %T", result)
	}

	// A: IVar{0} becomes I1
	if _, ok := glue.A.(ast.I1); !ok {
		t.Errorf("Glue.A: expected I1, got %v", glue.A)
	}
	// Phi: (i=0)[i1/i] = ⊥
	if _, ok := glue.System[0].Phi.(ast.FaceBot); !ok {
		t.Errorf("Glue branch Phi: expected FaceBot, got %T", glue.System[0].Phi)
	}
	// T: IVar{1} decremented
	if bt, ok := glue.System[0].T.(ast.IVar); !ok || bt.Ix != 0 {
		t.Errorf("Glue branch T: expected IVar{0}, got %v", glue.System[0].T)
	}
	// Equiv: IVar{2} decremented
	if eq, ok := glue.System[0].Equiv.(ast.IVar); !ok || eq.Ix != 1 {
		t.Errorf("Glue branch Equiv: expected IVar{1}, got %v", glue.System[0].Equiv)
	}
}

// --- GlueElem (non-binding) ---

func TestISubst_GlueElem(t *testing.T) {
	t.Parallel()
	term := ast.GlueElem{
		System: []ast.GlueElemBranch{
			{Phi: ast.FaceEq{IVar: 1, IsOne: true}, Term: ast.IVar{Ix: 2}},
		},
		Base: ast.IVar{Ix: 1},
	}
	result := ISubst(1, ast.I0{}, term)

	ge, ok := result.(ast.GlueElem)
	if !ok {
		t.Fatalf("expected GlueElem, got %T", result)
	}

	// Phi: (i=1)[i0/i] = ⊥
	if _, ok := ge.System[0].Phi.(ast.FaceBot); !ok {
		t.Errorf("GlueElem branch Phi: expected FaceBot, got %T", ge.System[0].Phi)
	}
	// Term: IVar{2} decremented
	if tm, ok := ge.System[0].Term.(ast.IVar); !ok || tm.Ix != 1 {
		t.Errorf("GlueElem branch Term: expected IVar{1}, got %v", ge.System[0].Term)
	}
	// Base: IVar{1} becomes I0
	if _, ok := ge.Base.(ast.I0); !ok {
		t.Errorf("GlueElem.Base: expected I0, got %v", ge.Base)
	}
}

// --- Unglue (non-binding) ---

func TestISubst_Unglue(t *testing.T) {
	t.Parallel()
	term := ast.Unglue{
		Ty: ast.IVar{Ix: 0},
		G:  ast.IVar{Ix: 1},
	}
	result := ISubst(0, ast.I1{}, term)

	ug, ok := result.(ast.Unglue)
	if !ok {
		t.Fatalf("expected Unglue, got %T", result)
	}

	if _, ok := ug.Ty.(ast.I1); !ok {
		t.Errorf("Unglue.Ty: expected I1, got %v", ug.Ty)
	}
	if g, ok := ug.G.(ast.IVar); !ok || g.Ix != 0 {
		t.Errorf("Unglue.G: expected IVar{0}, got %v", ug.G)
	}
}

// --- UA (non-binding) ---

func TestISubst_UA(t *testing.T) {
	t.Parallel()
	term := ast.UA{
		A:     ast.IVar{Ix: 0},
		B:     ast.IVar{Ix: 1},
		Equiv: ast.IVar{Ix: 2},
	}
	result := ISubst(1, ast.I0{}, term)

	ua, ok := result.(ast.UA)
	if !ok {
		t.Fatalf("expected UA, got %T", result)
	}

	// A: IVar{0} unchanged
	if a, ok := ua.A.(ast.IVar); !ok || a.Ix != 0 {
		t.Errorf("UA.A: expected IVar{0}, got %v", ua.A)
	}
	// B: IVar{1} becomes I0
	if _, ok := ua.B.(ast.I0); !ok {
		t.Errorf("UA.B: expected I0, got %v", ua.B)
	}
	// Equiv: IVar{2} decremented
	if e, ok := ua.Equiv.(ast.IVar); !ok || e.Ix != 1 {
		t.Errorf("UA.Equiv: expected IVar{1}, got %v", ua.Equiv)
	}
}

// --- UABeta (non-binding) ---

func TestISubst_UABeta(t *testing.T) {
	t.Parallel()
	term := ast.UABeta{
		Equiv: ast.IVar{Ix: 0},
		Arg:   ast.IVar{Ix: 1},
	}
	result := ISubst(0, ast.I1{}, term)

	uab, ok := result.(ast.UABeta)
	if !ok {
		t.Fatalf("expected UABeta, got %T", result)
	}

	if _, ok := uab.Equiv.(ast.I1); !ok {
		t.Errorf("UABeta.Equiv: expected I1, got %v", uab.Equiv)
	}
	if arg, ok := uab.Arg.(ast.IVar); !ok || arg.Ix != 0 {
		t.Errorf("UABeta.Arg: expected IVar{0}, got %v", uab.Arg)
	}
}

// ============================================================================
// ISubst Tests - Binding Term Types
// ============================================================================

// --- PathP (binds in A) ---

func TestISubst_PathP(t *testing.T) {
	t.Parallel()
	// PathP A binds interval variable; X, Y don't
	term := ast.PathP{
		A: ast.IVar{Ix: 1}, // Under binder: j+1=1
		X: ast.IVar{Ix: 0},
		Y: ast.IVar{Ix: 1},
	}
	result := ISubst(0, ast.I0{}, term)

	pathp, ok := result.(ast.PathP)
	if !ok {
		t.Fatalf("expected PathP, got %T", result)
	}

	// A: IVar{1} with j+1=1 matches, becomes I0 (shifted: IShift(1,0,I0)=I0)
	if _, ok := pathp.A.(ast.I0); !ok {
		t.Errorf("PathP.A: expected I0, got %v", pathp.A)
	}
	// X: IVar{0} at j=0, becomes I0
	if _, ok := pathp.X.(ast.I0); !ok {
		t.Errorf("PathP.X: expected I0, got %v", pathp.X)
	}
	// Y: IVar{1} above j=0, decremented to IVar{0}
	if y, ok := pathp.Y.(ast.IVar); !ok || y.Ix != 0 {
		t.Errorf("PathP.Y: expected IVar{0}, got %v", pathp.Y)
	}
}

// --- PathLam (binds in Body) ---

func TestISubst_PathLam(t *testing.T) {
	t.Parallel()
	term := ast.PathLam{
		Binder: "i",
		Body:   ast.IVar{Ix: 1}, // Under binder: j+1=1
	}
	result := ISubst(0, ast.I1{}, term)

	plam, ok := result.(ast.PathLam)
	if !ok {
		t.Fatalf("expected PathLam, got %T", result)
	}

	if plam.Binder != "i" {
		t.Errorf("PathLam.Binder should be preserved, got %q", plam.Binder)
	}
	// Body: IVar{1} with j+1=1, becomes I1
	if _, ok := plam.Body.(ast.I1); !ok {
		t.Errorf("PathLam.Body: expected I1, got %v", plam.Body)
	}
}

// --- Transport (binds in A) ---

func TestISubst_Transport(t *testing.T) {
	t.Parallel()
	term := ast.Transport{
		A: ast.IVar{Ix: 1}, // Under binder
		E: ast.IVar{Ix: 0}, // Not under binder
	}
	result := ISubst(0, ast.I0{}, term)

	tr, ok := result.(ast.Transport)
	if !ok {
		t.Fatalf("expected Transport, got %T", result)
	}

	// A: IVar{1} with j+1=1, becomes I0
	if _, ok := tr.A.(ast.I0); !ok {
		t.Errorf("Transport.A: expected I0, got %v", tr.A)
	}
	// E: IVar{0} at j=0, becomes I0
	if _, ok := tr.E.(ast.I0); !ok {
		t.Errorf("Transport.E: expected I0, got %v", tr.E)
	}
}

// --- Comp (binds in A, Phi, Tube) ---

func TestISubst_Comp(t *testing.T) {
	t.Parallel()
	term := ast.Comp{
		IBinder: "i",
		A:       ast.IVar{Ix: 1}, // Under binder
		Phi:     ast.FaceEq{IVar: 1, IsOne: false},
		Tube:    ast.IVar{Ix: 1},
		Base:    ast.IVar{Ix: 0}, // Not under binder
	}
	result := ISubst(0, ast.I1{}, term)

	comp, ok := result.(ast.Comp)
	if !ok {
		t.Fatalf("expected Comp, got %T", result)
	}

	// A, Phi, Tube: IVar{1} with j+1=1, becomes I1
	if _, ok := comp.A.(ast.I1); !ok {
		t.Errorf("Comp.A: expected I1, got %v", comp.A)
	}
	// Phi: (i=0)[i1/i] under binder - substitutes at j+1=1
	// FaceEq{1, false} at j+1=1 becomes FaceBot (i1=0 is false)
	if _, ok := comp.Phi.(ast.FaceBot); !ok {
		t.Errorf("Comp.Phi: expected FaceBot, got %T", comp.Phi)
	}
	if _, ok := comp.Tube.(ast.I1); !ok {
		t.Errorf("Comp.Tube: expected I1, got %v", comp.Tube)
	}
	// Base: IVar{0} at j=0, becomes I1
	if _, ok := comp.Base.(ast.I1); !ok {
		t.Errorf("Comp.Base: expected I1, got %v", comp.Base)
	}
}

// --- HComp (binds in Phi, Tube only) ---

func TestISubst_HComp(t *testing.T) {
	t.Parallel()
	term := ast.HComp{
		A:    ast.IVar{Ix: 0}, // NOT under binder
		Phi:  ast.FaceEq{IVar: 1, IsOne: true},
		Tube: ast.IVar{Ix: 1},
		Base: ast.IVar{Ix: 0},
	}
	result := ISubst(0, ast.I0{}, term)

	hcomp, ok := result.(ast.HComp)
	if !ok {
		t.Fatalf("expected HComp, got %T", result)
	}

	// A: IVar{0} at j=0, becomes I0
	if _, ok := hcomp.A.(ast.I0); !ok {
		t.Errorf("HComp.A: expected I0, got %v", hcomp.A)
	}
	// Phi: (i=1)[i0/i] at j+1=1 becomes FaceBot
	if _, ok := hcomp.Phi.(ast.FaceBot); !ok {
		t.Errorf("HComp.Phi: expected FaceBot, got %T", hcomp.Phi)
	}
	// Tube: IVar{1} at j+1=1, becomes I0
	if _, ok := hcomp.Tube.(ast.I0); !ok {
		t.Errorf("HComp.Tube: expected I0, got %v", hcomp.Tube)
	}
	// Base: IVar{0} at j=0, becomes I0
	if _, ok := hcomp.Base.(ast.I0); !ok {
		t.Errorf("HComp.Base: expected I0, got %v", hcomp.Base)
	}
}

// --- Fill (binds in A, Phi, Tube) ---

func TestISubst_Fill(t *testing.T) {
	t.Parallel()
	term := ast.Fill{
		IBinder: "i",
		A:       ast.IVar{Ix: 2}, // Under binder
		Phi:     ast.FaceEq{IVar: 0, IsOne: false},
		Tube:    ast.IVar{Ix: 1},
		Base:    ast.IVar{Ix: 1}, // Not under binder
	}
	result := ISubst(1, ast.I1{}, term)

	fill, ok := result.(ast.Fill)
	if !ok {
		t.Fatalf("expected Fill, got %T", result)
	}

	// A: IVar{2} at j+1=2, becomes I1
	if _, ok := fill.A.(ast.I1); !ok {
		t.Errorf("Fill.A: expected I1, got %v", fill.A)
	}
	// Phi: FaceEq{0,false} unchanged (0 < j+1=2)
	if phi, ok := fill.Phi.(ast.FaceEq); !ok || phi.IVar != 0 {
		t.Errorf("Fill.Phi: expected FaceEq{0,false}, got %v", fill.Phi)
	}
	// Tube: IVar{1} at j+1=2, below, unchanged
	if tube, ok := fill.Tube.(ast.IVar); !ok || tube.Ix != 1 {
		t.Errorf("Fill.Tube: expected IVar{1}, got %v", fill.Tube)
	}
	// Base: IVar{1} at j=1, becomes I1
	if _, ok := fill.Base.(ast.I1); !ok {
		t.Errorf("Fill.Base: expected I1, got %v", fill.Base)
	}
}

// --- Partial (non-binding) ---

func TestISubst_Partial(t *testing.T) {
	t.Parallel()
	term := ast.Partial{
		Phi: ast.FaceEq{IVar: 0, IsOne: true},
		A:   ast.IVar{Ix: 1},
	}
	result := ISubst(0, ast.I1{}, term)

	p, ok := result.(ast.Partial)
	if !ok {
		t.Fatalf("expected Partial, got %T", result)
	}

	// Phi: (i=1)[i1/i] = ⊤
	if _, ok := p.Phi.(ast.FaceTop); !ok {
		t.Errorf("Partial.Phi: expected FaceTop, got %T", p.Phi)
	}
	// A: IVar{1} above j=0, decremented
	if a, ok := p.A.(ast.IVar); !ok || a.Ix != 0 {
		t.Errorf("Partial.A: expected IVar{0}, got %v", p.A)
	}
}

// --- System (non-binding) ---

func TestISubst_System(t *testing.T) {
	t.Parallel()
	term := ast.System{
		Branches: []ast.SystemBranch{
			{Phi: ast.FaceEq{IVar: 0, IsOne: false}, Term: ast.IVar{Ix: 1}},
			{Phi: ast.FaceEq{IVar: 1, IsOne: true}, Term: ast.IVar{Ix: 0}},
		},
	}
	result := ISubst(0, ast.I0{}, term)

	sys, ok := result.(ast.System)
	if !ok {
		t.Fatalf("expected System, got %T", result)
	}

	if len(sys.Branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(sys.Branches))
	}

	// First branch: (i=0)[i0/i] = ⊤, Term decremented
	if _, ok := sys.Branches[0].Phi.(ast.FaceTop); !ok {
		t.Errorf("Branch 0 Phi: expected FaceTop, got %T", sys.Branches[0].Phi)
	}
	if tm, ok := sys.Branches[0].Term.(ast.IVar); !ok || tm.Ix != 0 {
		t.Errorf("Branch 0 Term: expected IVar{0}, got %v", sys.Branches[0].Term)
	}

	// Second branch: Phi decremented, Term unchanged
	if phi, ok := sys.Branches[1].Phi.(ast.FaceEq); !ok || phi.IVar != 0 {
		t.Errorf("Branch 1 Phi: expected FaceEq{0,true}, got %v", sys.Branches[1].Phi)
	}
	if _, ok := sys.Branches[1].Term.(ast.I0); !ok {
		t.Errorf("Branch 1 Term: expected I0, got %v", sys.Branches[1].Term)
	}
}

// ============================================================================
// ISubst Edge Cases
// ============================================================================

// --- Nested Binders ---

func TestISubst_NestedBinders(t *testing.T) {
	t.Parallel()
	// PathLam inside PathP - two levels of binders
	term := ast.PathP{
		A: ast.PathLam{
			Binder: "j",
			Body:   ast.IVar{Ix: 2}, // j+1+1=2 → external var
		},
		X: ast.IVar{Ix: 0},
		Y: ast.IVar{Ix: 0},
	}
	result := ISubst(0, ast.I1{}, term)

	pathp, ok := result.(ast.PathP)
	if !ok {
		t.Fatalf("expected PathP, got %T", result)
	}

	plam, ok := pathp.A.(ast.PathLam)
	if !ok {
		t.Fatalf("expected PathLam inside PathP.A, got %T", pathp.A)
	}

	// Body: IVar{2} with j+1+1=2 becomes I1
	if _, ok := plam.Body.(ast.I1); !ok {
		t.Errorf("nested body: expected I1, got %v", plam.Body)
	}
}

func TestISubst_TripleNestedBinders(t *testing.T) {
	t.Parallel()
	// Transport inside PathLam inside PathP
	term := ast.PathP{
		A: ast.PathLam{
			Binder: "j",
			Body: ast.Transport{
				A: ast.IVar{Ix: 3}, // j+1+1+1=3
				E: ast.IVar{Ix: 0}, // bound by Transport
			},
		},
		X: ast.I0{},
		Y: ast.I0{},
	}
	result := ISubst(0, ast.I0{}, term)

	pathp := result.(ast.PathP)
	plam := pathp.A.(ast.PathLam)
	tr := plam.Body.(ast.Transport)

	// A: IVar{3} with j+1+1+1=3 becomes I0
	if _, ok := tr.A.(ast.I0); !ok {
		t.Errorf("triple nested: expected I0, got %v", tr.A)
	}
}

// --- Deep Substitution Depth ---

func TestISubst_DeepSubstitution(t *testing.T) {
	t.Parallel()
	// Substitution at j=5
	term := ast.Path{
		A: ast.IVar{Ix: 4}, // below j=5, unchanged
		X: ast.IVar{Ix: 5}, // at j=5, substituted
		Y: ast.IVar{Ix: 6}, // above j=5, decremented
	}
	result := ISubst(5, ast.I1{}, term)

	path := result.(ast.Path)

	if a, ok := path.A.(ast.IVar); !ok || a.Ix != 4 {
		t.Errorf("A: expected IVar{4}, got %v", path.A)
	}
	if _, ok := path.X.(ast.I1); !ok {
		t.Errorf("X: expected I1, got %v", path.X)
	}
	if y, ok := path.Y.(ast.IVar); !ok || y.Ix != 5 {
		t.Errorf("Y: expected IVar{5}, got %v", path.Y)
	}
}

// --- FaceAnd Simplification ---

func TestISubst_FaceAnd_Simplify(t *testing.T) {
	t.Parallel()
	// ((i=0) ∧ (i=1))[i0/i] = ⊤ ∧ ⊥ = ⊥
	term := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceEq{IVar: 0, IsOne: true},
	}
	result := ISubst(0, ast.I0{}, term)

	if _, ok := result.(ast.FaceBot); !ok {
		t.Errorf("expected FaceBot, got %T", result)
	}
}

func TestISubst_FaceAnd_TopElimination(t *testing.T) {
	t.Parallel()
	// ((i=0) ∧ (j=1))[i0/i] = ⊤ ∧ (j=1) = (j=1) [j decremented to 0]
	term := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceEq{IVar: 1, IsOne: true},
	}
	result := ISubst(0, ast.I0{}, term)

	eq, ok := result.(ast.FaceEq)
	if !ok {
		t.Fatalf("expected FaceEq, got %T", result)
	}
	if eq.IVar != 0 || eq.IsOne != true {
		t.Errorf("expected FaceEq{0,true}, got FaceEq{%d,%v}", eq.IVar, eq.IsOne)
	}
}

// --- FaceOr Simplification ---

func TestISubst_FaceOr_Simplify(t *testing.T) {
	t.Parallel()
	// ((i=0) ∨ (i=1))[i0/i] = ⊤ ∨ ⊥ = ⊤
	term := ast.FaceOr{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceEq{IVar: 0, IsOne: true},
	}
	result := ISubst(0, ast.I0{}, term)

	if _, ok := result.(ast.FaceTop); !ok {
		t.Errorf("expected FaceTop, got %T", result)
	}
}

func TestISubst_FaceOr_BotElimination(t *testing.T) {
	t.Parallel()
	// ((i=1) ∨ (j=1))[i0/i] = ⊥ ∨ (j=1) = (j=1)
	term := ast.FaceOr{
		Left:  ast.FaceEq{IVar: 0, IsOne: true},
		Right: ast.FaceEq{IVar: 1, IsOne: true},
	}
	result := ISubst(0, ast.I0{}, term)

	eq, ok := result.(ast.FaceEq)
	if !ok {
		t.Fatalf("expected FaceEq, got %T", result)
	}
	if eq.IVar != 0 {
		t.Errorf("expected IVar 0 after decrement, got %d", eq.IVar)
	}
}

// --- Standard Terms (no interval binders) ---

func TestISubst_StandardTerms(t *testing.T) {
	t.Parallel()
	// Standard terms have no interval binders, just recurse
	term := ast.Pi{
		Binder: "x",
		A:      ast.IVar{Ix: 0},
		B:      ast.IVar{Ix: 1},
	}
	result := ISubst(0, ast.I1{}, term)

	pi := result.(ast.Pi)
	if _, ok := pi.A.(ast.I1); !ok {
		t.Errorf("Pi.A: expected I1, got %v", pi.A)
	}
	if b, ok := pi.B.(ast.IVar); !ok || b.Ix != 0 {
		t.Errorf("Pi.B: expected IVar{0}, got %v", pi.B)
	}
}

// ============================================================================
// shiftExtension Tests - Term Variable Shifting for Cubical Types
// ============================================================================

func TestShiftExtension_Interval_Unchanged(t *testing.T) {
	t.Parallel()
	terms := []ast.Term{ast.Interval{}, ast.I0{}, ast.I1{}, ast.IVar{Ix: 5}}

	for _, tm := range terms {
		result, ok := shiftExtension(1, 0, tm)
		if !ok {
			t.Errorf("shiftExtension should handle %T", tm)
		}
		if !reflect.DeepEqual(result, tm) {
			t.Errorf("interval terms should be unchanged, got %v", result)
		}
	}
}

func TestShiftExtension_Path(t *testing.T) {
	t.Parallel()
	term := ast.Path{
		A: ast.Var{Ix: 0},
		X: ast.Var{Ix: 1},
		Y: ast.Var{Ix: 2},
	}
	result, ok := shiftExtension(1, 1, term)
	if !ok {
		t.Fatal("shiftExtension should handle Path")
	}

	path := result.(ast.Path)
	if v, ok := path.A.(ast.Var); !ok || v.Ix != 0 {
		t.Errorf("Path.A: expected Var{0}, got %v", path.A)
	}
	if v, ok := path.X.(ast.Var); !ok || v.Ix != 2 {
		t.Errorf("Path.X: expected Var{2}, got %v", path.X)
	}
	if v, ok := path.Y.(ast.Var); !ok || v.Ix != 3 {
		t.Errorf("Path.Y: expected Var{3}, got %v", path.Y)
	}
}

func TestShiftExtension_PathP(t *testing.T) {
	t.Parallel()
	term := ast.PathP{
		A: ast.Var{Ix: 1},
		X: ast.Var{Ix: 0},
		Y: ast.Var{Ix: 2},
	}
	result, ok := shiftExtension(1, 1, term)
	if !ok {
		t.Fatal("shiftExtension should handle PathP")
	}

	pathp := result.(ast.PathP)
	if v, ok := pathp.A.(ast.Var); !ok || v.Ix != 2 {
		t.Errorf("PathP.A: expected Var{2}, got %v", pathp.A)
	}
}

func TestShiftExtension_PathLam(t *testing.T) {
	t.Parallel()
	term := ast.PathLam{Binder: "i", Body: ast.Var{Ix: 1}}
	result, ok := shiftExtension(1, 1, term)
	if !ok {
		t.Fatal("shiftExtension should handle PathLam")
	}

	plam := result.(ast.PathLam)
	if v, ok := plam.Body.(ast.Var); !ok || v.Ix != 2 {
		t.Errorf("PathLam.Body: expected Var{2}, got %v", plam.Body)
	}
}

func TestShiftExtension_PathApp(t *testing.T) {
	t.Parallel()
	term := ast.PathApp{P: ast.Var{Ix: 0}, R: ast.Var{Ix: 1}}
	result, ok := shiftExtension(1, 0, term)
	if !ok {
		t.Fatal("shiftExtension should handle PathApp")
	}

	papp := result.(ast.PathApp)
	if v, ok := papp.P.(ast.Var); !ok || v.Ix != 1 {
		t.Errorf("PathApp.P: expected Var{1}, got %v", papp.P)
	}
}

func TestShiftExtension_Transport(t *testing.T) {
	t.Parallel()
	term := ast.Transport{A: ast.Var{Ix: 1}, E: ast.Var{Ix: 0}}
	result, ok := shiftExtension(1, 1, term)
	if !ok {
		t.Fatal("shiftExtension should handle Transport")
	}

	tr := result.(ast.Transport)
	if v, ok := tr.A.(ast.Var); !ok || v.Ix != 2 {
		t.Errorf("Transport.A: expected Var{2}, got %v", tr.A)
	}
}

func TestShiftExtension_FaceFormulas(t *testing.T) {
	t.Parallel()
	faces := []ast.Term{
		ast.FaceTop{},
		ast.FaceBot{},
		ast.FaceEq{IVar: 0, IsOne: false},
	}

	for _, f := range faces {
		result, ok := shiftExtension(1, 0, f)
		if !ok {
			t.Errorf("shiftExtension should handle %T", f)
		}
		if !reflect.DeepEqual(result, f) {
			t.Errorf("face %T should be unchanged", f)
		}
	}
}

func TestShiftExtension_FaceAnd(t *testing.T) {
	t.Parallel()
	term := ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}
	result, ok := shiftExtension(1, 0, term)
	if !ok {
		t.Fatal("shiftExtension should handle FaceAnd")
	}

	and := result.(ast.FaceAnd)
	if _, ok := and.Left.(ast.FaceTop); !ok {
		t.Errorf("FaceAnd.Left should be FaceTop")
	}
}

func TestShiftExtension_FaceOr(t *testing.T) {
	t.Parallel()
	term := ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}}
	result, ok := shiftExtension(1, 0, term)
	if !ok {
		t.Fatal("shiftExtension should handle FaceOr")
	}
	if _, ok := result.(ast.FaceOr); !ok {
		t.Errorf("expected FaceOr, got %T", result)
	}
}

func TestShiftExtension_Partial(t *testing.T) {
	t.Parallel()
	term := ast.Partial{Phi: ast.FaceTop{}, A: ast.Var{Ix: 0}}
	result, ok := shiftExtension(1, 0, term)
	if !ok {
		t.Fatal("shiftExtension should handle Partial")
	}

	p := result.(ast.Partial)
	if v, ok := p.A.(ast.Var); !ok || v.Ix != 1 {
		t.Errorf("Partial.A: expected Var{1}, got %v", p.A)
	}
}

func TestShiftExtension_System(t *testing.T) {
	t.Parallel()
	term := ast.System{
		Branches: []ast.SystemBranch{
			{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 1}},
		},
	}
	result, ok := shiftExtension(1, 1, term)
	if !ok {
		t.Fatal("shiftExtension should handle System")
	}

	sys := result.(ast.System)
	if v, ok := sys.Branches[0].Term.(ast.Var); !ok || v.Ix != 2 {
		t.Errorf("System branch Term: expected Var{2}, got %v", sys.Branches[0].Term)
	}
}

func TestShiftExtension_Comp(t *testing.T) {
	t.Parallel()
	term := ast.Comp{
		IBinder: "i",
		A:       ast.Var{Ix: 0},
		Phi:     ast.FaceTop{},
		Tube:    ast.Var{Ix: 1},
		Base:    ast.Var{Ix: 2},
	}
	result, ok := shiftExtension(1, 1, term)
	if !ok {
		t.Fatal("shiftExtension should handle Comp")
	}

	comp := result.(ast.Comp)
	if v, ok := comp.Tube.(ast.Var); !ok || v.Ix != 2 {
		t.Errorf("Comp.Tube: expected Var{2}, got %v", comp.Tube)
	}
	if v, ok := comp.Base.(ast.Var); !ok || v.Ix != 3 {
		t.Errorf("Comp.Base: expected Var{3}, got %v", comp.Base)
	}
}

func TestShiftExtension_HComp(t *testing.T) {
	t.Parallel()
	term := ast.HComp{A: ast.Var{Ix: 1}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 2}}
	result, ok := shiftExtension(1, 1, term)
	if !ok {
		t.Fatal("shiftExtension should handle HComp")
	}

	hcomp := result.(ast.HComp)
	if v, ok := hcomp.A.(ast.Var); !ok || v.Ix != 2 {
		t.Errorf("HComp.A: expected Var{2}, got %v", hcomp.A)
	}
}

func TestShiftExtension_Fill(t *testing.T) {
	t.Parallel()
	term := ast.Fill{IBinder: "i", A: ast.Var{Ix: 1}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 2}}
	result, ok := shiftExtension(1, 1, term)
	if !ok {
		t.Fatal("shiftExtension should handle Fill")
	}

	fill := result.(ast.Fill)
	if v, ok := fill.A.(ast.Var); !ok || v.Ix != 2 {
		t.Errorf("Fill.A: expected Var{2}, got %v", fill.A)
	}
}

func TestShiftExtension_Glue(t *testing.T) {
	t.Parallel()
	term := ast.Glue{
		A: ast.Var{Ix: 1},
		System: []ast.GlueBranch{
			{Phi: ast.FaceTop{}, T: ast.Var{Ix: 2}, Equiv: ast.Var{Ix: 0}},
		},
	}
	result, ok := shiftExtension(1, 1, term)
	if !ok {
		t.Fatal("shiftExtension should handle Glue")
	}

	glue := result.(ast.Glue)
	if v, ok := glue.A.(ast.Var); !ok || v.Ix != 2 {
		t.Errorf("Glue.A: expected Var{2}, got %v", glue.A)
	}
	if v, ok := glue.System[0].T.(ast.Var); !ok || v.Ix != 3 {
		t.Errorf("Glue branch T: expected Var{3}, got %v", glue.System[0].T)
	}
}

func TestShiftExtension_GlueElem(t *testing.T) {
	t.Parallel()
	term := ast.GlueElem{
		System: []ast.GlueElemBranch{{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 1}}},
		Base:   ast.Var{Ix: 0},
	}
	result, ok := shiftExtension(1, 0, term)
	if !ok {
		t.Fatal("shiftExtension should handle GlueElem")
	}

	ge := result.(ast.GlueElem)
	if v, ok := ge.Base.(ast.Var); !ok || v.Ix != 1 {
		t.Errorf("GlueElem.Base: expected Var{1}, got %v", ge.Base)
	}
}

func TestShiftExtension_Unglue(t *testing.T) {
	t.Parallel()
	term := ast.Unglue{Ty: ast.Var{Ix: 0}, G: ast.Var{Ix: 1}}
	result, ok := shiftExtension(1, 0, term)
	if !ok {
		t.Fatal("shiftExtension should handle Unglue")
	}

	ug := result.(ast.Unglue)
	if v, ok := ug.Ty.(ast.Var); !ok || v.Ix != 1 {
		t.Errorf("Unglue.Ty: expected Var{1}, got %v", ug.Ty)
	}
}

func TestShiftExtension_UA(t *testing.T) {
	t.Parallel()
	term := ast.UA{A: ast.Var{Ix: 0}, B: ast.Var{Ix: 1}, Equiv: ast.Var{Ix: 2}}
	result, ok := shiftExtension(1, 1, term)
	if !ok {
		t.Fatal("shiftExtension should handle UA")
	}

	ua := result.(ast.UA)
	if v, ok := ua.B.(ast.Var); !ok || v.Ix != 2 {
		t.Errorf("UA.B: expected Var{2}, got %v", ua.B)
	}
}

func TestShiftExtension_UABeta(t *testing.T) {
	t.Parallel()
	term := ast.UABeta{Equiv: ast.Var{Ix: 1}, Arg: ast.Var{Ix: 0}}
	result, ok := shiftExtension(1, 0, term)
	if !ok {
		t.Fatal("shiftExtension should handle UABeta")
	}

	uab := result.(ast.UABeta)
	if v, ok := uab.Equiv.(ast.Var); !ok || v.Ix != 2 {
		t.Errorf("UABeta.Equiv: expected Var{2}, got %v", uab.Equiv)
	}
}

func TestShiftExtension_Unknown(t *testing.T) {
	t.Parallel()
	_, ok := shiftExtension(1, 0, ast.Global{Name: "test"})
	if ok {
		t.Error("shiftExtension should return false for unknown types")
	}
}

// ============================================================================
// substExtension Tests - Term Variable Substitution for Cubical Types
// ============================================================================

func TestSubstExtension_Interval_Unchanged(t *testing.T) {
	t.Parallel()
	terms := []ast.Term{ast.Interval{}, ast.I0{}, ast.I1{}, ast.IVar{Ix: 3}}

	for _, tm := range terms {
		result, ok := substExtension(0, ast.Sort{U: 0}, tm)
		if !ok {
			t.Errorf("substExtension should handle %T", tm)
		}
		if !reflect.DeepEqual(result, tm) {
			t.Errorf("interval terms should be unchanged, got %v", result)
		}
	}
}

func TestSubstExtension_Path(t *testing.T) {
	t.Parallel()
	term := ast.Path{
		A: ast.Var{Ix: 0},
		X: ast.Sort{U: 0},
		Y: ast.Var{Ix: 1},
	}
	result, ok := substExtension(0, ast.Sort{U: 1}, term)
	if !ok {
		t.Fatal("substExtension should handle Path")
	}

	path := result.(ast.Path)
	if s, ok := path.A.(ast.Sort); !ok || s.U != 1 {
		t.Errorf("Path.A: expected Sort{1}, got %v", path.A)
	}
	if v, ok := path.Y.(ast.Var); !ok || v.Ix != 0 {
		t.Errorf("Path.Y: expected Var{0}, got %v", path.Y)
	}
}

func TestSubstExtension_PathP(t *testing.T) {
	t.Parallel()
	term := ast.PathP{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 2}}
	result, ok := substExtension(0, ast.I0{}, term)
	if !ok {
		t.Fatal("substExtension should handle PathP")
	}

	pathp := result.(ast.PathP)
	if _, ok := pathp.A.(ast.I0); !ok {
		t.Errorf("PathP.A: expected I0, got %v", pathp.A)
	}
}

func TestSubstExtension_PathLam(t *testing.T) {
	t.Parallel()
	term := ast.PathLam{Binder: "i", Body: ast.Var{Ix: 0}}
	result, ok := substExtension(0, ast.Sort{U: 5}, term)
	if !ok {
		t.Fatal("substExtension should handle PathLam")
	}

	plam := result.(ast.PathLam)
	if s, ok := plam.Body.(ast.Sort); !ok || s.U != 5 {
		t.Errorf("PathLam.Body: expected Sort{5}, got %v", plam.Body)
	}
}

func TestSubstExtension_PathApp(t *testing.T) {
	t.Parallel()
	term := ast.PathApp{P: ast.Var{Ix: 0}, R: ast.IVar{Ix: 0}}
	result, ok := substExtension(0, ast.Sort{U: 0}, term)
	if !ok {
		t.Fatal("substExtension should handle PathApp")
	}

	papp := result.(ast.PathApp)
	if s, ok := papp.P.(ast.Sort); !ok || s.U != 0 {
		t.Errorf("PathApp.P: expected Sort{0}, got %v", papp.P)
	}
	// R is IVar, unchanged by term subst
	if _, ok := papp.R.(ast.IVar); !ok {
		t.Errorf("PathApp.R should be IVar, got %T", papp.R)
	}
}

func TestSubstExtension_Transport(t *testing.T) {
	t.Parallel()
	term := ast.Transport{A: ast.Var{Ix: 0}, E: ast.Var{Ix: 1}}
	result, ok := substExtension(0, ast.I0{}, term)
	if !ok {
		t.Fatal("substExtension should handle Transport")
	}

	tr := result.(ast.Transport)
	if _, ok := tr.A.(ast.I0); !ok {
		t.Errorf("Transport.A: expected I0, got %v", tr.A)
	}
}

func TestSubstExtension_FaceFormulas(t *testing.T) {
	t.Parallel()
	faces := []ast.Term{ast.FaceTop{}, ast.FaceBot{}, ast.FaceEq{IVar: 0, IsOne: false}}

	for _, f := range faces {
		result, ok := substExtension(0, ast.Sort{U: 0}, f)
		if !ok {
			t.Errorf("substExtension should handle %T", f)
		}
		if !reflect.DeepEqual(result, f) {
			t.Errorf("face %T should be unchanged", f)
		}
	}
}

func TestSubstExtension_FaceAnd(t *testing.T) {
	t.Parallel()
	term := ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}
	result, ok := substExtension(0, ast.Sort{U: 0}, term)
	if !ok {
		t.Fatal("substExtension should handle FaceAnd")
	}
	if _, ok := result.(ast.FaceAnd); !ok {
		t.Errorf("expected FaceAnd, got %T", result)
	}
}

func TestSubstExtension_FaceOr(t *testing.T) {
	t.Parallel()
	term := ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}}
	result, ok := substExtension(0, ast.Sort{U: 0}, term)
	if !ok {
		t.Fatal("substExtension should handle FaceOr")
	}
	if _, ok := result.(ast.FaceOr); !ok {
		t.Errorf("expected FaceOr, got %T", result)
	}
}

func TestSubstExtension_Partial(t *testing.T) {
	t.Parallel()
	term := ast.Partial{Phi: ast.FaceTop{}, A: ast.Var{Ix: 0}}
	result, ok := substExtension(0, ast.Sort{U: 3}, term)
	if !ok {
		t.Fatal("substExtension should handle Partial")
	}

	p := result.(ast.Partial)
	if s, ok := p.A.(ast.Sort); !ok || s.U != 3 {
		t.Errorf("Partial.A: expected Sort{3}, got %v", p.A)
	}
}

func TestSubstExtension_System(t *testing.T) {
	t.Parallel()
	term := ast.System{
		Branches: []ast.SystemBranch{
			{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}},
			{Phi: ast.FaceBot{}, Term: ast.Var{Ix: 1}},
		},
	}
	result, ok := substExtension(0, ast.Sort{U: 5}, term)
	if !ok {
		t.Fatal("substExtension should handle System")
	}

	sys := result.(ast.System)
	if s, ok := sys.Branches[0].Term.(ast.Sort); !ok || s.U != 5 {
		t.Errorf("Branch 0 Term: expected Sort{5}, got %v", sys.Branches[0].Term)
	}
	if v, ok := sys.Branches[1].Term.(ast.Var); !ok || v.Ix != 0 {
		t.Errorf("Branch 1 Term: expected Var{0}, got %v", sys.Branches[1].Term)
	}
}

func TestSubstExtension_Comp(t *testing.T) {
	t.Parallel()
	term := ast.Comp{
		IBinder: "i",
		A:       ast.Var{Ix: 0},
		Phi:     ast.FaceTop{},
		Tube:    ast.Var{Ix: 1},
		Base:    ast.Var{Ix: 2},
	}
	result, ok := substExtension(0, ast.I0{}, term)
	if !ok {
		t.Fatal("substExtension should handle Comp")
	}

	comp := result.(ast.Comp)
	if _, ok := comp.A.(ast.I0); !ok {
		t.Errorf("Comp.A: expected I0, got %v", comp.A)
	}
}

func TestSubstExtension_HComp(t *testing.T) {
	t.Parallel()
	term := ast.HComp{A: ast.Var{Ix: 1}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 2}}
	result, ok := substExtension(1, ast.I1{}, term)
	if !ok {
		t.Fatal("substExtension should handle HComp")
	}

	hcomp := result.(ast.HComp)
	if _, ok := hcomp.A.(ast.I1); !ok {
		t.Errorf("HComp.A: expected I1, got %v", hcomp.A)
	}
}

func TestSubstExtension_Fill(t *testing.T) {
	t.Parallel()
	term := ast.Fill{IBinder: "i", A: ast.Var{Ix: 0}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 1}, Base: ast.Var{Ix: 2}}
	result, ok := substExtension(0, ast.Sort{U: 0}, term)
	if !ok {
		t.Fatal("substExtension should handle Fill")
	}

	fill := result.(ast.Fill)
	if s, ok := fill.A.(ast.Sort); !ok || s.U != 0 {
		t.Errorf("Fill.A: expected Sort{0}, got %v", fill.A)
	}
}

func TestSubstExtension_Glue(t *testing.T) {
	t.Parallel()
	term := ast.Glue{
		A: ast.Var{Ix: 0},
		System: []ast.GlueBranch{
			{Phi: ast.FaceTop{}, T: ast.Var{Ix: 1}, Equiv: ast.Var{Ix: 2}},
		},
	}
	result, ok := substExtension(0, ast.I0{}, term)
	if !ok {
		t.Fatal("substExtension should handle Glue")
	}

	glue := result.(ast.Glue)
	if _, ok := glue.A.(ast.I0); !ok {
		t.Errorf("Glue.A: expected I0, got %v", glue.A)
	}
	if v, ok := glue.System[0].T.(ast.Var); !ok || v.Ix != 0 {
		t.Errorf("Glue branch T: expected Var{0}, got %v", glue.System[0].T)
	}
}

func TestSubstExtension_GlueElem(t *testing.T) {
	t.Parallel()
	term := ast.GlueElem{
		System: []ast.GlueElemBranch{{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}}},
		Base:   ast.Var{Ix: 1},
	}
	result, ok := substExtension(0, ast.I0{}, term)
	if !ok {
		t.Fatal("substExtension should handle GlueElem")
	}

	ge := result.(ast.GlueElem)
	if _, ok := ge.System[0].Term.(ast.I0); !ok {
		t.Errorf("GlueElem branch Term: expected I0, got %v", ge.System[0].Term)
	}
}

func TestSubstExtension_Unglue(t *testing.T) {
	t.Parallel()
	term := ast.Unglue{Ty: ast.Var{Ix: 0}, G: ast.Var{Ix: 1}}
	result, ok := substExtension(0, ast.Sort{U: 7}, term)
	if !ok {
		t.Fatal("substExtension should handle Unglue")
	}

	ug := result.(ast.Unglue)
	if s, ok := ug.Ty.(ast.Sort); !ok || s.U != 7 {
		t.Errorf("Unglue.Ty: expected Sort{7}, got %v", ug.Ty)
	}
}

func TestSubstExtension_UA(t *testing.T) {
	t.Parallel()
	term := ast.UA{A: ast.Var{Ix: 0}, B: ast.Var{Ix: 1}, Equiv: ast.Var{Ix: 2}}
	result, ok := substExtension(1, ast.I0{}, term)
	if !ok {
		t.Fatal("substExtension should handle UA")
	}

	ua := result.(ast.UA)
	if _, ok := ua.B.(ast.I0); !ok {
		t.Errorf("UA.B: expected I0, got %v", ua.B)
	}
}

func TestSubstExtension_UABeta(t *testing.T) {
	t.Parallel()
	term := ast.UABeta{Equiv: ast.Var{Ix: 0}, Arg: ast.Var{Ix: 2}}
	result, ok := substExtension(0, ast.Sort{U: 3}, term)
	if !ok {
		t.Fatal("substExtension should handle UABeta")
	}

	uab := result.(ast.UABeta)
	if s, ok := uab.Equiv.(ast.Sort); !ok || s.U != 3 {
		t.Errorf("UABeta.Equiv: expected Sort{3}, got %v", uab.Equiv)
	}
	if v, ok := uab.Arg.(ast.Var); !ok || v.Ix != 1 {
		t.Errorf("UABeta.Arg: expected Var{1}, got %v", uab.Arg)
	}
}

func TestSubstExtension_Unknown(t *testing.T) {
	t.Parallel()
	_, ok := substExtension(0, ast.Sort{U: 0}, ast.Global{Name: "test"})
	if ok {
		t.Error("substExtension should return false for unknown types")
	}
}

// ============================================================================
// ISubst Direct Face Formula Tests
// ============================================================================

func TestISubst_FaceTop_Direct(t *testing.T) {
	t.Parallel()
	// FaceTop as direct term (not in FaceAnd/FaceOr) should be unchanged
	result := ISubst(0, ast.I0{}, ast.FaceTop{})
	if _, ok := result.(ast.FaceTop); !ok {
		t.Errorf("expected FaceTop, got %T", result)
	}
}

func TestISubst_FaceBot_Direct(t *testing.T) {
	t.Parallel()
	result := ISubst(0, ast.I0{}, ast.FaceBot{})
	if _, ok := result.(ast.FaceBot); !ok {
		t.Errorf("expected FaceBot, got %T", result)
	}
}

func TestISubst_FaceEq_AtJ_SubstI0_IsZero(t *testing.T) {
	t.Parallel()
	// (i = 0)[i0/i] = ⊤
	term := ast.FaceEq{IVar: 0, IsOne: false}
	result := ISubst(0, ast.I0{}, term)
	if _, ok := result.(ast.FaceTop); !ok {
		t.Errorf("(i=0)[i0/i] should be FaceTop, got %T", result)
	}
}

func TestISubst_FaceEq_AtJ_SubstI0_IsOne(t *testing.T) {
	t.Parallel()
	// (i = 1)[i0/i] = ⊥
	term := ast.FaceEq{IVar: 0, IsOne: true}
	result := ISubst(0, ast.I0{}, term)
	if _, ok := result.(ast.FaceBot); !ok {
		t.Errorf("(i=1)[i0/i] should be FaceBot, got %T", result)
	}
}

func TestISubst_FaceEq_AtJ_SubstI1_IsZero(t *testing.T) {
	t.Parallel()
	// (i = 0)[i1/i] = ⊥
	term := ast.FaceEq{IVar: 0, IsOne: false}
	result := ISubst(0, ast.I1{}, term)
	if _, ok := result.(ast.FaceBot); !ok {
		t.Errorf("(i=0)[i1/i] should be FaceBot, got %T", result)
	}
}

func TestISubst_FaceEq_AtJ_SubstI1_IsOne(t *testing.T) {
	t.Parallel()
	// (i = 1)[i1/i] = ⊤
	term := ast.FaceEq{IVar: 0, IsOne: true}
	result := ISubst(0, ast.I1{}, term)
	if _, ok := result.(ast.FaceTop); !ok {
		t.Errorf("(i=1)[i1/i] should be FaceTop, got %T", result)
	}
}

func TestISubst_FaceEq_AtJ_SubstIVar(t *testing.T) {
	t.Parallel()
	// (i = 0)[j/i] = (j = 0)
	term := ast.FaceEq{IVar: 0, IsOne: false}
	result := ISubst(0, ast.IVar{Ix: 5}, term)
	eq, ok := result.(ast.FaceEq)
	if !ok {
		t.Fatalf("expected FaceEq, got %T", result)
	}
	if eq.IVar != 5 || eq.IsOne != false {
		t.Errorf("expected FaceEq{5, false}, got FaceEq{%d, %v}", eq.IVar, eq.IsOne)
	}
}

func TestISubst_FaceEq_AtJ_SubstOther(t *testing.T) {
	t.Parallel()
	// (i = 0)[Sort/i] - unusual but should return original
	term := ast.FaceEq{IVar: 0, IsOne: false}
	result := ISubst(0, ast.Sort{U: 0}, term)
	if eq, ok := result.(ast.FaceEq); !ok || eq.IVar != 0 {
		t.Errorf("expected original FaceEq{0, false}, got %v", result)
	}
}

func TestISubst_FaceEq_AboveJ(t *testing.T) {
	t.Parallel()
	// (j = 1) where j > substitution index → decrement
	term := ast.FaceEq{IVar: 2, IsOne: true}
	result := ISubst(1, ast.I0{}, term)
	eq, ok := result.(ast.FaceEq)
	if !ok {
		t.Fatalf("expected FaceEq, got %T", result)
	}
	if eq.IVar != 1 {
		t.Errorf("expected IVar 1 (decremented), got %d", eq.IVar)
	}
}

func TestISubst_FaceEq_BelowJ(t *testing.T) {
	t.Parallel()
	// (i = 0) where i < substitution index → unchanged
	term := ast.FaceEq{IVar: 0, IsOne: false}
	result := ISubst(2, ast.I0{}, term)
	eq, ok := result.(ast.FaceEq)
	if !ok {
		t.Fatalf("expected FaceEq, got %T", result)
	}
	if eq.IVar != 0 {
		t.Errorf("expected IVar 0 (unchanged), got %d", eq.IVar)
	}
}

// ============================================================================
// ISubstFace Tests
// ============================================================================

func TestISubstFace_Nil(t *testing.T) {
	t.Parallel()
	result := ISubstFace(0, ast.I0{}, nil)
	if result != nil {
		t.Errorf("ISubstFace(nil) should return nil, got %v", result)
	}
}

func TestISubstFace_FaceTop(t *testing.T) {
	t.Parallel()
	result := ISubstFace(0, ast.I0{}, ast.FaceTop{})
	if _, ok := result.(ast.FaceTop); !ok {
		t.Errorf("expected FaceTop, got %T", result)
	}
}

func TestISubstFace_FaceBot(t *testing.T) {
	t.Parallel()
	result := ISubstFace(0, ast.I0{}, ast.FaceBot{})
	if _, ok := result.(ast.FaceBot); !ok {
		t.Errorf("expected FaceBot, got %T", result)
	}
}

func TestISubstFace_FaceEq_SubstIVar(t *testing.T) {
	t.Parallel()
	// (i = 1)[j/i] = (j = 1)
	face := ast.FaceEq{IVar: 0, IsOne: true}
	result := ISubstFace(0, ast.IVar{Ix: 3}, face)
	eq, ok := result.(ast.FaceEq)
	if !ok {
		t.Fatalf("expected FaceEq, got %T", result)
	}
	if eq.IVar != 3 || eq.IsOne != true {
		t.Errorf("expected FaceEq{3, true}, got FaceEq{%d, %v}", eq.IVar, eq.IsOne)
	}
}

func TestISubstFace_FaceEq_UnknownSubst(t *testing.T) {
	t.Parallel()
	// (i = 0)[Sort/i] - unusual term, return unchanged
	face := ast.FaceEq{IVar: 0, IsOne: false}
	result := ISubstFace(0, ast.Sort{U: 0}, face)
	eq, ok := result.(ast.FaceEq)
	if !ok || eq.IVar != 0 {
		t.Errorf("expected original face, got %v", result)
	}
}

func TestISubstFace_FaceAnd(t *testing.T) {
	t.Parallel()
	// ((i=0) ∧ (j=1))[i0/i] = ⊤ ∧ (j-1=1)
	face := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 0, IsOne: false},
		Right: ast.FaceEq{IVar: 1, IsOne: true},
	}
	result := ISubstFace(0, ast.I0{}, face)
	// Should simplify to just the right side with decremented var
	eq, ok := result.(ast.FaceEq)
	if !ok {
		t.Fatalf("expected FaceEq after simplification, got %T", result)
	}
	if eq.IVar != 0 || eq.IsOne != true {
		t.Errorf("expected FaceEq{0, true}, got FaceEq{%d, %v}", eq.IVar, eq.IsOne)
	}
}

func TestISubstFace_FaceOr(t *testing.T) {
	t.Parallel()
	// ((i=1) ∨ (j=0))[i0/i] = ⊥ ∨ (j-1=0)
	face := ast.FaceOr{
		Left:  ast.FaceEq{IVar: 0, IsOne: true},
		Right: ast.FaceEq{IVar: 1, IsOne: false},
	}
	result := ISubstFace(0, ast.I0{}, face)
	// Should simplify to just the right side
	eq, ok := result.(ast.FaceEq)
	if !ok {
		t.Fatalf("expected FaceEq after simplification, got %T", result)
	}
	if eq.IVar != 0 {
		t.Errorf("expected IVar 0, got %d", eq.IVar)
	}
}

func TestISubstFace_Unknown(t *testing.T) {
	t.Parallel()
	// Unknown face type returns unchanged
	type CustomFace struct{}
	var custom ast.Face = nil // Edge case: type assertion handles unknown
	result := ISubstFace(0, ast.I0{}, custom)
	if result != custom {
		t.Errorf("expected custom face unchanged")
	}
}
