package ctx

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func TestLookupVar(t *testing.T) {
	ctx := Ctx{}
	// Add bindings
	ctx.Extend("x", &ast.Var{Ix: 0}) // dummy ty
	ctx.Extend("y", &ast.Var{Ix: 1})

	// Lookup most recent (ix=0)
	ty, ok := ctx.LookupVar(0)
	if !ok {
		t.Error("Expected to find binding for ix=0")
	}
	if ty == nil {
		t.Error("Expected non-nil type")
	}

	// Lookup second (ix=1)
	ty2, ok2 := ctx.LookupVar(1)
	if !ok2 {
		t.Error("Expected to find binding for ix=1")
	}
	if ty2 == nil {
		t.Error("Expected non-nil type")
	}

	// Lookup out of bounds
	_, ok3 := ctx.LookupVar(2)
	if ok3 {
		t.Error("Expected not to find binding for ix=2")
	}
}

func TestLen(t *testing.T) {
	ctx := Ctx{}
	if ctx.Len() != 0 {
		t.Error("Expected length 0")
	}
	ctx.Extend("x", &ast.Var{Ix: 0})
	if ctx.Len() != 1 {
		t.Error("Expected length 1")
	}
}

func TestDrop(t *testing.T) {
	ctx := Ctx{}
	ctx.Extend("x", &ast.Var{Ix: 0})
	ctx.Extend("y", &ast.Var{Ix: 1})
	if ctx.Len() != 2 {
		t.Error("Expected length 2")
	}
	ctx = ctx.Drop()
	if ctx.Len() != 1 {
		t.Error("Expected length 1 after drop")
	}
}
