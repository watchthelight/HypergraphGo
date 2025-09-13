package ctx

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func TestLookupVar(t *testing.T) {
	ctx := Ctx{}
	ty1 := ast.Sort{U: 0}
	ty2 := ast.Sort{U: 1}

	ctx.Extend("x", ty1)
	ctx.Extend("y", ty2)

	// ix=0 is most recent, y
	got, ok := ctx.LookupVar(0)
	if !ok {
		t.Errorf("LookupVar(0) should be ok")
	}
	if s, ok := got.(ast.Sort); !ok || s.U != 1 {
		t.Errorf("LookupVar(0) = %v; want Sort{U:1}", got)
	}

	// ix=1 is x
	got, ok = ctx.LookupVar(1)
	if !ok {
		t.Errorf("LookupVar(1) should be ok")
	}
	if s, ok := got.(ast.Sort); !ok || s.U != 0 {
		t.Errorf("LookupVar(1) = %v; want Sort{U:0}", got)
	}

	// out of bounds
	_, ok = ctx.LookupVar(2)
	if ok {
		t.Errorf("LookupVar(2) should be false")
	}
}
