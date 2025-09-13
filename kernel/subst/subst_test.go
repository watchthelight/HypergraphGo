package subst

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func TestShift(t *testing.T) {
	// Shift free var
	if got := Shift(1, 0, ast.Var{Ix: 0}); got != (ast.Var{Ix: 1}) {
		t.Errorf("Shift(1,0, Var{0}) = %v; want Var{1}", got)
	}

	// Don't shift bound var
	if got := Shift(1, 1, ast.Var{Ix: 0}); got != (ast.Var{Ix: 0}) {
		t.Errorf("Shift(1,1, Var{0}) = %v; want Var{0}", got)
	}

	// Shift free var in Lam body
	lam := ast.Lam{Binder: "x", Body: ast.Var{Ix: 1}}
	shifted := Shift(1, 0, lam)
	if l, ok := shifted.(ast.Lam); !ok || l.Body != (ast.Var{Ix: 2}) {
		t.Errorf("Shift in Lam body = %v; want body Var{2}", shifted)
	}
}

func TestSubst(t *testing.T) {
	s := ast.Sort{U: 0}

	// Subst free var
	if got := Subst(0, s, ast.Var{Ix: 0}); got != s {
		t.Errorf("Subst(0, s, Var{0}) = %v; want %v", got, s)
	}

	// Subst into Lam body: Var{1} is free, becomes shifted s
	lam := ast.Lam{Binder: "x", Body: ast.Var{Ix: 1}}
	got := Subst(0, s, lam)
	expectedBody := Shift(1, 0, s) // s shifted under binder
	if l, ok := got.(ast.Lam); !ok || l.Body != expectedBody {
		t.Errorf("Subst into Lam body = %v; want body %v", got, expectedBody)
	}

	// Subst free var in Pi codomain
	pi := ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 1}}
	gotPi := Subst(0, s, pi)
	expectedB := Shift(1, 0, s)
	if p, ok := gotPi.(ast.Pi); !ok || p.A != (ast.Sort{U: 0}) || p.B != expectedB {
		t.Errorf("Subst in Pi = %v; want B %v", gotPi, expectedB)
	}
}

func TestShiftSubstInteraction(t *testing.T) {
	// Shift then Subst
	v := ast.Var{Ix: 0}
	shifted := Shift(1, 0, v) // Var{1}
	substed := Subst(1, ast.Sort{U: 0}, shifted)
	if substed != (ast.Sort{U: 0}) {
		t.Errorf("Shift then Subst = %v; want Sort{0}", substed)
	}
}
