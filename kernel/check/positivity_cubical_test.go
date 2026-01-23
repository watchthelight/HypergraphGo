package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// TestCheckPositivity_Path tests positivity checking for Path types.
func TestCheckPositivity_Path(t *testing.T) {
	// Valid: Path doesn't contain the inductive type in negative position
	validPath := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.Path{
				A: ast.Global{Name: "Nat"},
				X: ast.Var{Ix: 0},
				Y: ast.Var{Ix: 0},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validPath); err != nil {
		t.Errorf("CheckPositivity(Path valid) unexpected error: %v", err)
	}

	// Invalid: Good in Path type component in negative position
	invalidPath := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Path{
					A: ast.Global{Name: "Bad"}, // Bad in A
					X: ast.Var{Ix: 0},
					Y: ast.Var{Ix: 0},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidPath); err == nil {
		t.Error("CheckPositivity(Path negative) expected error, got nil")
	}

	// Invalid: Bad in Path endpoint (X)
	invalidPathX := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Path{
					A: ast.Global{Name: "Nat"},
					X: ast.Global{Name: "Bad"}, // Bad in X
					Y: ast.Var{Ix: 0},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidPathX); err == nil {
		t.Error("CheckPositivity(Path X negative) expected error, got nil")
	}
}

// TestCheckPositivity_PathP tests positivity checking for PathP (dependent path types).
func TestCheckPositivity_PathP(t *testing.T) {
	// Valid: PathP without the inductive type
	validPathP := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.PathP{
				A: ast.Lam{Binder: "i", Ann: ast.Interval{}, Body: ast.Global{Name: "Nat"}},
				X: ast.Var{Ix: 0},
				Y: ast.Var{Ix: 0},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validPathP); err != nil {
		t.Errorf("CheckPositivity(PathP valid) unexpected error: %v", err)
	}

	// Invalid: Bad in PathP type family
	invalidPathP := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.PathP{
					A: ast.Global{Name: "Bad"}, // Bad in A
					X: ast.Var{Ix: 0},
					Y: ast.Var{Ix: 0},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidPathP); err == nil {
		t.Error("CheckPositivity(PathP negative) expected error, got nil")
	}
}

// TestCheckPositivity_PathLam tests positivity checking for path lambdas.
func TestCheckPositivity_PathLam(t *testing.T) {
	// Valid: PathLam without the inductive type
	validPathLam := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.PathLam{
				Binder: "i",
				Body:   ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validPathLam); err != nil {
		t.Errorf("CheckPositivity(PathLam valid) unexpected error: %v", err)
	}

	// Invalid: Bad in PathLam body in negative position
	invalidPathLam := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.PathLam{
					Binder: "i",
					Body:   ast.Global{Name: "Bad"},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidPathLam); err == nil {
		t.Error("CheckPositivity(PathLam negative) expected error, got nil")
	}
}

// TestCheckPositivity_PathApp tests positivity checking for path applications.
func TestCheckPositivity_PathApp(t *testing.T) {
	// Valid: PathApp without the inductive type
	validPathApp := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.PathApp{
				P: ast.Var{Ix: 0},
				R: ast.I0{},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validPathApp); err != nil {
		t.Errorf("CheckPositivity(PathApp valid) unexpected error: %v", err)
	}

	// Invalid: Bad in PathApp path
	invalidPathApp := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.PathApp{
					P: ast.Global{Name: "Bad"},
					R: ast.I0{},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidPathApp); err == nil {
		t.Error("CheckPositivity(PathApp negative) expected error, got nil")
	}
}

// TestCheckPositivity_Transport tests positivity checking for transport.
func TestCheckPositivity_Transport(t *testing.T) {
	// Valid: Transport without the inductive type in negative position
	validTransport := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.Transport{
				A: ast.Lam{Binder: "i", Ann: ast.Interval{}, Body: ast.Global{Name: "Nat"}},
				E: ast.Var{Ix: 0},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validTransport); err != nil {
		t.Errorf("CheckPositivity(Transport valid) unexpected error: %v", err)
	}

	// Invalid: Bad in Transport type family
	invalidTransport := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Transport{
					A: ast.Global{Name: "Bad"},
					E: ast.Var{Ix: 0},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidTransport); err == nil {
		t.Error("CheckPositivity(Transport negative) expected error, got nil")
	}

	// Invalid: Bad in Transport element
	invalidTransportE := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Transport{
					A: ast.Global{Name: "Nat"},
					E: ast.Global{Name: "Bad"},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidTransportE); err == nil {
		t.Error("CheckPositivity(Transport E negative) expected error, got nil")
	}
}

// TestCheckPositivity_Comp tests positivity checking for composition operations.
func TestCheckPositivity_Comp(t *testing.T) {
	// Valid: Comp without the inductive type
	validComp := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.Comp{
				A:    ast.Lam{Binder: "i", Ann: ast.Interval{}, Body: ast.Global{Name: "Nat"}},
				Phi:  ast.FaceTop{},
				Tube: ast.Var{Ix: 0},
				Base: ast.Var{Ix: 0},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validComp); err != nil {
		t.Errorf("CheckPositivity(Comp valid) unexpected error: %v", err)
	}

	// Invalid: Bad in Comp type family
	invalidComp := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Comp{
					A:    ast.Global{Name: "Bad"},
					Phi:  ast.FaceTop{},
					Tube: ast.Var{Ix: 0},
					Base: ast.Var{Ix: 0},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidComp); err == nil {
		t.Error("CheckPositivity(Comp negative) expected error, got nil")
	}

	// Invalid: Bad in Comp tube
	invalidCompTube := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Comp{
					A:    ast.Global{Name: "Nat"},
					Phi:  ast.FaceTop{},
					Tube: ast.Global{Name: "Bad"},
					Base: ast.Var{Ix: 0},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidCompTube); err == nil {
		t.Error("CheckPositivity(Comp tube negative) expected error, got nil")
	}
}

// TestCheckPositivity_HComp tests positivity checking for homogeneous composition.
func TestCheckPositivity_HComp(t *testing.T) {
	// Valid: HComp without the inductive type
	validHComp := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.HComp{
				A:    ast.Global{Name: "Nat"},
				Phi:  ast.FaceTop{},
				Tube: ast.Var{Ix: 0},
				Base: ast.Var{Ix: 0},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validHComp); err != nil {
		t.Errorf("CheckPositivity(HComp valid) unexpected error: %v", err)
	}

	// Invalid: Bad in HComp
	invalidHComp := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.HComp{
					A:    ast.Global{Name: "Bad"},
					Phi:  ast.FaceTop{},
					Tube: ast.Var{Ix: 0},
					Base: ast.Var{Ix: 0},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidHComp); err == nil {
		t.Error("CheckPositivity(HComp negative) expected error, got nil")
	}
}

// TestCheckPositivity_Fill tests positivity checking for fill operation.
func TestCheckPositivity_Fill(t *testing.T) {
	// Valid: Fill without the inductive type
	validFill := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.Fill{
				A:    ast.Lam{Binder: "i", Ann: ast.Interval{}, Body: ast.Global{Name: "Nat"}},
				Phi:  ast.FaceTop{},
				Tube: ast.Var{Ix: 0},
				Base: ast.Var{Ix: 0},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validFill); err != nil {
		t.Errorf("CheckPositivity(Fill valid) unexpected error: %v", err)
	}

	// Invalid: Bad in Fill base
	invalidFill := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Fill{
					A:    ast.Global{Name: "Nat"},
					Phi:  ast.FaceTop{},
					Tube: ast.Var{Ix: 0},
					Base: ast.Global{Name: "Bad"},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidFill); err == nil {
		t.Error("CheckPositivity(Fill negative) expected error, got nil")
	}
}

// TestCheckPositivity_Glue tests positivity checking for Glue types.
func TestCheckPositivity_Glue(t *testing.T) {
	// Valid: Glue without the inductive type
	validGlue := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.Glue{
				A: ast.Global{Name: "Nat"},
				System: []ast.GlueBranch{
					{
						Phi:   ast.FaceEq{IVar: 0, IsOne: true},
						T:     ast.Global{Name: "Bool"},
						Equiv: ast.Var{Ix: 0},
					},
				},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validGlue); err != nil {
		t.Errorf("CheckPositivity(Glue valid) unexpected error: %v", err)
	}

	// Invalid: Bad in Glue base type
	invalidGlue := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Glue{
					A:      ast.Global{Name: "Bad"},
					System: nil,
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidGlue); err == nil {
		t.Error("CheckPositivity(Glue negative) expected error, got nil")
	}

	// Invalid: Bad in Glue system branch type
	invalidGlueSystem := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Glue{
					A: ast.Global{Name: "Nat"},
					System: []ast.GlueBranch{
						{
							Phi:   ast.FaceTop{},
							T:     ast.Global{Name: "Bad"},
							Equiv: ast.Var{Ix: 0},
						},
					},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidGlueSystem); err == nil {
		t.Error("CheckPositivity(Glue system negative) expected error, got nil")
	}
}

// TestCheckPositivity_GlueElem tests positivity checking for GlueElem.
func TestCheckPositivity_GlueElem(t *testing.T) {
	// Valid: GlueElem without the inductive type
	validGlueElem := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.GlueElem{
				System: []ast.GlueElemBranch{
					{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}},
				},
				Base: ast.Var{Ix: 0},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validGlueElem); err != nil {
		t.Errorf("CheckPositivity(GlueElem valid) unexpected error: %v", err)
	}

	// Invalid: Bad in GlueElem base
	invalidGlueElem := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.GlueElem{
					System: nil,
					Base:   ast.Global{Name: "Bad"},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidGlueElem); err == nil {
		t.Error("CheckPositivity(GlueElem negative) expected error, got nil")
	}
}

// TestCheckPositivity_Unglue tests positivity checking for Unglue.
func TestCheckPositivity_Unglue(t *testing.T) {
	// Valid: Unglue without the inductive type
	validUnglue := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.Unglue{
				Ty: ast.Global{Name: "Nat"},
				G:  ast.Var{Ix: 0},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validUnglue); err != nil {
		t.Errorf("CheckPositivity(Unglue valid) unexpected error: %v", err)
	}

	// Invalid: Bad in Unglue
	invalidUnglue := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Unglue{
					Ty: ast.Global{Name: "Bad"},
					G:  ast.Var{Ix: 0},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidUnglue); err == nil {
		t.Error("CheckPositivity(Unglue negative) expected error, got nil")
	}
}

// TestCheckPositivity_UA tests positivity checking for univalence.
func TestCheckPositivity_UA(t *testing.T) {
	// Valid: UA without the inductive type
	validUA := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.UA{
				A:     ast.Global{Name: "Nat"},
				B:     ast.Global{Name: "Bool"},
				Equiv: ast.Var{Ix: 0},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validUA); err != nil {
		t.Errorf("CheckPositivity(UA valid) unexpected error: %v", err)
	}

	// Invalid: Bad in UA types
	invalidUA := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.UA{
					A:     ast.Global{Name: "Bad"},
					B:     ast.Global{Name: "Nat"},
					Equiv: ast.Var{Ix: 0},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidUA); err == nil {
		t.Error("CheckPositivity(UA negative) expected error, got nil")
	}
}

// TestCheckPositivity_UABeta tests positivity checking for UA beta.
func TestCheckPositivity_UABeta(t *testing.T) {
	// Valid: UABeta without the inductive type
	validUABeta := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.UABeta{
				Equiv: ast.Var{Ix: 0},
				Arg:   ast.Var{Ix: 0},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validUABeta); err != nil {
		t.Errorf("CheckPositivity(UABeta valid) unexpected error: %v", err)
	}

	// Invalid: Bad in UABeta
	invalidUABeta := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.UABeta{
					Equiv: ast.Global{Name: "Bad"},
					Arg:   ast.Var{Ix: 0},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidUABeta); err == nil {
		t.Error("CheckPositivity(UABeta negative) expected error, got nil")
	}
}

// TestCheckPositivity_Partial tests positivity checking for Partial types.
func TestCheckPositivity_Partial(t *testing.T) {
	// Valid: Partial without the inductive type
	validPartial := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.Partial{
				Phi: ast.FaceTop{},
				A:   ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validPartial); err != nil {
		t.Errorf("CheckPositivity(Partial valid) unexpected error: %v", err)
	}

	// Invalid: Bad in Partial type
	invalidPartial := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Partial{
					Phi: ast.FaceTop{},
					A:   ast.Global{Name: "Bad"},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidPartial); err == nil {
		t.Error("CheckPositivity(Partial negative) expected error, got nil")
	}
}

// TestCheckPositivity_System tests positivity checking for System terms.
func TestCheckPositivity_System(t *testing.T) {
	// Valid: System without the inductive type
	validSystem := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.System{
				Branches: []ast.SystemBranch{
					{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}},
					{Phi: ast.FaceBot{}, Term: ast.Var{Ix: 0}},
				},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", validSystem); err != nil {
		t.Errorf("CheckPositivity(System valid) unexpected error: %v", err)
	}

	// Invalid: Bad in System branch
	invalidSystem := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.System{
					Branches: []ast.SystemBranch{
						{Phi: ast.FaceTop{}, Term: ast.Global{Name: "Bad"}},
					},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	if err := CheckPositivity("Bad", invalidSystem); err == nil {
		t.Error("CheckPositivity(System negative) expected error, got nil")
	}
}

// TestCheckPositivity_IntervalTerms tests that interval terms don't cause issues.
func TestCheckPositivity_IntervalTerms(t *testing.T) {
	// Interval type, I0, I1, IVar should all pass positivity
	tests := []struct {
		name string
		term ast.Term
	}{
		{"Interval", ast.Interval{}},
		{"I0", ast.I0{}},
		{"I1", ast.I1{}},
		{"IVar", ast.IVar{Ix: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctors := []Constructor{
				{Name: "mk", Type: ast.Pi{
					Binder: "_",
					A:      tt.term,
					B:      ast.Global{Name: "Good"},
				}},
			}
			if err := CheckPositivity("Good", ctors); err != nil {
				t.Errorf("CheckPositivity(%s) unexpected error: %v", tt.name, err)
			}
		})
	}
}

// TestCheckPositivity_FaceFormulas tests positivity with face formulas.
func TestCheckPositivity_FaceFormulas(t *testing.T) {
	// FaceAnd with inductive type (should pass since faces don't contain types)
	faceAnd := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.Partial{
				Phi: ast.FaceAnd{
					Left:  ast.FaceEq{IVar: 0, IsOne: true},
					Right: ast.FaceEq{IVar: 1, IsOne: false},
				},
				A: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", faceAnd); err != nil {
		t.Errorf("CheckPositivity(FaceAnd) unexpected error: %v", err)
	}

	// FaceOr
	faceOr := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.Partial{
				Phi: ast.FaceOr{
					Left:  ast.FaceEq{IVar: 0, IsOne: true},
					Right: ast.FaceEq{IVar: 1, IsOne: false},
				},
				A: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	if err := CheckPositivity("Good", faceOr); err != nil {
		t.Errorf("CheckPositivity(FaceOr) unexpected error: %v", err)
	}
}

// TestOccursIn_CubicalTerms tests occurrence checking for cubical terms.
func TestOccursIn_CubicalTerms(t *testing.T) {
	tests := []struct {
		name     string
		global   string
		term     ast.Term
		expected bool
	}{
		// Interval terms - no occurrences
		{"Interval", "T", ast.Interval{}, false},
		{"I0", "T", ast.I0{}, false},
		{"I1", "T", ast.I1{}, false},
		{"IVar", "T", ast.IVar{Ix: 0}, false},

		// Path types
		{"Path A", "T", ast.Path{A: ast.Global{Name: "T"}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}, true},
		{"Path X", "T", ast.Path{A: ast.Global{Name: "Nat"}, X: ast.Global{Name: "T"}, Y: ast.Var{Ix: 0}}, true},
		{"Path Y", "T", ast.Path{A: ast.Global{Name: "Nat"}, X: ast.Var{Ix: 0}, Y: ast.Global{Name: "T"}}, true},
		{"Path none", "T", ast.Path{A: ast.Global{Name: "Nat"}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}, false},

		// PathP
		{"PathP A", "T", ast.PathP{A: ast.Global{Name: "T"}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}, true},
		{"PathP none", "T", ast.PathP{A: ast.Global{Name: "Nat"}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}, false},

		// PathLam
		{"PathLam Body", "T", ast.PathLam{Binder: "i", Body: ast.Global{Name: "T"}}, true},
		{"PathLam none", "T", ast.PathLam{Binder: "i", Body: ast.Var{Ix: 0}}, false},

		// PathApp
		{"PathApp P", "T", ast.PathApp{P: ast.Global{Name: "T"}, R: ast.I0{}}, true},
		{"PathApp R", "T", ast.PathApp{P: ast.Var{Ix: 0}, R: ast.Global{Name: "T"}}, true},
		{"PathApp none", "T", ast.PathApp{P: ast.Var{Ix: 0}, R: ast.I0{}}, false},

		// Transport
		{"Transport A", "T", ast.Transport{A: ast.Global{Name: "T"}, E: ast.Var{Ix: 0}}, true},
		{"Transport E", "T", ast.Transport{A: ast.Global{Name: "Nat"}, E: ast.Global{Name: "T"}}, true},
		{"Transport none", "T", ast.Transport{A: ast.Global{Name: "Nat"}, E: ast.Var{Ix: 0}}, false},

		// Comp
		{"Comp A", "T", ast.Comp{A: ast.Global{Name: "T"}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 0}}, true},
		{"Comp Tube", "T", ast.Comp{A: ast.Global{Name: "Nat"}, Phi: ast.FaceTop{}, Tube: ast.Global{Name: "T"}, Base: ast.Var{Ix: 0}}, true},
		{"Comp Base", "T", ast.Comp{A: ast.Global{Name: "Nat"}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Global{Name: "T"}}, true},
		{"Comp none", "T", ast.Comp{A: ast.Global{Name: "Nat"}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 0}}, false},

		// HComp
		{"HComp A", "T", ast.HComp{A: ast.Global{Name: "T"}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 0}}, true},
		{"HComp none", "T", ast.HComp{A: ast.Global{Name: "Nat"}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 0}}, false},

		// Fill
		{"Fill A", "T", ast.Fill{A: ast.Global{Name: "T"}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 0}}, true},
		{"Fill none", "T", ast.Fill{A: ast.Global{Name: "Nat"}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Var{Ix: 0}}, false},

		// Partial
		{"Partial A", "T", ast.Partial{Phi: ast.FaceTop{}, A: ast.Global{Name: "T"}}, true},
		{"Partial none", "T", ast.Partial{Phi: ast.FaceTop{}, A: ast.Global{Name: "Nat"}}, false},

		// System
		{"System branch", "T", ast.System{Branches: []ast.SystemBranch{{Phi: ast.FaceTop{}, Term: ast.Global{Name: "T"}}}}, true},
		{"System empty", "T", ast.System{Branches: nil}, false},
		{"System none", "T", ast.System{Branches: []ast.SystemBranch{{Phi: ast.FaceTop{}, Term: ast.Var{Ix: 0}}}}, false},

		// Glue
		{"Glue A", "T", ast.Glue{A: ast.Global{Name: "T"}, System: nil}, true},
		{"Glue system T", "T", ast.Glue{A: ast.Global{Name: "Nat"}, System: []ast.GlueBranch{{Phi: ast.FaceTop{}, T: ast.Global{Name: "T"}, Equiv: ast.Var{Ix: 0}}}}, true},
		{"Glue system Equiv", "T", ast.Glue{A: ast.Global{Name: "Nat"}, System: []ast.GlueBranch{{Phi: ast.FaceTop{}, T: ast.Global{Name: "Nat"}, Equiv: ast.Global{Name: "T"}}}}, true},
		{"Glue none", "T", ast.Glue{A: ast.Global{Name: "Nat"}, System: nil}, false},

		// GlueElem
		{"GlueElem Base", "T", ast.GlueElem{System: nil, Base: ast.Global{Name: "T"}}, true},
		{"GlueElem system", "T", ast.GlueElem{System: []ast.GlueElemBranch{{Phi: ast.FaceTop{}, Term: ast.Global{Name: "T"}}}, Base: ast.Var{Ix: 0}}, true},
		{"GlueElem none", "T", ast.GlueElem{System: nil, Base: ast.Var{Ix: 0}}, false},

		// Unglue
		{"Unglue Ty", "T", ast.Unglue{Ty: ast.Global{Name: "T"}, G: ast.Var{Ix: 0}}, true},
		{"Unglue G", "T", ast.Unglue{Ty: ast.Global{Name: "Nat"}, G: ast.Global{Name: "T"}}, true},
		{"Unglue none", "T", ast.Unglue{Ty: ast.Global{Name: "Nat"}, G: ast.Var{Ix: 0}}, false},

		// UA
		{"UA A", "T", ast.UA{A: ast.Global{Name: "T"}, B: ast.Global{Name: "Nat"}, Equiv: ast.Var{Ix: 0}}, true},
		{"UA B", "T", ast.UA{A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "T"}, Equiv: ast.Var{Ix: 0}}, true},
		{"UA Equiv", "T", ast.UA{A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "Bool"}, Equiv: ast.Global{Name: "T"}}, true},
		{"UA none", "T", ast.UA{A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "Bool"}, Equiv: ast.Var{Ix: 0}}, false},

		// UABeta
		{"UABeta Equiv", "T", ast.UABeta{Equiv: ast.Global{Name: "T"}, Arg: ast.Var{Ix: 0}}, true},
		{"UABeta Arg", "T", ast.UABeta{Equiv: ast.Var{Ix: 0}, Arg: ast.Global{Name: "T"}}, true},
		{"UABeta none", "T", ast.UABeta{Equiv: ast.Var{Ix: 0}, Arg: ast.Var{Ix: 0}}, false},

		// Face formulas (always false since they don't contain type references)
		{"FaceTop", "T", ast.FaceTop{}, false},
		{"FaceBot", "T", ast.FaceBot{}, false},
		{"FaceEq", "T", ast.FaceEq{IVar: 0, IsOne: true}, false},
		{"FaceAnd", "T", ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}, false},
		{"FaceOr", "T", ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OccursIn(tt.global, tt.term)
			if result != tt.expected {
				t.Errorf("OccursIn(%q, %T) = %v, want %v", tt.global, tt.term, result, tt.expected)
			}
		})
	}
}

// TestOccursInFace tests the occursInFace helper function.
func TestOccursInFace(t *testing.T) {
	// Face formulas never contain type references, so occursInFace should always return false
	tests := []struct {
		name string
		face ast.Face
	}{
		{"nil", nil},
		{"FaceTop", ast.FaceTop{}},
		{"FaceBot", ast.FaceBot{}},
		{"FaceEq", ast.FaceEq{IVar: 0, IsOne: true}},
		{"FaceAnd", ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}},
		{"FaceOr", ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}}},
		{"nested FaceAnd", ast.FaceAnd{
			Left:  ast.FaceAnd{Left: ast.FaceEq{IVar: 0, IsOne: true}, Right: ast.FaceBot{}},
			Right: ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceEq{IVar: 1, IsOne: false}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since faces don't contain types, this should always be false
			result := occursInFace("AnyType", tt.face)
			if result {
				t.Errorf("occursInFace(%q, %T) = true, want false", "AnyType", tt.face)
			}
		})
	}
}

// TestCheckArgTypePositivityFace tests face formula positivity checking.
func TestCheckArgTypePositivityFace(t *testing.T) {
	// Face formula positivity checking should always succeed since faces don't contain types
	tests := []struct {
		name string
		face ast.Face
	}{
		{"nil", nil},
		{"FaceTop", ast.FaceTop{}},
		{"FaceBot", ast.FaceBot{}},
		{"FaceEq", ast.FaceEq{IVar: 0, IsOne: true}},
		{"FaceAnd", ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}},
		{"FaceOr", ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkArgTypePositivityFace("Ind", "ctor", tt.face, Positive, 0)
			if err != nil {
				t.Errorf("checkArgTypePositivityFace() unexpected error: %v", err)
			}
		})
	}
}

// ============================================================================
// Extended Cubical Positivity Error Path Tests
// ============================================================================

// TestCheckPositivity_CubicalErrorPaths tests error paths in cubical positivity checking.
func TestCheckPositivity_CubicalErrorPaths(t *testing.T) {
	tests := []struct {
		name string
		term ast.Term
	}{
		// PathP endpoint errors
		{
			"PathP X negative",
			ast.Pi{Binder: "_", A: ast.Pi{Binder: "_", A: ast.PathP{A: ast.Sort{U: 0}, X: ast.Global{Name: "Bad"}, Y: ast.Var{Ix: 0}}, B: ast.Sort{U: 0}}, B: ast.Global{Name: "Bad"}},
		},
		{
			"PathP Y negative",
			ast.Pi{Binder: "_", A: ast.Pi{Binder: "_", A: ast.PathP{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Global{Name: "Bad"}}, B: ast.Sort{U: 0}}, B: ast.Global{Name: "Bad"}},
		},

		// PathApp R errors
		{
			"PathApp R negative",
			ast.Pi{Binder: "_", A: ast.Pi{Binder: "_", A: ast.PathApp{P: ast.Var{Ix: 0}, R: ast.Global{Name: "Bad"}}, B: ast.Sort{U: 0}}, B: ast.Global{Name: "Bad"}},
		},

		// Comp/HComp/Fill tube/base errors
		{
			"Comp Base negative",
			ast.Pi{Binder: "_", A: ast.Pi{Binder: "_", A: ast.Comp{A: ast.Sort{U: 0}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Global{Name: "Bad"}}, B: ast.Sort{U: 0}}, B: ast.Global{Name: "Bad"}},
		},
		{
			"HComp Tube negative",
			ast.Pi{Binder: "_", A: ast.Pi{Binder: "_", A: ast.HComp{A: ast.Sort{U: 0}, Phi: ast.FaceTop{}, Tube: ast.Global{Name: "Bad"}, Base: ast.Var{Ix: 0}}, B: ast.Sort{U: 0}}, B: ast.Global{Name: "Bad"}},
		},
		{
			"HComp Base negative",
			ast.Pi{Binder: "_", A: ast.Pi{Binder: "_", A: ast.HComp{A: ast.Sort{U: 0}, Phi: ast.FaceTop{}, Tube: ast.Var{Ix: 0}, Base: ast.Global{Name: "Bad"}}, B: ast.Sort{U: 0}}, B: ast.Global{Name: "Bad"}},
		},
		{
			"Fill Tube negative",
			ast.Pi{Binder: "_", A: ast.Pi{Binder: "_", A: ast.Fill{A: ast.Sort{U: 0}, Phi: ast.FaceTop{}, Tube: ast.Global{Name: "Bad"}, Base: ast.Var{Ix: 0}}, B: ast.Sort{U: 0}}, B: ast.Global{Name: "Bad"}},
		},

		// Glue system errors
		{
			"Glue Equiv negative",
			ast.Pi{Binder: "_", A: ast.Pi{Binder: "_", A: ast.Glue{A: ast.Sort{U: 0}, System: []ast.GlueBranch{{Phi: ast.FaceTop{}, T: ast.Sort{U: 0}, Equiv: ast.Global{Name: "Bad"}}}}, B: ast.Sort{U: 0}}, B: ast.Global{Name: "Bad"}},
		},

		// GlueElem system errors
		{
			"GlueElem system Term negative",
			ast.Pi{Binder: "_", A: ast.Pi{Binder: "_", A: ast.GlueElem{System: []ast.GlueElemBranch{{Phi: ast.FaceTop{}, Term: ast.Global{Name: "Bad"}}}, Base: ast.Var{Ix: 0}}, B: ast.Sort{U: 0}}, B: ast.Global{Name: "Bad"}},
		},

		// Unglue G error
		{
			"Unglue G negative",
			ast.Pi{Binder: "_", A: ast.Pi{Binder: "_", A: ast.Unglue{Ty: ast.Sort{U: 0}, G: ast.Global{Name: "Bad"}}, B: ast.Sort{U: 0}}, B: ast.Global{Name: "Bad"}},
		},

		// UA B and Equiv errors
		{
			"UA B negative",
			ast.Pi{Binder: "_", A: ast.Pi{Binder: "_", A: ast.UA{A: ast.Sort{U: 0}, B: ast.Global{Name: "Bad"}, Equiv: ast.Var{Ix: 0}}, B: ast.Sort{U: 0}}, B: ast.Global{Name: "Bad"}},
		},
		{
			"UA Equiv negative",
			ast.Pi{Binder: "_", A: ast.Pi{Binder: "_", A: ast.UA{A: ast.Sort{U: 0}, B: ast.Sort{U: 0}, Equiv: ast.Global{Name: "Bad"}}, B: ast.Sort{U: 0}}, B: ast.Global{Name: "Bad"}},
		},

		// UABeta Arg error
		{
			"UABeta Arg negative",
			ast.Pi{Binder: "_", A: ast.Pi{Binder: "_", A: ast.UABeta{Equiv: ast.Var{Ix: 0}, Arg: ast.Global{Name: "Bad"}}, B: ast.Sort{U: 0}}, B: ast.Global{Name: "Bad"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctors := []Constructor{{Name: "mk", Type: tt.term}}
			err := CheckPositivity("Bad", ctors)
			if err == nil {
				t.Errorf("CheckPositivity expected error for %s", tt.name)
			}
		})
	}
}

// TestOccursInFace_AllBranches tests all branches of occursInFace.
func TestOccursInFace_AllBranches(t *testing.T) {
	// Test nested face formulas
	tests := []struct {
		name string
		face ast.Face
	}{
		{"deeply nested FaceAnd left", ast.FaceAnd{
			Left:  ast.FaceAnd{Left: ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}}, Right: ast.FaceEq{IVar: 0, IsOne: true}},
			Right: ast.FaceBot{},
		}},
		{"deeply nested FaceOr right", ast.FaceOr{
			Left:  ast.FaceTop{},
			Right: ast.FaceOr{Left: ast.FaceAnd{Left: ast.FaceEq{IVar: 1, IsOne: false}, Right: ast.FaceTop{}}, Right: ast.FaceBot{}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := occursInFace("T", tt.face)
			if result {
				t.Errorf("occursInFace should return false for face formulas")
			}
		})
	}
}
