package ctx

import (
	"testing"

	"github.com/watchthelight/hypergraphgo/internal/ast"
)

func TestLookupVar(t *testing.T) {
	var c Ctx
	ty1 := ast.Sort{U: 0}
	ty2 := ast.Sort{U: 1}

	c.Extend("x", ty1)
	c.Extend("y", ty2)

	// ix=0 is most recent, y
	got, ok := c.LookupVar(0)
	if !ok || got != ty2 {
		t.Errorf("LookupVar(0) = %v, %v; want %v, true", got, ok, ty2)
	}

	// ix=1 is x
	got, ok = c.LookupVar(1)
	if !ok || got != ty1 {
		t.Errorf("LookupVar(1) = %v, %v; want %v, true", got, ok, ty1)
	}

	// ix=2 out of range
	_, ok = c.LookupVar(2)
	if ok {
		t.Errorf("LookupVar(2) should be false")
	}

	// negative
	_, ok = c.LookupVar(-1)
	if ok {
		t.Errorf("LookupVar(-1) should be false")
	}
}
