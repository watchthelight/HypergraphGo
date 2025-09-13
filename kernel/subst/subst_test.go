package subst

import (
	"testing"

	"github.com/watchthelight/hypergraphgo/internal/ast"
)

func TestShift(t *testing.T) {
	// Shift free variable
	v0 := ast.Var{Ix: 0}
	shifted := Shift(1, 0, v0)
	if shifted.(ast.Var).Ix != 1 {
		t.Errorf("Expected Ix=1, got %v", shifted)
	}

	// Don't shift bound
	shifted2 := Shift(1, 1, v0)
	if shifted2.(ast.Var).Ix != 0 {
		t.Errorf("Expected Ix=0, got %v", shifted2)
	}

	// Shift free var in lambda body
	lam := ast.Lam{Binder: "x", Body: ast.Var{Ix: 1}}
	shiftedLam := Shift(1, 0, lam)
	if shiftedLam.(ast.Lam).Body.(ast.Var).Ix != 2 {
		t.Errorf("Expected body Ix=2, got %v", shiftedLam)
	}
}

func TestSubst(t *testing.T) {
	// Substitute in var
	v0 := ast.Var{Ix: 0}
	s := ast.Sort{U: 0}
	subst := Subst(0, s, v0)
	if _, ok := subst.(ast.Sort); !ok {
		t.Errorf("Expected Sort, got %v", subst)
	}

	// Don't substitute bound
	lam := ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}}
	substLam := Subst(0, s, lam)
	if substLam.(ast.Lam).Body.(ast.Var).Ix != 0 {
		t.Errorf("Expected body Ix=0, got %v", substLam)
	}

	// Substitute free in body
	lam2 := ast.Lam{Binder: "x", Body: ast.Var{Ix: 1}}
	substLam2 := Subst(0, s, lam2)
	if _, ok := substLam2.(ast.Lam).Body.(ast.Sort); !ok {
		t.Errorf("Expected body Sort, got %v", substLam2)
	}
}

func TestShiftSubstInteraction(t *testing.T) {
	// Shift then Subst
	v1 := ast.Var{Ix: 1}
	shifted := Shift(1, 0, v1) // Ix=2
	s := ast.Sort{U: 0}
	subst := Subst(0, s, shifted.(ast.Var)) // Ix=2 >0, Ix=1
	if subst.(ast.Var).Ix != 1 {
		t.Errorf("Expected Ix=1, got %v", subst)
	}
}

func TestNestedPiLam(t *testing.T) {
	// Pi x:Type0. Pi y:x . z where z is free Var{3}
	pi := ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Pi{Binder: "y", A: ast.Var{Ix: 0}, B: ast.Var{Ix: 3}}}
	shifted := Shift(1, 0, pi)
	// Check that free var in inner is shifted
	innerPi := shifted.(ast.Pi).B.(ast.Pi)
	if innerPi.A.(ast.Var).Ix != 0 || innerPi.B.(ast.Var).Ix != 4 {
		t.Errorf("Expected A:Ix=0, B:Ix=4, got A:%v, B:%v", innerPi.A, innerPi.B)
	}

	// Subst free var
	substPi := Subst(0, ast.Sort{U: 1}, pi)
	// Outer A remains Sort{0}, inner A Var{0}, inner B Var{3} >0 becomes Var{2}
	innerSubst := substPi.(ast.Pi).B.(ast.Pi)
	if innerSubst.A.(ast.Var).Ix != 0 || innerSubst.B.(ast.Var).Ix != 2 {
		t.Errorf("Expected A:Ix=0, B:Ix=2, got A:%v, B:%v", innerSubst.A, innerSubst.B)
	}
}
