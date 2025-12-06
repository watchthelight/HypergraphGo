package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
)

// TestEndToEnd_DeclareAndEvaluate tests the full pipeline:
// 1. Declare an inductive type with DeclareInductive
// 2. Verify the eliminator is generated correctly
// 3. Verify the recursor is registered for reduction
// 4. Test that evaluation works correctly
func TestEndToEnd_DeclareAndEvaluate(t *testing.T) {
	// Clear the recursor registry for clean test
	eval.ClearRecursorRegistry()

	// Create a fresh environment
	env := NewGlobalEnv()

	// Declare a Unit type
	err := env.DeclareInductive("Unit", ast.Sort{U: 0}, []Constructor{
		{Name: "tt", Type: ast.Global{Name: "Unit"}},
	}, "unitElim")
	if err != nil {
		t.Fatalf("DeclareInductive(Unit) failed: %v", err)
	}

	// Verify eliminator type was generated
	elimType := env.LookupType("unitElim")
	if elimType == nil {
		t.Fatal("unitElim type not found in environment")
	}

	// Verify it's a Pi type (P : Unit -> Type) -> ...
	pi, ok := elimType.(ast.Pi)
	if !ok {
		t.Fatalf("unitElim expected Pi type, got %T", elimType)
	}
	if pi.Binder != "P" {
		t.Errorf("unitElim expected motive binder 'P', got %q", pi.Binder)
	}

	// Verify recursor is registered
	info := eval.LookupRecursor("unitElim")
	if info == nil {
		t.Fatal("unitElim not registered in recursor registry")
	}
	if info.IndName != "Unit" {
		t.Errorf("RecursorInfo.IndName = %q, want 'Unit'", info.IndName)
	}
	if info.NumCases != 1 {
		t.Errorf("RecursorInfo.NumCases = %d, want 1", info.NumCases)
	}
	if len(info.Ctors) != 1 || info.Ctors[0].Name != "tt" {
		t.Error("RecursorInfo.Ctors should have one constructor 'tt'")
	}

	// Test evaluation: unitElim P ptt tt --> ptt
	unitElim := ast.Global{Name: "unitElim"}
	motive := ast.Lam{Binder: "_", Body: ast.Sort{U: 0}}
	ptt := ast.Global{Name: "result"}
	tt := ast.Global{Name: "tt"}

	term := ast.MkApps(unitElim, motive, ptt, tt)
	normalized := eval.NormalizeNBE(term)

	if normalized != "result" {
		t.Errorf("unitElim P result tt evaluated to %q, want 'result'", normalized)
	}

	// Clean up
	eval.ClearRecursorRegistry()
}

// TestEndToEnd_CustomNatLike tests a custom Nat-like inductive.
func TestEndToEnd_CustomNatLike(t *testing.T) {
	eval.ClearRecursorRegistry()

	env := NewGlobalEnvWithPrimitives() // Need Nat for the test

	// Declare MyNat : Type0 with constructors mzero : MyNat, msucc : MyNat -> MyNat
	err := env.DeclareInductive("MyNat", ast.Sort{U: 0}, []Constructor{
		{Name: "mzero", Type: ast.Global{Name: "MyNat"}},
		{Name: "msucc", Type: ast.Pi{
			Binder: "_",
			A:      ast.Global{Name: "MyNat"},
			B:      ast.Global{Name: "MyNat"},
		}},
	}, "myNatElim")
	if err != nil {
		t.Fatalf("DeclareInductive(MyNat) failed: %v", err)
	}

	// Verify recursor info
	info := eval.LookupRecursor("myNatElim")
	if info == nil {
		t.Fatal("myNatElim not registered")
	}
	if info.NumCases != 2 {
		t.Errorf("myNatElim NumCases = %d, want 2", info.NumCases)
	}
	if len(info.Ctors[1].RecursiveIdx) != 1 || info.Ctors[1].RecursiveIdx[0] != 0 {
		t.Error("msucc should have one recursive arg at index 0")
	}

	// Test mzero case: myNatElim P pz ps mzero --> pz
	myNatElim := ast.Global{Name: "myNatElim"}
	motive := ast.Lam{Binder: "_", Body: ast.Sort{U: 0}}
	pz := ast.Global{Name: "zeroCase"}
	ps := ast.Lam{Binder: "n", Body: ast.Lam{Binder: "ih", Body: ast.Global{Name: "succCase"}}}
	mzero := ast.Global{Name: "mzero"}

	term := ast.MkApps(myNatElim, motive, pz, ps, mzero)
	normalized := eval.NormalizeNBE(term)

	if normalized != "zeroCase" {
		t.Errorf("myNatElim _ zeroCase _ mzero = %q, want 'zeroCase'", normalized)
	}

	// Test msucc mzero case: myNatElim P pz ps (msucc mzero) --> ps mzero (myNatElim P pz ps mzero)
	one := ast.App{T: ast.Global{Name: "msucc"}, U: mzero}
	term = ast.MkApps(myNatElim, motive, pz, ps, one)
	normalized = eval.NormalizeNBE(term)

	if normalized != "succCase" {
		t.Errorf("myNatElim _ _ _ (msucc mzero) = %q, want 'succCase'", normalized)
	}

	eval.ClearRecursorRegistry()
}

// TestEndToEnd_PositivityRejection verifies negative inductives are rejected.
func TestEndToEnd_PositivityRejection(t *testing.T) {
	env := NewGlobalEnvWithPrimitives()

	// Try to declare a negative inductive
	err := env.DeclareInductive("Bad", ast.Sort{U: 0}, []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A:      ast.Global{Name: "Bad"}, // Negative occurrence!
				B:      ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}, "badElim")

	if err == nil {
		t.Error("DeclareInductive should reject negative occurrence")
	}

	// Verify Bad was not added to environment
	if env.LookupType("Bad") != nil {
		t.Error("Bad should not be in environment after rejection")
	}
}

// TestEndToEnd_IllFormedConstructor verifies ill-formed constructors are rejected.
func TestEndToEnd_IllFormedConstructor(t *testing.T) {
	env := NewGlobalEnv()

	// Try to declare with constructor that references unknown type
	err := env.DeclareInductive("Test", ast.Sort{U: 0}, []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A:      ast.Global{Name: "UnknownType"}, // Not in environment!
			B:      ast.Global{Name: "Test"},
		}},
	}, "testElim")

	if err == nil {
		t.Error("DeclareInductive should reject unknown type in constructor")
	}

	// Verify Test was not added
	if env.LookupType("Test") != nil {
		t.Error("Test should not be in environment after rejection")
	}
}

// TestEndToEnd_RecursorTypeStructure verifies the generated recursor type structure.
func TestEndToEnd_RecursorTypeStructure(t *testing.T) {
	env := NewGlobalEnv()

	// Declare Bool
	err := env.DeclareInductive("Bool", ast.Sort{U: 0}, []Constructor{
		{Name: "true", Type: ast.Global{Name: "Bool"}},
		{Name: "false", Type: ast.Global{Name: "Bool"}},
	}, "boolElim")
	if err != nil {
		t.Fatalf("DeclareInductive(Bool) failed: %v", err)
	}

	// Get eliminator type
	elimType := env.LookupType("boolElim")

	// boolElim : (P : Bool -> Type) -> P true -> P false -> (b : Bool) -> P b
	// Structure: Pi P . Pi case_true . Pi case_false . Pi b . P b

	// Verify outer Pi (motive P)
	pi1, ok := elimType.(ast.Pi)
	if !ok {
		t.Fatalf("boolElim level 1: expected Pi, got %T", elimType)
	}
	if pi1.Binder != "P" {
		t.Errorf("boolElim binder 1 = %q, want 'P'", pi1.Binder)
	}

	// Verify second Pi (case_true)
	pi2, ok := pi1.B.(ast.Pi)
	if !ok {
		t.Fatalf("boolElim level 2: expected Pi, got %T", pi1.B)
	}
	if pi2.Binder != "case_true" {
		t.Errorf("boolElim binder 2 = %q, want 'case_true'", pi2.Binder)
	}

	// Verify third Pi (case_false)
	pi3, ok := pi2.B.(ast.Pi)
	if !ok {
		t.Fatalf("boolElim level 3: expected Pi, got %T", pi2.B)
	}
	if pi3.Binder != "case_false" {
		t.Errorf("boolElim binder 3 = %q, want 'case_false'", pi3.Binder)
	}

	// Verify fourth Pi (target b)
	pi4, ok := pi3.B.(ast.Pi)
	if !ok {
		t.Fatalf("boolElim level 4: expected Pi, got %T", pi3.B)
	}
	if pi4.Binder != "t" {
		t.Errorf("boolElim binder 4 = %q, want 't'", pi4.Binder)
	}

	// Verify target domain is Bool
	if g, ok := pi4.A.(ast.Global); !ok || g.Name != "Bool" {
		t.Errorf("boolElim target domain = %v, want Global{Bool}", pi4.A)
	}
}
