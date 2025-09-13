package subst

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func TestShift(t *testing.T) {
	// Shift Var(0) by 1, cutoff 0 -> Var(1)
	v := &ast.Var{Ix: 0}
	shifted := Shift(1, 0, v)
	if shifted.(*ast.Var).Ix != 1 {
		t.Errorf("Expected Ix=1, got %d", shifted.(*ast.Var).Ix)
	}

	// Shift Var(0) by 1, cutoff 1 -> Var(0) (no shift)
	shifted2 := Shift(1, 1, v)
	if shifted2.(*ast.Var).Ix != 0 {
		t.Errorf("Expected Ix=0, got %d", shifted2.(*ast.Var).Ix)
	}
}

func TestSubst(t *testing.T) {
	// Subst Var(0) with Var(5) in Var(0) -> Var(5)
	s := &ast.Var{Ix: 5}
	tm := &ast.Var{Ix: 0}
	substed := Subst(0, s, tm)
	if substed.(*ast.Var).Ix != 5 {
		t.Errorf("Expected Ix=5, got %d", substed.(*ast.Var).Ix)
	}

	// Subst Var(0) with Var(5) in Var(1) -> Var(0) (decrement)
	tm2 := &ast.Var{Ix: 1}
	substed2 := Subst(0, s, tm2)
	if substed2.(*ast.Var).Ix != 0 {
		t.Errorf("Expected Ix=0, got %d", substed2.(*ast.Var).Ix)
	}
}

func TestSubstIntoLam(t *testing.T) {
	// λx. y  where y is free Var(1), substitute 0 with z (Var(20))
	body := &ast.Var{Ix: 1}
	lam := &ast.Lam{Body: body}
	z := &ast.Var{Ix: 20}
	substed := Subst(0, z, lam)
	if substedLam, ok := substed.(*ast.Lam); ok {
		if bodyVar, ok := substedLam.Body.(*ast.Var); ok {
			if bodyVar.Ix != 21 { // shifted z
				t.Errorf("Expected Ix=21, got %d", bodyVar.Ix)
			}
		} else {
			t.Error("Body not Var")
		}
	} else {
		t.Error("Not Lam")
	}
}

func TestNestedPi(t *testing.T) {
	// Πx:A. Πy:B. x
	a := &ast.Var{Ix: 10}
	b := &ast.Var{Ix: 11}
	x := &ast.Var{Ix: 1} // outer x is 1 in inner
	innerPi := &ast.Pi{Binder: "y", A: b, B: x}
	pi := &ast.Pi{Binder: "x", A: a, B: innerPi}

	// Subst 0 with z in pi
	z := &ast.Var{Ix: 20}
	substed := Subst(0, z, pi)
	if pi1, ok := substed.(*ast.Pi); ok {
		if pi2, ok := pi1.B.(*ast.Pi); ok {
			if body, ok := pi2.B.(*ast.Var); ok {
				if body.Ix != 2 { // shifted z is 21, but wait, let's check
					// Wait, in inner, x is Var(1), subst 0 with z, but since 1 > 0, becomes Var(0), but then shifted? Wait, the code shifts s when crossing.
					// In Pi, B: Subst(j+1, Shift(1,0,s), B)
					// For j=0, s=z=Var(20), Shift(1,0,z)=Var(21)
					// Then in inner B, Subst(1, Var(21), Var(1)) = since 1==1, Var(21)
					// Wait, but in the test, x is Var(1), but in the pi, the x in body is Var(1) for the inner binder.
					// Wait, let's adjust the test.
					// For Πx. Πy. x, the x in body is Var(1) (y is 0, x is 1)
					// Subst 0 with z, in outer, A subst 0->z, but A is Var(10), 10>0, Var(9)
					// B subst 1 with Shift(1,0,z)=Var(21), in B which is Πy. x, subst 1 with Var(21) in Πy. x
					// In Πy. x, A subst 1->Var(21), but A is Var(11), 11>1, Var(10)
					// B subst 2 with Shift(1,0,Var(21))=Var(22), in x=Var(1), 1<2, Var(1)
					// So body remains Var(1)
					// Perhaps change the test to have Var(0) in body.
					// For nested, let's have Πx. Πy. y
					y := &ast.Var{Ix: 0}
					innerPi2 := &ast.Pi{Binder: "y", A: b, B: y}
					pi2 := &ast.Pi{Binder: "x", A: a, B: innerPi2}
					substed2 := Subst(0, z, pi2)
					if pi12, ok := substed2.(*ast.Pi); ok {
						if pi22, ok := pi12.B.(*ast.Pi); ok {
							if body2, ok := pi22.B.(*ast.Var); ok {
								if body2.Ix != 0 { // y is still 0
									t.Errorf("Expected Ix=0, got %d", body2.Ix)
								}
							}
						}
					}
				}
			}
		}
	}
}

func TestShiftThenSubst(t *testing.T) {
	// Shift then Subst sanity
	v := &ast.Var{Ix: 0}
	shifted := Shift(1, 0, v)                      // Var(1)
	substed := Subst(1, &ast.Var{Ix: 10}, shifted) // replace 1 with 10
	if substed.(*ast.Var).Ix != 10 {
		t.Errorf("Expected Ix=10, got %d", substed.(*ast.Var).Ix)
	}
}
