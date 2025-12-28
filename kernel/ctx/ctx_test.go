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

func TestLen(t *testing.T) {
	// Test empty context
	c := Ctx{}
	if got := c.Len(); got != 0 {
		t.Errorf("Len() on empty context = %d; want 0", got)
	}

	// Test after single Extend
	c.Extend("x", ast.Sort{U: 0})
	if got := c.Len(); got != 1 {
		t.Errorf("Len() after one Extend = %d; want 1", got)
	}

	// Test after multiple Extends
	c.Extend("y", ast.Sort{U: 1})
	c.Extend("z", ast.Sort{U: 2})
	if got := c.Len(); got != 3 {
		t.Errorf("Len() after three Extends = %d; want 3", got)
	}
}

func TestDrop(t *testing.T) {
	// Test Drop on empty context returns empty
	c := Ctx{}
	dropped := c.Drop()
	if dropped.Len() != 0 {
		t.Errorf("Drop() on empty context should return empty, got Len=%d", dropped.Len())
	}

	// Test Drop restores previous state
	c.Extend("x", ast.Sort{U: 0})
	c.Extend("y", ast.Sort{U: 1})

	dropped = c.Drop()
	if dropped.Len() != 1 {
		t.Errorf("Drop() should reduce Len to 1, got %d", dropped.Len())
	}

	// Verify the remaining binding is "x"
	if ty, ok := dropped.LookupVar(0); !ok {
		t.Errorf("LookupVar(0) after Drop should succeed")
	} else if s, ok := ty.(ast.Sort); !ok || s.U != 0 {
		t.Errorf("LookupVar(0) after Drop = %v; want Sort{U:0}", ty)
	}

	// Original context unchanged (value semantics)
	if c.Len() != 2 {
		t.Errorf("Original context should still have Len=2, got %d", c.Len())
	}
}

func TestLookupVarNegativeIndex(t *testing.T) {
	c := Ctx{}
	c.Extend("x", ast.Sort{U: 0})

	// Negative index should return false
	if _, ok := c.LookupVar(-1); ok {
		t.Errorf("LookupVar(-1) should return false")
	}
}

func TestLookupVarEmptyContext(t *testing.T) {
	c := Ctx{}

	// Any index on empty context should return false
	if _, ok := c.LookupVar(0); ok {
		t.Errorf("LookupVar(0) on empty context should return false")
	}
}

func TestChainedExtendDrop(t *testing.T) {
	c := Ctx{}

	// Build up context
	c.Extend("a", ast.Sort{U: 0})
	c.Extend("b", ast.Sort{U: 1})
	c.Extend("c", ast.Sort{U: 2})

	if c.Len() != 3 {
		t.Fatalf("Expected Len=3, got %d", c.Len())
	}

	// Drop one at a time and verify
	c1 := c.Drop()
	if c1.Len() != 2 {
		t.Errorf("After first Drop, expected Len=2, got %d", c1.Len())
	}

	c2 := c1.Drop()
	if c2.Len() != 1 {
		t.Errorf("After second Drop, expected Len=1, got %d", c2.Len())
	}

	c3 := c2.Drop()
	if c3.Len() != 0 {
		t.Errorf("After third Drop, expected Len=0, got %d", c3.Len())
	}

	// Dropping from empty should stay empty
	c4 := c3.Drop()
	if c4.Len() != 0 {
		t.Errorf("Drop on empty should stay empty, got Len=%d", c4.Len())
	}
}

func TestExtendNilTypePanics(t *testing.T) {
	c := Ctx{}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Extend(nil) should panic")
		}
	}()

	c.Extend("x", nil)
}
