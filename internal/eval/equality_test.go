package eval

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// ============================================================================
// alphaEqFace Tests
// ============================================================================

func TestAlphaEqFace_Nil(t *testing.T) {
	t.Parallel()
	if !alphaEqFace(nil, nil) {
		t.Error("nil faces should be alpha-equal")
	}
	if alphaEqFace(ast.FaceTop{}, nil) {
		t.Error("FaceTop and nil should not be alpha-equal")
	}
	if alphaEqFace(nil, ast.FaceTop{}) {
		t.Error("nil and FaceTop should not be alpha-equal")
	}
}

func TestAlphaEqFace_FaceTop(t *testing.T) {
	t.Parallel()
	if !alphaEqFace(ast.FaceTop{}, ast.FaceTop{}) {
		t.Error("FaceTop should equal FaceTop")
	}
	if alphaEqFace(ast.FaceTop{}, ast.FaceBot{}) {
		t.Error("FaceTop should not equal FaceBot")
	}
}

func TestAlphaEqFace_FaceBot(t *testing.T) {
	t.Parallel()
	if !alphaEqFace(ast.FaceBot{}, ast.FaceBot{}) {
		t.Error("FaceBot should equal FaceBot")
	}
	if alphaEqFace(ast.FaceBot{}, ast.FaceTop{}) {
		t.Error("FaceBot should not equal FaceTop")
	}
}

func TestAlphaEqFace_FaceEq(t *testing.T) {
	t.Parallel()
	eq1 := ast.FaceEq{IVar: 0, IsOne: false}
	eq2 := ast.FaceEq{IVar: 0, IsOne: false}
	eq3 := ast.FaceEq{IVar: 0, IsOne: true}
	eq4 := ast.FaceEq{IVar: 1, IsOne: false}

	if !alphaEqFace(eq1, eq2) {
		t.Error("identical FaceEq should be alpha-equal")
	}
	if alphaEqFace(eq1, eq3) {
		t.Error("FaceEq with different IsOne should not be alpha-equal")
	}
	if alphaEqFace(eq1, eq4) {
		t.Error("FaceEq with different IVar should not be alpha-equal")
	}
	if alphaEqFace(eq1, ast.FaceTop{}) {
		t.Error("FaceEq should not equal FaceTop")
	}
}

func TestAlphaEqFace_FaceAnd(t *testing.T) {
	t.Parallel()
	and1 := ast.FaceAnd{Left: ast.FaceEq{IVar: 0}, Right: ast.FaceEq{IVar: 1}}
	and2 := ast.FaceAnd{Left: ast.FaceEq{IVar: 0}, Right: ast.FaceEq{IVar: 1}}
	and3 := ast.FaceAnd{Left: ast.FaceEq{IVar: 0}, Right: ast.FaceEq{IVar: 2}}

	if !alphaEqFace(and1, and2) {
		t.Error("identical FaceAnd should be alpha-equal")
	}
	if alphaEqFace(and1, and3) {
		t.Error("FaceAnd with different Right should not be alpha-equal")
	}
	if alphaEqFace(and1, ast.FaceTop{}) {
		t.Error("FaceAnd should not equal FaceTop")
	}
}

func TestAlphaEqFace_FaceOr(t *testing.T) {
	t.Parallel()
	or1 := ast.FaceOr{Left: ast.FaceEq{IVar: 0}, Right: ast.FaceEq{IVar: 1}}
	or2 := ast.FaceOr{Left: ast.FaceEq{IVar: 0}, Right: ast.FaceEq{IVar: 1}}
	or3 := ast.FaceOr{Left: ast.FaceEq{IVar: 0}, Right: ast.FaceEq{IVar: 2}}

	if !alphaEqFace(or1, or2) {
		t.Error("identical FaceOr should be alpha-equal")
	}
	if alphaEqFace(or1, or3) {
		t.Error("FaceOr with different Right should not be alpha-equal")
	}
	if alphaEqFace(or1, ast.FaceTop{}) {
		t.Error("FaceOr should not equal FaceTop")
	}
}

// ============================================================================
// ValueEqual Tests
// ============================================================================

func TestValueEqual_VSort(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v1       Value
		v2       Value
		expected bool
	}{
		{"same level", VSort{Level: 0}, VSort{Level: 0}, true},
		{"different levels", VSort{Level: 0}, VSort{Level: 1}, false},
		{"sort vs global", VSort{Level: 0}, VGlobal{Name: "x"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValueEqual(tt.v1, tt.v2); got != tt.expected {
				t.Errorf("ValueEqual(%v, %v) = %v, want %v", tt.v1, tt.v2, got, tt.expected)
			}
		})
	}
}

func TestValueEqual_VGlobal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		v1       Value
		v2       Value
		expected bool
	}{
		{"same name", VGlobal{Name: "x"}, VGlobal{Name: "x"}, true},
		{"different names", VGlobal{Name: "x"}, VGlobal{Name: "y"}, false},
		{"global vs sort", VGlobal{Name: "x"}, VSort{Level: 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValueEqual(tt.v1, tt.v2); got != tt.expected {
				t.Errorf("ValueEqual(%v, %v) = %v, want %v", tt.v1, tt.v2, got, tt.expected)
			}
		})
	}
}

func TestValueEqual_VPair(t *testing.T) {
	t.Parallel()
	pair1 := VPair{Fst: VSort{Level: 0}, Snd: VGlobal{Name: "x"}}
	pair2 := VPair{Fst: VSort{Level: 0}, Snd: VGlobal{Name: "x"}}
	pair3 := VPair{Fst: VSort{Level: 1}, Snd: VGlobal{Name: "x"}}

	if !ValueEqual(pair1, pair2) {
		t.Error("identical pairs should be equal")
	}
	if ValueEqual(pair1, pair3) {
		t.Error("pairs with different Fst should not be equal")
	}
	if ValueEqual(pair1, VSort{Level: 0}) {
		t.Error("pair vs non-pair should not be equal")
	}
}

func TestValueEqual_VId(t *testing.T) {
	t.Parallel()
	id1 := VId{A: VSort{Level: 0}, X: VGlobal{Name: "x"}, Y: VGlobal{Name: "y"}}
	id2 := VId{A: VSort{Level: 0}, X: VGlobal{Name: "x"}, Y: VGlobal{Name: "y"}}
	id3 := VId{A: VSort{Level: 0}, X: VGlobal{Name: "x"}, Y: VGlobal{Name: "z"}}

	if !ValueEqual(id1, id2) {
		t.Error("identical VIds should be equal")
	}
	if ValueEqual(id1, id3) {
		t.Error("VIds with different Y should not be equal")
	}
	if ValueEqual(id1, VSort{Level: 0}) {
		t.Error("VId vs non-VId should not be equal")
	}
}

func TestValueEqual_VRefl(t *testing.T) {
	t.Parallel()
	refl1 := VRefl{A: VSort{Level: 0}, X: VGlobal{Name: "x"}}
	refl2 := VRefl{A: VSort{Level: 0}, X: VGlobal{Name: "x"}}
	refl3 := VRefl{A: VSort{Level: 1}, X: VGlobal{Name: "x"}}

	if !ValueEqual(refl1, refl2) {
		t.Error("identical VRefl should be equal")
	}
	if ValueEqual(refl1, refl3) {
		t.Error("VRefl with different A should not be equal")
	}
}

// ============================================================================
// Cubical Value Equality Tests
// ============================================================================

func TestValueEqual_Intervals(t *testing.T) {
	t.Parallel()

	if !ValueEqual(VI0{}, VI0{}) {
		t.Error("VI0 should equal VI0")
	}
	if !ValueEqual(VI1{}, VI1{}) {
		t.Error("VI1 should equal VI1")
	}
	if ValueEqual(VI0{}, VI1{}) {
		t.Error("VI0 should not equal VI1")
	}
	if ValueEqual(VI0{}, VSort{Level: 0}) {
		t.Error("VI0 should not equal VSort")
	}
}

func TestValueEqual_VIVar(t *testing.T) {
	t.Parallel()

	if !ValueEqual(VIVar{Level: 0}, VIVar{Level: 0}) {
		t.Error("same VIVar should be equal")
	}
	if ValueEqual(VIVar{Level: 0}, VIVar{Level: 1}) {
		t.Error("different VIVar levels should not be equal")
	}
	if ValueEqual(VIVar{Level: 0}, VI0{}) {
		t.Error("VIVar should not equal VI0")
	}
}

func TestValueEqual_VPath(t *testing.T) {
	t.Parallel()
	path1 := VPath{A: VSort{Level: 0}, X: VGlobal{Name: "x"}, Y: VGlobal{Name: "y"}}
	path2 := VPath{A: VSort{Level: 0}, X: VGlobal{Name: "x"}, Y: VGlobal{Name: "y"}}
	path3 := VPath{A: VSort{Level: 0}, X: VGlobal{Name: "x"}, Y: VGlobal{Name: "z"}}

	if !ValueEqual(path1, path2) {
		t.Error("identical VPath should be equal")
	}
	if ValueEqual(path1, path3) {
		t.Error("VPath with different Y should not be equal")
	}
}

func TestValueEqual_VNeutral(t *testing.T) {
	t.Parallel()
	n1 := VNeutral{N: Neutral{Head: Head{Glob: "x"}}}
	n2 := VNeutral{N: Neutral{Head: Head{Glob: "x"}}}
	n3 := VNeutral{N: Neutral{Head: Head{Glob: "y"}}}

	if !ValueEqual(n1, n2) {
		t.Error("identical VNeutral should be equal")
	}
	if ValueEqual(n1, n3) {
		t.Error("VNeutral with different head should not be equal")
	}
	if ValueEqual(n1, VGlobal{Name: "x"}) {
		t.Error("VNeutral should not equal VGlobal")
	}
}

func TestValueEqual_VLam(t *testing.T) {
	t.Parallel()
	lam1 := VLam{Body: &Closure{Env: nil, Term: ast.Var{Ix: 0}}}
	lam2 := VLam{Body: &Closure{Env: nil, Term: ast.Var{Ix: 0}}}
	lam3 := VLam{Body: &Closure{Env: nil, Term: ast.Var{Ix: 1}}}

	if !ValueEqual(lam1, lam2) {
		t.Error("identical VLam should be equal")
	}
	if ValueEqual(lam1, lam3) {
		t.Error("VLam with different body should not be equal")
	}
	if ValueEqual(lam1, VGlobal{Name: "x"}) {
		t.Error("VLam should not equal VGlobal")
	}
}

func TestValueEqual_VPi(t *testing.T) {
	t.Parallel()
	pi1 := VPi{A: VSort{Level: 0}, B: &Closure{Env: nil, Term: ast.Var{Ix: 0}}}
	pi2 := VPi{A: VSort{Level: 0}, B: &Closure{Env: nil, Term: ast.Var{Ix: 0}}}
	pi3 := VPi{A: VSort{Level: 1}, B: &Closure{Env: nil, Term: ast.Var{Ix: 0}}}

	if !ValueEqual(pi1, pi2) {
		t.Error("identical VPi should be equal")
	}
	if ValueEqual(pi1, pi3) {
		t.Error("VPi with different A should not be equal")
	}
	if ValueEqual(pi1, VGlobal{Name: "x"}) {
		t.Error("VPi should not equal VGlobal")
	}
}

func TestValueEqual_VSigma(t *testing.T) {
	t.Parallel()
	sig1 := VSigma{A: VSort{Level: 0}, B: &Closure{Env: nil, Term: ast.Var{Ix: 0}}}
	sig2 := VSigma{A: VSort{Level: 0}, B: &Closure{Env: nil, Term: ast.Var{Ix: 0}}}
	sig3 := VSigma{A: VSort{Level: 1}, B: &Closure{Env: nil, Term: ast.Var{Ix: 0}}}

	if !ValueEqual(sig1, sig2) {
		t.Error("identical VSigma should be equal")
	}
	if ValueEqual(sig1, sig3) {
		t.Error("VSigma with different A should not be equal")
	}
}

// ============================================================================
// NeutralEqual Tests
// ============================================================================

func TestNeutralEqual_SameHead(t *testing.T) {
	t.Parallel()
	n1 := Neutral{Head: Head{Glob: "x"}, Sp: nil}
	n2 := Neutral{Head: Head{Glob: "x"}, Sp: nil}

	if !NeutralEqual(n1, n2) {
		t.Error("neutrals with same head should be equal")
	}
}

func TestNeutralEqual_DifferentHead(t *testing.T) {
	t.Parallel()
	n1 := Neutral{Head: Head{Glob: "x"}, Sp: nil}
	n2 := Neutral{Head: Head{Glob: "y"}, Sp: nil}

	if NeutralEqual(n1, n2) {
		t.Error("neutrals with different head should not be equal")
	}
}

func TestNeutralEqual_DifferentSpineLength(t *testing.T) {
	t.Parallel()
	n1 := Neutral{Head: Head{Glob: "x"}, Sp: []Value{VSort{Level: 0}}}
	n2 := Neutral{Head: Head{Glob: "x"}, Sp: nil}

	if NeutralEqual(n1, n2) {
		t.Error("neutrals with different spine length should not be equal")
	}
}

func TestNeutralEqual_SameSpine(t *testing.T) {
	t.Parallel()
	n1 := Neutral{Head: Head{Glob: "f"}, Sp: []Value{VSort{Level: 0}, VGlobal{Name: "x"}}}
	n2 := Neutral{Head: Head{Glob: "f"}, Sp: []Value{VSort{Level: 0}, VGlobal{Name: "x"}}}

	if !NeutralEqual(n1, n2) {
		t.Error("neutrals with same spine should be equal")
	}
}

func TestNeutralEqual_DifferentSpineValues(t *testing.T) {
	t.Parallel()
	n1 := Neutral{Head: Head{Glob: "f"}, Sp: []Value{VSort{Level: 0}}}
	n2 := Neutral{Head: Head{Glob: "f"}, Sp: []Value{VSort{Level: 1}}}

	if NeutralEqual(n1, n2) {
		t.Error("neutrals with different spine values should not be equal")
	}
}

// ============================================================================
// FaceValue Equality Tests
// ============================================================================

func TestFaceValueEqual_VFaceAnd(t *testing.T) {
	t.Parallel()
	f1 := VFaceAnd{Left: VFaceEq{ILevel: 0, IsOne: false}, Right: VFaceEq{ILevel: 1, IsOne: true}}
	f2 := VFaceAnd{Left: VFaceEq{ILevel: 0, IsOne: false}, Right: VFaceEq{ILevel: 1, IsOne: true}}
	f3 := VFaceAnd{Left: VFaceEq{ILevel: 0, IsOne: false}, Right: VFaceEq{ILevel: 2, IsOne: true}}

	if !faceValueEqual(f1, f2) {
		t.Error("identical VFaceAnd should be equal")
	}
	if faceValueEqual(f1, f3) {
		t.Error("VFaceAnd with different Right should not be equal")
	}
}

func TestFaceValueEqual_VFaceOr(t *testing.T) {
	t.Parallel()
	f1 := VFaceOr{Left: VFaceEq{ILevel: 0, IsOne: false}, Right: VFaceEq{ILevel: 1, IsOne: true}}
	f2 := VFaceOr{Left: VFaceEq{ILevel: 0, IsOne: false}, Right: VFaceEq{ILevel: 1, IsOne: true}}
	f3 := VFaceOr{Left: VFaceEq{ILevel: 0, IsOne: false}, Right: VFaceEq{ILevel: 2, IsOne: true}}

	if !faceValueEqual(f1, f2) {
		t.Error("identical VFaceOr should be equal")
	}
	if faceValueEqual(f1, f3) {
		t.Error("VFaceOr with different Right should not be equal")
	}
}

func TestFaceValueEqual_VFaceEq(t *testing.T) {
	t.Parallel()
	if !faceValueEqual(VFaceEq{ILevel: 0, IsOne: false}, VFaceEq{ILevel: 0, IsOne: false}) {
		t.Error("same VFaceEq should be equal")
	}
	if faceValueEqual(VFaceEq{ILevel: 0, IsOne: false}, VFaceEq{ILevel: 1, IsOne: false}) {
		t.Error("different VFaceEq levels should not be equal")
	}
	if faceValueEqual(VFaceEq{ILevel: 0, IsOne: false}, VFaceEq{ILevel: 0, IsOne: true}) {
		t.Error("different VFaceEq IsOne should not be equal")
	}
}

func TestFaceValueEqual_VFaceTopBot(t *testing.T) {
	t.Parallel()
	if !faceValueEqual(VFaceTop{}, VFaceTop{}) {
		t.Error("VFaceTop should equal VFaceTop")
	}
	if !faceValueEqual(VFaceBot{}, VFaceBot{}) {
		t.Error("VFaceBot should equal VFaceBot")
	}
	if faceValueEqual(VFaceTop{}, VFaceBot{}) {
		t.Error("VFaceTop should not equal VFaceBot")
	}
	if faceValueEqual(VFaceTop{}, VFaceEq{ILevel: 0, IsOne: false}) {
		t.Error("VFaceTop should not equal VFaceEq")
	}
}

// ============================================================================
// VPathP and VPathLam Equality Tests
// ============================================================================

func TestValueEqual_VPathP(t *testing.T) {
	t.Parallel()
	pp1 := VPathP{
		A: &IClosure{Env: nil, IEnv: nil, Term: ast.Sort{U: 0}},
		X: VGlobal{Name: "x"},
		Y: VGlobal{Name: "y"},
	}
	pp2 := VPathP{
		A: &IClosure{Env: nil, IEnv: nil, Term: ast.Sort{U: 0}},
		X: VGlobal{Name: "x"},
		Y: VGlobal{Name: "y"},
	}
	pp3 := VPathP{
		A: &IClosure{Env: nil, IEnv: nil, Term: ast.Sort{U: 1}},
		X: VGlobal{Name: "x"},
		Y: VGlobal{Name: "y"},
	}

	if !ValueEqual(pp1, pp2) {
		t.Error("identical VPathP should be equal")
	}
	if ValueEqual(pp1, pp3) {
		t.Error("VPathP with different A should not be equal")
	}
	if ValueEqual(pp1, VGlobal{Name: "x"}) {
		t.Error("VPathP should not equal VGlobal")
	}
}

func TestValueEqual_VPathLam(t *testing.T) {
	t.Parallel()
	pl1 := VPathLam{Body: &IClosure{Env: nil, IEnv: nil, Term: ast.IVar{Ix: 0}}}
	pl2 := VPathLam{Body: &IClosure{Env: nil, IEnv: nil, Term: ast.IVar{Ix: 0}}}
	pl3 := VPathLam{Body: &IClosure{Env: nil, IEnv: nil, Term: ast.IVar{Ix: 1}}}

	if !ValueEqual(pl1, pl2) {
		t.Error("identical VPathLam should be equal")
	}
	if ValueEqual(pl1, pl3) {
		t.Error("VPathLam with different body should not be equal")
	}
}

// ============================================================================
// Additional ValueEqual Tests for Cubical Types
// ============================================================================

func TestValueEqual_VTransport(t *testing.T) {
	t.Parallel()
	tr1 := VTransport{A: &IClosure{Term: ast.Sort{U: 0}}, E: VGlobal{Name: "x"}}
	tr2 := VTransport{A: &IClosure{Term: ast.Sort{U: 0}}, E: VGlobal{Name: "x"}}
	tr3 := VTransport{A: &IClosure{Term: ast.Sort{U: 1}}, E: VGlobal{Name: "x"}}

	if !ValueEqual(tr1, tr2) {
		t.Error("identical VTransport should be equal")
	}
	if ValueEqual(tr1, tr3) {
		t.Error("VTransport with different A should not be equal")
	}
	if ValueEqual(tr1, VGlobal{Name: "x"}) {
		t.Error("VTransport should not equal VGlobal")
	}
}

func TestValueEqual_VFaceTop(t *testing.T) {
	t.Parallel()
	if !ValueEqual(VFaceTop{}, VFaceTop{}) {
		t.Error("VFaceTop should equal VFaceTop")
	}
	if ValueEqual(VFaceTop{}, VFaceBot{}) {
		t.Error("VFaceTop should not equal VFaceBot")
	}
}

func TestValueEqual_VFaceBot(t *testing.T) {
	t.Parallel()
	if !ValueEqual(VFaceBot{}, VFaceBot{}) {
		t.Error("VFaceBot should equal VFaceBot")
	}
	if ValueEqual(VFaceBot{}, VFaceTop{}) {
		t.Error("VFaceBot should not equal VFaceTop")
	}
}

func TestValueEqual_VFaceEq(t *testing.T) {
	t.Parallel()
	eq1 := VFaceEq{ILevel: 0, IsOne: false}
	eq2 := VFaceEq{ILevel: 0, IsOne: false}
	eq3 := VFaceEq{ILevel: 0, IsOne: true}
	eq4 := VFaceEq{ILevel: 1, IsOne: false}

	if !ValueEqual(eq1, eq2) {
		t.Error("identical VFaceEq should be equal")
	}
	if ValueEqual(eq1, eq3) {
		t.Error("VFaceEq with different IsOne should not be equal")
	}
	if ValueEqual(eq1, eq4) {
		t.Error("VFaceEq with different ILevel should not be equal")
	}
	if ValueEqual(eq1, VFaceTop{}) {
		t.Error("VFaceEq should not equal VFaceTop")
	}
}

func TestValueEqual_VFaceAnd(t *testing.T) {
	t.Parallel()
	and1 := VFaceAnd{Left: VFaceEq{ILevel: 0}, Right: VFaceEq{ILevel: 1}}
	and2 := VFaceAnd{Left: VFaceEq{ILevel: 0}, Right: VFaceEq{ILevel: 1}}
	and3 := VFaceAnd{Left: VFaceEq{ILevel: 0}, Right: VFaceEq{ILevel: 2}}

	if !ValueEqual(and1, and2) {
		t.Error("identical VFaceAnd should be equal")
	}
	if ValueEqual(and1, and3) {
		t.Error("VFaceAnd with different Right should not be equal")
	}
	if ValueEqual(and1, VFaceTop{}) {
		t.Error("VFaceAnd should not equal VFaceTop")
	}
}

func TestValueEqual_VFaceOr(t *testing.T) {
	t.Parallel()
	or1 := VFaceOr{Left: VFaceEq{ILevel: 0}, Right: VFaceEq{ILevel: 1}}
	or2 := VFaceOr{Left: VFaceEq{ILevel: 0}, Right: VFaceEq{ILevel: 1}}
	or3 := VFaceOr{Left: VFaceEq{ILevel: 0}, Right: VFaceEq{ILevel: 2}}

	if !ValueEqual(or1, or2) {
		t.Error("identical VFaceOr should be equal")
	}
	if ValueEqual(or1, or3) {
		t.Error("VFaceOr with different Right should not be equal")
	}
}

func TestValueEqual_VPartial(t *testing.T) {
	t.Parallel()
	p1 := VPartial{Phi: VFaceTop{}, A: VSort{Level: 0}}
	p2 := VPartial{Phi: VFaceTop{}, A: VSort{Level: 0}}
	p3 := VPartial{Phi: VFaceBot{}, A: VSort{Level: 0}}

	if !ValueEqual(p1, p2) {
		t.Error("identical VPartial should be equal")
	}
	if ValueEqual(p1, p3) {
		t.Error("VPartial with different Phi should not be equal")
	}
	if ValueEqual(p1, VSort{Level: 0}) {
		t.Error("VPartial should not equal VSort")
	}
}

func TestValueEqual_VSystem(t *testing.T) {
	t.Parallel()
	sys1 := VSystem{Branches: []VSystemBranch{{Phi: VFaceTop{}, Term: VGlobal{Name: "x"}}}}
	sys2 := VSystem{Branches: []VSystemBranch{{Phi: VFaceTop{}, Term: VGlobal{Name: "x"}}}}
	sys3 := VSystem{Branches: []VSystemBranch{{Phi: VFaceBot{}, Term: VGlobal{Name: "x"}}}}
	sys4 := VSystem{Branches: []VSystemBranch{}}

	if !ValueEqual(sys1, sys2) {
		t.Error("identical VSystem should be equal")
	}
	if ValueEqual(sys1, sys3) {
		t.Error("VSystem with different Phi should not be equal")
	}
	if ValueEqual(sys1, sys4) {
		t.Error("VSystem with different branch count should not be equal")
	}
	if ValueEqual(sys1, VSort{Level: 0}) {
		t.Error("VSystem should not equal VSort")
	}
}
