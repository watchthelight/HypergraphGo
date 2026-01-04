package eval

import (
	"strings"
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// ============================================================================
// SprintValue Tests
// ============================================================================

func TestSprintValue_VSort(t *testing.T) {
	t.Parallel()
	tests := []struct {
		level    int
		expected string
	}{
		{0, "Type0"},
		{1, "Type1"},
		{42, "Type42"},
	}
	for _, tt := range tests {
		got := SprintValue(VSort{Level: tt.level})
		if got != tt.expected {
			t.Errorf("SprintValue(VSort{%d}) = %q, want %q", tt.level, got, tt.expected)
		}
	}
}

func TestSprintValue_VGlobal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		expected string
	}{
		{"x", "x"},
		{"Nat", "Nat"},
		{"very-long-name", "very-long-name"},
	}
	for _, tt := range tests {
		got := SprintValue(VGlobal{Name: tt.name})
		if got != tt.expected {
			t.Errorf("SprintValue(VGlobal{%q}) = %q, want %q", tt.name, got, tt.expected)
		}
	}
}

func TestSprintValue_VPair(t *testing.T) {
	t.Parallel()
	pair := VPair{Fst: VSort{Level: 0}, Snd: VGlobal{Name: "x"}}
	got := SprintValue(pair)
	if !strings.Contains(got, "Type0") || !strings.Contains(got, "x") {
		t.Errorf("SprintValue(VPair) = %q, should contain 'Type0' and 'x'", got)
	}
}

func TestSprintValue_VLam(t *testing.T) {
	t.Parallel()
	lam := VLam{Body: &Closure{Env: nil, Term: ast.Var{Ix: 0}}}
	got := SprintValue(lam)
	if !strings.Contains(got, "closure") {
		t.Errorf("SprintValue(VLam) = %q, should contain 'closure'", got)
	}
}

func TestSprintValue_VPi(t *testing.T) {
	t.Parallel()
	pi := VPi{A: VSort{Level: 0}, B: &Closure{Env: nil, Term: ast.Var{Ix: 0}}}
	got := SprintValue(pi)
	if !strings.Contains(got, "Pi") || !strings.Contains(got, "Type0") {
		t.Errorf("SprintValue(VPi) = %q, should contain 'Pi' and 'Type0'", got)
	}
}

func TestSprintValue_VSigma(t *testing.T) {
	t.Parallel()
	sig := VSigma{A: VSort{Level: 0}, B: &Closure{Env: nil, Term: ast.Var{Ix: 0}}}
	got := SprintValue(sig)
	if !strings.Contains(got, "Sigma") || !strings.Contains(got, "Type0") {
		t.Errorf("SprintValue(VSigma) = %q, should contain 'Sigma' and 'Type0'", got)
	}
}

func TestSprintValue_VId(t *testing.T) {
	t.Parallel()
	id := VId{A: VSort{Level: 0}, X: VGlobal{Name: "x"}, Y: VGlobal{Name: "y"}}
	got := SprintValue(id)
	if !strings.Contains(got, "Id") || !strings.Contains(got, "x") || !strings.Contains(got, "y") {
		t.Errorf("SprintValue(VId) = %q, should contain 'Id', 'x', 'y'", got)
	}
}

func TestSprintValue_VRefl(t *testing.T) {
	t.Parallel()
	refl := VRefl{A: VSort{Level: 0}, X: VGlobal{Name: "x"}}
	got := SprintValue(refl)
	if !strings.Contains(got, "refl") || !strings.Contains(got, "x") {
		t.Errorf("SprintValue(VRefl) = %q, should contain 'refl' and 'x'", got)
	}
}

// ============================================================================
// Cubical Value Sprint Tests
// ============================================================================

func TestSprintValue_VI0(t *testing.T) {
	t.Parallel()
	got := SprintValue(VI0{})
	if got != "i0" {
		t.Errorf("SprintValue(VI0) = %q, want 'i0'", got)
	}
}

func TestSprintValue_VI1(t *testing.T) {
	t.Parallel()
	got := SprintValue(VI1{})
	if got != "i1" {
		t.Errorf("SprintValue(VI1) = %q, want 'i1'", got)
	}
}

func TestSprintValue_VIVar(t *testing.T) {
	t.Parallel()
	tests := []struct {
		level    int
		expected string
	}{
		{0, "i{0}"},
		{1, "i{1}"},
	}
	for _, tt := range tests {
		got := SprintValue(VIVar{Level: tt.level})
		if got != tt.expected {
			t.Errorf("SprintValue(VIVar{%d}) = %q, want %q", tt.level, got, tt.expected)
		}
	}
}

func TestSprintValue_VPath(t *testing.T) {
	t.Parallel()
	path := VPath{A: VSort{Level: 0}, X: VGlobal{Name: "x"}, Y: VGlobal{Name: "y"}}
	got := SprintValue(path)
	if !strings.Contains(got, "Path") {
		t.Errorf("SprintValue(VPath) = %q, should contain 'Path'", got)
	}
}

func TestSprintValue_VPathP(t *testing.T) {
	t.Parallel()
	pp := VPathP{
		A: &IClosure{Env: nil, IEnv: nil, Term: ast.Sort{U: 0}},
		X: VGlobal{Name: "x"},
		Y: VGlobal{Name: "y"},
	}
	got := SprintValue(pp)
	if !strings.Contains(got, "PathP") {
		t.Errorf("SprintValue(VPathP) = %q, should contain 'PathP'", got)
	}
}

func TestSprintValue_VPathLam(t *testing.T) {
	t.Parallel()
	pl := VPathLam{Body: &IClosure{Env: nil, IEnv: nil, Term: ast.IVar{Ix: 0}}}
	got := SprintValue(pl)
	if !strings.Contains(got, "closure") {
		t.Errorf("SprintValue(VPathLam) = %q, should contain 'closure'", got)
	}
}

func TestSprintValue_VTransport(t *testing.T) {
	t.Parallel()
	tr := VTransport{
		A: &IClosure{Env: nil, IEnv: nil, Term: ast.Sort{U: 0}},
		E: VGlobal{Name: "e"},
	}
	got := SprintValue(tr)
	if !strings.Contains(got, "transport") {
		t.Errorf("SprintValue(VTransport) = %q, should contain 'transport'", got)
	}
}

// ============================================================================
// Face Value Sprint Tests
// ============================================================================

func TestSprintValue_VFaceTop(t *testing.T) {
	t.Parallel()
	got := SprintValue(VFaceTop{})
	if got != "⊤" {
		t.Errorf("SprintValue(VFaceTop) = %q, want '⊤'", got)
	}
}

func TestSprintValue_VFaceBot(t *testing.T) {
	t.Parallel()
	got := SprintValue(VFaceBot{})
	if got != "⊥" {
		t.Errorf("SprintValue(VFaceBot) = %q, want '⊥'", got)
	}
}

func TestSprintValue_VFaceEq(t *testing.T) {
	t.Parallel()
	tests := []struct {
		level  int
		isOne  bool
		substr string
	}{
		{0, false, "= 0"},
		{1, true, "= 1"},
	}
	for _, tt := range tests {
		got := SprintValue(VFaceEq{ILevel: tt.level, IsOne: tt.isOne})
		if !strings.Contains(got, tt.substr) {
			t.Errorf("SprintValue(VFaceEq{%d, %v}) = %q, should contain %q", tt.level, tt.isOne, got, tt.substr)
		}
	}
}

func TestSprintValue_VFaceAnd(t *testing.T) {
	t.Parallel()
	and := VFaceAnd{
		Left:  VFaceEq{ILevel: 0, IsOne: false},
		Right: VFaceEq{ILevel: 1, IsOne: true},
	}
	got := SprintValue(and)
	if !strings.Contains(got, "∧") {
		t.Errorf("SprintValue(VFaceAnd) = %q, should contain '∧'", got)
	}
}

func TestSprintValue_VFaceOr(t *testing.T) {
	t.Parallel()
	or := VFaceOr{
		Left:  VFaceEq{ILevel: 0, IsOne: false},
		Right: VFaceEq{ILevel: 1, IsOne: true},
	}
	got := SprintValue(or)
	if !strings.Contains(got, "∨") {
		t.Errorf("SprintValue(VFaceOr) = %q, should contain '∨'", got)
	}
}

// ============================================================================
// Partial and System Sprint Tests
// ============================================================================

func TestSprintValue_VPartial(t *testing.T) {
	t.Parallel()
	partial := VPartial{Phi: VFaceTop{}, A: VSort{Level: 0}}
	got := SprintValue(partial)
	if !strings.Contains(got, "Partial") {
		t.Errorf("SprintValue(VPartial) = %q, should contain 'Partial'", got)
	}
}

func TestSprintValue_VSystem(t *testing.T) {
	t.Parallel()
	sys := VSystem{
		Branches: []VSystemBranch{
			{Phi: VFaceEq{ILevel: 0, IsOne: false}, Term: VGlobal{Name: "x"}},
			{Phi: VFaceEq{ILevel: 0, IsOne: true}, Term: VGlobal{Name: "y"}},
		},
	}
	got := SprintValue(sys)
	if !strings.Contains(got, "[") || !strings.Contains(got, "]") {
		t.Errorf("SprintValue(VSystem) = %q, should contain brackets", got)
	}
}

// ============================================================================
// Glue and Comp Sprint Tests
// ============================================================================

func TestSprintValue_VGlue(t *testing.T) {
	t.Parallel()
	glue := VGlue{
		A: VSort{Level: 0},
		System: []VGlueBranch{
			{Phi: VFaceTop{}, T: VSort{Level: 0}, Equiv: VGlobal{Name: "e"}},
		},
	}
	got := SprintValue(glue)
	if !strings.Contains(got, "Glue") {
		t.Errorf("SprintValue(VGlue) = %q, should contain 'Glue'", got)
	}
}

func TestSprintValue_VGlueElem(t *testing.T) {
	t.Parallel()
	elem := VGlueElem{
		Base: VGlobal{Name: "x"},
		System: []VGlueElemBranch{
			{Phi: VFaceTop{}, Term: VGlobal{Name: "t"}},
		},
	}
	got := SprintValue(elem)
	if !strings.Contains(got, "glue") {
		t.Errorf("SprintValue(VGlueElem) = %q, should contain 'glue'", got)
	}
}

func TestSprintValue_VComp(t *testing.T) {
	t.Parallel()
	comp := VComp{
		A:    &IClosure{Env: nil, IEnv: nil, Term: ast.Sort{U: 0}},
		Phi:  VFaceTop{},
		Tube: &IClosure{Env: nil, IEnv: nil, Term: ast.Var{Ix: 0}},
		Base: VGlobal{Name: "base"},
	}
	got := SprintValue(comp)
	if !strings.Contains(got, "comp") {
		t.Errorf("SprintValue(VComp) = %q, should contain 'comp'", got)
	}
}

func TestSprintValue_VHComp(t *testing.T) {
	t.Parallel()
	hcomp := VHComp{
		A:    VSort{Level: 0},
		Phi:  VFaceTop{},
		Tube: &IClosure{Env: nil, IEnv: nil, Term: ast.Var{Ix: 0}},
		Base: VGlobal{Name: "base"},
	}
	got := SprintValue(hcomp)
	if !strings.Contains(got, "hcomp") {
		t.Errorf("SprintValue(VHComp) = %q, should contain 'hcomp'", got)
	}
}

// ============================================================================
// HIT Value Sprint Tests
// ============================================================================

func TestSprintValue_VHITPathCtor(t *testing.T) {
	t.Parallel()
	hit := VHITPathCtor{
		HITName:  "S1",
		CtorName: "loop",
		Args:     nil,
		IArgs:    []Value{VIVar{Level: 0}},
	}
	got := SprintValue(hit)
	// The output is "(loop @ i{0})" - it shows the ctor name and interval arg
	if !strings.Contains(got, "loop") {
		t.Errorf("SprintValue(VHITPathCtor) = %q, should contain 'loop'", got)
	}
}

// ============================================================================
// Neutral Value Sprint Tests
// ============================================================================

func TestSprintValue_VNeutral(t *testing.T) {
	t.Parallel()
	neutral := VNeutral{N: Neutral{Head: Head{Glob: "f"}, Sp: nil}}
	got := SprintValue(neutral)
	if !strings.Contains(got, "f") {
		t.Errorf("SprintValue(VNeutral) = %q, should contain 'f'", got)
	}
}

func TestSprintValue_VNeutralWithSpine(t *testing.T) {
	t.Parallel()
	neutral := VNeutral{
		N: Neutral{
			Head: Head{Glob: "f"},
			Sp:   []Value{VGlobal{Name: "x"}, VSort{Level: 0}},
		},
	}
	got := SprintValue(neutral)
	if !strings.Contains(got, "f") {
		t.Errorf("SprintValue(VNeutral with spine) = %q, should contain 'f'", got)
	}
}

// ============================================================================
// WriteFaceValue Tests
// ============================================================================

func TestWriteFaceValue_Nested(t *testing.T) {
	t.Parallel()
	// (i0 ∧ i1) ∨ i2
	nested := VFaceOr{
		Left: VFaceAnd{
			Left:  VFaceEq{ILevel: 0, IsOne: false},
			Right: VFaceEq{ILevel: 1, IsOne: true},
		},
		Right: VFaceEq{ILevel: 2, IsOne: false},
	}
	got := SprintValue(nested)
	if !strings.Contains(got, "∧") || !strings.Contains(got, "∨") {
		t.Errorf("SprintValue(nested face) = %q, should contain '∧' and '∨'", got)
	}
}
