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
