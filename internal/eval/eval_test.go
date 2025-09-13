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
