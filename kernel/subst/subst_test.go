package subst

import (
	"testing"

	"github.com/watchthelight/hypergraphgo/internal/ast"
)

func TestShift(t *testing.T) {
	// Shift(1, 0, Var{0}) = Var{1}
	v0 := ast.Var{Ix: 0}
	result := Shift(1, 0, v0)
	var expected ast.Term = ast.Var{Ix: 1}
	if ast.Sprint(result) != ast.Sprint(expected) {
		t.Errorf("Shift(1, 0, Var{0}) = %s, want %s", ast.Sprint(result), ast.Sprint(expected))
	}

	// Shift(1, 0, Lam{Body: Var{0}}) = Lam{Body: Var{0}}
	lam := ast.Lam{Body: v0}
	result = Shift(1, 0, lam)
	expected = lam
	if ast.Sprint(result) != ast.Sprint(expected) {
		t.Errorf("Shift(1, 0, Lam{Body: Var{0}}) = %s, want %s", ast.Sprint(result), ast.Sprint(expected))
	}

	// Shift(1, 0, Lam{Body: Var{1}}) = Lam{Body: Var{2}}
	lam2 := ast.Lam{Body: ast.Var{Ix: 1}}
	result = Shift(1, 0, lam2)
	expected = ast.Lam{Body: ast.Var{Ix: 2}}
	if ast.Sprint(result) != ast.Sprint(expected) {
		t.Errorf("Shift(1, 0, Lam{Body: Var{1}}) = %s, want %s", ast.Sprint(result), ast.Sprint(expected))
	}
}

func TestSubst(t *testing.T) {
	s := ast.Sort{U: 0}
	v0 := ast.Var{Ix: 0}
	v1 := ast.Var{Ix: 1}

	// Subst(0, s, Var{0}) = s
	result := Subst(0, s, v0)
	if ast.Sprint(result) != ast.Sprint(s) {
		t.Errorf("Subst(0, Sort{0}, Var{0}) = %s, want %s", ast.Sprint(result), ast.Sprint(s))
	}

	// Subst(0, s, Var{1}) = Var{0}
	result = Subst(0, s, v1)
	var expected ast.Term = v0
	if ast.Sprint(result) != ast.Sprint(expected) {
		t.Errorf("Subst(0, Sort{0}, Var{1}) = %s, want %s", ast.Sprint(result), ast.Sprint(expected))
	}

	// Subst(0, s, Lam{Body: Var{0}}) = Lam{Body: Var{0}}
	lam := ast.Lam{Body: v0}
	result = Subst(0, s, lam)
	expected = lam
	if ast.Sprint(result) != ast.Sprint(expected) {
		t.Errorf("Subst(0, Sort{0}, Lam{Body: Var{0}}) = %s, want %s", ast.Sprint(result), ast.Sprint(expected))
	}

	// Subst(0, s, Lam{Body: Var{1}}) = Lam{Body: Shift(1,0,s)}
	lam2 := ast.Lam{Body: v1}
	result = Subst(0, s, lam2)
	expected = ast.Lam{Body: Shift(1, 0, s)}
	if ast.Sprint(result) != ast.Sprint(expected) {
		t.Errorf("Subst(0, Sort{0}, Lam{Body: Var{1}}) = %s, want %s", ast.Sprint(result), ast.Sprint(expected))
	}
}

func TestShiftSubstInteraction(t *testing.T) {
	// Shift then Subst: Shift(1,0, Subst(0, Var{1}, Lam{Body: Var{0}}))
	// First, Subst(0, Var{1}, Lam{Body: Var{0}}) = Lam{Body: Var{0}}
	inner := Subst(0, ast.Var{Ix: 1}, ast.Lam{Body: ast.Var{Ix: 0}})
	// Then Shift(1,0, inner) = Lam{Body: Var{0}}
	result := Shift(1, 0, inner)
	expected := ast.Lam{Body: ast.Var{Ix: 0}}
	if ast.Sprint(result) != ast.Sprint(expected) {
		t.Errorf("Shift after Subst = %s, want %s", ast.Sprint(result), ast.Sprint(expected))
	}
}

func TestNestedPiLam(t *testing.T) {
	// Pi x:Type0. Lam y:x. Var{1}  (after shift)
	// Subst into nested
	pi := ast.Pi{
		A: ast.Sort{U: 0},
		B: ast.Lam{Body: ast.Var{Ix: 1}},
	}
	s := ast.Sort{U: 1}
	result := Subst(0, s, pi)
	expected := ast.Pi{
		A: ast.Sort{U: 0},
		B: ast.Lam{Body: ast.Var{Ix: 1}}, // since Subst(1, Shift(1,0,s), Var{1}) = Var{1}
	}
	if ast.Sprint(result) != ast.Sprint(expected) {
		t.Errorf("Subst in nested Pi Lam = %s, want %s", ast.Sprint(result), ast.Sprint(expected))
	}
}
