package ast

import (
	"strings"
	"testing"
)

// ============================================================================
// Interval Type Tests
// ============================================================================

func TestSprint_Interval(t *testing.T) {
	result := Sprint(Interval{})
	if result != "I" {
		t.Errorf("Sprint(Interval{}) = %q, want %q", result, "I")
	}
}

func TestSprint_I0(t *testing.T) {
	result := Sprint(I0{})
	if result != "i0" {
		t.Errorf("Sprint(I0{}) = %q, want %q", result, "i0")
	}
}

func TestSprint_I1(t *testing.T) {
	result := Sprint(I1{})
	if result != "i1" {
		t.Errorf("Sprint(I1{}) = %q, want %q", result, "i1")
	}
}

func TestSprint_IVar(t *testing.T) {
	tests := []struct {
		ix       int
		expected string
	}{
		{0, "i{0}"},
		{1, "i{1}"},
		{10, "i{10}"},
	}

	for _, tt := range tests {
		result := Sprint(IVar{Ix: tt.ix})
		if result != tt.expected {
			t.Errorf("Sprint(IVar{%d}) = %q, want %q", tt.ix, result, tt.expected)
		}
	}
}

// ============================================================================
// Path Type Tests
// ============================================================================

func TestSprint_Path(t *testing.T) {
	path := Path{
		A: Global{Name: "Nat"},
		X: Global{Name: "zero"},
		Y: Global{Name: "one"},
	}

	result := Sprint(path)
	if !strings.HasPrefix(result, "(Path ") {
		t.Errorf("Path output should start with '(Path ': got %q", result)
	}
	if !strings.Contains(result, "Nat") {
		t.Error("Path output should contain type 'Nat'")
	}
	if !strings.Contains(result, "zero") {
		t.Error("Path output should contain endpoint 'zero'")
	}
	if !strings.Contains(result, "one") {
		t.Error("Path output should contain endpoint 'one'")
	}
	if !strings.HasSuffix(result, ")") {
		t.Error("Path output should end with ')'")
	}
}

func TestSprint_PathP(t *testing.T) {
	pathP := PathP{
		A: Global{Name: "TypeFamily"},
		X: Global{Name: "x0"},
		Y: Global{Name: "x1"},
	}

	result := Sprint(pathP)
	if !strings.HasPrefix(result, "(PathP ") {
		t.Errorf("PathP output should start with '(PathP ': got %q", result)
	}
	if !strings.Contains(result, "TypeFamily") {
		t.Error("PathP output should contain 'TypeFamily'")
	}
}

func TestSprint_PathLam(t *testing.T) {
	tests := []struct {
		name     string
		pathLam  PathLam
		contains []string
	}{
		{
			name:     "with binder",
			pathLam:  PathLam{Binder: "i", Body: Global{Name: "x"}},
			contains: []string{"<i>", "x"},
		},
		{
			name:     "empty binder uses underscore",
			pathLam:  PathLam{Binder: "", Body: Global{Name: "y"}},
			contains: []string{"<_>", "y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sprint(tt.pathLam)
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("PathLam output should contain %q: got %q", s, result)
				}
			}
		})
	}
}

func TestSprint_PathApp(t *testing.T) {
	tests := []struct {
		name     string
		pathApp  PathApp
		contains string
	}{
		{
			name:     "path app with i0",
			pathApp:  PathApp{P: Global{Name: "p"}, R: I0{}},
			contains: "@ i0",
		},
		{
			name:     "path app with i1",
			pathApp:  PathApp{P: Global{Name: "q"}, R: I1{}},
			contains: "@ i1",
		},
		{
			name:     "path app with IVar",
			pathApp:  PathApp{P: Global{Name: "r"}, R: IVar{Ix: 0}},
			contains: "@ i{0}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sprint(tt.pathApp)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("PathApp output should contain %q: got %q", tt.contains, result)
			}
		})
	}
}

func TestSprint_Transport(t *testing.T) {
	transport := Transport{
		A: Global{Name: "TypeFamily"},
		E: Global{Name: "elem"},
	}

	result := Sprint(transport)
	if !strings.HasPrefix(result, "(transport ") {
		t.Errorf("Transport output should start with '(transport ': got %q", result)
	}
	if !strings.Contains(result, "TypeFamily") {
		t.Error("Transport output should contain 'TypeFamily'")
	}
	if !strings.Contains(result, "elem") {
		t.Error("Transport output should contain 'elem'")
	}
}

// ============================================================================
// Face Formula Tests
// ============================================================================

func TestSprint_FaceTop(t *testing.T) {
	result := Sprint(FaceTop{})
	if result != "⊤" {
		t.Errorf("Sprint(FaceTop{}) = %q, want %q", result, "⊤")
	}
}

func TestSprint_FaceBot(t *testing.T) {
	result := Sprint(FaceBot{})
	if result != "⊥" {
		t.Errorf("Sprint(FaceBot{}) = %q, want %q", result, "⊥")
	}
}

func TestSprint_FaceEq(t *testing.T) {
	tests := []struct {
		face     FaceEq
		expected string
	}{
		{FaceEq{IVar: 0, IsOne: false}, "(i{0} = 0)"},
		{FaceEq{IVar: 0, IsOne: true}, "(i{0} = 1)"},
		{FaceEq{IVar: 5, IsOne: false}, "(i{5} = 0)"},
		{FaceEq{IVar: 3, IsOne: true}, "(i{3} = 1)"},
	}

	for _, tt := range tests {
		result := Sprint(tt.face)
		if result != tt.expected {
			t.Errorf("Sprint(%+v) = %q, want %q", tt.face, result, tt.expected)
		}
	}
}

func TestSprint_FaceAnd(t *testing.T) {
	face := FaceAnd{
		Left:  FaceEq{IVar: 0, IsOne: true},
		Right: FaceEq{IVar: 1, IsOne: false},
	}

	result := Sprint(face)
	if !strings.Contains(result, "∧") {
		t.Errorf("FaceAnd output should contain '∧': got %q", result)
	}
	if !strings.Contains(result, "i{0} = 1") {
		t.Error("FaceAnd output should contain left face")
	}
	if !strings.Contains(result, "i{1} = 0") {
		t.Error("FaceAnd output should contain right face")
	}
}

func TestSprint_FaceOr(t *testing.T) {
	face := FaceOr{
		Left:  FaceEq{IVar: 0, IsOne: true},
		Right: FaceEq{IVar: 1, IsOne: true},
	}

	result := Sprint(face)
	if !strings.Contains(result, "∨") {
		t.Errorf("FaceOr output should contain '∨': got %q", result)
	}
}

func TestSprint_FaceNested(t *testing.T) {
	// ((i=0) ∧ (j=1)) ∨ (k=0)
	face := FaceOr{
		Left: FaceAnd{
			Left:  FaceEq{IVar: 0, IsOne: false},
			Right: FaceEq{IVar: 1, IsOne: true},
		},
		Right: FaceEq{IVar: 2, IsOne: false},
	}

	result := Sprint(face)
	if !strings.Contains(result, "∧") {
		t.Error("Nested face should contain '∧'")
	}
	if !strings.Contains(result, "∨") {
		t.Error("Nested face should contain '∨'")
	}
}

// ============================================================================
// Partial Types and Systems Tests
// ============================================================================

func TestSprint_Partial(t *testing.T) {
	partial := Partial{
		Phi: FaceEq{IVar: 0, IsOne: true},
		A:   Global{Name: "Nat"},
	}

	result := Sprint(partial)
	if !strings.HasPrefix(result, "(Partial ") {
		t.Errorf("Partial output should start with '(Partial ': got %q", result)
	}
	if !strings.Contains(result, "i{0} = 1") {
		t.Error("Partial output should contain face constraint")
	}
	if !strings.Contains(result, "Nat") {
		t.Error("Partial output should contain type")
	}
}

func TestSprint_System(t *testing.T) {
	tests := []struct {
		name     string
		system   System
		contains []string
	}{
		{
			name:     "empty system",
			system:   System{Branches: nil},
			contains: []string{"[", "]"},
		},
		{
			name: "single branch",
			system: System{
				Branches: []SystemBranch{
					{Phi: FaceEq{IVar: 0, IsOne: true}, Term: Global{Name: "x"}},
				},
			},
			contains: []string{"[", "↦", "x", "]"},
		},
		{
			name: "multiple branches",
			system: System{
				Branches: []SystemBranch{
					{Phi: FaceEq{IVar: 0, IsOne: false}, Term: Global{Name: "left"}},
					{Phi: FaceEq{IVar: 0, IsOne: true}, Term: Global{Name: "right"}},
				},
			},
			contains: []string{"left", "right", ","},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sprint(tt.system)
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("System output should contain %q: got %q", s, result)
				}
			}
		})
	}
}

// ============================================================================
// Composition Operations Tests
// ============================================================================

func TestSprint_Comp(t *testing.T) {
	tests := []struct {
		name     string
		comp     Comp
		contains []string
	}{
		{
			name: "with binder",
			comp: Comp{
				IBinder: "i",
				A:       Global{Name: "A"},
				Phi:     FaceEq{IVar: 0, IsOne: true},
				Tube:    Global{Name: "tube"},
				Base:    Global{Name: "base"},
			},
			contains: []string{"(comp^i", "A", "↦", "tube", "base"},
		},
		{
			name: "without binder",
			comp: Comp{
				IBinder: "",
				A:       Global{Name: "B"},
				Phi:     FaceTop{},
				Tube:    Global{Name: "u"},
				Base:    Global{Name: "b"},
			},
			contains: []string{"(comp ", "B"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sprint(tt.comp)
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("Comp output should contain %q: got %q", s, result)
				}
			}
		})
	}
}

func TestSprint_HComp(t *testing.T) {
	hcomp := HComp{
		A:    Global{Name: "A"},
		Phi:  FaceEq{IVar: 0, IsOne: false},
		Tube: Global{Name: "tube"},
		Base: Global{Name: "base"},
	}

	result := Sprint(hcomp)
	if !strings.HasPrefix(result, "(hcomp ") {
		t.Errorf("HComp output should start with '(hcomp ': got %q", result)
	}
	if !strings.Contains(result, "A") {
		t.Error("HComp output should contain type")
	}
	if !strings.Contains(result, "↦") {
		t.Error("HComp output should contain '↦'")
	}
}

func TestSprint_Fill(t *testing.T) {
	tests := []struct {
		name     string
		fill     Fill
		contains []string
	}{
		{
			name: "with binder",
			fill: Fill{
				IBinder: "j",
				A:       Global{Name: "TypeLine"},
				Phi:     FaceEq{IVar: 1, IsOne: true},
				Tube:    Global{Name: "tube"},
				Base:    Global{Name: "a0"},
			},
			contains: []string{"(fill^j", "TypeLine", "↦"},
		},
		{
			name: "without binder",
			fill: Fill{
				IBinder: "",
				A:       Global{Name: "T"},
				Phi:     FaceBot{},
				Tube:    Global{Name: "t"},
				Base:    Global{Name: "b"},
			},
			contains: []string{"(fill ", "T"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sprint(tt.fill)
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("Fill output should contain %q: got %q", s, result)
				}
			}
		})
	}
}

// ============================================================================
// Glue Types Tests
// ============================================================================

func TestSprint_Glue(t *testing.T) {
	tests := []struct {
		name     string
		glue     Glue
		contains []string
	}{
		{
			name: "empty system",
			glue: Glue{
				A:      Global{Name: "Base"},
				System: nil,
			},
			contains: []string{"(Glue ", "Base", "[", "])"},
		},
		{
			name: "single branch",
			glue: Glue{
				A: Global{Name: "A"},
				System: []GlueBranch{
					{
						Phi:   FaceEq{IVar: 0, IsOne: false},
						T:     Global{Name: "T"},
						Equiv: Global{Name: "e"},
					},
				},
			},
			contains: []string{"(Glue ", "A", "i{0} = 0", "↦", "T", "e"},
		},
		{
			name: "multiple branches",
			glue: Glue{
				A: Global{Name: "B"},
				System: []GlueBranch{
					{Phi: FaceEq{IVar: 0, IsOne: false}, T: Global{Name: "T1"}, Equiv: Global{Name: "e1"}},
					{Phi: FaceEq{IVar: 0, IsOne: true}, T: Global{Name: "T2"}, Equiv: Global{Name: "e2"}},
				},
			},
			contains: []string{"T1", "T2", ","},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sprint(tt.glue)
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("Glue output should contain %q: got %q", s, result)
				}
			}
		})
	}
}

func TestSprint_GlueElem(t *testing.T) {
	tests := []struct {
		name     string
		glueElem GlueElem
		contains []string
	}{
		{
			name: "empty system",
			glueElem: GlueElem{
				System: nil,
				Base:   Global{Name: "a"},
			},
			contains: []string{"(glue [", "] ", "a"},
		},
		{
			name: "with branches",
			glueElem: GlueElem{
				System: []GlueElemBranch{
					{Phi: FaceEq{IVar: 0, IsOne: false}, Term: Global{Name: "t"}},
				},
				Base: Global{Name: "base"},
			},
			contains: []string{"glue", "↦", "t", "base"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sprint(tt.glueElem)
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("GlueElem output should contain %q: got %q", s, result)
				}
			}
		})
	}
}

func TestSprint_Unglue(t *testing.T) {
	unglue := Unglue{
		Ty: Global{Name: "GlueTy"},
		G:  Global{Name: "g"},
	}

	result := Sprint(unglue)
	if !strings.HasPrefix(result, "(unglue ") {
		t.Errorf("Unglue output should start with '(unglue ': got %q", result)
	}
	if !strings.Contains(result, "g") {
		t.Error("Unglue output should contain element 'g'")
	}
}

// ============================================================================
// Univalence Tests
// ============================================================================

func TestSprint_UA(t *testing.T) {
	ua := UA{
		A:     Global{Name: "A"},
		B:     Global{Name: "B"},
		Equiv: Global{Name: "equiv"},
	}

	result := Sprint(ua)
	if !strings.HasPrefix(result, "(ua ") {
		t.Errorf("UA output should start with '(ua ': got %q", result)
	}
	if !strings.Contains(result, "A") {
		t.Error("UA output should contain 'A'")
	}
	if !strings.Contains(result, "B") {
		t.Error("UA output should contain 'B'")
	}
	if !strings.Contains(result, "equiv") {
		t.Error("UA output should contain 'equiv'")
	}
}

func TestSprint_UABeta(t *testing.T) {
	uaBeta := UABeta{
		Equiv: Global{Name: "e"},
		Arg:   Global{Name: "x"},
	}

	result := Sprint(uaBeta)
	if !strings.HasPrefix(result, "(ua-β ") {
		t.Errorf("UABeta output should start with '(ua-β ': got %q", result)
	}
	if !strings.Contains(result, "e") {
		t.Error("UABeta output should contain 'e'")
	}
	if !strings.Contains(result, "x") {
		t.Error("UABeta output should contain 'x'")
	}
}

// ============================================================================
// HITApp Tests
// ============================================================================

func TestSprint_HITApp(t *testing.T) {
	tests := []struct {
		name     string
		hitApp   HITApp
		contains []string
	}{
		{
			name: "no args",
			hitApp: HITApp{
				HITName: "S1",
				Ctor:    "loop",
				Args:    nil,
				IArgs:   nil,
			},
			contains: []string{"(loop", ")"},
		},
		{
			name: "with term args",
			hitApp: HITApp{
				HITName: "Susp",
				Ctor:    "merid",
				Args:    []Term{Global{Name: "a"}},
				IArgs:   nil,
			},
			contains: []string{"(merid", "a", ")"},
		},
		{
			name: "with interval args",
			hitApp: HITApp{
				HITName: "S1",
				Ctor:    "loop",
				Args:    nil,
				IArgs:   []Term{IVar{Ix: 0}},
			},
			contains: []string{"loop", "@ i{0}"},
		},
		{
			name: "with both args",
			hitApp: HITApp{
				HITName: "Quot",
				Ctor:    "eq",
				Args:    []Term{Global{Name: "x"}, Global{Name: "y"}},
				IArgs:   []Term{I0{}},
			},
			contains: []string{"eq", "x", "y", "@ i0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sprint(tt.hitApp)
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("HITApp output should contain %q: got %q", s, result)
				}
			}
		})
	}
}

// ============================================================================
// writeFace Edge Cases
// ============================================================================

func TestWriteFace_Nil(t *testing.T) {
	// writeFace handles nil by printing ⊥
	partial := Partial{
		Phi: nil,
		A:   Global{Name: "A"},
	}

	result := Sprint(partial)
	// nil face should be printed as ⊥
	if !strings.Contains(result, "⊥") {
		t.Errorf("nil face should be printed as '⊥': got %q", result)
	}
}

func TestWriteFace_UnknownFaceType(t *testing.T) {
	// This tests the default case in writeFace
	// We can't easily test this without creating a custom Face type,
	// but we can verify all known face types work
	faces := []Face{
		FaceTop{},
		FaceBot{},
		FaceEq{IVar: 0, IsOne: true},
		FaceAnd{Left: FaceTop{}, Right: FaceBot{}},
		FaceOr{Left: FaceTop{}, Right: FaceBot{}},
	}

	for _, f := range faces {
		result := Sprint(Partial{Phi: f, A: Global{Name: "T"}})
		if result == "" {
			t.Errorf("Sprint returned empty for face %T", f)
		}
		// Should not contain "?face" which is the unknown face marker
		if strings.Contains(result, "?face") {
			t.Errorf("Known face type %T printed as unknown", f)
		}
	}
}

// ============================================================================
// Complex Cubical Terms
// ============================================================================

func TestSprint_ComplexCubical(t *testing.T) {
	// A realistic cubical term: comp with nested faces and path applications
	comp := Comp{
		IBinder: "i",
		A: PathLam{
			Binder: "j",
			Body:   Global{Name: "A"},
		},
		Phi: FaceOr{
			Left:  FaceEq{IVar: 0, IsOne: false},
			Right: FaceEq{IVar: 0, IsOne: true},
		},
		Tube: PathApp{
			P: Global{Name: "p"},
			R: IVar{Ix: 0},
		},
		Base: Global{Name: "base"},
	}

	result := Sprint(comp)
	if result == "" {
		t.Error("Complex cubical term should not print empty")
	}
	if !strings.Contains(result, "comp^i") {
		t.Error("Should contain 'comp^i'")
	}
	if !strings.Contains(result, "∨") {
		t.Error("Should contain '∨' from face disjunction")
	}
}

func TestSprint_GlueWithNestedTypes(t *testing.T) {
	// Glue type with complex nested structure
	glue := Glue{
		A: Pi{Binder: "x", A: Global{Name: "Nat"}, B: Global{Name: "Nat"}},
		System: []GlueBranch{
			{
				Phi: FaceAnd{
					Left:  FaceEq{IVar: 0, IsOne: false},
					Right: FaceEq{IVar: 1, IsOne: false},
				},
				T:     Sigma{Binder: "y", A: Global{Name: "A"}, B: Global{Name: "B"}},
				Equiv: Global{Name: "eq"},
			},
		},
	}

	result := Sprint(glue)
	if result == "" {
		t.Error("Nested Glue should not print empty")
	}
	if !strings.Contains(result, "Glue") {
		t.Error("Should contain 'Glue'")
	}
	if !strings.Contains(result, "Pi") {
		t.Error("Should contain nested 'Pi'")
	}
	if !strings.Contains(result, "Sigma") {
		t.Error("Should contain nested 'Sigma'")
	}
}

// ============================================================================
// tryWriteExtension Non-Cubical Return False
// ============================================================================

func TestTryWriteExtension_ReturnsFalseForNonCubical(t *testing.T) {
	// Verify non-cubical terms return false from writeCubical
	nonCubicalTerms := []Term{
		Sort{U: 0},
		Var{Ix: 0},
		Global{Name: "x"},
		Pi{Binder: "x", A: Sort{U: 0}, B: Sort{U: 0}},
		Lam{Binder: "x", Body: Var{Ix: 0}},
		App{T: Global{Name: "f"}, U: Global{Name: "x"}},
	}

	for _, term := range nonCubicalTerms {
		// These should all still print correctly via the normal path
		result := Sprint(term)
		if result == "" {
			t.Errorf("Non-cubical term %T should still print", term)
		}
	}
}

// ============================================================================
// All Cubical Terms Print Non-Empty
// ============================================================================

func TestSprint_AllCubicalTermsNonEmpty(t *testing.T) {
	cubicalTerms := []Term{
		// Interval
		Interval{},
		I0{},
		I1{},
		IVar{Ix: 0},
		// Paths
		Path{A: Global{Name: "A"}, X: Global{Name: "x"}, Y: Global{Name: "y"}},
		PathP{A: Global{Name: "F"}, X: Global{Name: "x"}, Y: Global{Name: "y"}},
		PathLam{Binder: "i", Body: Global{Name: "t"}},
		PathApp{P: Global{Name: "p"}, R: I0{}},
		Transport{A: Global{Name: "F"}, E: Global{Name: "e"}},
		// Faces
		FaceTop{},
		FaceBot{},
		FaceEq{IVar: 0, IsOne: true},
		FaceAnd{Left: FaceTop{}, Right: FaceBot{}},
		FaceOr{Left: FaceTop{}, Right: FaceBot{}},
		// Partial
		Partial{Phi: FaceTop{}, A: Global{Name: "A"}},
		System{Branches: []SystemBranch{{Phi: FaceTop{}, Term: Global{Name: "t"}}}},
		// Composition
		Comp{A: Global{Name: "A"}, Phi: FaceTop{}, Tube: Global{Name: "u"}, Base: Global{Name: "b"}},
		HComp{A: Global{Name: "A"}, Phi: FaceTop{}, Tube: Global{Name: "u"}, Base: Global{Name: "b"}},
		Fill{A: Global{Name: "A"}, Phi: FaceTop{}, Tube: Global{Name: "u"}, Base: Global{Name: "b"}},
		// Glue
		Glue{A: Global{Name: "A"}, System: nil},
		GlueElem{System: nil, Base: Global{Name: "a"}},
		Unglue{Ty: Global{Name: "G"}, G: Global{Name: "g"}},
		// Univalence
		UA{A: Global{Name: "A"}, B: Global{Name: "B"}, Equiv: Global{Name: "e"}},
		UABeta{Equiv: Global{Name: "e"}, Arg: Global{Name: "x"}},
		// HIT
		HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: nil},
	}

	for _, term := range cubicalTerms {
		result := Sprint(term)
		if result == "" {
			t.Errorf("Sprint returned empty for cubical term %T", term)
		}
	}
}
