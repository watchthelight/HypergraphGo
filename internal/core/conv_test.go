package core_test

import (
	"testing"

	"github.com/watchthelight/hypergraphgo/internal/ast"
	"github.com/watchthelight/hypergraphgo/internal/core"
)

func lamId() ast.Term { return ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}} }

func TestConv_Beta(t *testing.T) {
	id := lamId()
	app := ast.App{T: id, U: ast.Global{Name: "y"}}
	if !core.Conv(app, ast.Global{Name: "y"}, core.EtaFlags{}) {
		t.Fatal("beta reduction should yield convertibility")
	}
}

func TestConv_NonEqual(t *testing.T) {
	a := ast.Pair{Fst: ast.Global{Name: "a"}, Snd: ast.Global{Name: "b"}}
	b := ast.Pair{Fst: ast.Global{Name: "a"}, Snd: ast.Global{Name: "c"}}
	if core.Conv(a, b, core.EtaFlags{}) {
		t.Fatal("different pairs must not be convertible")
	}
}

func TestConv_AppSpine(t *testing.T) {
	f := ast.Global{Name: "f"}
	l := ast.MkApps(f, ast.Var{Ix: 0}, ast.Var{Ix: 1})
	r := ast.App{T: ast.App{T: f, U: ast.Var{Ix: 0}}, U: ast.Var{Ix: 1}}
	if !core.Conv(l, r, core.EtaFlags{}) {
		t.Fatal("application spine associativity should normalize")
	}
}
