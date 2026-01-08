package eval_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
)

func nf(t ast.Term) string { return ast.Sprint(eval.Normalize(t)) }

func TestNormalize_Beta(t *testing.T) {
	id := ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}}
	app := ast.App{T: id, U: ast.Global{Name: "y"}}
	if got, want := nf(app), "y"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNormalize_Proj(t *testing.T) {
	p := ast.Pair{Fst: ast.Global{Name: "a"}, Snd: ast.Global{Name: "b"}}
	if got := nf(ast.Fst{P: p}); got != "a" {
		t.Fatalf("fst: got %q, want a", got)
	}
	if got := nf(ast.Snd{P: p}); got != "b" {
		t.Fatalf("snd: got %q, want b", got)
	}
}

func TestNormalize_Neutral(t *testing.T) {
	f := ast.Global{Name: "f"}
	arg := ast.Var{Ix: 0}
	got := nf(ast.App{T: f, U: arg})
	if got != "(f {0})" {
		t.Fatalf("neutral app printed as %q", got)
	}
}

func TestNormalize_AppSpine(t *testing.T) {
	f := ast.Global{Name: "f"}
	l := ast.MkApps(f, ast.Var{Ix: 0}, ast.Var{Ix: 1})
	r := ast.App{T: ast.App{T: f, U: ast.Var{Ix: 0}}, U: ast.Var{Ix: 1}}
	if nf(l) != nf(r) {
		t.Fatalf("application spine should normalize to same form")
	}
}

// TestNormalize_FstNeutral tests Fst on a non-pair (stuck case).
func TestNormalize_FstNeutral(t *testing.T) {
	// fst x where x is a variable (neutral) - should stay stuck
	neutralTerm := ast.Global{Name: "x"}
	fstTerm := ast.Fst{P: neutralTerm}
	got := nf(fstTerm)
	// Should produce "(fst x)"
	if !strings.Contains(got, "fst") || !strings.Contains(got, "x") {
		t.Fatalf("fst on neutral: got %q, expected to contain fst and x", got)
	}
}

// TestNormalize_SndNeutral tests Snd on a non-pair (stuck case).
func TestNormalize_SndNeutral(t *testing.T) {
	// snd x where x is a variable (neutral) - should stay stuck
	neutralTerm := ast.Global{Name: "x"}
	sndTerm := ast.Snd{P: neutralTerm}
	got := nf(sndTerm)
	// Should produce "(snd x)"
	if !strings.Contains(got, "snd") || !strings.Contains(got, "x") {
		t.Fatalf("snd on neutral: got %q, expected to contain snd and x", got)
	}
}

// TestNormalize_Default tests the default case for non-reducible terms.
func TestNormalize_Default(t *testing.T) {
	// Sort should pass through unchanged
	sortTerm := ast.Sort{U: 0}
	got := nf(sortTerm)
	if got != "Type0" {
		t.Fatalf("Sort normalization: got %q, want Type0", got)
	}

	// Pi should pass through unchanged
	piTerm := ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	got = nf(piTerm)
	if !strings.Contains(got, "Pi") || !strings.Contains(got, "Type0") {
		t.Fatalf("Pi normalization: got %q, expected to contain Pi and Type0", got)
	}

	// Sigma should pass through unchanged
	sigmaTerm := ast.Sigma{Binder: "x", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	got = nf(sigmaTerm)
	if !strings.Contains(got, "Sigma") || !strings.Contains(got, "Type0") {
		t.Fatalf("Sigma normalization: got %q, expected to contain Sigma and Type0", got)
	}

	// Lambda should pass through unchanged
	lamTerm := ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}}
	got = nf(lamTerm)
	if !strings.Contains(got, "\\") || !strings.Contains(got, "x") {
		t.Fatalf("Lambda normalization: got %q, expected to contain lambda", got)
	}
}

// TestNormalize_NestedBeta tests nested beta reductions.
func TestNormalize_NestedBeta(t *testing.T) {
	// (λx. (λy. y) x) z -> z
	inner := ast.Lam{Binder: "y", Body: ast.Var{Ix: 0}}
	outer := ast.Lam{Binder: "x", Body: ast.App{T: inner, U: ast.Var{Ix: 0}}}
	app := ast.App{T: outer, U: ast.Global{Name: "z"}}
	got := nf(app)
	if got != "z" {
		t.Fatalf("nested beta: got %q, want z", got)
	}
}

// Golden tests
func TestGolden_NormalForms(t *testing.T) {
	dir := filepath.Join("testdata", "nf")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skip("no golden directory:", err)
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".in") {
			continue
		}
		inPath := filepath.Join(dir, e.Name())
		goldenPath := strings.TrimSuffix(inPath, ".in") + ".golden"
		src, _ := os.ReadFile(inPath)
		var tm ast.Term
		switch strings.TrimSpace(string(src)) {
		case "beta-id-y":
			tm = ast.App{T: ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}}, U: ast.Global{Name: "y"}}
		case "fst-pair":
			tm = ast.Fst{P: ast.Pair{Fst: ast.Global{Name: "a"}, Snd: ast.Global{Name: "b"}}}
		case "spine-f-01":
			tm = ast.MkApps(ast.Global{Name: "f"}, ast.Var{Ix: 0}, ast.Var{Ix: 1})
		default:
			t.Fatalf("unknown golden label %q", src)
		}
		got := nf(tm)
		want, _ := os.ReadFile(goldenPath)
		if got != strings.TrimSpace(string(want)) {
			t.Fatalf("%s: got %q, want %q", e.Name(), got, strings.TrimSpace(string(want)))
		}
	}
}
