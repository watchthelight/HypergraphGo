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

// TestCheckPositivity_DoubleDomain tests positivity in nested domains.
// The rule is: once in a domain (negative position), we stay negative.
func TestCheckPositivity_DoubleDomain(t *testing.T) {
	tests := []struct {
		name         string
		indName      string
		constructors []Constructor
		valid        bool
	}{
		{
			// (A -> B) -> C: In A position, we're negative
			// mkBad : ((Bad -> Nat) -> Nat) -> Bad
			// Bad appears at depth 2 (domain of domain), which is NEGATIVE
			name:    "Double domain - still negative",
			indName: "Bad",
			constructors: []Constructor{
				{Name: "mk", Type: ast.Pi{
					Binder: "f",
					A: ast.Pi{
						Binder: "_",
						A: ast.Pi{
							Binder: "_",
							A:      ast.Global{Name: "Bad"}, // Negative!
							B:      ast.Global{Name: "Nat"},
						},
						B: ast.Global{Name: "Nat"},
					},
					B: ast.Global{Name: "Bad"},
				}},
			},
			valid: false,
		},
		{
			// mk : (Nat -> Good) -> Good
			// Good appears only in positive position (codomain)
			name:    "Only in positive positions",
			indName: "Good",
			constructors: []Constructor{
				{Name: "mk", Type: ast.Pi{
					Binder: "f",
					A: ast.Pi{
						Binder: "_",
						A:      ast.Global{Name: "Nat"},
						B:      ast.Global{Name: "Good"}, // Positive (codomain)
					},
					B: ast.Global{Name: "Good"},
				}},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPositivity(tt.indName, tt.constructors)
			if tt.valid && err != nil {
				t.Errorf("CheckPositivity() unexpected error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("CheckPositivity() expected error, got nil")
			}
		})
	}
}

// TestCheckPositivity_Sigma tests positivity in Sigma types.
func TestCheckPositivity_Sigma(t *testing.T) {
	// mk : (Σ(x : Nat) . Bad) -> Bad
	// Bad appears in positive position (second component of Sigma)
	validSigma := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "p",
			A: ast.Sigma{
				Binder: "x",
				A:      ast.Global{Name: "Nat"},
				B:      ast.Global{Name: "Good"}, // Positive
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	err := CheckPositivity("Good", validSigma)
	if err != nil {
		t.Errorf("CheckPositivity(Sigma positive) unexpected error: %v", err)
	}

	// mk : ((Σ(x : Bad) . Nat) -> Nat) -> Bad
	// Bad appears in negative position (inside domain, even within Sigma)
	invalidSigma := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Sigma{
					Binder: "x",
					A:      ast.Global{Name: "Bad"}, // Negative (domain of outer domain)
					B:      ast.Global{Name: "Nat"},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	err = CheckPositivity("Bad", invalidSigma)
	if err == nil {
		t.Error("CheckPositivity(Sigma negative) expected error, got nil")
	}
}

// TestCheckPositivity_Id tests positivity with identity types.
func TestCheckPositivity_Id(t *testing.T) {
	// mk : Id Nat zero zero -> Good
	// Good doesn't appear in Id type components, so this is valid
	validId := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "p",
			A: ast.Id{
				A: ast.Global{Name: "Nat"},
				X: ast.Global{Name: "zero"},
				Y: ast.Global{Name: "zero"},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	err := CheckPositivity("Good", validId)
	if err != nil {
		t.Errorf("CheckPositivity(Id no occurrence) unexpected error: %v", err)
	}

	// mk : (Id Bad x y -> Nat) -> Bad
	// Bad appears in negative position
	invalidId := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Id{
					A: ast.Global{Name: "Bad"}, // Negative (in domain)
					X: ast.Var{Ix: 0},
					Y: ast.Var{Ix: 0},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	err = CheckPositivity("Bad", invalidId)
	if err == nil {
		t.Error("CheckPositivity(Id negative) expected error, got nil")
	}
}

// TestCheckPositivity_MultipleRecursiveArgs tests constructors with multiple recursive arguments.
func TestCheckPositivity_MultipleRecursiveArgs(t *testing.T) {
	// mk : T -> T -> T (two recursive args - valid)
	valid := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "x",
			A:      ast.Global{Name: "T"},
			B: ast.Pi{
				Binder: "y",
				A:      ast.Global{Name: "T"},
				B:      ast.Global{Name: "T"},
			},
		}},
	}

	err := CheckPositivity("T", valid)
	if err != nil {
		t.Errorf("CheckPositivity(multiple recursive) unexpected error: %v", err)
	}

	// mk : (T -> T) -> T (T in domain of function arg - invalid)
	invalid := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A:      ast.Global{Name: "Bad"},
				B:      ast.Global{Name: "Bad"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	err = CheckPositivity("Bad", invalid)
	if err == nil {
		t.Error("CheckPositivity(T -> T as arg) expected error, got nil")
	}
}

// TestCheckPositivity_Let tests positivity with Let expressions.
func TestCheckPositivity_Let(t *testing.T) {
	// mk : (let x = Nat in x) -> Good (no occurrence of Good in domain)
	valid := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.Let{
				Binder: "x",
				Ann:    ast.Sort{U: 0},
				Val:    ast.Global{Name: "Nat"},
				Body:   ast.Var{Ix: 0},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	err := CheckPositivity("Good", valid)
	if err != nil {
		t.Errorf("CheckPositivity(Let no occurrence) unexpected error: %v", err)
	}

	// mk : ((let x = Bad in x) -> Nat) -> Bad (Bad in domain through Let)
	invalid := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Let{
					Binder: "x",
					Ann:    ast.Sort{U: 0},
					Val:    ast.Global{Name: "Bad"},
					Body:   ast.Var{Ix: 0},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	err = CheckPositivity("Bad", invalid)
	if err == nil {
		t.Error("CheckPositivity(Let with Bad in domain) expected error, got nil")
	}
}

// TestCheckPositivity_Projections tests positivity with Fst/Snd.
func TestCheckPositivity_Projections(t *testing.T) {
	// These shouldn't normally appear in constructor types, but test anyway
	// mk : (fst p -> Nat) -> Good where p doesn't contain Good
	valid := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.Pi{
				Binder: "_",
				A:      ast.Fst{P: ast.Var{Ix: 0}}, // fst of a variable
				B:      ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	err := CheckPositivity("Good", valid)
	if err != nil {
		t.Errorf("CheckPositivity(Fst/Snd) unexpected error: %v", err)
	}
}

// TestCheckPositivity_Pair tests positivity with Pair expressions.
func TestCheckPositivity_Pair(t *testing.T) {
	// mk : (Nat, Good) -> Good - Good in positive position
	valid := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "_",
			A: ast.Pair{
				Fst: ast.Global{Name: "Nat"},
				Snd: ast.Global{Name: "Good"},
			},
			B: ast.Global{Name: "Good"},
		}},
	}

	err := CheckPositivity("Good", valid)
	if err != nil {
		t.Errorf("CheckPositivity(Pair positive) unexpected error: %v", err)
	}

	// mk : ((Nat, Bad) -> Nat) -> Bad - Bad in negative position
	invalid := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Pair{
					Fst: ast.Global{Name: "Nat"},
					Snd: ast.Global{Name: "Bad"},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	err = CheckPositivity("Bad", invalid)
	if err == nil {
		t.Error("CheckPositivity(Pair negative) expected error, got nil")
	}
}

// TestCheckPositivity_Refl tests positivity with Refl.
func TestCheckPositivity_Refl(t *testing.T) {
	// mk : (refl Good x -> Nat) -> Good - Good occurs in refl type
	invalid := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.Refl{
					A: ast.Global{Name: "Bad"},
					X: ast.Var{Ix: 0},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	err := CheckPositivity("Bad", invalid)
	if err == nil {
		t.Error("CheckPositivity(Refl negative) expected error, got nil")
	}
}

// TestCheckPositivity_J tests positivity with J eliminator.
func TestCheckPositivity_J(t *testing.T) {
	// J with Good in one of its arguments, in negative position
	invalid := []Constructor{
		{Name: "mk", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_",
				A: ast.J{
					A: ast.Global{Name: "Bad"}, // Bad in A position
					C: ast.Var{Ix: 0},
					D: ast.Var{Ix: 0},
					X: ast.Var{Ix: 0},
					Y: ast.Var{Ix: 0},
					P: ast.Var{Ix: 0},
				},
				B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "Bad"},
		}},
	}

	err := CheckPositivity("Bad", invalid)
	if err == nil {
		t.Error("CheckPositivity(J negative) expected error, got nil")
	}
}

// TestCheckMutualPositivity_Valid tests valid mutual recursion.
func TestCheckMutualPositivity_Valid(t *testing.T) {
	// Even/Odd mutual recursion
	// Even : Type, Odd : Type
	// evenZero : Even
	// evenSucc : Odd -> Even
	// oddSucc : Even -> Odd
	constrs := map[string][]Constructor{
		"Even": {
			{Name: "evenZero", Type: ast.Global{Name: "Even"}},
			{Name: "evenSucc", Type: ast.Pi{
				Binder: "_",
				A:      ast.Global{Name: "Odd"},
				B:      ast.Global{Name: "Even"},
			}},
		},
		"Odd": {
			{Name: "oddSucc", Type: ast.Pi{
				Binder: "_",
				A:      ast.Global{Name: "Even"},
				B:      ast.Global{Name: "Odd"},
			}},
		},
	}

	err := CheckMutualPositivity([]string{"Even", "Odd"}, constrs)
	if err != nil {
		t.Errorf("CheckMutualPositivity(Even/Odd) unexpected error: %v", err)
	}
}

// TestCheckMutualPositivity_Invalid tests invalid mutual recursion.
func TestCheckMutualPositivity_Invalid(t *testing.T) {
	// A and B where A constructor has B in negative position
	// mk : (B -> Nat) -> A
	constrs := map[string][]Constructor{
		"A": {
			{Name: "mk", Type: ast.Pi{
				Binder: "f",
				A: ast.Pi{
					Binder: "_",
					A:      ast.Global{Name: "B"}, // B in negative position
					B:      ast.Global{Name: "Nat"},
				},
				B: ast.Global{Name: "A"},
			}},
		},
		"B": {
			{Name: "mkB", Type: ast.Global{Name: "B"}},
		},
	}

	err := CheckMutualPositivity([]string{"A", "B"}, constrs)
	if err == nil {
		t.Error("CheckMutualPositivity(A has B negative) expected error, got nil")
	}
}

// TestCheckMutualPositivity_ThreeWay tests three-way mutual recursion.
func TestCheckMutualPositivity_ThreeWay(t *testing.T) {
	// A -> B -> C -> A cycle (all positive)
	constrs := map[string][]Constructor{
		"A": {
			{Name: "mkA", Type: ast.Pi{
				Binder: "_", A: ast.Global{Name: "C"}, B: ast.Global{Name: "A"},
			}},
		},
		"B": {
			{Name: "mkB", Type: ast.Pi{
				Binder: "_", A: ast.Global{Name: "A"}, B: ast.Global{Name: "B"},
			}},
		},
		"C": {
			{Name: "mkC", Type: ast.Pi{
				Binder: "_", A: ast.Global{Name: "B"}, B: ast.Global{Name: "C"},
			}},
		},
	}

	err := CheckMutualPositivity([]string{"A", "B", "C"}, constrs)
	if err != nil {
		t.Errorf("CheckMutualPositivity(three-way) unexpected error: %v", err)
	}

	// Now make one negative: mkA : (C -> Nat) -> A
	constrs["A"] = []Constructor{
		{Name: "mkA", Type: ast.Pi{
			Binder: "f",
			A: ast.Pi{
				Binder: "_", A: ast.Global{Name: "C"}, B: ast.Global{Name: "Nat"},
			},
			B: ast.Global{Name: "A"},
		}},
	}

	err = CheckMutualPositivity([]string{"A", "B", "C"}, constrs)
	if err == nil {
		t.Error("CheckMutualPositivity(three-way negative) expected error, got nil")
	}
}

// TestPositivityError tests PositivityError formatting.
func TestPositivityError(t *testing.T) {
	err := &PositivityError{
		IndName:     "Bad",
		Constructor: "mk",
		Position:    "argument 1",
		Polarity:    Negative,
	}

	msg := err.Error()
	if msg == "" {
		t.Error("PositivityError.Error() returned empty string")
	}
	if !contains(msg, "Bad") || !contains(msg, "mk") || !contains(msg, "negative") {
		t.Errorf("PositivityError message missing expected content: %s", msg)
	}
}

// TestOccursIn_Additional tests OccursIn for additional term types.
func TestOccursIn_Additional(t *testing.T) {
	tests := []struct {
		name     string
		global   string
		term     ast.Term
		expected bool
	}{
		{
			name:     "In Sigma first",
			global:   "T",
			term:     ast.Sigma{Binder: "x", A: ast.Global{Name: "T"}, B: ast.Sort{U: 0}},
			expected: true,
		},
		{
			name:     "In Sigma second",
			global:   "T",
			term:     ast.Sigma{Binder: "x", A: ast.Sort{U: 0}, B: ast.Global{Name: "T"}},
			expected: true,
		},
		{
			name:     "In Lam Ann",
			global:   "T",
			term:     ast.Lam{Binder: "x", Ann: ast.Global{Name: "T"}, Body: ast.Var{Ix: 0}},
			expected: true,
		},
		{
			name:     "In Lam Body",
			global:   "T",
			term:     ast.Lam{Binder: "x", Ann: ast.Sort{U: 0}, Body: ast.Global{Name: "T"}},
			expected: true,
		},
		{
			name:     "In Pair Fst",
			global:   "T",
			term:     ast.Pair{Fst: ast.Global{Name: "T"}, Snd: ast.Var{Ix: 0}},
			expected: true,
		},
		{
			name:     "In Pair Snd",
			global:   "T",
			term:     ast.Pair{Fst: ast.Var{Ix: 0}, Snd: ast.Global{Name: "T"}},
			expected: true,
		},
		{
			name:     "In Fst",
			global:   "T",
			term:     ast.Fst{P: ast.Global{Name: "T"}},
			expected: true,
		},
		{
			name:     "In Snd",
			global:   "T",
			term:     ast.Snd{P: ast.Global{Name: "T"}},
			expected: true,
		},
		{
			name:     "In Let Ann",
			global:   "T",
			term:     ast.Let{Binder: "x", Ann: ast.Global{Name: "T"}, Val: ast.Var{Ix: 0}, Body: ast.Var{Ix: 0}},
			expected: true,
		},
		{
			name:     "In Let Val",
			global:   "T",
			term:     ast.Let{Binder: "x", Ann: ast.Sort{U: 0}, Val: ast.Global{Name: "T"}, Body: ast.Var{Ix: 0}},
			expected: true,
		},
		{
			name:     "In Let Body",
			global:   "T",
			term:     ast.Let{Binder: "x", Ann: ast.Sort{U: 0}, Val: ast.Var{Ix: 0}, Body: ast.Global{Name: "T"}},
			expected: true,
		},
		{
			name:     "In Id A",
			global:   "T",
			term:     ast.Id{A: ast.Global{Name: "T"}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}},
			expected: true,
		},
		{
			name:     "In Refl A",
			global:   "T",
			term:     ast.Refl{A: ast.Global{Name: "T"}, X: ast.Var{Ix: 0}},
			expected: true,
		},
		{
			name:     "In Refl X",
			global:   "T",
			term:     ast.Refl{A: ast.Sort{U: 0}, X: ast.Global{Name: "T"}},
			expected: true,
		},
		{
			name:   "In J A",
			global: "T",
			term: ast.J{
				A: ast.Global{Name: "T"},
				C: ast.Var{Ix: 0}, D: ast.Var{Ix: 0},
				X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}, P: ast.Var{Ix: 0},
			},
			expected: true,
		},
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

// contains is a simple helper for string containment check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
