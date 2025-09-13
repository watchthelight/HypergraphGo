package ast_test

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

type fakeGlobals map[string]struct{}

func (g fakeGlobals) Has(name string) bool {
	_, ok := g[name]
	return ok
}

func TestResolve_PiLamApp(t *testing.T) {
	// Raw: (\x => x) y  where y is global
	raw := ast.RApp{
		T: ast.RLam{Binder: "x", Body: ast.RVar{Name: "x"}},
		U: ast.RGlobal{Name: "y"},
	}
	core, err := ast.Resolve(fakeGlobals{"y": {}}, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	got := ast.Sprint(core)
	want := "((\\x => {0}) y)"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolve_ScopeShadow(t *testing.T) {
	// let x = {global x} in (\x => {0})  should resolve inner x to Var0 and let body Var0 to bound x
	raw := ast.RLet{
		Binder: "x",
		Val:    ast.RGlobal{Name: "x"},
		Body:   ast.RLam{Binder: "x", Body: ast.RVar{Name: "x"}},
	}
	core, err := ast.Resolve(fakeGlobals{"x": {}}, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	got := ast.Sprint(core)
	want := "(let x = x in (\\x => {0}))"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolve_Unbound(t *testing.T) {
	_, err := ast.Resolve(nil, nil, ast.RVar{Name: "ghost"})
	if err == nil {
		t.Fatal("expected error for unbound variable")
	}
}
