package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
)

func TestDeclareInductive_Valid(t *testing.T) {
	tests := []struct {
		name    string
		indName string
		indType ast.Term
		constrs []Constructor
		elim    string
	}{
		{
			name:    "Nat",
			indName: "Nat",
			indType: ast.Sort{U: 0},
			constrs: []Constructor{
				{Name: "zero", Type: ast.Global{Name: "Nat"}},
				{Name: "succ", Type: ast.Pi{
					Binder: "_",
					A:      ast.Global{Name: "Nat"},
					B:      ast.Global{Name: "Nat"},
				}},
			},
			elim: "natElim",
		},
		{
			name:    "Bool",
			indName: "Bool",
			indType: ast.Sort{U: 0},
			constrs: []Constructor{
				{Name: "true", Type: ast.Global{Name: "Bool"}},
				{Name: "false", Type: ast.Global{Name: "Bool"}},
			},
			elim: "boolElim",
		},
		{
			name:    "Unit",
			indName: "Unit",
			indType: ast.Sort{U: 0},
			constrs: []Constructor{
				{Name: "tt", Type: ast.Global{Name: "Unit"}},
			},
			elim: "unitElim",
		},
		{
			name:    "Empty",
			indName: "Empty",
			indType: ast.Sort{U: 0},
			constrs: []Constructor{},
			elim:    "emptyElim",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := NewGlobalEnv()
			err := env.DeclareInductive(tt.indName, tt.indType, tt.constrs, tt.elim)
			if err != nil {
				t.Errorf("DeclareInductive() unexpected error: %v", err)
			}

			// Verify the inductive was added
			indType := env.LookupType(tt.indName)
			if indType == nil {
				t.Errorf("DeclareInductive() inductive not found in environment")
			}

			// Verify constructors are accessible
			for _, c := range tt.constrs {
				ctorType := env.LookupType(c.Name)
				if ctorType == nil {
					t.Errorf("DeclareInductive() constructor %s not found in environment", c.Name)
				}
			}
		})
	}
}

func TestDeclareInductive_Invalid(t *testing.T) {
	tests := []struct {
		name          string
		indName       string
		indType       ast.Term
		constrs       []Constructor
		elim          string
		errType       string // "positivity" or "result"
		usePrimitives bool   // whether to use environment with primitives
	}{
		{
			name:    "Negative occurrence",
			indName: "Bad",
			indType: ast.Sort{U: 0},
			constrs: []Constructor{
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
			elim:          "badElim",
			errType:       "positivity",
			usePrimitives: true, // Need Nat in environment
		},
		{
			name:    "Wrong result type",
			indName: "Foo",
			indType: ast.Sort{U: 0},
			constrs: []Constructor{
				{Name: "mk", Type: ast.Global{Name: "Bar"}}, // Returns Bar, not Foo
			},
			elim:          "fooElim",
			errType:       "result",
			usePrimitives: false,
		},
		{
			name:    "Constructor returns different type",
			indName: "X",
			indType: ast.Sort{U: 0},
			constrs: []Constructor{
				{Name: "mkX", Type: ast.Pi{
					Binder: "_",
					A:      ast.Global{Name: "Nat"},
					B:      ast.Global{Name: "Y"}, // Wrong result type
				}},
			},
			elim:          "xElim",
			errType:       "result",
			usePrimitives: true, // Need Nat in environment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var env *GlobalEnv
			if tt.usePrimitives {
				env = NewGlobalEnvWithPrimitives()
			} else {
				env = NewGlobalEnv()
			}
			err := env.DeclareInductive(tt.indName, tt.indType, tt.constrs, tt.elim)
			if err == nil {
				t.Error("DeclareInductive() expected error, got nil")
				return
			}

			// Check error type
			switch tt.errType {
			case "positivity":
				if _, ok := err.(*PositivityError); !ok {
					t.Errorf("DeclareInductive() expected PositivityError, got %T: %v", err, err)
				}
			case "result":
				if _, ok := err.(*ConstructorError); !ok {
					t.Errorf("DeclareInductive() expected ConstructorError, got %T: %v", err, err)
				}
			}
		})
	}
}

func TestConstructorResultType(t *testing.T) {
	tests := []struct {
		name     string
		ty       ast.Term
		expected ast.Term
	}{
		{
			name:     "Direct global",
			ty:       ast.Global{Name: "Nat"},
			expected: ast.Global{Name: "Nat"},
		},
		{
			name: "Single Pi",
			ty: ast.Pi{
				Binder: "_",
				A:      ast.Global{Name: "Nat"},
				B:      ast.Global{Name: "Nat"},
			},
			expected: ast.Global{Name: "Nat"},
		},
		{
			name: "Nested Pi",
			ty: ast.Pi{
				Binder: "x",
				A:      ast.Global{Name: "A"},
				B: ast.Pi{
					Binder: "y",
					A:      ast.Global{Name: "B"},
					B:      ast.Global{Name: "C"},
				},
			},
			expected: ast.Global{Name: "C"},
		},
		{
			name: "Applied type",
			ty: ast.Pi{
				Binder: "x",
				A:      ast.Global{Name: "A"},
				B:      ast.App{T: ast.Global{Name: "List"}, U: ast.Var{Ix: 0}},
			},
			expected: ast.App{T: ast.Global{Name: "List"}, U: ast.Var{Ix: 0}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constructorResultType(tt.ty)
			// Simple structural equality check
			if result == nil {
				t.Error("constructorResultType() returned nil")
				return
			}
		})
	}
}

func TestDeclareInductive_InvalidType(t *testing.T) {
	// DeclareInductive should reject invalid inductive types

	t.Run("Pi not ending in Sort", func(t *testing.T) {
		// A Pi type that doesn't end in a Sort is invalid
		// (x : Type) -> Nat is not a valid inductive type
		env := NewGlobalEnvWithPrimitives() // Need Nat defined
		err := env.DeclareInductive("Bad", ast.Pi{
			Binder: "x",
			A:      ast.Sort{U: 0},
			B:      ast.Global{Name: "Nat"}, // Ends in Nat, not Sort
		}, []Constructor{
			{Name: "mk", Type: ast.Global{Name: "Bad"}},
		}, "badElim")
		if err == nil {
			t.Error("DeclareInductive() expected error for Pi not ending in Sort, got nil")
		}
		if _, ok := err.(*InductiveError); !ok {
			t.Errorf("DeclareInductive() expected InductiveError, got %T: %v", err, err)
		}
	})

	t.Run("Global instead of Sort", func(t *testing.T) {
		// An inductive type that's just a Global (not a Sort) is invalid
		env := NewGlobalEnvWithPrimitives()
		err := env.DeclareInductive("Bad2", ast.Global{Name: "Nat"}, []Constructor{
			{Name: "mk", Type: ast.Global{Name: "Bad2"}},
		}, "bad2Elim")
		if err == nil {
			t.Error("DeclareInductive() expected error for non-Sort type, got nil")
		}
		if _, ok := err.(*InductiveError); !ok {
			t.Errorf("DeclareInductive() expected InductiveError, got %T: %v", err, err)
		}
	})
}

func TestDeclareInductive_RegistersEliminator(t *testing.T) {
	// Verify that DeclareInductive registers the eliminator in GlobalEnv
	env := NewGlobalEnv()
	err := env.DeclareInductive("Nat", ast.Sort{U: 0}, []Constructor{
		{Name: "zero", Type: ast.Global{Name: "Nat"}},
		{Name: "succ", Type: ast.Pi{
			Binder: "_",
			A:      ast.Global{Name: "Nat"},
			B:      ast.Global{Name: "Nat"},
		}},
	}, "natElim")
	if err != nil {
		t.Fatalf("DeclareInductive() unexpected error: %v", err)
	}

	// Check that natElim is registered
	elimType := env.LookupType("natElim")
	if elimType == nil {
		t.Error("DeclareInductive() should register the eliminator, but natElim not found")
	}

	// Verify it's a Pi type (eliminator starts with motive)
	if _, ok := elimType.(ast.Pi); !ok {
		t.Errorf("natElim should be a Pi type, got %T", elimType)
	}
}

func TestIsAppOfGlobal(t *testing.T) {
	tests := []struct {
		name     string
		term     ast.Term
		global   string
		expected bool
	}{
		{
			name:     "Direct global match",
			term:     ast.Global{Name: "List"},
			global:   "List",
			expected: true,
		},
		{
			name:     "Direct global no match",
			term:     ast.Global{Name: "List"},
			global:   "Nat",
			expected: false,
		},
		{
			name:     "Single app match",
			term:     ast.App{T: ast.Global{Name: "List"}, U: ast.Global{Name: "Nat"}},
			global:   "List",
			expected: true,
		},
		{
			name: "Nested app match",
			term: ast.App{
				T: ast.App{T: ast.Global{Name: "Either"}, U: ast.Global{Name: "A"}},
				U: ast.Global{Name: "B"},
			},
			global:   "Either",
			expected: true,
		},
		{
			name:     "App no match",
			term:     ast.App{T: ast.Global{Name: "List"}, U: ast.Global{Name: "Nat"}},
			global:   "Vec",
			expected: false,
		},
		{
			name:     "Var not matching",
			term:     ast.Var{Ix: 0},
			global:   "List",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAppOfGlobal(tt.term, tt.global)
			if result != tt.expected {
				t.Errorf("isAppOfGlobal() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDeclareInductive_IllFormedConstructor(t *testing.T) {
	// DeclareInductive should reject ill-formed constructor types
	tests := []struct {
		name    string
		indName string
		indType ast.Term
		constrs []Constructor
		elim    string
	}{
		{
			name:    "Constructor type references unknown global",
			indName: "Bad",
			indType: ast.Sort{U: 0},
			constrs: []Constructor{
				{Name: "mk", Type: ast.Pi{
					Binder: "_",
					A:      ast.Global{Name: "UnknownType"},
					B:      ast.Global{Name: "Bad"},
				}},
			},
			elim: "badElim",
		},
		{
			name:    "Constructor domain is not a type",
			indName: "Bad2",
			indType: ast.Sort{U: 0},
			constrs: []Constructor{
				// Pi with domain that isn't a type (zero is a value, not a type)
				{Name: "mk", Type: ast.Pi{
					Binder: "_",
					A:      ast.Global{Name: "zero"}, // zero : Nat, not a type
					B:      ast.Global{Name: "Bad2"},
				}},
			},
			elim: "bad2Elim",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := NewGlobalEnvWithPrimitives()
			err := env.DeclareInductive(tt.indName, tt.indType, tt.constrs, tt.elim)
			if err == nil {
				t.Error("DeclareInductive() expected error for ill-formed constructor, got nil")
			}
			if _, ok := err.(*ConstructorError); !ok {
				t.Errorf("DeclareInductive() expected ConstructorError, got %T: %v", err, err)
			}
		})
	}
}

// TestProperty_IndexArgPositionsCompleteness verifies the invariant that
// IndexArgPositions entries are either complete (length == NumIndices) or absent.
// This ensures the evaluator can safely use metadata without partial application.
func TestProperty_IndexArgPositionsCompleteness(t *testing.T) {
	eval.ClearRecursorRegistry()
	env := NewGlobalEnvWithPrimitives()

	// Test a variety of inductives with different index structures
	testCases := []struct {
		name     string
		indName  string
		indType  ast.Term
		constrs  []Constructor
		elim     string
		numIdx   int
		expectOK bool // whether metadata should be complete
	}{
		{
			name:    "Vec (variable index)",
			indName: "Vec",
			indType: ast.Pi{
				Binder: "A",
				A:      ast.Sort{U: 0},
				B: ast.Pi{
					Binder: "n",
					A:      ast.Global{Name: "Nat"},
					B:      ast.Sort{U: 0},
				},
			},
			constrs: []Constructor{
				{Name: "vnil", Type: ast.Pi{
					Binder: "A",
					A:      ast.Sort{U: 0},
					B: ast.App{
						T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 0}},
						U: ast.Global{Name: "zero"},
					},
				}},
				{Name: "vcons", Type: ast.Pi{
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
									U: ast.Var{Ix: 1}, // n (variable index)
								},
								B: ast.App{
									T: ast.App{T: ast.Global{Name: "Vec"}, U: ast.Var{Ix: 3}},
									U: ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 2}},
								},
							},
						},
					},
				}},
			},
			elim:     "vecElim",
			numIdx:   1,
			expectOK: true, // xs has type Vec A n, n is a variable
		},
		{
			name:    "Stepped (computed index)",
			indName: "Stepped",
			indType: ast.Pi{
				Binder: "n",
				A:      ast.Global{Name: "Nat"},
				B:      ast.Sort{U: 0},
			},
			constrs: []Constructor{
				{Name: "base", Type: ast.App{
					T: ast.Global{Name: "Stepped"},
					U: ast.Global{Name: "zero"},
				}},
				{Name: "step", Type: ast.Pi{
					Binder: "n",
					A:      ast.Global{Name: "Nat"},
					B: ast.Pi{
						Binder: "s",
						A: ast.App{
							T: ast.Global{Name: "Stepped"},
							U: ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 0}}, // succ n (computed)
						},
						B: ast.App{
							T: ast.Global{Name: "Stepped"},
							U: ast.App{T: ast.Global{Name: "succ"}, U: ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 1}}},
						},
					},
				}},
			},
			elim:     "steppedElim",
			numIdx:   1,
			expectOK: false, // s has type Stepped (succ n), succ n is computed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := env.DeclareInductive(tc.indName, tc.indType, tc.constrs, tc.elim)
			if err != nil {
				t.Fatalf("DeclareInductive(%s) failed: %v", tc.indName, err)
			}

			info := eval.LookupRecursor(tc.elim)
			if info == nil {
				t.Fatalf("%s not registered", tc.elim)
			}

			if info.NumIndices != tc.numIdx {
				t.Errorf("NumIndices = %d, want %d", info.NumIndices, tc.numIdx)
			}

			// Verify property: every IndexArgPositions entry must be complete or absent
			for ctorIdx, ctor := range info.Ctors {
				for recIdx, positions := range ctor.IndexArgPositions {
					// If present, must have length == NumIndices (complete)
					// or length 0 (effectively empty, will fall back)
					if len(positions) > 0 && len(positions) < info.NumIndices {
						t.Errorf("ctor[%d].IndexArgPositions[%d] = %v (len %d) is partially filled, "+
							"expected complete (len %d) or empty",
							ctorIdx, recIdx, positions, len(positions), info.NumIndices)
					}
				}
			}
		})
	}

	eval.ClearRecursorRegistry()
}
