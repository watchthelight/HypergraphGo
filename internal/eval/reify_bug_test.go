package eval

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// TestReifyFstWithSpineGt1 tests Bug #3: fst neutral with spine > 1
// If p : Σ(f: A → B). C, then fst p : A → B, so (fst p) arg is valid.
// The reified result should be App{Fst{p}, arg}, NOT App{App{Global{fst}, p}, arg}.
func TestReifyFstWithSpineGt1(t *testing.T) {
	// Create a fst neutral with 2 args: fst(p) applied to arg
	neutral := VNeutral{
		N: Neutral{
			Head: Head{Glob: "fst"},
			Sp:   []Value{VGlobal{Name: "p"}, VGlobal{Name: "arg"}},
		},
	}

	result := Reify(neutral)

	// Should be App{Fst{P: Global{p}}, Global{arg}}
	app, ok := result.(ast.App)
	if !ok {
		t.Fatalf("expected App, got %T: %v", result, result)
	}

	// The function part should be Fst{P: ...}, not Global{fst} applied to p
	if fstTerm, ok := app.T.(ast.Fst); ok {
		// Good - it's Fst
		if g, ok := fstTerm.P.(ast.Global); !ok || g.Name != "p" {
			t.Errorf("expected Fst{P: Global{p}}, got Fst{P: %v}", fstTerm.P)
		}
	} else if innerApp, ok := app.T.(ast.App); ok {
		// Bug: it created App{Global{fst}, p} instead of Fst{p}
		t.Errorf("BUG CONFIRMED: got nested App instead of Fst: App{%T, ...}", innerApp.T)
	} else {
		t.Errorf("unexpected structure: got %T for inner term", app.T)
	}
}

// TestReifySndWithSpineGt1 tests the same bug for snd
func TestReifySndWithSpineGt1(t *testing.T) {
	neutral := VNeutral{
		N: Neutral{
			Head: Head{Glob: "snd"},
			Sp:   []Value{VGlobal{Name: "p"}, VGlobal{Name: "arg"}},
		},
	}

	result := Reify(neutral)

	app, ok := result.(ast.App)
	if !ok {
		t.Fatalf("expected App, got %T: %v", result, result)
	}

	if sndTerm, ok := app.T.(ast.Snd); ok {
		if g, ok := sndTerm.P.(ast.Global); !ok || g.Name != "p" {
			t.Errorf("expected Snd{P: Global{p}}, got Snd{P: %v}", sndTerm.P)
		}
	} else if innerApp, ok := app.T.(ast.App); ok {
		t.Errorf("BUG CONFIRMED: got nested App instead of Snd: App{%T, ...}", innerApp.T)
	} else {
		t.Errorf("unexpected structure: got %T for inner term", app.T)
	}
}

// TestReifyNestedPi tests Bug #1/#2: nested Pi reification with de Bruijn indices
// Π(A:Type). Π(x:A). A should reify to Pi{Sort{0}, Pi{Var{0}, Var{1}}}
func TestReifyNestedPi(t *testing.T) {
	// Build: Π(A:Type). Π(x:A). A
	// In de Bruijn: Pi{A: Sort{0}, B: Pi{A: Var{0}, B: Var{1}}}
	// The inner A (in body position) refers to index 1 (skip x, get A)

	innerPiTerm := ast.Pi{
		Binder: "x",
		A:      ast.Var{Ix: 0}, // A (outer bound var)
		B:      ast.Var{Ix: 1}, // A again (skip x, get A)
	}

	outerPi := VPi{
		A: VSort{Level: 0},
		B: &Closure{Env: &Env{Bindings: nil}, Term: innerPiTerm},
	}

	result := Reify(outerPi)

	pi, ok := result.(ast.Pi)
	if !ok {
		t.Fatalf("expected outer Pi, got %T: %v", result, result)
	}

	// Domain should be Sort{0}
	if s, ok := pi.A.(ast.Sort); !ok || s.U != 0 {
		t.Errorf("expected Sort{0} for domain, got %v", pi.A)
	}

	innerPi, ok := pi.B.(ast.Pi)
	if !ok {
		t.Fatalf("expected inner Pi, got %T: %v", pi.B, pi.B)
	}

	// Inner domain should be Var{0} (referring to outer A)
	if v, ok := innerPi.A.(ast.Var); !ok || v.Ix != 0 {
		t.Errorf("expected Var{0} for inner domain, got %v", innerPi.A)
	}

	// Inner body should be Var{1} (A under two binders)
	if v, ok := innerPi.B.(ast.Var); ok {
		if v.Ix != 1 {
			t.Errorf("BUG: inner body should be Var{1}, got Var{%d}", v.Ix)
		}
	} else {
		t.Errorf("expected Var for inner body, got %T: %v", innerPi.B, innerPi.B)
	}
}

// TestReifyNestedPiThroughEval tests the same via full Eval->Reify cycle
func TestReifyNestedPiThroughEval(t *testing.T) {
	// Π(A:Type). Π(x:A). A
	term := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "x",
			A:      ast.Var{Ix: 0}, // A
			B:      ast.Var{Ix: 1}, // A (skip x)
		},
	}

	val := Eval(&Env{}, term)
	result := Reify(val)

	// Parse result structure
	pi, ok := result.(ast.Pi)
	if !ok {
		t.Fatalf("expected Pi, got %T", result)
	}

	innerPi, ok := pi.B.(ast.Pi)
	if !ok {
		t.Fatalf("expected inner Pi, got %T", pi.B)
	}

	// Check inner body is Var{1}
	if v, ok := innerPi.B.(ast.Var); ok {
		if v.Ix != 1 {
			t.Errorf("BUG: inner body should be Var{1}, got Var{%d}", v.Ix)
		}
	} else {
		t.Errorf("expected Var, got %T", innerPi.B)
	}
}
