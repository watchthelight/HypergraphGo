package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func TestExtractPiArgs(t *testing.T) {
	tests := []struct {
		name     string
		ty       ast.Term
		expected int // number of args
	}{
		{
			name:     "No args (direct type)",
			ty:       ast.Global{Name: "Nat"},
			expected: 0,
		},
		{
			name: "One arg",
			ty: ast.Pi{
				Binder: "_",
				A:      ast.Global{Name: "Nat"},
				B:      ast.Global{Name: "Nat"},
			},
			expected: 1,
		},
		{
			name: "Two args",
			ty: ast.Pi{
				Binder: "x",
				A:      ast.Global{Name: "A"},
				B: ast.Pi{
					Binder: "xs",
					A:      ast.Global{Name: "List"},
					B:      ast.Global{Name: "List"},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := extractPiArgs(tt.ty)
			if len(args) != tt.expected {
				t.Errorf("extractPiArgs() got %d args, want %d", len(args), tt.expected)
			}
		})
	}
}

func TestCountRecursiveArgs(t *testing.T) {
	tests := []struct {
		name     string
		indName  string
		args     []PiArg
		expected int
	}{
		{
			name:     "No recursive args",
			indName:  "Nat",
			args:     []PiArg{},
			expected: 0,
		},
		{
			name:    "One recursive arg (Nat)",
			indName: "Nat",
			args: []PiArg{
				{Name: "n", Type: ast.Global{Name: "Nat"}},
			},
			expected: 1,
		},
		{
			name:    "Mixed args",
			indName: "List",
			args: []PiArg{
				{Name: "x", Type: ast.Var{Ix: 0}},            // A (non-recursive)
				{Name: "xs", Type: ast.Global{Name: "List"}}, // List (recursive)
			},
			expected: 1,
		},
		{
			name:    "No match",
			indName: "Nat",
			args: []PiArg{
				{Name: "x", Type: ast.Global{Name: "Bool"}},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countRecursiveArgs(tt.indName, tt.args)
			if result != tt.expected {
				t.Errorf("countRecursiveArgs() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestIsRecursiveArgType(t *testing.T) {
	tests := []struct {
		name     string
		indName  string
		ty       ast.Term
		expected bool
	}{
		{
			name:     "Direct match",
			indName:  "Nat",
			ty:       ast.Global{Name: "Nat"},
			expected: true,
		},
		{
			name:     "No match",
			indName:  "Nat",
			ty:       ast.Global{Name: "Bool"},
			expected: false,
		},
		{
			name:     "Applied type match",
			indName:  "List",
			ty:       ast.App{T: ast.Global{Name: "List"}, U: ast.Global{Name: "A"}},
			expected: true,
		},
		{
			name:     "Variable (non-recursive)",
			indName:  "Nat",
			ty:       ast.Var{Ix: 0},
			expected: false,
		},
		// Higher-order recursive detection tests
		{
			name:    "Higher-order: A -> T",
			indName: "Tree",
			ty: ast.Pi{
				Binder: "_",
				A:      ast.Global{Name: "A"},
				B:      ast.Global{Name: "Tree"},
			},
			expected: true,
		},
		{
			name:    "Higher-order: A -> B -> T",
			indName: "Tree",
			ty: ast.Pi{
				Binder: "_",
				A:      ast.Global{Name: "A"},
				B: ast.Pi{
					Binder: "_",
					A:      ast.Global{Name: "B"},
					B:      ast.Global{Name: "Tree"},
				},
			},
			expected: true,
		},
		{
			name:    "Higher-order: A -> List T (applied)",
			indName: "Tree",
			ty: ast.Pi{
				Binder: "_",
				A:      ast.Global{Name: "A"},
				B:      ast.App{T: ast.Global{Name: "List"}, U: ast.Global{Name: "Tree"}},
			},
			expected: true,
		},
		{
			name:    "Not higher-order: A -> B (no T in codomain)",
			indName: "Tree",
			ty: ast.Pi{
				Binder: "_",
				A:      ast.Global{Name: "A"},
				B:      ast.Global{Name: "B"},
			},
			expected: false,
		},
		{
			name:    "Not recursive even with T in domain",
			indName: "Tree",
			ty: ast.Pi{
				Binder: "_",
				A:      ast.Global{Name: "Tree"}, // T in domain, not codomain
				B:      ast.Global{Name: "B"},
			},
			expected: false, // Only codomain matters for isRecursiveArgType
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRecursiveArgType(tt.indName, tt.ty)
			if result != tt.expected {
				t.Errorf("isRecursiveArgType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateRecursorTypeSimple(t *testing.T) {
	// Test that we can generate recursor types for known inductives
	tests := []struct {
		name string
		ind  *Inductive
	}{
		{
			name: "Nat",
			ind: &Inductive{
				Name: "Nat",
				Type: ast.Sort{U: 0},
				Constructors: []Constructor{
					{Name: "zero", Type: ast.Global{Name: "Nat"}},
					{Name: "succ", Type: ast.Pi{
						Binder: "_",
						A:      ast.Global{Name: "Nat"},
						B:      ast.Global{Name: "Nat"},
					}},
				},
				Eliminator: "natElim",
			},
		},
		{
			name: "Bool",
			ind: &Inductive{
				Name: "Bool",
				Type: ast.Sort{U: 0},
				Constructors: []Constructor{
					{Name: "true", Type: ast.Global{Name: "Bool"}},
					{Name: "false", Type: ast.Global{Name: "Bool"}},
				},
				Eliminator: "boolElim",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRecursorTypeSimple(tt.ind)
			if result == nil {
				t.Error("GenerateRecursorTypeSimple() returned nil")
			}
			// The result should be a Pi type (outermost binder is the motive)
			if _, ok := result.(ast.Pi); !ok {
				t.Errorf("GenerateRecursorTypeSimple() expected Pi type, got %T", result)
			}
		})
	}
}

func TestGenerateRecursorType_Unit(t *testing.T) {
	// Unit type has a very simple eliminator
	// unitElim : (P : Unit -> Type) -> P tt -> (u : Unit) -> P u
	unit := &Inductive{
		Name: "Unit",
		Type: ast.Sort{U: 0},
		Constructors: []Constructor{
			{Name: "tt", Type: ast.Global{Name: "Unit"}},
		},
		Eliminator: "unitElim",
	}

	result := GenerateRecursorType(unit)
	if result == nil {
		t.Error("GenerateRecursorType() returned nil")
	}

	// Check structure: Pi P . Pi case_tt . Pi u . P u
	pi1, ok := result.(ast.Pi)
	if !ok {
		t.Fatalf("Expected outer Pi, got %T", result)
	}
	if pi1.Binder != "P" {
		t.Errorf("Expected binder 'P', got %q", pi1.Binder)
	}

	pi2, ok := pi1.B.(ast.Pi)
	if !ok {
		t.Fatalf("Expected second Pi, got %T", pi1.B)
	}
	if pi2.Binder != "case_tt" {
		t.Errorf("Expected binder 'case_tt', got %q", pi2.Binder)
	}

	pi3, ok := pi2.B.(ast.Pi)
	if !ok {
		t.Fatalf("Expected third Pi, got %T", pi2.B)
	}
	if pi3.Binder != "t" {
		t.Errorf("Expected binder 't', got %q", pi3.Binder)
	}
}

func TestGenerateRecursorType_Empty(t *testing.T) {
	// Empty type (no constructors) has eliminator:
	// emptyElim : (P : Empty -> Type) -> (e : Empty) -> P e
	empty := &Inductive{
		Name:         "Empty",
		Type:         ast.Sort{U: 0},
		Constructors: []Constructor{},
		Eliminator:   "emptyElim",
	}

	result := GenerateRecursorType(empty)
	if result == nil {
		t.Error("GenerateRecursorType() returned nil")
	}

	// Check structure: Pi P . Pi e . P e
	pi1, ok := result.(ast.Pi)
	if !ok {
		t.Fatalf("Expected outer Pi, got %T", result)
	}
	if pi1.Binder != "P" {
		t.Errorf("Expected binder 'P', got %q", pi1.Binder)
	}

	pi2, ok := pi1.B.(ast.Pi)
	if !ok {
		t.Fatalf("Expected target Pi, got %T", pi1.B)
	}
	if pi2.Binder != "t" {
		t.Errorf("Expected binder 't', got %q", pi2.Binder)
	}
}

// TestBuildCaseType_Nat verifies buildCaseType for Nat's succ constructor.
// succ : Nat -> Nat
// case_succ : (n : Nat) -> P n -> P (succ n)
func TestBuildCaseType_Nat(t *testing.T) {
	indName := "Nat"
	args := []PiArg{
		{Name: "n", Type: ast.Global{Name: "Nat"}},
	}
	numRecursive := 1
	pBaseIdx := 1 // P is one case after this

	result := buildCaseType(indName, args, numRecursive, pBaseIdx, "succ", 0)

	// Expected: (n : Nat) -> (ih_n : P n) -> P (succ n)
	pi1, ok := result.(ast.Pi)
	if !ok {
		t.Fatalf("Expected outer Pi, got %T", result)
	}
	if pi1.Binder != "n" {
		t.Errorf("Expected binder 'n', got %q", pi1.Binder)
	}

	pi2, ok := pi1.B.(ast.Pi)
	if !ok {
		t.Fatalf("Expected IH Pi, got %T", pi1.B)
	}
	if pi2.Binder != "ih_n" {
		t.Errorf("Expected binder 'ih_n', got %q", pi2.Binder)
	}

	// Result type should be App{P, App{succ, n}}
	resultApp, ok := pi2.B.(ast.App)
	if !ok {
		t.Fatalf("Expected App for result, got %T", pi2.B)
	}
	// Check that result is P applied to something
	if pVar, ok := resultApp.T.(ast.Var); !ok || pVar.Ix != 3 {
		t.Errorf("Expected P (Var 3), got %v", resultApp.T)
	}
}

// TestBuildCaseType_List verifies buildCaseType for List's cons constructor.
// cons : Nat -> List -> List (monomorphic list)
// case_cons : (x : Nat) -> (xs : List) -> P xs -> P (cons x xs)
func TestBuildCaseType_List(t *testing.T) {
	indName := "List"
	args := []PiArg{
		{Name: "x", Type: ast.Global{Name: "Nat"}},   // non-recursive
		{Name: "xs", Type: ast.Global{Name: "List"}}, // recursive
	}
	numRecursive := 1
	pBaseIdx := 1

	result := buildCaseType(indName, args, numRecursive, pBaseIdx, "cons", 0)

	// Expected: (x : Nat) -> (xs : List) -> (ih_xs : P xs) -> P (cons x xs)
	// Structure: Pi x . Pi xs . Pi ih_xs . P (cons x xs)
	pi1, ok := result.(ast.Pi)
	if !ok {
		t.Fatalf("Expected outer Pi for x, got %T", result)
	}
	if pi1.Binder != "x" {
		t.Errorf("Expected binder 'x', got %q", pi1.Binder)
	}

	pi2, ok := pi1.B.(ast.Pi)
	if !ok {
		t.Fatalf("Expected Pi for xs, got %T", pi1.B)
	}
	if pi2.Binder != "xs" {
		t.Errorf("Expected binder 'xs', got %q", pi2.Binder)
	}

	pi3, ok := pi2.B.(ast.Pi)
	if !ok {
		t.Fatalf("Expected Pi for ih_xs, got %T", pi2.B)
	}
	if pi3.Binder != "ih_xs" {
		t.Errorf("Expected binder 'ih_xs', got %q", pi3.Binder)
	}

	// Verify IH type references xs (Var 0 at the ih binder)
	ihApp, ok := pi3.A.(ast.App)
	if !ok {
		t.Fatalf("Expected IH type to be App, got %T", pi3.A)
	}
	if xsRef, ok := ihApp.U.(ast.Var); !ok || xsRef.Ix != 0 {
		t.Errorf("Expected IH argument to be xs (Var 0), got %v", ihApp.U)
	}
}

// TestBuildCaseType_Tree verifies buildCaseType for a binary Tree's branch constructor.
// branch : Tree -> Tree -> Tree
// case_branch : (l : Tree) -> P l -> (r : Tree) -> P r -> P (branch l r)
func TestBuildCaseType_Tree(t *testing.T) {
	indName := "Tree"
	args := []PiArg{
		{Name: "l", Type: ast.Global{Name: "Tree"}}, // recursive
		{Name: "r", Type: ast.Global{Name: "Tree"}}, // recursive
	}
	numRecursive := 2
	pBaseIdx := 1

	result := buildCaseType(indName, args, numRecursive, pBaseIdx, "branch", 0)

	// Expected: (l : Tree) -> (ih_l : P l) -> (r : Tree) -> (ih_r : P r) -> P (branch l r)
	// Count the Pi levels: should be 4 binders + result
	count := 0
	current := result
	binders := []string{}
	for {
		if pi, ok := current.(ast.Pi); ok {
			count++
			binders = append(binders, pi.Binder)
			current = pi.B
		} else {
			break
		}
	}

	if count != 4 {
		t.Errorf("Expected 4 Pi binders, got %d: %v", count, binders)
	}

	expectedBinders := []string{"l", "ih_l", "r", "ih_r"}
	for i, expected := range expectedBinders {
		if i < len(binders) && binders[i] != expected {
			t.Errorf("Binder %d: expected %q, got %q", i, expected, binders[i])
		}
	}
}

// ============================================================================
// extractUniverseLevel Tests
// ============================================================================

func TestExtractUniverseLevel_DirectSort(t *testing.T) {
	ty := ast.Sort{U: 5}
	level := extractUniverseLevel(ty)
	if level != 5 {
		t.Errorf("extractUniverseLevel(Sort{5}) = %d, want 5", level)
	}
}

func TestExtractUniverseLevel_SinglePi(t *testing.T) {
	// (A : Type) -> Type_3
	ty := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B:      ast.Sort{U: 3},
	}
	level := extractUniverseLevel(ty)
	if level != 3 {
		t.Errorf("extractUniverseLevel expected 3, got %d", level)
	}
}

func TestExtractUniverseLevel_NestedPi(t *testing.T) {
	// (A : Type) -> (B : Type) -> Type_7
	ty := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "B",
			A:      ast.Sort{U: 0},
			B:      ast.Sort{U: 7},
		},
	}
	level := extractUniverseLevel(ty)
	if level != 7 {
		t.Errorf("extractUniverseLevel expected 7, got %d", level)
	}
}

func TestExtractUniverseLevel_DeeplyNestedPi(t *testing.T) {
	// (A : Type) -> (B : Type) -> (C : Type) -> Type_10
	ty := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "B",
			A:      ast.Sort{U: 0},
			B: ast.Pi{
				Binder: "C",
				A:      ast.Sort{U: 0},
				B:      ast.Sort{U: 10},
			},
		},
	}
	level := extractUniverseLevel(ty)
	if level != 10 {
		t.Errorf("extractUniverseLevel expected 10, got %d", level)
	}
}

func TestExtractUniverseLevel_Fallback(t *testing.T) {
	// Unexpected type should return 0
	ty := ast.Global{Name: "Nat"}
	level := extractUniverseLevel(ty)
	if level != 0 {
		t.Errorf("extractUniverseLevel fallback expected 0, got %d", level)
	}
}

// ============================================================================
// isRecursiveArgTypeMulti Tests
// ============================================================================

func TestIsRecursiveArgTypeMulti_SingleName_Match(t *testing.T) {
	names := []string{"Nat"}
	ty := ast.Global{Name: "Nat"}
	if !isRecursiveArgTypeMulti(names, ty) {
		t.Error("expected true for direct match")
	}
}

func TestIsRecursiveArgTypeMulti_MultipleNames_FirstMatch(t *testing.T) {
	names := []string{"Even", "Odd"}
	ty := ast.Global{Name: "Even"}
	if !isRecursiveArgTypeMulti(names, ty) {
		t.Error("expected true for first name match")
	}
}

func TestIsRecursiveArgTypeMulti_MultipleNames_SecondMatch(t *testing.T) {
	names := []string{"Even", "Odd"}
	ty := ast.Global{Name: "Odd"}
	if !isRecursiveArgTypeMulti(names, ty) {
		t.Error("expected true for second name match")
	}
}

func TestIsRecursiveArgTypeMulti_NoMatch(t *testing.T) {
	names := []string{"Even", "Odd"}
	ty := ast.Global{Name: "Nat"}
	if isRecursiveArgTypeMulti(names, ty) {
		t.Error("expected false for no match")
	}
}

func TestIsRecursiveArgTypeMulti_Applied_Match(t *testing.T) {
	// List A where List is in our names
	names := []string{"List"}
	ty := ast.App{T: ast.Global{Name: "List"}, U: ast.Global{Name: "A"}}
	if !isRecursiveArgTypeMulti(names, ty) {
		t.Error("expected true for applied type match")
	}
}

func TestIsRecursiveArgTypeMulti_Applied_NoMatch(t *testing.T) {
	// Option A where Option is not in our names
	names := []string{"List"}
	ty := ast.App{T: ast.Global{Name: "Option"}, U: ast.Global{Name: "A"}}
	if isRecursiveArgTypeMulti(names, ty) {
		t.Error("expected false for applied type no match")
	}
}

func TestIsRecursiveArgTypeMulti_HigherOrder_Match(t *testing.T) {
	// (A -> Even) where Even is in our mutual names
	names := []string{"Even", "Odd"}
	ty := ast.Pi{
		Binder: "_",
		A:      ast.Global{Name: "Nat"},
		B:      ast.Global{Name: "Even"},
	}
	if !isRecursiveArgTypeMulti(names, ty) {
		t.Error("expected true for higher-order recursive match")
	}
}

func TestIsRecursiveArgTypeMulti_HigherOrder_DeepMatch(t *testing.T) {
	// (A -> List Odd) where Odd is nested
	names := []string{"Even", "Odd"}
	ty := ast.Pi{
		Binder: "_",
		A:      ast.Global{Name: "Nat"},
		B:      ast.App{T: ast.Global{Name: "List"}, U: ast.Global{Name: "Odd"}},
	}
	if !isRecursiveArgTypeMulti(names, ty) {
		t.Error("expected true for deeply nested higher-order match")
	}
}

func TestIsRecursiveArgTypeMulti_HigherOrder_NoMatch(t *testing.T) {
	// (A -> B) where neither A nor B is in our names
	names := []string{"Even", "Odd"}
	ty := ast.Pi{
		Binder: "_",
		A:      ast.Global{Name: "A"},
		B:      ast.Global{Name: "B"},
	}
	if isRecursiveArgTypeMulti(names, ty) {
		t.Error("expected false for higher-order no match")
	}
}

func TestIsRecursiveArgTypeMulti_Var(t *testing.T) {
	// Variable should not match
	names := []string{"Nat"}
	ty := ast.Var{Ix: 0}
	if isRecursiveArgTypeMulti(names, ty) {
		t.Error("expected false for Var")
	}
}

// ============================================================================
// buildAppliedInductiveFull Tests
// ============================================================================

func TestBuildAppliedInductiveFull_NoParams_NoIndices(t *testing.T) {
	result := buildAppliedInductiveFull("Nat", 0, 0, 1)
	global, ok := result.(ast.Global)
	if !ok {
		t.Fatalf("expected Global, got %T", result)
	}
	if global.Name != "Nat" {
		t.Errorf("expected Nat, got %s", global.Name)
	}
}

func TestBuildAppliedInductiveFull_OneParam_NoIndices(t *testing.T) {
	// List A: numParams=1, numIndices=0, numCases=2
	result := buildAppliedInductiveFull("List", 1, 0, 2)
	// Should be: List param0
	app, ok := result.(ast.App)
	if !ok {
		t.Fatalf("expected App, got %T", result)
	}
	global, ok := app.T.(ast.Global)
	if !ok || global.Name != "List" {
		t.Errorf("expected List, got %v", app.T)
	}
	// param_0 is at: numCases + 1 + numIndices + (numParams - 0 - 1) = 2 + 1 + 0 + 0 = 3
	varArg, ok := app.U.(ast.Var)
	if !ok {
		t.Fatalf("expected Var for param, got %T", app.U)
	}
	if varArg.Ix != 3 {
		t.Errorf("expected param at index 3, got %d", varArg.Ix)
	}
}

func TestBuildAppliedInductiveFull_OneParam_OneIndex(t *testing.T) {
	// Vec A n: numParams=1, numIndices=1, numCases=2
	result := buildAppliedInductiveFull("Vec", 1, 1, 2)
	// Should be: Vec param0 idx0 = App(App(Vec, param0), idx0)
	app1, ok := result.(ast.App)
	if !ok {
		t.Fatalf("expected outer App, got %T", result)
	}
	// idx_0 is at: numIndices - 0 - 1 = 0
	idxVar, ok := app1.U.(ast.Var)
	if !ok || idxVar.Ix != 0 {
		t.Errorf("expected idx at index 0, got %v", app1.U)
	}

	app2, ok := app1.T.(ast.App)
	if !ok {
		t.Fatalf("expected inner App, got %T", app1.T)
	}
	// param_0 is at: numCases + 1 + numIndices + (numParams - 0 - 1) = 2 + 1 + 1 + 0 = 4
	paramVar, ok := app2.U.(ast.Var)
	if !ok || paramVar.Ix != 4 {
		t.Errorf("expected param at index 4, got %v", app2.U)
	}
}

func TestBuildAppliedInductiveFull_TwoParams_TwoIndices(t *testing.T) {
	// DepPair A B n m: numParams=2, numIndices=2, numCases=1
	result := buildAppliedInductiveFull("DepPair", 2, 2, 1)

	// Extract all applications
	apps := []ast.Var{}
	current := result
	for {
		if app, ok := current.(ast.App); ok {
			if v, ok := app.U.(ast.Var); ok {
				apps = append([]ast.Var{v}, apps...)
			}
			current = app.T
		} else {
			break
		}
	}

	// Should have 4 apps: param0, param1, idx0, idx1
	if len(apps) != 4 {
		t.Fatalf("expected 4 apps, got %d", len(apps))
	}
	// Last arg should be idx_1 at position 0 (innermost)
	if apps[3].Ix != 0 {
		t.Errorf("idx_1 expected at 0, got %d", apps[3].Ix)
	}
	// Second to last should be idx_0 at position 1
	if apps[2].Ix != 1 {
		t.Errorf("idx_0 expected at 1, got %d", apps[2].Ix)
	}
}

// ============================================================================
// buildMotiveTypeFull Tests
// ============================================================================

func TestBuildMotiveTypeFull_NoIndices(t *testing.T) {
	ind := &Inductive{
		Name:       "Nat",
		NumParams:  0,
		NumIndices: 0,
		ParamTypes: nil,
		IndexTypes: nil,
	}
	result := buildMotiveTypeFull(ind, 1)

	// Expected: Nat -> Type_1
	pi, ok := result.(ast.Pi)
	if !ok {
		t.Fatalf("expected Pi, got %T", result)
	}
	if pi.Binder != "_" {
		t.Errorf("expected binder '_', got %q", pi.Binder)
	}
	// Domain should be Nat
	if global, ok := pi.A.(ast.Global); !ok || global.Name != "Nat" {
		t.Errorf("expected domain Nat, got %v", pi.A)
	}
	// Codomain should be Type_1
	if sort, ok := pi.B.(ast.Sort); !ok || sort.U != 1 {
		t.Errorf("expected codomain Type_1, got %v", pi.B)
	}
}

func TestBuildMotiveTypeFull_WithIndices(t *testing.T) {
	// Vec A n: motive should be (n : Nat) -> Vec A n -> Type
	ind := &Inductive{
		Name:       "Vec",
		NumParams:  1,
		NumIndices: 1,
		ParamTypes: []ast.Term{ast.Sort{U: 0}},
		IndexTypes: []ast.Term{ast.Global{Name: "Nat"}},
	}
	result := buildMotiveTypeFull(ind, 0)

	// Expected: (n : Nat) -> Vec A n -> Type
	pi1, ok := result.(ast.Pi)
	if !ok {
		t.Fatalf("expected outer Pi for index, got %T", result)
	}
	if pi1.Binder != "n" {
		t.Errorf("expected binder 'n', got %q", pi1.Binder)
	}

	pi2, ok := pi1.B.(ast.Pi)
	if !ok {
		t.Fatalf("expected inner Pi for target, got %T", pi1.B)
	}
	if pi2.Binder != "_" {
		t.Errorf("expected binder '_', got %q", pi2.Binder)
	}

	// Codomain should be Type
	if sort, ok := pi2.B.(ast.Sort); !ok || sort.U != 0 {
		t.Errorf("expected codomain Type_0, got %v", pi2.B)
	}
}

// ============================================================================
// paramName and indexName Tests
// ============================================================================

func TestParamName(t *testing.T) {
	tests := []struct {
		i        int
		expected string
	}{
		{0, "A"},
		{1, "B"},
		{2, "C"},
		{25, "Z"},
		{26, "P0"},
		{27, "P1"},
	}

	for _, tt := range tests {
		result := paramName(tt.i)
		if result != tt.expected {
			t.Errorf("paramName(%d) = %q, want %q", tt.i, result, tt.expected)
		}
	}
}

func TestIndexName(t *testing.T) {
	tests := []struct {
		i        int
		expected string
	}{
		{0, "n"},
		{1, "m"},
		{2, "k"},
		{3, "j"},
		{4, "i"},
		{5, "i5"},
		{6, "i6"},
	}

	for _, tt := range tests {
		result := indexName(tt.i)
		if result != tt.expected {
			t.Errorf("indexName(%d) = %q, want %q", tt.i, result, tt.expected)
		}
	}
}

// ============================================================================
// buildIHType Tests
// ============================================================================

func TestBuildIHType_NoIndices(t *testing.T) {
	// Simple case: P x where x is at index 0
	result := buildIHType(5, ast.Global{Name: "Nat"}, 0, 0, 0)

	// Expected: App{Var{5}, Var{0}}
	app, ok := result.(ast.App)
	if !ok {
		t.Fatalf("expected App, got %T", result)
	}
	pVar, ok := app.T.(ast.Var)
	if !ok || pVar.Ix != 5 {
		t.Errorf("expected P at index 5, got %v", app.T)
	}
	xVar, ok := app.U.(ast.Var)
	if !ok || xVar.Ix != 0 {
		t.Errorf("expected x at index 0, got %v", app.U)
	}
}

func TestBuildIHType_WithIndices(t *testing.T) {
	// Vec A n: IH for xs : Vec A n should be P n xs
	// argType = Vec A n = App(App(Vec, A), n)
	argType := ast.App{
		T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 1}}, // Vec A
		U: ast.Var{Ix: 0}, // n
	}
	result := buildIHType(10, argType, 1, 1, 0)

	// Expected: App(App(P, n'), xs) where n' is shifted by 1+ihCount = 1
	// So n (Var{0}) becomes Var{1}, and xs is Var{0}
	app1, ok := result.(ast.App)
	if !ok {
		t.Fatalf("expected outer App, got %T", result)
	}
	// Inner should be xs at 0
	xsVar, ok := app1.U.(ast.Var)
	if !ok || xsVar.Ix != 0 {
		t.Errorf("expected xs at 0, got %v", app1.U)
	}

	app2, ok := app1.T.(ast.App)
	if !ok {
		t.Fatalf("expected inner App, got %T", app1.T)
	}
	// Index n (was Var{0}) should be shifted to Var{1}
	nVar, ok := app2.U.(ast.Var)
	if !ok || nVar.Ix != 1 {
		t.Errorf("expected n at 1 (shifted), got %v", app2.U)
	}
}

// ============================================================================
// extractIndicesFromType Tests
// ============================================================================

func TestExtractIndicesFromType_NoIndices(t *testing.T) {
	// List A: no indices
	ty := ast.App{T: ast.Global{Name: "List"}, U: ast.Var{Ix: 0}}
	result := extractIndicesFromType(ty, 1, 0)
	if len(result) != 0 {
		t.Errorf("expected 0 indices, got %d", len(result))
	}
}

func TestExtractIndicesFromType_OneIndex(t *testing.T) {
	// Vec A n: one index
	ty := ast.App{
		T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 1}}, // Vec A
		U: ast.Var{Ix: 0}, // n
	}
	result := extractIndicesFromType(ty, 1, 1)
	if len(result) != 1 {
		t.Fatalf("expected 1 index, got %d", len(result))
	}
	if v, ok := result[0].(ast.Var); !ok || v.Ix != 0 {
		t.Errorf("expected index Var{0}, got %v", result[0])
	}
}

func TestExtractIndicesFromType_TwoIndices(t *testing.T) {
	// Matrix A n m: two indices
	ty := ast.App{
		T: ast.App{
			T: ast.App{T: ast.Global{Name: "Matrix"}, U: ast.Var{Ix: 2}}, // Matrix A
			U: ast.Var{Ix: 1}, // n
		},
		U: ast.Var{Ix: 0}, // m
	}
	result := extractIndicesFromType(ty, 1, 2)
	if len(result) != 2 {
		t.Fatalf("expected 2 indices, got %d", len(result))
	}
}

func TestExtractIndicesFromType_TooFewArgs(t *testing.T) {
	// Not enough args
	ty := ast.Global{Name: "Nat"}
	result := extractIndicesFromType(ty, 1, 1)
	if result != nil {
		t.Errorf("expected nil for too few args, got %v", result)
	}
}

// ============================================================================
// extractConstructorIndices Tests
// ============================================================================

func TestExtractConstructorIndices_Simple(t *testing.T) {
	// vnil : (A : Type) -> Vec A 0
	ctorType := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.App{
			T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 0}}, // Vec A
			U: ast.Global{Name: "zero"}, // 0
		},
	}
	result := extractConstructorIndices(ctorType, 1, 1)
	if len(result) != 1 {
		t.Fatalf("expected 1 index, got %d", len(result))
	}
	if global, ok := result[0].(ast.Global); !ok || global.Name != "zero" {
		t.Errorf("expected index zero, got %v", result[0])
	}
}

// ============================================================================
// shiftIndexExpr Tests
// ============================================================================

func TestShiftIndexExpr_Variable(t *testing.T) {
	expr := ast.Var{Ix: 2}
	result := shiftIndexExpr(expr, 4, []int{0, 1, 2}, 1, 2)
	// With ihCount=2, shift by 2
	v, ok := result.(ast.Var)
	if !ok {
		t.Fatalf("expected Var, got %T", result)
	}
	if v.Ix != 4 { // 2 + 2 = 4
		t.Errorf("expected Var{4}, got Var{%d}", v.Ix)
	}
}

func TestShiftIndexExpr_Global(t *testing.T) {
	expr := ast.Global{Name: "zero"}
	result := shiftIndexExpr(expr, 4, []int{0}, 1, 1)
	// Global should be unchanged
	g, ok := result.(ast.Global)
	if !ok || g.Name != "zero" {
		t.Errorf("expected Global{zero}, got %v", result)
	}
}

// ============================================================================
// GenerateRecursorType for Indexed Inductives
// ============================================================================

func TestGenerateRecursorType_Vec(t *testing.T) {
	// Vec : Type -> Nat -> Type
	// vecElim : (A : Type) -> (P : (n : Nat) -> Vec A n -> Type) ->
	//           case_vnil -> case_vcons -> (n : Nat) -> (xs : Vec A n) -> P n xs
	vec := &Inductive{
		Name:       "Vec",
		Type:       ast.Pi{Binder: "A", A: ast.Sort{U: 0}, B: ast.Pi{Binder: "n", A: ast.Global{Name: "Nat"}, B: ast.Sort{U: 0}}},
		NumParams:  1,
		NumIndices: 1,
		ParamTypes: []ast.Term{ast.Sort{U: 0}},
		IndexTypes: []ast.Term{ast.Global{Name: "Nat"}},
		Constructors: []Constructor{
			{Name: "vnil", Type: ast.Pi{Binder: "A", A: ast.Sort{U: 0}, B: ast.App{T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 0}}, U: ast.Global{Name: "zero"}}}},
			{Name: "vcons", Type: ast.Pi{
				Binder: "A", A: ast.Sort{U: 0},
				B: ast.Pi{Binder: "n", A: ast.Global{Name: "Nat"},
					B: ast.Pi{Binder: "x", A: ast.Var{Ix: 1},
						B: ast.Pi{Binder: "xs", A: ast.App{T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 2}}, U: ast.Var{Ix: 1}},
							B: ast.App{T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 3}}, U: ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 2}}}}}},
			}},
		},
		Eliminator: "vecElim",
	}

	result := GenerateRecursorType(vec)
	if result == nil {
		t.Fatal("GenerateRecursorType returned nil")
	}

	// Check outermost binder is A (parameter)
	pi, ok := result.(ast.Pi)
	if !ok {
		t.Fatalf("expected Pi, got %T", result)
	}
	if pi.Binder != "A" {
		t.Errorf("expected binder 'A', got %q", pi.Binder)
	}
}

func TestGenerateRecursorType_Fin(t *testing.T) {
	// Fin : Nat -> Type (no params, one index)
	// finElim : (P : (n : Nat) -> Fin n -> Type) ->
	//           case_fzero -> case_fsucc -> (n : Nat) -> (i : Fin n) -> P n i
	fin := &Inductive{
		Name:       "Fin",
		Type:       ast.Pi{Binder: "n", A: ast.Global{Name: "Nat"}, B: ast.Sort{U: 0}},
		NumParams:  0,
		NumIndices: 1,
		ParamTypes: nil,
		IndexTypes: []ast.Term{ast.Global{Name: "Nat"}},
		Constructors: []Constructor{
			{Name: "fzero", Type: ast.Pi{Binder: "n", A: ast.Global{Name: "Nat"}, B: ast.App{T: ast.Global{Name: "Fin"}, U: ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 0}}}}},
			{Name: "fsucc", Type: ast.Pi{
				Binder: "n", A: ast.Global{Name: "Nat"},
				B: ast.Pi{Binder: "i", A: ast.App{T: ast.Global{Name: "Fin"}, U: ast.Var{Ix: 0}},
					B: ast.App{T: ast.Global{Name: "Fin"}, U: ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 1}}}},
			}},
		},
		Eliminator: "finElim",
	}

	result := GenerateRecursorType(fin)
	if result == nil {
		t.Fatal("GenerateRecursorType returned nil")
	}

	// Check outermost binder is P (motive, since no params)
	pi, ok := result.(ast.Pi)
	if !ok {
		t.Fatalf("expected Pi, got %T", result)
	}
	if pi.Binder != "P" {
		t.Errorf("expected binder 'P', got %q", pi.Binder)
	}
}

// ============================================================================
// buildCaseTypeFull Tests (for indexed inductives)
// ============================================================================

func TestBuildCaseTypeFull_VecCons(t *testing.T) {
	// vcons data args (after params): n : Nat, x : A, xs : Vec A n
	// case_vcons : (n : Nat) -> (x : A) -> (xs : Vec A n) -> (ih : P n xs) -> P (succ n) (vcons A n x xs)
	ind := &Inductive{
		Name:       "Vec",
		NumParams:  1,
		NumIndices: 1,
		ParamTypes: []ast.Term{ast.Sort{U: 0}},
		IndexTypes: []ast.Term{ast.Global{Name: "Nat"}},
	}

	args := []PiArg{
		{Name: "n", Type: ast.Global{Name: "Nat"}},
		{Name: "x", Type: ast.Var{Ix: 1}}, // A
		{Name: "xs", Type: ast.App{T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 2}}, U: ast.Var{Ix: 0}}},
	}

	ctorResultIndices := []ast.Term{ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 2}}}

	result := buildCaseTypeFull(ind, args, 1, 1, "vcons", ctorResultIndices)

	// Count binders: n, x, xs, ih_xs = 4
	count := 0
	binders := []string{}
	current := result
	for {
		if pi, ok := current.(ast.Pi); ok {
			count++
			binders = append(binders, pi.Binder)
			current = pi.B
		} else {
			break
		}
	}

	if count != 4 {
		t.Errorf("expected 4 binders, got %d: %v", count, binders)
	}
	expectedBinders := []string{"n", "x", "xs", "ih_xs"}
	for i, expected := range expectedBinders {
		if i < len(binders) && binders[i] != expected {
			t.Errorf("binder %d: expected %q, got %q", i, expected, binders[i])
		}
	}
}
