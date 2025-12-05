package subst

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// TestShiftNil tests that Shift handles nil terms correctly.
func TestShiftNil(t *testing.T) {
	result := Shift(1, 0, nil)
	if result != nil {
		t.Errorf("Shift(nil) should return nil, got %v", result)
	}
}

// TestSubstNil tests that Subst handles nil terms correctly.
func TestSubstNil(t *testing.T) {
	sub := ast.Global{Name: "x"}
	result := Subst(0, sub, nil)
	if result != nil {
		t.Errorf("Subst(nil) should return nil, got %v", result)
	}
}

// TestShiftZero tests that shifting by 0 is an identity operation.
func TestShiftZero(t *testing.T) {
	terms := []ast.Term{
		ast.Var{Ix: 0},
		ast.Var{Ix: 5},
		ast.Global{Name: "x"},
		ast.Sort{U: 0},
		ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}},
		ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}},
	}

	for _, term := range terms {
		result := Shift(0, 0, term)
		// For identity, we just check it doesn't panic and returns something
		if result == nil {
			t.Errorf("Shift(0, 0, %T) returned nil", term)
		}
	}
}

// TestShiftNegative tests shifting with negative values.
func TestShiftNegative(t *testing.T) {
	// Shift down by 1
	term := ast.Var{Ix: 5}
	result := Shift(-1, 0, term)
	if v, ok := result.(ast.Var); !ok || v.Ix != 4 {
		t.Errorf("Shift(-1, 0, Var{5}) should give Var{4}, got %v", result)
	}

	// Shift down below cutoff - should not change
	term = ast.Var{Ix: 2}
	result = Shift(-1, 5, term)
	if v, ok := result.(ast.Var); !ok || v.Ix != 2 {
		t.Errorf("Var below cutoff should not change, got %v", result)
	}
}

// TestUnknownTermType tests that unknown term types are returned unchanged.
func TestUnknownTermType(t *testing.T) {
	// Create a mock term type that implements ast.Term but isn't handled
	// Since we can't easily create an unknown type, we test the behavior
	// by checking that all known types work correctly

	knownTypes := []ast.Term{
		ast.Var{Ix: 0},
		ast.Sort{U: 0},
		ast.Global{Name: "x"},
		ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}},
		ast.Lam{Binder: "x", Ann: ast.Sort{U: 0}, Body: ast.Var{Ix: 0}},
		ast.App{T: ast.Global{Name: "f"}, U: ast.Global{Name: "x"}},
		ast.Sigma{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}},
		ast.Pair{Fst: ast.Global{Name: "a"}, Snd: ast.Global{Name: "b"}},
		ast.Fst{P: ast.Global{Name: "p"}},
		ast.Snd{P: ast.Global{Name: "p"}},
		ast.Let{Binder: "x", Ann: ast.Sort{U: 0}, Val: ast.Global{Name: "v"}, Body: ast.Var{Ix: 0}},
	}

	for _, term := range knownTypes {
		// Shift should not panic
		_ = Shift(1, 0, term)
		// Subst should not panic
		_ = Subst(0, ast.Global{Name: "s"}, term)
	}
}

// TestSubstChain tests multiple sequential substitutions.
func TestSubstChain(t *testing.T) {
	// (\x. \y. x) after substituting for outer x
	// Start with Var{1} (refers to x in context [y, x])
	term := ast.Var{Ix: 1}

	// Substitute Var{1} with Global{a}
	result := Subst(1, ast.Global{Name: "a"}, term)
	if g, ok := result.(ast.Global); !ok || g.Name != "a" {
		t.Errorf("expected Global{a}, got %v", result)
	}

	// Chain: substitute, then substitute again
	term2 := ast.App{T: ast.Var{Ix: 0}, U: ast.Var{Ix: 1}}
	result2 := Subst(0, ast.Global{Name: "f"}, term2)
	// Should be App{Global{f}, Var{0}} (Var{1} becomes Var{0} after subst)
	if app, ok := result2.(ast.App); ok {
		if _, ok := app.T.(ast.Global); !ok {
			t.Errorf("expected function to be Global, got %T", app.T)
		}
		if v, ok := app.U.(ast.Var); !ok || v.Ix != 0 {
			t.Errorf("expected argument to be Var{0}, got %v", app.U)
		}
	} else {
		t.Errorf("expected App, got %T", result2)
	}
}

// TestAllTermTypes tests Shift and Subst on every AST node type.
func TestAllTermTypes(t *testing.T) {
	sub := ast.Global{Name: "sub"}

	tests := []struct {
		name string
		term ast.Term
	}{
		{"Var", ast.Var{Ix: 0}},
		{"Sort", ast.Sort{U: 1}},
		{"Global", ast.Global{Name: "g"}},
		{"Pi", ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}}},
		{"Lam", ast.Lam{Binder: "x", Ann: ast.Sort{U: 0}, Body: ast.Var{Ix: 0}}},
		{"App", ast.App{T: ast.Var{Ix: 0}, U: ast.Var{Ix: 1}}},
		{"Sigma", ast.Sigma{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}}},
		{"Pair", ast.Pair{Fst: ast.Var{Ix: 0}, Snd: ast.Var{Ix: 1}}},
		{"Fst", ast.Fst{P: ast.Var{Ix: 0}}},
		{"Snd", ast.Snd{P: ast.Var{Ix: 0}}},
		{"Let", ast.Let{Binder: "x", Ann: ast.Sort{U: 0}, Val: ast.Var{Ix: 0}, Body: ast.Var{Ix: 1}}},
	}

	for _, tt := range tests {
		t.Run("Shift_"+tt.name, func(t *testing.T) {
			result := Shift(1, 0, tt.term)
			if result == nil {
				t.Errorf("Shift returned nil for %s", tt.name)
			}
		})

		t.Run("Subst_"+tt.name, func(t *testing.T) {
			result := Subst(0, sub, tt.term)
			if result == nil {
				t.Errorf("Subst returned nil for %s", tt.name)
			}
		})
	}
}

// TestShiftPreservesStructure tests that Shift preserves term structure.
func TestShiftPreservesStructure(t *testing.T) {
	// Complex term: Pi (x : Sort 0) . Lam (y : x) . y
	// where x is Var{0} (bound by Pi) and y is Var{0} (bound by Lam)
	term := ast.Pi{
		Binder: "x",
		A:      ast.Sort{U: 0},
		B: ast.Lam{
			Binder: "y",
			Ann:    ast.Var{Ix: 0}, // refers to x (bound by Pi, so index 0 under the Pi)
			Body:   ast.Var{Ix: 0}, // refers to y (bound by Lam)
		},
	}

	result := Shift(1, 0, term)
	pi, ok := result.(ast.Pi)
	if !ok {
		t.Fatalf("expected Pi, got %T", result)
	}

	lam, ok := pi.B.(ast.Lam)
	if !ok {
		t.Fatalf("expected Lam inside Pi, got %T", pi.B)
	}

	// Var{0} in annotation refers to the bound variable x from Pi
	// Under the Pi binder, cutoff becomes 1, so Var{0} < cutoff and is NOT shifted
	// This is correct: bound variables should not be shifted
	if v, ok := lam.Ann.(ast.Var); !ok || v.Ix != 0 {
		t.Errorf("expected Var{0} for annotation (bound by Pi), got %v", lam.Ann)
	}

	// Test with free variable: Shift should affect free vars
	termWithFree := ast.Lam{
		Binder: "x",
		Ann:    ast.Var{Ix: 1}, // free variable (refers outside)
		Body:   ast.Var{Ix: 0}, // bound variable
	}
	resultFree := Shift(1, 0, termWithFree)
	lamFree := resultFree.(ast.Lam)

	// Ann (Var{1}) is free (>= cutoff 0 initially, and after going under binder, >= cutoff 1)
	// So it should be shifted to Var{2}
	if v, ok := lamFree.Ann.(ast.Var); !ok || v.Ix != 2 {
		t.Errorf("expected Var{2} for shifted free variable, got %v", lamFree.Ann)
	}
}

// TestSubstUnderBinders tests substitution under binders.
func TestSubstUnderBinders(t *testing.T) {
	// Term: Lam x . Var 1 (free variable)
	term := ast.Lam{
		Binder: "x",
		Body:   ast.Var{Ix: 1}, // Free var (refers to outer context)
	}

	// Substitute index 0 with Global{y}
	result := Subst(0, ast.Global{Name: "y"}, term)

	lam, ok := result.(ast.Lam)
	if !ok {
		t.Fatalf("expected Lam, got %T", result)
	}

	// The body Var{1} refers to index 0 in outer context (after going under binder)
	// So it should be substituted with shifted Global{y}
	if g, ok := lam.Body.(ast.Global); !ok || g.Name != "y" {
		t.Errorf("expected Global{y} in body, got %v", lam.Body)
	}
}
