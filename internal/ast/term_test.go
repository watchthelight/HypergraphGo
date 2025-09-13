package ast_test

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func TestSprint_Id(t *testing.T) {
	// id : Π(A:Type0). A → A  represented as λ . Var0
	id := ast.Lam{
		Binder: "x",
		Ann:    nil,
		Body:   ast.Var{Ix: 0},
	}
	got := ast.Sprint(id)
	want := "(\\x => {0})"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSprint_Pi(t *testing.T) {
	// (Pi x:Type0 . x)
	pi := ast.Pi{
		Binder: "x",
		A:      ast.Sort{U: 0},
		B:      ast.Var{Ix: 0},
	}
	got := ast.Sprint(pi)
	want := "(Pi x: Type0 . {0})"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSprint_AppChain(t *testing.T) {
	tm := ast.MkApps(ast.Global{Name: "f"}, ast.Var{Ix: 0}, ast.Var{Ix: 1})
	got := ast.Sprint(tm)
	want := "(f {0} {1})"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
