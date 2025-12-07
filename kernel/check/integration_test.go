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

// TestEndToEnd_List tests a List inductive with two constructors.
// List A : Type with nil : List A, cons : A -> List A -> List A
func TestEndToEnd_List(t *testing.T) {
	eval.ClearRecursorRegistry()

	env := NewGlobalEnvWithPrimitives()

	// For simplicity, we use a monomorphic List (no type parameter A)
	// List : Type with nil : List, cons : Nat -> List -> List
	err := env.DeclareInductive("List", ast.Sort{U: 0}, []Constructor{
		{Name: "nil", Type: ast.Global{Name: "List"}},
		{Name: "cons", Type: ast.Pi{
			Binder: "x",
			A:      ast.Global{Name: "Nat"},
			B: ast.Pi{
				Binder: "xs",
				A:      ast.Global{Name: "List"},
				B:      ast.Global{Name: "List"},
			},
		}},
	}, "listElim")
	if err != nil {
		t.Fatalf("DeclareInductive(List) failed: %v", err)
	}

	// Verify recursor info
	info := eval.LookupRecursor("listElim")
	if info == nil {
		t.Fatal("listElim not registered")
	}
	if info.NumCases != 2 {
		t.Errorf("listElim NumCases = %d, want 2", info.NumCases)
	}
	// cons has 2 args: x:Nat (non-recursive), xs:List (recursive at index 1)
	if len(info.Ctors[1].RecursiveIdx) != 1 || info.Ctors[1].RecursiveIdx[0] != 1 {
		t.Errorf("cons recursive indices = %v, want [1]", info.Ctors[1].RecursiveIdx)
	}

	// Test nil case: listElim P pnil pcons nil --> pnil
	listElim := ast.Global{Name: "listElim"}
	motive := ast.Lam{Binder: "_", Body: ast.Sort{U: 0}}
	pnil := ast.Global{Name: "nilCase"}
	// pcons takes x, xs, ih and returns something
	pcons := ast.Lam{Binder: "x", Body: ast.Lam{Binder: "xs", Body: ast.Lam{Binder: "ih", Body: ast.Global{Name: "consCase"}}}}
	nil_ := ast.Global{Name: "nil"}

	term := ast.MkApps(listElim, motive, pnil, pcons, nil_)
	normalized := eval.NormalizeNBE(term)

	if normalized != "nilCase" {
		t.Errorf("listElim _ nilCase _ nil = %q, want 'nilCase'", normalized)
	}

	// Test cons case: listElim P pnil pcons (cons x nil) --> pcons x nil (listElim P pnil pcons nil)
	// Since pcons = λx.λxs.λih. consCase, this reduces to consCase
	zero := ast.Global{Name: "zero"}
	oneElem := ast.App{T: ast.App{T: ast.Global{Name: "cons"}, U: zero}, U: nil_}
	term = ast.MkApps(listElim, motive, pnil, pcons, oneElem)
	normalized = eval.NormalizeNBE(term)

	if normalized != "consCase" {
		t.Errorf("listElim _ _ _ (cons zero nil) = %q, want 'consCase'", normalized)
	}

	eval.ClearRecursorRegistry()
}

// TestEndToEnd_Tree tests a Tree inductive with nested List usage.
// Tree : Type with leaf : Nat -> Tree, node : List Tree -> Tree
func TestEndToEnd_Tree(t *testing.T) {
	eval.ClearRecursorRegistry()

	env := NewGlobalEnvWithPrimitives()

	// First declare List of Tree: we need List in env
	// For this test, we use a simple Tree without nested List since we don't have
	// parameterized types. Instead: Tree : Type, leaf : Tree, branch : Tree -> Tree -> Tree
	err := env.DeclareInductive("Tree", ast.Sort{U: 0}, []Constructor{
		{Name: "leaf", Type: ast.Global{Name: "Tree"}},
		{Name: "branch", Type: ast.Pi{
			Binder: "l",
			A:      ast.Global{Name: "Tree"},
			B: ast.Pi{
				Binder: "r",
				A:      ast.Global{Name: "Tree"},
				B:      ast.Global{Name: "Tree"},
			},
		}},
	}, "treeElim")
	if err != nil {
		t.Fatalf("DeclareInductive(Tree) failed: %v", err)
	}

	// Verify recursor info
	info := eval.LookupRecursor("treeElim")
	if info == nil {
		t.Fatal("treeElim not registered")
	}
	if info.NumCases != 2 {
		t.Errorf("treeElim NumCases = %d, want 2", info.NumCases)
	}
	// branch has 2 recursive args at indices 0 and 1
	if len(info.Ctors[1].RecursiveIdx) != 2 {
		t.Errorf("branch recursive indices = %v, want 2 indices", info.Ctors[1].RecursiveIdx)
	}

	// Test leaf case: treeElim P pleaf pbranch leaf --> pleaf
	treeElim := ast.Global{Name: "treeElim"}
	motive := ast.Lam{Binder: "_", Body: ast.Sort{U: 0}}
	pleaf := ast.Global{Name: "leafCase"}
	// pbranch takes l, ihl, r, ihr and returns something
	pbranch := ast.Lam{Binder: "l", Body: ast.Lam{Binder: "ihl", Body: ast.Lam{Binder: "r", Body: ast.Lam{Binder: "ihr", Body: ast.Global{Name: "branchCase"}}}}}
	leaf := ast.Global{Name: "leaf"}

	term := ast.MkApps(treeElim, motive, pleaf, pbranch, leaf)
	normalized := eval.NormalizeNBE(term)

	if normalized != "leafCase" {
		t.Errorf("treeElim _ leafCase _ leaf = %q, want 'leafCase'", normalized)
	}

	// Test branch case: treeElim P pleaf pbranch (branch leaf leaf)
	// --> pbranch leaf (treeElim P pleaf pbranch leaf) leaf (treeElim P pleaf pbranch leaf)
	// --> pbranch leaf leafCase leaf leafCase --> branchCase
	twoLeaves := ast.App{T: ast.App{T: ast.Global{Name: "branch"}, U: leaf}, U: leaf}
	term = ast.MkApps(treeElim, motive, pleaf, pbranch, twoLeaves)
	normalized = eval.NormalizeNBE(term)

	if normalized != "branchCase" {
		t.Errorf("treeElim _ _ _ (branch leaf leaf) = %q, want 'branchCase'", normalized)
	}

	eval.ClearRecursorRegistry()
}

// TestEndToEnd_NestedRecursion tests nested recursive calls.
// Uses MyNat to test succ (succ (succ mzero)) reduction.
func TestEndToEnd_NestedRecursion(t *testing.T) {
	eval.ClearRecursorRegistry()

	env := NewGlobalEnv()

	// Declare MyNat
	err := env.DeclareInductive("MyNat", ast.Sort{U: 0}, []Constructor{
		{Name: "mzero", Type: ast.Global{Name: "MyNat"}},
		{Name: "msucc", Type: ast.Pi{
			Binder: "n",
			A:      ast.Global{Name: "MyNat"},
			B:      ast.Global{Name: "MyNat"},
		}},
	}, "myNatElim")
	if err != nil {
		t.Fatalf("DeclareInductive(MyNat) failed: %v", err)
	}

	// Build msucc (msucc (msucc mzero)) = 3
	mzero := ast.Global{Name: "mzero"}
	one := ast.App{T: ast.Global{Name: "msucc"}, U: mzero}
	two := ast.App{T: ast.Global{Name: "msucc"}, U: one}
	three := ast.App{T: ast.Global{Name: "msucc"}, U: two}

	// Test that we can reduce recursively
	// myNatElim P pz ps three should reduce to ps two (ps one (ps zero pz))
	// With ps = λn.λih. succCase, this fully reduces to succCase
	myNatElim := ast.Global{Name: "myNatElim"}
	motive := ast.Lam{Binder: "_", Body: ast.Sort{U: 0}}
	pz := ast.Global{Name: "zeroResult"}
	ps := ast.Lam{Binder: "n", Body: ast.Lam{Binder: "ih", Body: ast.Global{Name: "succResult"}}}

	term := ast.MkApps(myNatElim, motive, pz, ps, three)
	normalized := eval.NormalizeNBE(term)

	if normalized != "succResult" {
		t.Errorf("myNatElim _ _ _ (msucc (msucc (msucc mzero))) = %q, want 'succResult'", normalized)
	}

	eval.ClearRecursorRegistry()
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

// TestEndToEnd_ParameterizedList tests parameterized inductive types.
// List : Type -> Type with nil and cons constructors.
func TestEndToEnd_ParameterizedList(t *testing.T) {
	eval.ClearRecursorRegistry()
	env := NewGlobalEnv()

	// List : Type -> Type
	listType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B:      ast.Sort{U: 0},
	}

	// nil : (A : Type) -> List A
	nilType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B:      ast.App{T: ast.Global{Name: "List"}, U: ast.Var{Ix: 0}},
	}

	// cons : (A : Type) -> A -> List A -> List A
	consType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "x",
			A:      ast.Var{Ix: 0}, // A
			B: ast.Pi{
				Binder: "xs",
				A:      ast.App{T: ast.Global{Name: "List"}, U: ast.Var{Ix: 1}}, // List A
				B:      ast.App{T: ast.Global{Name: "List"}, U: ast.Var{Ix: 2}}, // List A
			},
		},
	}

	err := env.DeclareInductive("List", listType, []Constructor{
		{Name: "nil", Type: nilType},
		{Name: "cons", Type: consType},
	}, "listElim")
	if err != nil {
		t.Fatalf("DeclareInductive(List) failed: %v", err)
	}

	// Verify NumParams was extracted
	ind := env.inductives["List"]
	if ind.NumParams != 1 {
		t.Errorf("List.NumParams = %d, want 1", ind.NumParams)
	}

	// Verify RecursorInfo
	info := eval.LookupRecursor("listElim")
	if info == nil {
		t.Fatal("listElim not registered")
	}
	if info.NumParams != 1 {
		t.Errorf("RecursorInfo.NumParams = %d, want 1", info.NumParams)
	}
	if info.NumCases != 2 {
		t.Errorf("RecursorInfo.NumCases = %d, want 2", info.NumCases)
	}

	// nil has 0 data args (1 param skipped)
	if info.Ctors[0].NumArgs != 0 {
		t.Errorf("nil.NumArgs = %d, want 0", info.Ctors[0].NumArgs)
	}

	// cons has 2 data args (x and xs, with 1 param skipped)
	if info.Ctors[1].NumArgs != 2 {
		t.Errorf("cons.NumArgs = %d, want 2", info.Ctors[1].NumArgs)
	}

	// cons.xs is recursive (index 1)
	if len(info.Ctors[1].RecursiveIdx) != 1 || info.Ctors[1].RecursiveIdx[0] != 1 {
		t.Errorf("cons.RecursiveIdx = %v, want [1]", info.Ctors[1].RecursiveIdx)
	}

	// Verify eliminator type structure
	// listElim : (A : Type) -> (P : List A -> Type) -> P (nil A) -> (...) -> (xs : List A) -> P xs
	elimType := env.LookupType("listElim")
	if elimType == nil {
		t.Fatal("listElim type not found")
	}

	// First binder should be parameter A
	pi1, ok := elimType.(ast.Pi)
	if !ok {
		t.Fatalf("listElim level 1: expected Pi, got %T", elimType)
	}
	if pi1.Binder != "A" {
		t.Errorf("listElim binder 1 = %q, want 'A'", pi1.Binder)
	}

	// Second binder should be motive P
	pi2, ok := pi1.B.(ast.Pi)
	if !ok {
		t.Fatalf("listElim level 2: expected Pi, got %T", pi1.B)
	}
	if pi2.Binder != "P" {
		t.Errorf("listElim binder 2 = %q, want 'P'", pi2.Binder)
	}

	eval.ClearRecursorRegistry()
}

// TestEndToEnd_IndexedVec tests indexed inductive types.
// Vec : Type -> Nat -> Type with vnil and vcons constructors.
func TestEndToEnd_IndexedVec(t *testing.T) {
	eval.ClearRecursorRegistry()
	env := NewGlobalEnvWithPrimitives()

	// Vec : Type -> Nat -> Type
	// (A : Type) -> (n : Nat) -> Type
	vecType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "n",
			A:      ast.Global{Name: "Nat"},
			B:      ast.Sort{U: 0},
		},
	}

	// vnil : (A : Type) -> Vec A zero
	vnilType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.App{
			T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 0}},
			U: ast.Global{Name: "zero"},
		},
	}

	// vcons : (A : Type) -> (n : Nat) -> A -> Vec A n -> Vec A (succ n)
	vconsType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "n",
			A:      ast.Global{Name: "Nat"},
			B: ast.Pi{
				Binder: "x",
				A:      ast.Var{Ix: 1}, // A
				B: ast.Pi{
					Binder: "xs",
					A: ast.App{
						T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 2}}, // Vec A
						U: ast.Var{Ix: 1},                                         // n
					},
					B: ast.App{
						T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 3}},  // Vec A
						U: ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 2}}, // succ n
					},
				},
			},
		},
	}

	err := env.DeclareInductive("Vec", vecType, []Constructor{
		{Name: "vnil", Type: vnilType},
		{Name: "vcons", Type: vconsType},
	}, "vecElim")
	if err != nil {
		t.Fatalf("DeclareInductive(Vec) failed: %v", err)
	}

	// Verify NumParams and NumIndices were correctly extracted
	ind := env.inductives["Vec"]
	if ind.NumParams != 1 {
		t.Errorf("Vec.NumParams = %d, want 1", ind.NumParams)
	}
	if ind.NumIndices != 1 {
		t.Errorf("Vec.NumIndices = %d, want 1", ind.NumIndices)
	}

	// Verify RecursorInfo
	info := eval.LookupRecursor("vecElim")
	if info == nil {
		t.Fatal("vecElim not registered")
	}
	if info.NumParams != 1 {
		t.Errorf("RecursorInfo.NumParams = %d, want 1", info.NumParams)
	}
	if info.NumIndices != 1 {
		t.Errorf("RecursorInfo.NumIndices = %d, want 1", info.NumIndices)
	}
	if info.NumCases != 2 {
		t.Errorf("RecursorInfo.NumCases = %d, want 2", info.NumCases)
	}

	// vnil has 0 data args (1 param skipped)
	if info.Ctors[0].NumArgs != 0 {
		t.Errorf("vnil.NumArgs = %d, want 0", info.Ctors[0].NumArgs)
	}

	// vcons has 3 data args: n, x, xs (1 param skipped)
	if info.Ctors[1].NumArgs != 3 {
		t.Errorf("vcons.NumArgs = %d, want 3", info.Ctors[1].NumArgs)
	}

	// Verify eliminator type structure
	// vecElim : (A : Type) -> (P : (n : Nat) -> Vec A n -> Type) -> ...
	elimType := env.LookupType("vecElim")
	if elimType == nil {
		t.Fatal("vecElim type not found")
	}

	// First binder should be parameter A
	pi1, ok := elimType.(ast.Pi)
	if !ok {
		t.Fatalf("vecElim level 1: expected Pi, got %T", elimType)
	}
	if pi1.Binder != "A" {
		t.Errorf("vecElim binder 1 = %q, want 'A'", pi1.Binder)
	}

	// Second binder should be motive P
	pi2, ok := pi1.B.(ast.Pi)
	if !ok {
		t.Fatalf("vecElim level 2: expected Pi, got %T", pi1.B)
	}
	if pi2.Binder != "P" {
		t.Errorf("vecElim binder 2 = %q, want 'P'", pi2.Binder)
	}

	// The motive should be: (n : Nat) -> Vec A n -> Type
	// which is a Pi type
	motiveType := pi2.A
	pi_m1, ok := motiveType.(ast.Pi)
	if !ok {
		t.Fatalf("vecElim motive: expected Pi (for index), got %T", motiveType)
	}
	if pi_m1.Binder != "n" {
		t.Errorf("vecElim motive index binder = %q, want 'n'", pi_m1.Binder)
	}

	eval.ClearRecursorRegistry()
}

// TestEndToEnd_IndexedVecReduction tests that vecElim reduces correctly.
func TestEndToEnd_IndexedVecReduction(t *testing.T) {
	eval.ClearRecursorRegistry()
	env := NewGlobalEnvWithPrimitives()

	// Vec : Type -> Nat -> Type
	vecType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "n",
			A:      ast.Global{Name: "Nat"},
			B:      ast.Sort{U: 0},
		},
	}

	// vnil : (A : Type) -> Vec A zero
	vnilType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.App{
			T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 0}},
			U: ast.Global{Name: "zero"},
		},
	}

	// vcons : (A : Type) -> (n : Nat) -> A -> Vec A n -> Vec A (succ n)
	vconsType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "n",
			A:      ast.Global{Name: "Nat"},
			B: ast.Pi{
				Binder: "x",
				A:      ast.Var{Ix: 1},
				B: ast.Pi{
					Binder: "xs",
					A: ast.App{
						T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 2}},
						U: ast.Var{Ix: 1},
					},
					B: ast.App{
						T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 3}},
						U: ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 2}},
					},
				},
			},
		},
	}

	err := env.DeclareInductive("Vec", vecType, []Constructor{
		{Name: "vnil", Type: vnilType},
		{Name: "vcons", Type: vconsType},
	}, "vecElim")
	if err != nil {
		t.Fatalf("DeclareInductive(Vec) failed: %v", err)
	}

	// Build: vecElim Nat P pvnil pvcons zero (vnil Nat)
	// Should reduce to pvnil
	nat := ast.Global{Name: "Nat"}
	vecElim := ast.Global{Name: "vecElim"}
	motive := ast.Lam{Binder: "n", Body: ast.Lam{Binder: "_", Body: ast.Sort{U: 0}}}
	pvnil := ast.Global{Name: "vnilResult"}
	pvcons := ast.Lam{
		Binder: "n",
		Body: ast.Lam{
			Binder: "x",
			Body: ast.Lam{
				Binder: "xs",
				Body: ast.Lam{
					Binder: "ih",
					Body:   ast.Global{Name: "vconsResult"},
				},
			},
		},
	}
	zero := ast.Global{Name: "zero"}
	vnilNat := ast.App{T: ast.Global{Name: "vnil"}, U: nat}

	// vecElim Nat P pvnil pvcons zero (vnil Nat)
	term := ast.MkApps(vecElim, nat, motive, pvnil, pvcons, zero, vnilNat)
	normalized := eval.NormalizeNBE(term)

	if normalized != "vnilResult" {
		t.Errorf("vecElim Nat _ vnilResult _ zero (vnil Nat) = %q, want 'vnilResult'", normalized)
	}

	// Test vcons reduction:
	// vecElim Nat P pvnil pvcons (succ zero) (vcons Nat zero x (vnil Nat))
	// Should reduce to: pvcons zero x (vnil Nat) (vecElim Nat P pvnil pvcons zero (vnil Nat))
	// With our pvcons = λn.λx.λxs.λih. vconsResult, this becomes vconsResult
	succZero := ast.App{T: ast.Global{Name: "succ"}, U: zero}
	x := ast.Global{Name: "someElement"}
	// vcons Nat zero x (vnil Nat)
	vconsApp := ast.MkApps(ast.Global{Name: "vcons"}, nat, zero, x, vnilNat)

	term2 := ast.MkApps(vecElim, nat, motive, pvnil, pvcons, succZero, vconsApp)
	normalized2 := eval.NormalizeNBE(term2)

	if normalized2 != "vconsResult" {
		t.Errorf("vecElim Nat _ _ _ (succ zero) (vcons Nat zero x (vnil Nat)) = %q, want 'vconsResult'", normalized2)
	}

	eval.ClearRecursorRegistry()
}

// TestEndToEnd_IndexArgPositionsMetadata verifies that IndexArgPositions is computed correctly.
func TestEndToEnd_IndexArgPositionsMetadata(t *testing.T) {
	eval.ClearRecursorRegistry()
	env := NewGlobalEnvWithPrimitives()

	// Vec : Type -> Nat -> Type
	vecType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "n",
			A:      ast.Global{Name: "Nat"},
			B:      ast.Sort{U: 0},
		},
	}

	// vnil : (A : Type) -> Vec A zero
	vnilType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.App{
			T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 0}},
			U: ast.Global{Name: "zero"},
		},
	}

	// vcons : (A : Type) -> (n : Nat) -> A -> Vec A n -> Vec A (succ n)
	// Data args after param A: [n, x, xs] at positions [0, 1, 2]
	// xs (at position 2) has type Vec A n, where n is at position 0
	// So IndexArgPositions[2] should be [0]
	vconsType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "n",
			A:      ast.Global{Name: "Nat"},
			B: ast.Pi{
				Binder: "x",
				A:      ast.Var{Ix: 1},
				B: ast.Pi{
					Binder: "xs",
					A: ast.App{
						T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 2}},
						U: ast.Var{Ix: 1}, // n
					},
					B: ast.App{
						T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 3}},
						U: ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 2}},
					},
				},
			},
		},
	}

	err := env.DeclareInductive("Vec", vecType, []Constructor{
		{Name: "vnil", Type: vnilType},
		{Name: "vcons", Type: vconsType},
	}, "vecElim")
	if err != nil {
		t.Fatalf("DeclareInductive(Vec) failed: %v", err)
	}

	// Verify IndexArgPositions metadata
	info := eval.LookupRecursor("vecElim")
	if info == nil {
		t.Fatal("vecElim not registered")
	}

	// vnil has no recursive args, so IndexArgPositions should be empty/nil
	vnilCtor := info.Ctors[0]
	if len(vnilCtor.IndexArgPositions) != 0 {
		t.Errorf("vnil.IndexArgPositions = %v, want empty", vnilCtor.IndexArgPositions)
	}

	// vcons has recursive arg xs at position 2, with index n at position 0
	vconsCtor := info.Ctors[1]
	if vconsCtor.IndexArgPositions == nil {
		t.Fatal("vcons.IndexArgPositions should not be nil")
	}
	idxPos, ok := vconsCtor.IndexArgPositions[2]
	if !ok {
		t.Fatal("vcons.IndexArgPositions[2] not found")
	}
	if len(idxPos) != 1 || idxPos[0] != 0 {
		t.Errorf("vcons.IndexArgPositions[2] = %v, want [0]", idxPos)
	}

	eval.ClearRecursorRegistry()
}

// TestEndToEnd_NestedVecReduction tests Vec reduction with nested vcons (length 2 vector).
// This exercises the IH construction to ensure indices are extracted correctly.
func TestEndToEnd_NestedVecReduction(t *testing.T) {
	eval.ClearRecursorRegistry()
	env := NewGlobalEnvWithPrimitives()

	// Vec : Type -> Nat -> Type
	vecType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "n",
			A:      ast.Global{Name: "Nat"},
			B:      ast.Sort{U: 0},
		},
	}

	// vnil : (A : Type) -> Vec A zero
	vnilType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.App{
			T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 0}},
			U: ast.Global{Name: "zero"},
		},
	}

	// vcons : (A : Type) -> (n : Nat) -> A -> Vec A n -> Vec A (succ n)
	vconsType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "n",
			A:      ast.Global{Name: "Nat"},
			B: ast.Pi{
				Binder: "x",
				A:      ast.Var{Ix: 1},
				B: ast.Pi{
					Binder: "xs",
					A: ast.App{
						T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 2}},
						U: ast.Var{Ix: 1},
					},
					B: ast.App{
						T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 3}},
						U: ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 2}},
					},
				},
			},
		},
	}

	err := env.DeclareInductive("Vec", vecType, []Constructor{
		{Name: "vnil", Type: vnilType},
		{Name: "vcons", Type: vconsType},
	}, "vecElim")
	if err != nil {
		t.Fatalf("DeclareInductive(Vec) failed: %v", err)
	}

	// Build a Vec Nat 2 = vcons Nat 1 x (vcons Nat 0 y (vnil Nat))
	nat := ast.Global{Name: "Nat"}
	zero := ast.Global{Name: "zero"}
	one := ast.App{T: ast.Global{Name: "succ"}, U: zero}
	two := ast.App{T: ast.Global{Name: "succ"}, U: one}
	x := ast.Global{Name: "elem1"}
	y := ast.Global{Name: "elem2"}
	vnil := ast.App{T: ast.Global{Name: "vnil"}, U: nat}
	vcons1 := ast.MkApps(ast.Global{Name: "vcons"}, nat, zero, y, vnil)
	vcons2 := ast.MkApps(ast.Global{Name: "vcons"}, nat, one, x, vcons1)

	// Create a motive that ignores the index and vector
	// P : (n : Nat) -> Vec Nat n -> Type
	// P n v = Nat
	motive := ast.Lam{Binder: "n", Body: ast.Lam{Binder: "_", Body: nat}}

	// pvnil : P zero (vnil Nat) = Nat, we'll return zero
	pvnil := zero

	// pvcons : (n : Nat) -> (x : Nat) -> (xs : Vec Nat n) -> P n xs -> P (succ n) (vcons Nat n x xs)
	// Return: succ ih (count elements)
	pvcons := ast.Lam{
		Binder: "n",
		Body: ast.Lam{
			Binder: "x",
			Body: ast.Lam{
				Binder: "xs",
				Body: ast.Lam{
					Binder: "ih",
					Body:   ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 0}}, // succ ih
				},
			},
		},
	}

	// vecElim Nat P pvnil pvcons 2 (vcons Nat 1 x (vcons Nat 0 y (vnil Nat)))
	// Should compute: succ (succ zero) = 2
	vecElim := ast.Global{Name: "vecElim"}
	term := ast.MkApps(vecElim, nat, motive, pvnil, pvcons, two, vcons2)

	// Normalize and check the result
	normalized := eval.NormalizeNBE(term)

	// The result should be (succ (succ zero)) which normalizes to that form
	expected := "(succ (succ zero))"
	if normalized != expected {
		t.Errorf("vecElim counting length 2 vec = %q, want %q", normalized, expected)
	}

	eval.ClearRecursorRegistry()
}

// TestEndToEnd_ParameterizedListReduction tests that listElim reduces correctly.
func TestEndToEnd_ParameterizedListReduction(t *testing.T) {
	eval.ClearRecursorRegistry()
	env := NewGlobalEnvWithPrimitives()

	// Declare List
	listType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B:      ast.Sort{U: 0},
	}

	nilType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B:      ast.App{T: ast.Global{Name: "List"}, U: ast.Var{Ix: 0}},
	}

	consType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "x",
			A:      ast.Var{Ix: 0},
			B: ast.Pi{
				Binder: "xs",
				A:      ast.App{T: ast.Global{Name: "List"}, U: ast.Var{Ix: 1}},
				B:      ast.App{T: ast.Global{Name: "List"}, U: ast.Var{Ix: 2}},
			},
		},
	}

	err := env.DeclareInductive("List", listType, []Constructor{
		{Name: "nil", Type: nilType},
		{Name: "cons", Type: consType},
	}, "listElim")
	if err != nil {
		t.Fatalf("DeclareInductive(List) failed: %v", err)
	}

	// Build: listElim Nat P pnil pcons (nil Nat)
	// Should reduce to pnil
	nat := ast.Global{Name: "Nat"}
	listElim := ast.Global{Name: "listElim"}
	motive := ast.Lam{Binder: "_", Body: ast.Sort{U: 0}}
	pnil := ast.Global{Name: "nilResult"}
	pcons := ast.Lam{
		Binder: "x",
		Body: ast.Lam{
			Binder: "xs",
			Body: ast.Lam{
				Binder: "ih",
				Body:   ast.Global{Name: "consResult"},
			},
		},
	}
	nilNat := ast.App{T: ast.Global{Name: "nil"}, U: nat}

	term := ast.MkApps(listElim, nat, motive, pnil, pcons, nilNat)
	normalized := eval.NormalizeNBE(term)

	if normalized != "nilResult" {
		t.Errorf("listElim Nat _ nilResult _ (nil Nat) = %q, want 'nilResult'", normalized)
	}

	// Build: listElim Nat P pnil pcons (cons Nat zero (nil Nat))
	// Should reduce to pcons applied to zero, nil, and IH (pnil)
	// With pcons = λx.λxs.λih. consResult, this fully reduces to consResult
	zero := ast.Global{Name: "zero"}
	consNat := ast.MkApps(ast.Global{Name: "cons"}, nat, zero, nilNat)

	term2 := ast.MkApps(listElim, nat, motive, pnil, pcons, consNat)
	normalized2 := eval.NormalizeNBE(term2)

	if normalized2 != "consResult" {
		t.Errorf("listElim Nat _ _ _ (cons Nat zero (nil Nat)) = %q, want 'consResult'", normalized2)
	}

	eval.ClearRecursorRegistry()
}
