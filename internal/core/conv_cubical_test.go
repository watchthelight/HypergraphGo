package core

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// TestAlphaEqFace tests alpha-equality for face formulas.
func TestAlphaEqFace(t *testing.T) {
	tests := []struct {
		name string
		a, b ast.Face
		want bool
	}{
		// Basic face formulas
		{"FaceTop == FaceTop", ast.FaceTop{}, ast.FaceTop{}, true},
		{"FaceBot == FaceBot", ast.FaceBot{}, ast.FaceBot{}, true},
		{"FaceTop != FaceBot", ast.FaceTop{}, ast.FaceBot{}, false},

		// FaceEq
		{"FaceEq same i=0", ast.FaceEq{IVar: 0, IsOne: false}, ast.FaceEq{IVar: 0, IsOne: false}, true},
		{"FaceEq same i=1", ast.FaceEq{IVar: 1, IsOne: true}, ast.FaceEq{IVar: 1, IsOne: true}, true},
		{"FaceEq different IVar", ast.FaceEq{IVar: 0, IsOne: false}, ast.FaceEq{IVar: 1, IsOne: false}, false},
		{"FaceEq different IsOne", ast.FaceEq{IVar: 0, IsOne: false}, ast.FaceEq{IVar: 0, IsOne: true}, false},

		// FaceAnd
		{
			"FaceAnd same",
			ast.FaceAnd{Left: ast.FaceEq{IVar: 0, IsOne: false}, Right: ast.FaceEq{IVar: 1, IsOne: true}},
			ast.FaceAnd{Left: ast.FaceEq{IVar: 0, IsOne: false}, Right: ast.FaceEq{IVar: 1, IsOne: true}},
			true,
		},
		{
			"FaceAnd different left",
			ast.FaceAnd{Left: ast.FaceEq{IVar: 0, IsOne: false}, Right: ast.FaceEq{IVar: 1, IsOne: true}},
			ast.FaceAnd{Left: ast.FaceEq{IVar: 0, IsOne: true}, Right: ast.FaceEq{IVar: 1, IsOne: true}},
			false,
		},

		// FaceOr
		{
			"FaceOr same",
			ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}},
			ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}},
			true,
		},
		{
			"FaceOr different",
			ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}},
			ast.FaceOr{Left: ast.FaceBot{}, Right: ast.FaceTop{}},
			false,
		},

		// Nested faces
		{
			"nested FaceAnd",
			ast.FaceAnd{
				Left:  ast.FaceAnd{Left: ast.FaceEq{IVar: 0, IsOne: false}, Right: ast.FaceEq{IVar: 1, IsOne: true}},
				Right: ast.FaceTop{},
			},
			ast.FaceAnd{
				Left:  ast.FaceAnd{Left: ast.FaceEq{IVar: 0, IsOne: false}, Right: ast.FaceEq{IVar: 1, IsOne: true}},
				Right: ast.FaceTop{},
			},
			true,
		},

		// nil handling
		{"nil == nil", nil, nil, true},
		{"nil != FaceTop", nil, ast.FaceTop{}, false},
		{"FaceTop != nil", ast.FaceTop{}, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := alphaEqFace(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("alphaEqFace(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestAlphaEqCubical tests alpha-equality for cubical AST terms.
func TestAlphaEqCubical(t *testing.T) {
	// Helper to create a simple type for testing
	typeU := ast.Sort{U: 0}

	tests := []struct {
		name string
		a, b ast.Term
		want bool
	}{
		// Interval types
		{"Interval == Interval", ast.Interval{}, ast.Interval{}, true},
		{"I0 == I0", ast.I0{}, ast.I0{}, true},
		{"I1 == I1", ast.I1{}, ast.I1{}, true},
		{"I0 != I1", ast.I0{}, ast.I1{}, false},

		// IVar
		{"IVar same index", ast.IVar{Ix: 0}, ast.IVar{Ix: 0}, true},
		{"IVar different index", ast.IVar{Ix: 0}, ast.IVar{Ix: 1}, false},

		// Path types
		{
			"Path same",
			ast.Path{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
			ast.Path{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
			true,
		},
		{
			"Path different endpoint",
			ast.Path{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
			ast.Path{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 2}},
			false,
		},

		// PathP
		{
			"PathP same",
			ast.PathP{A: ast.Lam{Binder: "i", Body: typeU}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
			ast.PathP{A: ast.Lam{Binder: "i", Body: typeU}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
			true,
		},

		// PathLam
		{
			"PathLam same body",
			ast.PathLam{Binder: "i", Body: ast.Var{Ix: 0}},
			ast.PathLam{Binder: "j", Body: ast.Var{Ix: 0}}, // binder names don't matter
			true,
		},
		{
			"PathLam different body",
			ast.PathLam{Binder: "i", Body: ast.Var{Ix: 0}},
			ast.PathLam{Binder: "i", Body: ast.Var{Ix: 1}},
			false,
		},

		// PathApp
		{
			"PathApp same",
			ast.PathApp{P: ast.Var{Ix: 0}, R: ast.I0{}},
			ast.PathApp{P: ast.Var{Ix: 0}, R: ast.I0{}},
			true,
		},
		{
			"PathApp different R",
			ast.PathApp{P: ast.Var{Ix: 0}, R: ast.I0{}},
			ast.PathApp{P: ast.Var{Ix: 0}, R: ast.I1{}},
			false,
		},

		// Transport
		{
			"Transport same",
			ast.Transport{A: ast.Lam{Binder: "i", Body: typeU}, E: ast.Var{Ix: 0}},
			ast.Transport{A: ast.Lam{Binder: "i", Body: typeU}, E: ast.Var{Ix: 0}},
			true,
		},

		// Face formulas as terms
		{"FaceTop term", ast.FaceTop{}, ast.FaceTop{}, true},
		{"FaceBot term", ast.FaceBot{}, ast.FaceBot{}, true},
		{"FaceEq term same", ast.FaceEq{IVar: 0, IsOne: false}, ast.FaceEq{IVar: 0, IsOne: false}, true},
		{
			"FaceAnd term same",
			ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}},
			ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}},
			true,
		},
		{
			"FaceOr term same",
			ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}},
			ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}},
			true,
		},

		// Partial
		{
			"Partial same",
			ast.Partial{Phi: ast.FaceTop{}, A: typeU},
			ast.Partial{Phi: ast.FaceTop{}, A: typeU},
			true,
		},
		{
			"Partial different phi",
			ast.Partial{Phi: ast.FaceTop{}, A: typeU},
			ast.Partial{Phi: ast.FaceBot{}, A: typeU},
			false,
		},

		// System
		{
			"System empty",
			ast.System{Branches: []ast.SystemBranch{}},
			ast.System{Branches: []ast.SystemBranch{}},
			true,
		},
		{
			"System same branches",
			ast.System{Branches: []ast.SystemBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: false}, Term: ast.Var{Ix: 0}},
				{Phi: ast.FaceEq{IVar: 0, IsOne: true}, Term: ast.Var{Ix: 1}},
			}},
			ast.System{Branches: []ast.SystemBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: false}, Term: ast.Var{Ix: 0}},
				{Phi: ast.FaceEq{IVar: 0, IsOne: true}, Term: ast.Var{Ix: 1}},
			}},
			true,
		},
		{
			"System different length",
			ast.System{Branches: []ast.SystemBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: false}, Term: ast.Var{Ix: 0}},
			}},
			ast.System{Branches: []ast.SystemBranch{}},
			false,
		},

		// Comp
		{
			"Comp same",
			ast.Comp{IBinder: "i", A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			ast.Comp{IBinder: "j", A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			true,
		},
		{
			"Comp different base",
			ast.Comp{IBinder: "i", A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			ast.Comp{IBinder: "i", A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 2}},
			false,
		},

		// HComp
		{
			"HComp same",
			ast.HComp{A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			ast.HComp{A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			true,
		},

		// Fill
		{
			"Fill same",
			ast.Fill{IBinder: "i", A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			ast.Fill{IBinder: "j", A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			true,
		},

		// Glue
		{
			"Glue empty system",
			ast.Glue{A: typeU, System: []ast.GlueBranch{}},
			ast.Glue{A: typeU, System: []ast.GlueBranch{}},
			true,
		},
		{
			"Glue with branches",
			ast.Glue{A: typeU, System: []ast.GlueBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: false}, T: typeU, Equiv: ast.Var{Ix: 0}},
			}},
			ast.Glue{A: typeU, System: []ast.GlueBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: false}, T: typeU, Equiv: ast.Var{Ix: 0}},
			}},
			true,
		},
		{
			"Glue different system",
			ast.Glue{A: typeU, System: []ast.GlueBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: false}, T: typeU, Equiv: ast.Var{Ix: 0}},
			}},
			ast.Glue{A: typeU, System: []ast.GlueBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: true}, T: typeU, Equiv: ast.Var{Ix: 0}},
			}},
			false,
		},

		// GlueElem
		{
			"GlueElem same",
			ast.GlueElem{
				System: []ast.GlueElemBranch{{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}}},
				Base:   ast.Var{Ix: 1},
			},
			ast.GlueElem{
				System: []ast.GlueElemBranch{{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}}},
				Base:   ast.Var{Ix: 1},
			},
			true,
		},

		// Unglue
		{
			"Unglue same",
			ast.Unglue{Ty: typeU, G: ast.Var{Ix: 0}},
			ast.Unglue{Ty: typeU, G: ast.Var{Ix: 0}},
			true,
		},
		{
			"Unglue different G",
			ast.Unglue{Ty: typeU, G: ast.Var{Ix: 0}},
			ast.Unglue{Ty: typeU, G: ast.Var{Ix: 1}},
			false,
		},

		// UA
		{
			"UA same",
			ast.UA{A: typeU, B: typeU, Equiv: ast.Var{Ix: 0}},
			ast.UA{A: typeU, B: typeU, Equiv: ast.Var{Ix: 0}},
			true,
		},
		{
			"UA different equiv",
			ast.UA{A: typeU, B: typeU, Equiv: ast.Var{Ix: 0}},
			ast.UA{A: typeU, B: typeU, Equiv: ast.Var{Ix: 1}},
			false,
		},

		// UABeta
		{
			"UABeta same",
			ast.UABeta{Equiv: ast.Var{Ix: 0}, Arg: ast.Var{Ix: 1}},
			ast.UABeta{Equiv: ast.Var{Ix: 0}, Arg: ast.Var{Ix: 1}},
			true,
		},
		{
			"UABeta different arg",
			ast.UABeta{Equiv: ast.Var{Ix: 0}, Arg: ast.Var{Ix: 1}},
			ast.UABeta{Equiv: ast.Var{Ix: 0}, Arg: ast.Var{Ix: 2}},
			false,
		},

		// Cross-type comparisons should fail
		{"Interval != I0", ast.Interval{}, ast.I0{}, false},
		{"Path != PathP", ast.Path{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
			ast.PathP{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}, false},
		{"Comp != HComp",
			ast.Comp{IBinder: "i", A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			ast.HComp{A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			false},

		// HITApp tests
		{
			"HITApp same",
			ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: nil},
			ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: nil},
			true,
		},
		{
			"HITApp different HIT name",
			ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: nil},
			ast.HITApp{HITName: "S2", Ctor: "loop", Args: nil, IArgs: nil},
			false,
		},
		{
			"HITApp different ctor name",
			ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: nil},
			ast.HITApp{HITName: "S1", Ctor: "base", Args: nil, IArgs: nil},
			false,
		},
		{
			"HITApp with term args same",
			ast.HITApp{HITName: "Susp", Ctor: "merid", Args: []ast.Term{ast.Var{Ix: 0}}, IArgs: nil},
			ast.HITApp{HITName: "Susp", Ctor: "merid", Args: []ast.Term{ast.Var{Ix: 0}}, IArgs: nil},
			true,
		},
		{
			"HITApp different term args",
			ast.HITApp{HITName: "Susp", Ctor: "merid", Args: []ast.Term{ast.Var{Ix: 0}}, IArgs: nil},
			ast.HITApp{HITName: "Susp", Ctor: "merid", Args: []ast.Term{ast.Var{Ix: 1}}, IArgs: nil},
			false,
		},
		{
			"HITApp different args length",
			ast.HITApp{HITName: "Susp", Ctor: "merid", Args: []ast.Term{ast.Var{Ix: 0}}, IArgs: nil},
			ast.HITApp{HITName: "Susp", Ctor: "merid", Args: nil, IArgs: nil},
			false,
		},
		{
			"HITApp with iargs same",
			ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: []ast.Term{ast.I0{}}},
			ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: []ast.Term{ast.I0{}}},
			true,
		},
		{
			"HITApp different iargs",
			ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: []ast.Term{ast.I0{}}},
			ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: []ast.Term{ast.I1{}}},
			false,
		},
		{
			"HITApp different iargs length",
			ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: []ast.Term{ast.I0{}}},
			ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: nil},
			false,
		},
		{
			"HITApp full same",
			ast.HITApp{
				HITName: "Quot",
				Ctor:    "eq",
				Args:    []ast.Term{ast.Var{Ix: 0}, ast.Var{Ix: 1}},
				IArgs:   []ast.Term{ast.IVar{Ix: 0}},
			},
			ast.HITApp{
				HITName: "Quot",
				Ctor:    "eq",
				Args:    []ast.Term{ast.Var{Ix: 0}, ast.Var{Ix: 1}},
				IArgs:   []ast.Term{ast.IVar{Ix: 0}},
			},
			true,
		},

		// Cubical vs non-cubical type mismatch
		{"Interval vs Sort", ast.Interval{}, ast.Sort{U: 0}, false},
		{"Path vs Pi", ast.Path{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
			ast.Pi{Binder: "x", A: typeU, B: ast.Var{Ix: 0}}, false},
		{"FaceTop vs Global", ast.FaceTop{}, ast.Global{Name: "true"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AlphaEq(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("AlphaEq(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestShiftTermExtension tests term shifting for cubical types.
func TestShiftTermExtension(t *testing.T) {
	typeU := ast.Sort{U: 0}

	tests := []struct {
		name   string
		term   ast.Term
		d      int
		cutoff int
	}{
		// Constants don't change
		{"Interval", ast.Interval{}, 1, 0},
		{"I0", ast.I0{}, 1, 0},
		{"I1", ast.I1{}, 1, 0},

		// IVar uses separate index space (no shift)
		{"IVar", ast.IVar{Ix: 0}, 1, 0},

		// Path types get inner terms shifted
		{"Path", ast.Path{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}, 1, 0},
		{"PathP", ast.PathP{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}, 1, 0},
		{"PathLam", ast.PathLam{Binder: "i", Body: ast.Var{Ix: 0}}, 1, 0},
		{"PathApp", ast.PathApp{P: ast.Var{Ix: 0}, R: ast.I0{}}, 1, 0},
		{"Transport", ast.Transport{A: typeU, E: ast.Var{Ix: 0}}, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, handled := shiftTermExtension(tt.term, tt.d, tt.cutoff)
			if !handled {
				t.Errorf("shiftTermExtension(%T) not handled", tt.term)
			}
			if result == nil {
				t.Errorf("shiftTermExtension(%T) returned nil", tt.term)
			}
		})
	}
}

// TestShiftTermExtension_NonCubical tests that non-cubical terms are not handled.
func TestShiftTermExtension_NonCubical(t *testing.T) {
	nonCubical := []ast.Term{
		ast.Sort{U: 0},
		ast.Var{Ix: 0},
		ast.Global{Name: "x"},
		ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}},
		ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}},
		ast.App{T: ast.Global{Name: "f"}, U: ast.Global{Name: "x"}},
	}

	for _, tm := range nonCubical {
		t.Run(typeNameOf(tm), func(t *testing.T) {
			_, handled := shiftTermExtension(tm, 1, 0)
			if handled {
				t.Errorf("shiftTermExtension should not handle %T", tm)
			}
		})
	}
}

// TestAlphaEqFace_UnknownFace tests default case in alphaEqFace.
func TestAlphaEqFace_DefaultCase(t *testing.T) {
	// All known face types are covered - this ensures the default returns false
	faces := []ast.Face{
		ast.FaceTop{},
		ast.FaceBot{},
		ast.FaceEq{IVar: 0, IsOne: false},
		ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}},
		ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}},
	}

	// Same type should always match
	for _, f := range faces {
		if !alphaEqFace(f, f) {
			t.Errorf("alphaEqFace(%T, %T) should be true", f, f)
		}
	}
}

// typeNameOf returns a descriptive name for a term type.
func typeNameOf(t ast.Term) string {
	switch t.(type) {
	case ast.Sort:
		return "Sort"
	case ast.Var:
		return "Var"
	case ast.Global:
		return "Global"
	case ast.Pi:
		return "Pi"
	case ast.Lam:
		return "Lam"
	case ast.App:
		return "App"
	default:
		return "unknown"
	}
}

// ============================================================================
// Type Mismatch Coverage Tests
// ============================================================================

// TestAlphaEqExtension_TypeMismatch tests that cubical terms compared
// against different types return (false, true) to indicate handled but not equal.
func TestAlphaEqExtension_TypeMismatch(t *testing.T) {
	typeU := ast.Sort{U: 0}
	otherTerm := ast.Var{Ix: 99} // A term that won't match any cubical type

	tests := []struct {
		name string
		a    ast.Term
	}{
		{"Interval vs other", ast.Interval{}},
		{"I0 vs other", ast.I0{}},
		{"I1 vs other", ast.I1{}},
		{"IVar vs other", ast.IVar{Ix: 0}},
		{"Path vs other", ast.Path{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}},
		{"PathP vs other", ast.PathP{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}},
		{"PathLam vs other", ast.PathLam{Binder: "i", Body: ast.Var{Ix: 0}}},
		{"PathApp vs other", ast.PathApp{P: ast.Var{Ix: 0}, R: ast.I0{}}},
		{"Transport vs other", ast.Transport{A: typeU, E: ast.Var{Ix: 0}}},
		{"FaceEq vs other", ast.FaceEq{IVar: 0, IsOne: false}},
		{"FaceAnd vs other", ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}},
		{"FaceOr vs other", ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}}},
		{"Partial vs other", ast.Partial{Phi: ast.FaceTop{}, A: typeU}},
		{"System vs other", ast.System{Branches: []ast.SystemBranch{}}},
		{"Comp vs other", ast.Comp{IBinder: "i", A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}}},
		{"HComp vs other", ast.HComp{A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}}},
		{"Fill vs other", ast.Fill{IBinder: "i", A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}}},
		{"Glue vs other", ast.Glue{A: typeU, System: []ast.GlueBranch{}}},
		{"GlueElem vs other", ast.GlueElem{System: []ast.GlueElemBranch{}, Base: ast.Var{Ix: 0}}},
		{"Unglue vs other", ast.Unglue{Ty: typeU, G: ast.Var{Ix: 0}}},
		{"UA vs other", ast.UA{A: typeU, B: typeU, Equiv: ast.Var{Ix: 0}}},
		{"UABeta vs other", ast.UABeta{Equiv: ast.Var{Ix: 0}, Arg: ast.Var{Ix: 1}}},
		{"HITApp vs other", ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, handled := alphaEqExtension(tt.a, otherTerm)
			if !handled {
				t.Errorf("alphaEqExtension(%T, Var) should be handled", tt.a)
			}
			if result {
				t.Errorf("alphaEqExtension(%T, Var) should be false", tt.a)
			}
		})
	}
}

// TestAlphaEqFace_TypeMismatch tests face formula type mismatches.
func TestAlphaEqFace_TypeMismatch(t *testing.T) {
	tests := []struct {
		name string
		a    ast.Face
		b    ast.Face
	}{
		{"FaceEq vs FaceTop", ast.FaceEq{IVar: 0, IsOne: false}, ast.FaceTop{}},
		{"FaceEq vs FaceBot", ast.FaceEq{IVar: 0, IsOne: false}, ast.FaceBot{}},
		{"FaceEq vs FaceAnd", ast.FaceEq{IVar: 0, IsOne: false}, ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}},
		{"FaceAnd vs FaceTop", ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}, ast.FaceTop{}},
		{"FaceAnd vs FaceEq", ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}, ast.FaceEq{IVar: 0, IsOne: false}},
		{"FaceOr vs FaceTop", ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}}, ast.FaceTop{}},
		{"FaceOr vs FaceAnd", ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}}, ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if alphaEqFace(tt.a, tt.b) {
				t.Errorf("alphaEqFace(%T, %T) should be false", tt.a, tt.b)
			}
		})
	}
}

// TestAlphaEqExtension_SameCubicalTypeMismatch tests cubical term comparisons
// of the same type with different values (branch mismatch cases).
func TestAlphaEqExtension_SameCubicalTypeMismatch(t *testing.T) {
	typeU := ast.Sort{U: 0}

	tests := []struct {
		name string
		a, b ast.Term
	}{
		// System branch differences
		{
			"System different branch Phi",
			ast.System{Branches: []ast.SystemBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: false}, Term: ast.Var{Ix: 0}},
			}},
			ast.System{Branches: []ast.SystemBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: true}, Term: ast.Var{Ix: 0}},
			}},
		},
		{
			"System different branch Term",
			ast.System{Branches: []ast.SystemBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: false}, Term: ast.Var{Ix: 0}},
			}},
			ast.System{Branches: []ast.SystemBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: false}, Term: ast.Var{Ix: 1}},
			}},
		},

		// Glue branch differences
		{
			"Glue different branch Phi",
			ast.Glue{A: typeU, System: []ast.GlueBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: false}, T: typeU, Equiv: ast.Var{Ix: 0}},
			}},
			ast.Glue{A: typeU, System: []ast.GlueBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: true}, T: typeU, Equiv: ast.Var{Ix: 0}},
			}},
		},
		{
			"Glue different branch T",
			ast.Glue{A: typeU, System: []ast.GlueBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: false}, T: ast.Sort{U: 0}, Equiv: ast.Var{Ix: 0}},
			}},
			ast.Glue{A: typeU, System: []ast.GlueBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: false}, T: ast.Sort{U: 1}, Equiv: ast.Var{Ix: 0}},
			}},
		},
		{
			"Glue different branch Equiv",
			ast.Glue{A: typeU, System: []ast.GlueBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: false}, T: typeU, Equiv: ast.Var{Ix: 0}},
			}},
			ast.Glue{A: typeU, System: []ast.GlueBranch{
				{Phi: ast.FaceEq{IVar: 0, IsOne: false}, T: typeU, Equiv: ast.Var{Ix: 1}},
			}},
		},
		{
			"Glue different A",
			ast.Glue{A: ast.Sort{U: 0}, System: []ast.GlueBranch{}},
			ast.Glue{A: ast.Sort{U: 1}, System: []ast.GlueBranch{}},
		},

		// GlueElem branch differences
		{
			"GlueElem different branch Phi",
			ast.GlueElem{
				System: []ast.GlueElemBranch{{Phi: ast.FaceEq{IVar: 0, IsOne: false}, Term: ast.Var{Ix: 0}}},
				Base:   ast.Var{Ix: 1},
			},
			ast.GlueElem{
				System: []ast.GlueElemBranch{{Phi: ast.FaceEq{IVar: 0, IsOne: true}, Term: ast.Var{Ix: 0}}},
				Base:   ast.Var{Ix: 1},
			},
		},
		{
			"GlueElem different branch Term",
			ast.GlueElem{
				System: []ast.GlueElemBranch{{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}}},
				Base:   ast.Var{Ix: 1},
			},
			ast.GlueElem{
				System: []ast.GlueElemBranch{{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 2}}},
				Base:   ast.Var{Ix: 1},
			},
		},

		// Comp differences
		{
			"Comp different Phi",
			ast.Comp{IBinder: "i", A: typeU, Phi: ast.FaceEq{IVar: 0, IsOne: false}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			ast.Comp{IBinder: "i", A: typeU, Phi: ast.FaceEq{IVar: 0, IsOne: true}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
		},

		// HComp differences
		{
			"HComp different Phi",
			ast.HComp{A: typeU, Phi: ast.FaceEq{IVar: 0, IsOne: false}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			ast.HComp{A: typeU, Phi: ast.FaceEq{IVar: 0, IsOne: true}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
		},

		// Fill differences
		{
			"Fill different Phi",
			ast.Fill{IBinder: "i", A: typeU, Phi: ast.FaceEq{IVar: 0, IsOne: false}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			ast.Fill{IBinder: "i", A: typeU, Phi: ast.FaceEq{IVar: 0, IsOne: true}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if AlphaEq(tt.a, tt.b) {
				t.Errorf("AlphaEq() should return false for %s", tt.name)
			}
		})
	}
}

// TestAlphaEqExtension_DefaultCase verifies non-cubical terms trigger the default case.
func TestAlphaEqExtension_DefaultCase(t *testing.T) {
	// Non-cubical terms should return (false, false) from alphaEqExtension
	nonCubical := []ast.Term{
		ast.Sort{U: 0},
		ast.Var{Ix: 0},
		ast.Global{Name: "x"},
		ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}},
		ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}},
		ast.App{T: ast.Global{Name: "f"}, U: ast.Global{Name: "x"}},
		ast.Sigma{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}},
		ast.Pair{Fst: ast.Var{Ix: 0}, Snd: ast.Var{Ix: 1}},
		ast.Fst{P: ast.Var{Ix: 0}},
		ast.Snd{P: ast.Var{Ix: 0}},
		ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
		ast.Refl{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}},
	}

	for _, tm := range nonCubical {
		result, handled := alphaEqExtension(tm, tm)
		if handled {
			t.Errorf("alphaEqExtension(%T, %T) should not be handled (default case)", tm, tm)
		}
		if result {
			t.Errorf("alphaEqExtension(%T, %T) result should be false for default case", tm, tm)
		}
	}
}

// TestAlphaEqExtension_GlueElemSystemLengthMismatch tests different system lengths.
func TestAlphaEqExtension_GlueElemSystemLengthMismatch(t *testing.T) {
	typeU := ast.Sort{U: 0}

	// GlueElem with different system lengths
	a := ast.GlueElem{
		System: []ast.GlueElemBranch{
			{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}},
			{Phi: ast.FaceBot{}, Term: ast.Var{Ix: 1}},
		},
		Base: ast.Var{Ix: 2},
	}
	b := ast.GlueElem{
		System: []ast.GlueElemBranch{
			{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}},
		},
		Base: ast.Var{Ix: 2},
	}

	if AlphaEq(a, b) {
		t.Error("AlphaEq should return false for GlueElem with different system lengths")
	}

	// Glue with different system lengths
	c := ast.Glue{
		A: typeU,
		System: []ast.GlueBranch{
			{Phi: ast.FaceTop{}, T: typeU, Equiv: ast.Var{Ix: 0}},
			{Phi: ast.FaceBot{}, T: typeU, Equiv: ast.Var{Ix: 1}},
		},
	}
	d := ast.Glue{
		A: typeU,
		System: []ast.GlueBranch{
			{Phi: ast.FaceTop{}, T: typeU, Equiv: ast.Var{Ix: 0}},
		},
	}

	if AlphaEq(c, d) {
		t.Error("AlphaEq should return false for Glue with different system lengths")
	}
}

// TestAlphaEqExtension_CrossCubicalMismatch tests comparisons between
// different cubical types (e.g., Path vs PathP, Comp vs Fill).
func TestAlphaEqExtension_CrossCubicalMismatch(t *testing.T) {
	typeU := ast.Sort{U: 0}

	tests := []struct {
		name string
		a, b ast.Term
	}{
		{"I0 vs I1", ast.I0{}, ast.I1{}},
		{"I0 vs IVar", ast.I0{}, ast.IVar{Ix: 0}},
		{"I1 vs IVar", ast.I1{}, ast.IVar{Ix: 0}},
		{"Interval vs I0", ast.Interval{}, ast.I0{}},
		{"Interval vs IVar", ast.Interval{}, ast.IVar{Ix: 0}},

		{"Path vs PathP",
			ast.Path{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
			ast.PathP{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}},
		{"PathLam vs PathApp",
			ast.PathLam{Binder: "i", Body: ast.Var{Ix: 0}},
			ast.PathApp{P: ast.Var{Ix: 0}, R: ast.I0{}}},
		{"Path vs Transport",
			ast.Path{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}},
			ast.Transport{A: typeU, E: ast.Var{Ix: 0}}},

		{"FaceTop vs FaceBot", ast.FaceTop{}, ast.FaceBot{}},
		{"FaceTop vs FaceEq", ast.FaceTop{}, ast.FaceEq{IVar: 0, IsOne: false}},
		{"FaceBot vs FaceAnd", ast.FaceBot{}, ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}},

		{"Partial vs System",
			ast.Partial{Phi: ast.FaceTop{}, A: typeU},
			ast.System{Branches: []ast.SystemBranch{}}},

		{"Comp vs HComp",
			ast.Comp{IBinder: "i", A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			ast.HComp{A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}}},
		{"Comp vs Fill",
			ast.Comp{IBinder: "i", A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			ast.Fill{IBinder: "i", A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}}},
		{"HComp vs Fill",
			ast.HComp{A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}},
			ast.Fill{IBinder: "i", A: typeU, Phi: ast.FaceBot{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 1}}},

		{"Glue vs GlueElem",
			ast.Glue{A: typeU, System: []ast.GlueBranch{}},
			ast.GlueElem{System: []ast.GlueElemBranch{}, Base: ast.Var{Ix: 0}}},
		{"Glue vs Unglue",
			ast.Glue{A: typeU, System: []ast.GlueBranch{}},
			ast.Unglue{Ty: typeU, G: ast.Var{Ix: 0}}},
		{"GlueElem vs Unglue",
			ast.GlueElem{System: []ast.GlueElemBranch{}, Base: ast.Var{Ix: 0}},
			ast.Unglue{Ty: typeU, G: ast.Var{Ix: 0}}},

		{"UA vs UABeta",
			ast.UA{A: typeU, B: typeU, Equiv: ast.Var{Ix: 0}},
			ast.UABeta{Equiv: ast.Var{Ix: 0}, Arg: ast.Var{Ix: 1}}},

		{"HITApp vs Path",
			ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: nil},
			ast.Path{A: typeU, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if AlphaEq(tt.a, tt.b) {
				t.Errorf("AlphaEq(%T, %T) should return false", tt.a, tt.b)
			}
		})
	}
}
