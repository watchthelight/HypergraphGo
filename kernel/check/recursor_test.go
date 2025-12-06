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
				{Name: "x", Type: ast.Var{Ix: 0}},                // A (non-recursive)
				{Name: "xs", Type: ast.Global{Name: "List"}},     // List (recursive)
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
