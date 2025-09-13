package subst

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func TestShift(t *testing.T) {
	// Shift Var 0 by 1, cutoff 0: becomes Var 1
	v := ast.Var{Ix: 0}
	got := Shift(1, 0, v)
	if got.(ast.Var).Ix != 1 {
		t.Errorf("Shift(1, 0, Var{0}) = Ix %d; want 1", got.(ast.Var).Ix)
	}

	// Shift under binder: Lam body Var 0 should remain 0
	lam := ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}}
	got = Shift(1, 0, lam)
	body := got.(ast.Lam).Body
	if body.(ast.Var).Ix != 0 {
		t.Errorf("Shift under Lam: body Ix = %d; want 0", body.(ast.Var).Ix)
	}
}

func TestSubst(t *testing.T) {
	// Subst 0 with Sort in Var 0: becomes Sort
	s := ast.Sort{U: 0}
	v := ast.Var{Ix: 0}
	got := Subst(0, s, v)
	if _, ok := got.(ast.Sort); !ok {
		t.Errorf("Subst(0, Sort, Var{0}) = %T; want Sort", got)
	}

	// Subst 0 with Var 0 in Var 1: Var 1 becomes Var 0
	got = Subst(0, ast.Var{Ix: 0}, ast.Var{Ix: 1})
	if got.(ast.Var).Ix != 0 {
		t.Errorf("Subst(0, Var{0}, Var{1}) = Ix %d; want 0", got.(ast.Var).Ix)
	}
}

func TestSubstIntoLam(t *testing.T) {
	// Subst 0 with Var 0 in Lam{Body: Var{1}}: body becomes Var{1} (shifted)
	lam := ast.Lam{Binder: "x", Body: ast.Var{Ix: 1}}
	s := ast.Var{Ix: 0}
	got := Subst(0, s, lam)
	body := got.(ast.Lam).Body
	if body.(ast.Var).Ix != 1 {
		t.Errorf("Subst into Lam body: Ix = %d; want 1", body.(ast.Var).Ix)
	}
}

func TestNestedPi(t *testing.T) {
	// Pi x: Sort0. Pi y: Sort0. Var{1}  (x)
	pi := ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Pi{Binder: "y", A: ast.Sort{U: 0}, B: ast.Var{Ix: 1}}}
	// Subst 0 with Global in pi
	s := ast.Global{Name: "A"}
	got := Subst(0, s, pi)
	// Inner B should remain Var{1} (x is still index 1)
	innerB := got.(ast.Pi).B.(ast.Pi).B
	if v, ok := innerB.(ast.Var); !ok || v.Ix != 1 {
		t.Errorf("Nested Pi Subst: inner B = %v; want Var{Ix:1}", innerB)
	}
}

func TestShiftThenSubst(t *testing.T) {
	// Shift Var 0 to Var 1, then Subst 1 with Sort: becomes Sort
	v := ast.Var{Ix: 0}
	shifted := Shift(1, 0, v)
	substed := Subst(1, ast.Sort{U: 0}, shifted)
	if _, ok := substed.(ast.Sort); !ok {
		t.Errorf("Shift then Subst: %T; want Sort", substed)
	}
}
