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
	var current ast.Term = result
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
