package eval_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/watchthelight/hypergraphgo/internal/ast"
	"github.com/watchthelight/hypergraphgo/internal/eval"
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
	if got, want := nf(ast.Fst{P: p}), "a"; got != want {
		t.Fatalf("fst got %q", got)
	}
	if got, want := nf(ast.Snd{P: p}), "b"; got != want {
		t.Fatalf("snd got %q", got)
	}
}

func TestNormalize_NeutralStaysNeutral(t *testing.T) {
	f := ast.Global{Name: "f"}
	arg := ast.Var{Ix: 0}
	got := nf(ast.App{T: f, U: arg})
	// Expect printed neutral app: (f {0})
	if got != "(f {0})" {
		t.Fatalf("neutral app printed as %q", got)
	}
}

func TestNormalize_AppSpineAssociativity(t *testing.T) {
	f := ast.Global{Name: "f"}
	l := ast.MkApps(f, ast.Var{Ix: 0}, ast.Var{Ix: 1})
	r := ast.App{T: ast.App{T: f, U: ast.Var{Ix: 0}}, U: ast.Var{Ix: 1}}
	if got, want := nf(l), "(f {0} {1}"; got != want {
		t.Fatalf("lhs: got %q, want %q", got, want)
	}
	if got, want := nf(r), "(f {0} {1}"; got != want {
		t.Fatalf("rhs: got %q, want %q", got, want)
	}
}

// ------- Golden tests -------

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
		src, err := os.ReadFile(inPath)
		if err != nil {
			t.Fatal(err)
		}
		// Trivial "parser": the file contains one S-expr-like core literal name we assemble by hand for now.
		// Cases are simple and built in code below; we map known labels to terms.
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
		want, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Fatalf("missing golden: %s", goldenPath)
		}
		if got != strings.TrimSpace(string(want)) {
			t.Fatalf("%s: got %q, want %q", e.Name(), got, strings.TrimSpace(string(want)))
		}
	}
}
