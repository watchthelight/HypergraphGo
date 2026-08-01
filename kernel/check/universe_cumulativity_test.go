package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/core"
)

// These tests lock in the kernel's universe semantics: the tower is
// predicative (Sort i inhabits Sort i+1) and there is no cumulativity.
// NewChecker disables CumulativeUniv on purpose; path endpoint comparison
// needs exact equality, and folding subtyping into conversion without a
// separate judgment would be unsound. If cumulativity ever lands, it must
// arrive as its own subtyping judgment and these expectations must be
// revisited deliberately rather than drifting.

func TestCheckAcceptsTowerStep(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())

	// Sort 0 : Sort 1 is the tower itself and must hold.
	if err := checker.Check(emptyCtx(), NoSpan(), ast.Sort{U: 0}, ast.Sort{U: 1}); err != nil {
		t.Fatalf("Sort 0 should check against Sort 1: %v", err)
	}
}

func TestCheckRejectsCumulativeSortLift(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())

	// In a cumulative system Sort 0 : Sort 2 would hold. Here it must not:
	// Sort 0 synthesizes exactly Sort 1, and conversion is exact equality.
	if err := checker.Check(emptyCtx(), NoSpan(), ast.Sort{U: 0}, ast.Sort{U: 2}); err == nil {
		t.Fatal("Sort 0 checked against Sort 2; cumulative lifting must be rejected")
	}
}

func TestCheckRejectsInhabitantLift(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())
	checker.Globals().AddAxiom("A", ast.Sort{U: 0})

	// A lives in Sort 0 and must keep living there.
	if err := checker.Check(emptyCtx(), NoSpan(), ast.Global{Name: "A"}, ast.Sort{U: 0}); err != nil {
		t.Fatalf("A : Sort 0 should check: %v", err)
	}

	// Lifting A into Sort 1 is exactly what cumulativity would permit.
	if err := checker.Check(emptyCtx(), NoSpan(), ast.Global{Name: "A"}, ast.Sort{U: 1}); err == nil {
		t.Fatal("A : Sort 1 checked; inhabitant lifting must be rejected")
	}
}

func TestConvCumulativeFlagStaysReachable(t *testing.T) {
	env := core.NewEnv()
	exact := core.ConvOptions{}
	cumulative := core.ConvOptions{CumulativeUniv: true}

	// Without the flag, distinct levels never convert.
	if core.Conv(env, ast.Sort{U: 0}, ast.Sort{U: 1}, exact) {
		t.Fatal("Sort 0 and Sort 1 converted under exact equality")
	}

	// Equal levels convert under either mode.
	if !core.Conv(env, ast.Sort{U: 1}, ast.Sort{U: 1}, exact) {
		t.Fatal("Sort 1 failed to convert with itself")
	}

	// The flagged mode implements directional lifting (left into right).
	// The kernel never enables it, but future subtyping work builds here,
	// so pin down what the flag currently means.
	if !core.Conv(env, ast.Sort{U: 0}, ast.Sort{U: 1}, cumulative) {
		t.Fatal("cumulative mode rejected Sort 0 into Sort 1")
	}
	if core.Conv(env, ast.Sort{U: 1}, ast.Sort{U: 0}, cumulative) {
		t.Fatal("cumulative mode accepted Sort 1 into Sort 0; lifting must stay directional")
	}
}
