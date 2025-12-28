package eval

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func TestAlphaEqVar(t *testing.T) {
	tests := []struct {
		name string
		a, b ast.Term
		want bool
	}{
		{"same var", ast.Var{Ix: 0}, ast.Var{Ix: 0}, true},
		{"diff var", ast.Var{Ix: 0}, ast.Var{Ix: 1}, false},
		{"var vs global", ast.Var{Ix: 0}, ast.Global{Name: "x"}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AlphaEq(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("AlphaEq(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestAlphaEqGlobal(t *testing.T) {
	tests := []struct {
		name string
		a, b ast.Term
		want bool
	}{
		{"same global", ast.Global{Name: "foo"}, ast.Global{Name: "foo"}, true},
		{"diff global", ast.Global{Name: "foo"}, ast.Global{Name: "bar"}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AlphaEq(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("AlphaEq(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestAlphaEqSort(t *testing.T) {
	tests := []struct {
		name string
		a, b ast.Term
		want bool
	}{
		{"same sort", ast.Sort{U: 0}, ast.Sort{U: 0}, true},
		{"diff sort", ast.Sort{U: 0}, ast.Sort{U: 1}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AlphaEq(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("AlphaEq(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestAlphaEqLam(t *testing.T) {
	tests := []struct {
		name string
		a, b ast.Term
		want bool
	}{
		{
			"same lam diff binder names",
			ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}},
			ast.Lam{Binder: "y", Body: ast.Var{Ix: 0}},
			true, // binder names irrelevant
		},
		{
			"diff bodies",
			ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}},
			ast.Lam{Binder: "x", Body: ast.Var{Ix: 1}},
			false,
		},
		{
			"nested lam same",
			ast.Lam{Binder: "x", Body: ast.Lam{Binder: "y", Body: ast.Var{Ix: 1}}},
			ast.Lam{Binder: "a", Body: ast.Lam{Binder: "b", Body: ast.Var{Ix: 1}}},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AlphaEq(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("AlphaEq(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestAlphaEqApp(t *testing.T) {
	tests := []struct {
		name string
		a, b ast.Term
		want bool
	}{
		{
			"same app",
			ast.App{T: ast.Var{Ix: 0}, U: ast.Var{Ix: 1}},
			ast.App{T: ast.Var{Ix: 0}, U: ast.Var{Ix: 1}},
			true,
		},
		{
			"diff app func",
			ast.App{T: ast.Var{Ix: 0}, U: ast.Var{Ix: 1}},
			ast.App{T: ast.Var{Ix: 2}, U: ast.Var{Ix: 1}},
			false,
		},
		{
			"diff app arg",
			ast.App{T: ast.Var{Ix: 0}, U: ast.Var{Ix: 1}},
			ast.App{T: ast.Var{Ix: 0}, U: ast.Var{Ix: 2}},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AlphaEq(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("AlphaEq(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestAlphaEqPi(t *testing.T) {
	tests := []struct {
		name string
		a, b ast.Term
		want bool
	}{
		{
			"same pi diff binder names",
			ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}},
			ast.Pi{Binder: "y", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}},
			true, // binder names irrelevant
		},
		{
			"diff domain",
			ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}},
			ast.Pi{Binder: "x", A: ast.Sort{U: 1}, B: ast.Var{Ix: 0}},
			false,
		},
		{
			"diff codomain",
			ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}},
			ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 1}},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AlphaEq(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("AlphaEq(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestAlphaEqSigma(t *testing.T) {
	tests := []struct {
		name string
		a, b ast.Term
		want bool
	}{
		{
			"same sigma diff binder names",
			ast.Sigma{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}},
			ast.Sigma{Binder: "y", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}},
			true,
		},
		{
			"diff sigma",
			ast.Sigma{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}},
			ast.Sigma{Binder: "x", A: ast.Sort{U: 1}, B: ast.Var{Ix: 0}},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AlphaEq(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("AlphaEq(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestAlphaEqPair(t *testing.T) {
	tests := []struct {
		name string
		a, b ast.Term
		want bool
	}{
		{
			"same pair",
			ast.Pair{Fst: ast.Var{Ix: 0}, Snd: ast.Var{Ix: 1}},
			ast.Pair{Fst: ast.Var{Ix: 0}, Snd: ast.Var{Ix: 1}},
			true,
		},
		{
			"diff pair",
			ast.Pair{Fst: ast.Var{Ix: 0}, Snd: ast.Var{Ix: 1}},
			ast.Pair{Fst: ast.Var{Ix: 0}, Snd: ast.Var{Ix: 2}},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AlphaEq(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("AlphaEq(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestAlphaEqFstSnd(t *testing.T) {
	tests := []struct {
		name string
		a, b ast.Term
		want bool
	}{
		{"same fst", ast.Fst{P: ast.Var{Ix: 0}}, ast.Fst{P: ast.Var{Ix: 0}}, true},
		{"diff fst", ast.Fst{P: ast.Var{Ix: 0}}, ast.Fst{P: ast.Var{Ix: 1}}, false},
		{"same snd", ast.Snd{P: ast.Var{Ix: 0}}, ast.Snd{P: ast.Var{Ix: 0}}, true},
		{"diff snd", ast.Snd{P: ast.Var{Ix: 0}}, ast.Snd{P: ast.Var{Ix: 1}}, false},
		{"fst vs snd", ast.Fst{P: ast.Var{Ix: 0}}, ast.Snd{P: ast.Var{Ix: 0}}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AlphaEq(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("AlphaEq(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestAlphaEqLet(t *testing.T) {
	tests := []struct {
		name string
		a, b ast.Term
		want bool
	}{
		{
			"same let diff binder names",
			ast.Let{Binder: "x", Val: ast.Var{Ix: 0}, Body: ast.Var{Ix: 0}},
			ast.Let{Binder: "y", Val: ast.Var{Ix: 0}, Body: ast.Var{Ix: 0}},
			true, // binder names irrelevant
		},
		{
			"diff val",
			ast.Let{Binder: "x", Val: ast.Var{Ix: 0}, Body: ast.Var{Ix: 0}},
			ast.Let{Binder: "x", Val: ast.Var{Ix: 1}, Body: ast.Var{Ix: 0}},
			false,
		},
		{
			"diff body",
			ast.Let{Binder: "x", Val: ast.Var{Ix: 0}, Body: ast.Var{Ix: 0}},
			ast.Let{Binder: "x", Val: ast.Var{Ix: 0}, Body: ast.Var{Ix: 1}},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AlphaEq(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("AlphaEq(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestAlphaEqId(t *testing.T) {
	tests := []struct {
		name string
		a, b ast.Term
		want bool
	}{
		{
			"same id",
			ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
			ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
			true,
		},
		{
			"diff id type",
			ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
			ast.Id{A: ast.Sort{U: 1}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
			false,
		},
		{
			"diff id endpoints",
			ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
			ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 2}, Y: ast.Var{Ix: 1}},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AlphaEq(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("AlphaEq(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestAlphaEqRefl(t *testing.T) {
	tests := []struct {
		name string
		a, b ast.Term
		want bool
	}{
		{
			"same refl",
			ast.Refl{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}},
			ast.Refl{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}},
			true,
		},
		{
			"diff refl",
			ast.Refl{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}},
			ast.Refl{A: ast.Sort{U: 0}, X: ast.Var{Ix: 1}},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AlphaEq(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("AlphaEq(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestAlphaEqJ(t *testing.T) {
	j1 := ast.J{
		A: ast.Sort{U: 0},
		C: ast.Var{Ix: 0},
		D: ast.Var{Ix: 1},
		X: ast.Var{Ix: 2},
		Y: ast.Var{Ix: 3},
		P: ast.Var{Ix: 4},
	}
	j2 := ast.J{
		A: ast.Sort{U: 0},
		C: ast.Var{Ix: 0},
		D: ast.Var{Ix: 1},
		X: ast.Var{Ix: 2},
		Y: ast.Var{Ix: 3},
		P: ast.Var{Ix: 4},
	}
	j3 := ast.J{
		A: ast.Sort{U: 1}, // different
		C: ast.Var{Ix: 0},
		D: ast.Var{Ix: 1},
		X: ast.Var{Ix: 2},
		Y: ast.Var{Ix: 3},
		P: ast.Var{Ix: 4},
	}

	if !AlphaEq(j1, j2) {
		t.Error("expected identical J terms to be alpha-equal")
	}
	if AlphaEq(j1, j3) {
		t.Error("expected different J terms to not be alpha-equal")
	}
}

func TestAlphaEqNil(t *testing.T) {
	if !AlphaEq(nil, nil) {
		t.Error("expected nil, nil to be alpha-equal")
	}
	if AlphaEq(nil, ast.Var{Ix: 0}) {
		t.Error("expected nil, term to not be alpha-equal")
	}
	if AlphaEq(ast.Var{Ix: 0}, nil) {
		t.Error("expected term, nil to not be alpha-equal")
	}
}

// --- Cubical Terms ---

func TestAlphaEqInterval(t *testing.T) {
	if !AlphaEq(ast.Interval{}, ast.Interval{}) {
		t.Error("expected Interval to be alpha-equal to Interval")
	}
	if !AlphaEq(ast.I0{}, ast.I0{}) {
		t.Error("expected I0 to be alpha-equal to I0")
	}
	if !AlphaEq(ast.I1{}, ast.I1{}) {
		t.Error("expected I1 to be alpha-equal to I1")
	}
	if AlphaEq(ast.I0{}, ast.I1{}) {
		t.Error("expected I0 to not be alpha-equal to I1")
	}
}

func TestAlphaEqIVar(t *testing.T) {
	if !AlphaEq(ast.IVar{Ix: 0}, ast.IVar{Ix: 0}) {
		t.Error("expected same IVar to be alpha-equal")
	}
	if AlphaEq(ast.IVar{Ix: 0}, ast.IVar{Ix: 1}) {
		t.Error("expected different IVar to not be alpha-equal")
	}
}

func TestAlphaEqPath(t *testing.T) {
	p1 := ast.Path{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}
	p2 := ast.Path{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}
	p3 := ast.Path{A: ast.Sort{U: 1}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}

	if !AlphaEq(p1, p2) {
		t.Error("expected identical Path to be alpha-equal")
	}
	if AlphaEq(p1, p3) {
		t.Error("expected different Path to not be alpha-equal")
	}
}

func TestAlphaEqPathP(t *testing.T) {
	pp1 := ast.PathP{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 2}}
	pp2 := ast.PathP{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 2}}
	pp3 := ast.PathP{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 3}}

	if !AlphaEq(pp1, pp2) {
		t.Error("expected identical PathP to be alpha-equal")
	}
	if AlphaEq(pp1, pp3) {
		t.Error("expected different PathP to not be alpha-equal")
	}
}

func TestAlphaEqPathLam(t *testing.T) {
	pl1 := ast.PathLam{Binder: "i", Body: ast.IVar{Ix: 0}}
	pl2 := ast.PathLam{Binder: "j", Body: ast.IVar{Ix: 0}} // diff binder name
	pl3 := ast.PathLam{Binder: "i", Body: ast.IVar{Ix: 1}}

	if !AlphaEq(pl1, pl2) {
		t.Error("expected PathLam with same body to be alpha-equal (binder names irrelevant)")
	}
	if AlphaEq(pl1, pl3) {
		t.Error("expected PathLam with different body to not be alpha-equal")
	}
}

func TestAlphaEqPathApp(t *testing.T) {
	pa1 := ast.PathApp{P: ast.Var{Ix: 0}, R: ast.I0{}}
	pa2 := ast.PathApp{P: ast.Var{Ix: 0}, R: ast.I0{}}
	pa3 := ast.PathApp{P: ast.Var{Ix: 0}, R: ast.I1{}}

	if !AlphaEq(pa1, pa2) {
		t.Error("expected identical PathApp to be alpha-equal")
	}
	if AlphaEq(pa1, pa3) {
		t.Error("expected different PathApp to not be alpha-equal")
	}
}

func TestAlphaEqTransport(t *testing.T) {
	tr1 := ast.Transport{A: ast.Var{Ix: 0}, E: ast.Var{Ix: 1}}
	tr2 := ast.Transport{A: ast.Var{Ix: 0}, E: ast.Var{Ix: 1}}
	tr3 := ast.Transport{A: ast.Var{Ix: 0}, E: ast.Var{Ix: 2}}

	if !AlphaEq(tr1, tr2) {
		t.Error("expected identical Transport to be alpha-equal")
	}
	if AlphaEq(tr1, tr3) {
		t.Error("expected different Transport to not be alpha-equal")
	}
}

// --- Face Formulas ---

func TestAlphaEqFace(t *testing.T) {
	if !AlphaEq(ast.FaceTop{}, ast.FaceTop{}) {
		t.Error("expected FaceTop to be alpha-equal")
	}
	if !AlphaEq(ast.FaceBot{}, ast.FaceBot{}) {
		t.Error("expected FaceBot to be alpha-equal")
	}
	if AlphaEq(ast.FaceTop{}, ast.FaceBot{}) {
		t.Error("expected FaceTop and FaceBot to not be alpha-equal")
	}

	eq1 := ast.FaceEq{IVar: 0, IsOne: true}
	eq2 := ast.FaceEq{IVar: 0, IsOne: true}
	eq3 := ast.FaceEq{IVar: 0, IsOne: false}
	eq4 := ast.FaceEq{IVar: 1, IsOne: true}

	if !AlphaEq(eq1, eq2) {
		t.Error("expected same FaceEq to be alpha-equal")
	}
	if AlphaEq(eq1, eq3) {
		t.Error("expected FaceEq with different IsOne to not be alpha-equal")
	}
	if AlphaEq(eq1, eq4) {
		t.Error("expected FaceEq with different IVar to not be alpha-equal")
	}
}

func TestAlphaEqFaceAnd(t *testing.T) {
	and1 := ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}
	and2 := ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}
	and3 := ast.FaceAnd{Left: ast.FaceBot{}, Right: ast.FaceTop{}}

	if !AlphaEq(and1, and2) {
		t.Error("expected identical FaceAnd to be alpha-equal")
	}
	if AlphaEq(and1, and3) {
		t.Error("expected different FaceAnd to not be alpha-equal")
	}
}

func TestAlphaEqFaceOr(t *testing.T) {
	or1 := ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}}
	or2 := ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}}
	or3 := ast.FaceOr{Left: ast.FaceBot{}, Right: ast.FaceTop{}}

	if !AlphaEq(or1, or2) {
		t.Error("expected identical FaceOr to be alpha-equal")
	}
	if AlphaEq(or1, or3) {
		t.Error("expected different FaceOr to not be alpha-equal")
	}
}

func TestAlphaEqPartial(t *testing.T) {
	p1 := ast.Partial{Phi: ast.FaceTop{}, A: ast.Sort{U: 0}}
	p2 := ast.Partial{Phi: ast.FaceTop{}, A: ast.Sort{U: 0}}
	p3 := ast.Partial{Phi: ast.FaceBot{}, A: ast.Sort{U: 0}}

	if !AlphaEq(p1, p2) {
		t.Error("expected identical Partial to be alpha-equal")
	}
	if AlphaEq(p1, p3) {
		t.Error("expected different Partial to not be alpha-equal")
	}
}

func TestAlphaEqSystem(t *testing.T) {
	sys1 := ast.System{Branches: []ast.SystemBranch{
		{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}},
	}}
	sys2 := ast.System{Branches: []ast.SystemBranch{
		{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}},
	}}
	sys3 := ast.System{Branches: []ast.SystemBranch{
		{Phi: ast.FaceBot{}, Term: ast.Var{Ix: 0}},
	}}
	sys4 := ast.System{Branches: []ast.SystemBranch{
		{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}},
		{Phi: ast.FaceBot{}, Term: ast.Var{Ix: 1}},
	}}

	if !AlphaEq(sys1, sys2) {
		t.Error("expected identical System to be alpha-equal")
	}
	if AlphaEq(sys1, sys3) {
		t.Error("expected System with different face to not be alpha-equal")
	}
	if AlphaEq(sys1, sys4) {
		t.Error("expected System with different branch count to not be alpha-equal")
	}
}

func TestAlphaEqComp(t *testing.T) {
	comp1 := ast.Comp{IBinder: "i", A: ast.Sort{U: 0}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}}
	comp2 := ast.Comp{IBinder: "j", A: ast.Sort{U: 0}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}}
	comp3 := ast.Comp{IBinder: "i", A: ast.Sort{U: 1}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}}

	if !AlphaEq(comp1, comp2) {
		t.Error("expected Comp with same structure to be alpha-equal (binder names irrelevant)")
	}
	if AlphaEq(comp1, comp3) {
		t.Error("expected Comp with different type to not be alpha-equal")
	}
}

func TestAlphaEqHComp(t *testing.T) {
	hcomp1 := ast.HComp{A: ast.Sort{U: 0}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}}
	hcomp2 := ast.HComp{A: ast.Sort{U: 0}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}}
	hcomp3 := ast.HComp{A: ast.Sort{U: 0}, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}}

	if !AlphaEq(hcomp1, hcomp2) {
		t.Error("expected identical HComp to be alpha-equal")
	}
	if AlphaEq(hcomp1, hcomp3) {
		t.Error("expected HComp with different face to not be alpha-equal")
	}
}

func TestAlphaEqFill(t *testing.T) {
	fill1 := ast.Fill{IBinder: "i", A: ast.Sort{U: 0}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}}
	fill2 := ast.Fill{IBinder: "j", A: ast.Sort{U: 0}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}}
	fill3 := ast.Fill{IBinder: "i", A: ast.Sort{U: 0}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 2}}

	if !AlphaEq(fill1, fill2) {
		t.Error("expected Fill with same structure to be alpha-equal")
	}
	if AlphaEq(fill1, fill3) {
		t.Error("expected Fill with different base to not be alpha-equal")
	}
}

func TestAlphaEqGlue(t *testing.T) {
	glue1 := ast.Glue{
		A: ast.Sort{U: 0},
		System: []ast.GlueBranch{
			{Phi: ast.FaceTop{}, T: ast.Var{Ix: 0}, Equiv: ast.Var{Ix: 1}},
		},
	}
	glue2 := ast.Glue{
		A: ast.Sort{U: 0},
		System: []ast.GlueBranch{
			{Phi: ast.FaceTop{}, T: ast.Var{Ix: 0}, Equiv: ast.Var{Ix: 1}},
		},
	}
	glue3 := ast.Glue{
		A: ast.Sort{U: 0},
		System: []ast.GlueBranch{
			{Phi: ast.FaceBot{}, T: ast.Var{Ix: 0}, Equiv: ast.Var{Ix: 1}},
		},
	}

	if !AlphaEq(glue1, glue2) {
		t.Error("expected identical Glue to be alpha-equal")
	}
	if AlphaEq(glue1, glue3) {
		t.Error("expected Glue with different system to not be alpha-equal")
	}
}

func TestAlphaEqGlueElem(t *testing.T) {
	ge1 := ast.GlueElem{
		System: []ast.GlueElemBranch{{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}}},
		Base:   ast.Var{Ix: 1},
	}
	ge2 := ast.GlueElem{
		System: []ast.GlueElemBranch{{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}}},
		Base:   ast.Var{Ix: 1},
	}
	ge3 := ast.GlueElem{
		System: []ast.GlueElemBranch{{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}}},
		Base:   ast.Var{Ix: 2},
	}

	if !AlphaEq(ge1, ge2) {
		t.Error("expected identical GlueElem to be alpha-equal")
	}
	if AlphaEq(ge1, ge3) {
		t.Error("expected GlueElem with different base to not be alpha-equal")
	}
}

func TestAlphaEqUnglue(t *testing.T) {
	ug1 := ast.Unglue{Ty: ast.Var{Ix: 0}, G: ast.Var{Ix: 1}}
	ug2 := ast.Unglue{Ty: ast.Var{Ix: 0}, G: ast.Var{Ix: 1}}
	ug3 := ast.Unglue{Ty: ast.Var{Ix: 0}, G: ast.Var{Ix: 2}}

	if !AlphaEq(ug1, ug2) {
		t.Error("expected identical Unglue to be alpha-equal")
	}
	if AlphaEq(ug1, ug3) {
		t.Error("expected Unglue with different G to not be alpha-equal")
	}
}

func TestAlphaEqUA(t *testing.T) {
	ua1 := ast.UA{A: ast.Sort{U: 0}, B: ast.Sort{U: 0}, Equiv: ast.Var{Ix: 0}}
	ua2 := ast.UA{A: ast.Sort{U: 0}, B: ast.Sort{U: 0}, Equiv: ast.Var{Ix: 0}}
	ua3 := ast.UA{A: ast.Sort{U: 0}, B: ast.Sort{U: 1}, Equiv: ast.Var{Ix: 0}}

	if !AlphaEq(ua1, ua2) {
		t.Error("expected identical UA to be alpha-equal")
	}
	if AlphaEq(ua1, ua3) {
		t.Error("expected UA with different B to not be alpha-equal")
	}
}

func TestAlphaEqUABeta(t *testing.T) {
	uab1 := ast.UABeta{Equiv: ast.Var{Ix: 0}, Arg: ast.Var{Ix: 1}}
	uab2 := ast.UABeta{Equiv: ast.Var{Ix: 0}, Arg: ast.Var{Ix: 1}}
	uab3 := ast.UABeta{Equiv: ast.Var{Ix: 0}, Arg: ast.Var{Ix: 2}}

	if !AlphaEq(uab1, uab2) {
		t.Error("expected identical UABeta to be alpha-equal")
	}
	if AlphaEq(uab1, uab3) {
		t.Error("expected UABeta with different arg to not be alpha-equal")
	}
}
