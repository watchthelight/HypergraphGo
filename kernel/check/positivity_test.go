package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func TestCheckPositivity_Valid(t *testing.T) {
	tests := []struct {
		name         string
		indName      string
		constructors []Constructor
	}{
		{
			name:    "Nat zero",
			indName: "Nat",
			constructors: []Constructor{
				{Name: "zero", Type: ast.Global{Name: "Nat"}},
			},
		},
		{
			name:    "Nat succ",
			indName: "Nat",
			constructors: []Constructor{
				{Name: "zero", Type: ast.Global{Name: "Nat"}},
				{Name: "succ", Type: ast.Pi{
					Binder: "_",
					A:      ast.Global{Name: "Nat"},
					B:      ast.Global{Name: "Nat"},
				}},
			},
		},
		{
			name:    "Bool",
			indName: "Bool",
			constructors: []Constructor{
				{Name: "true", Type: ast.Global{Name: "Bool"}},
				{Name: "false", Type: ast.Global{Name: "Bool"}},
			},
		},
		{
			name:    "List nil",
			indName: "List",
			constructors: []Constructor{
				// nil : List (assuming List is already applied to A)
				{Name: "nil", Type: ast.Global{Name: "List"}},
			},
		},
		{
			name:    "List cons - positive",
			indName: "List",
			constructors: []Constructor{
				// cons : A -> List -> List
				{Name: "cons", Type: ast.Pi{
					Binder: "x",
					A:      ast.Var{Ix: 0}, // A (parameter)
					B: ast.Pi{
						Binder: "xs",
						A:      ast.Global{Name: "List"},
						B:      ast.Global{Name: "List"},
					},
				}},
			},
		},
		{
			name:    "Tree with nested positive occurrence",
			indName: "Tree",
			constructors: []Constructor{
				// node : List Tree -> Tree
				// Tree occurs positively as argument to List
				{Name: "node", Type: ast.Pi{
					Binder: "_",
					A:      ast.App{T: ast.Global{Name: "List"}, U: ast.Global{Name: "Tree"}},
					B:      ast.Global{Name: "Tree"},
				}},
			},
		},
		{
			name:    "No self-reference",
			indName: "Unit",
			constructors: []Constructor{
				{Name: "tt", Type: ast.Global{Name: "Unit"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPositivity(tt.indName, tt.constructors)
			if err != nil {
				t.Errorf("CheckPositivity() unexpected error: %v", err)
			}
		})
	}
}

func TestCheckPositivity_Invalid(t *testing.T) {
	tests := []struct {
		name         string
		indName      string
		constructors []Constructor
	}{
		{
			name:    "Negative occurrence in domain",
			indName: "Bad",
			constructors: []Constructor{
				// mk : (Bad -> Nat) -> Bad
				{Name: "mk", Type: ast.Pi{
					Binder: "f",
					A: ast.Pi{
						Binder: "_",
						A:      ast.Global{Name: "Bad"},
						B:      ast.Global{Name: "Nat"},
					},
					B: ast.Global{Name: "Bad"},
				}},
			},
		},
		{
			name:    "Nested negative occurrence",
			indName: "Evil",
			constructors: []Constructor{
				// mk : ((Evil -> Nat) -> Nat) -> Evil
				// Evil appears in negative position (domain of domain)
				{Name: "mk", Type: ast.Pi{
					Binder: "f",
					A: ast.Pi{
						Binder: "_",
						A: ast.Pi{
							Binder: "_",
							A:      ast.Global{Name: "Evil"},
							B:      ast.Global{Name: "Nat"},
						},
						B: ast.Global{Name: "Nat"},
					},
					B: ast.Global{Name: "Evil"},
				}},
			},
		},
		{
			name:    "Direct negative in simple arrow",
			indName: "Neg",
			constructors: []Constructor{
				// mk : Neg -> Neg -> Neg
				// First Neg is in negative position
				{Name: "mk", Type: ast.Pi{
					Binder: "_",
					A: ast.Pi{
						Binder: "_",
						A:      ast.Global{Name: "Neg"}, // negative!
						B:      ast.Global{Name: "Nat"},
					},
					B: ast.Global{Name: "Neg"},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPositivity(tt.indName, tt.constructors)
			if err == nil {
				t.Error("CheckPositivity() expected error, got nil")
			}
			if _, ok := err.(*PositivityError); !ok {
				t.Errorf("CheckPositivity() expected PositivityError, got %T", err)
			}
		})
	}
}

func TestOccursIn(t *testing.T) {
	tests := []struct {
		name     string
		global   string
		term     ast.Term
		expected bool
	}{
		{
			name:     "Global match",
			global:   "Nat",
			term:     ast.Global{Name: "Nat"},
			expected: true,
		},
		{
			name:     "Global no match",
			global:   "Nat",
			term:     ast.Global{Name: "Bool"},
			expected: false,
		},
		{
			name:     "In Pi domain",
			global:   "Nat",
			term:     ast.Pi{Binder: "x", A: ast.Global{Name: "Nat"}, B: ast.Sort{U: 0}},
			expected: true,
		},
		{
			name:     "In Pi codomain",
			global:   "Nat",
			term:     ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Global{Name: "Nat"}},
			expected: true,
		},
		{
			name:     "Not in Pi",
			global:   "Nat",
			term:     ast.Pi{Binder: "x", A: ast.Global{Name: "Bool"}, B: ast.Sort{U: 0}},
			expected: false,
		},
		{
			name:     "In App function",
			global:   "List",
			term:     ast.App{T: ast.Global{Name: "List"}, U: ast.Global{Name: "Nat"}},
			expected: true,
		},
		{
			name:     "In App argument",
			global:   "Nat",
			term:     ast.App{T: ast.Global{Name: "List"}, U: ast.Global{Name: "Nat"}},
			expected: true,
		},
		{
			name:     "Var",
			global:   "Nat",
			term:     ast.Var{Ix: 0},
			expected: false,
		},
		{
			name:     "Sort",
			global:   "Nat",
			term:     ast.Sort{U: 0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OccursIn(tt.global, tt.term)
			if result != tt.expected {
				t.Errorf("OccursIn(%q, %v) = %v, want %v", tt.global, tt.term, result, tt.expected)
			}
		})
	}
}

func TestIsRecursiveArg(t *testing.T) {
	tests := []struct {
		name     string
		indName  string
		argType  ast.Term
		expected bool
	}{
		{
			name:     "Direct reference",
			indName:  "Nat",
			argType:  ast.Global{Name: "Nat"},
			expected: true,
		},
		{
			name:     "No reference",
			indName:  "Nat",
			argType:  ast.Global{Name: "Bool"},
			expected: false,
		},
		{
			name:     "Nested in App",
			indName:  "Tree",
			argType:  ast.App{T: ast.Global{Name: "List"}, U: ast.Global{Name: "Tree"}},
			expected: true,
		},
		{
			name:     "Variable",
			indName:  "Nat",
			argType:  ast.Var{Ix: 0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRecursiveArg(tt.indName, tt.argType)
			if result != tt.expected {
				t.Errorf("IsRecursiveArg(%q, %v) = %v, want %v", tt.indName, tt.argType, result, tt.expected)
			}
		})
	}
}

func TestPolarity(t *testing.T) {
	if Positive.Flip() != Negative {
		t.Error("Positive.Flip() should be Negative")
	}
	if Negative.Flip() != Positive {
		t.Error("Negative.Flip() should be Positive")
	}
	if Positive.String() != "positive" {
		t.Error("Positive.String() should be 'positive'")
	}
	if Negative.String() != "negative" {
		t.Error("Negative.String() should be 'negative'")
	}
}
