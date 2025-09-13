package ctx

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func TestLookupVar(t *testing.T) {
	c := Ctx{}
	ty1 := ast.Sort{U: 0}
	ty2 := ast.Sort{U: 1}

	c.Extend("x", ty1)
	c.Extend("y", ty2)

	// ix=0: most recent, y
	if got, ok := c.LookupVar(0); !ok {
		t.Errorf("LookupVar(0) should succeed")
	} else if s, ok := got.(ast.Sort); !ok || s.U != 1 {
		t.Errorf("LookupVar(0) = %v; want Sort{U:1}", got)
	}

	// ix=1: x
	if got, ok := c.LookupVar(1); !ok {
		t.Errorf("LookupVar(1) should succeed")
	} else if s, ok := got.(ast.Sort); !ok || s.U != 0 {
		t.Errorf("LookupVar(1) = %v; want Sort{U:0}", got)
	}

	// ix=2: out of bounds
	if _, ok := c.LookupVar(2); ok {
		t.Errorf("LookupVar(2) should be false")
	}
}
