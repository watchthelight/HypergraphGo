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

// ============================================================================
// Len Tests
// ============================================================================

func TestLen_EmptyContext(t *testing.T) {
	t.Parallel()
	c := Ctx{}
	if got := c.Len(); got != 0 {
		t.Errorf("Len() = %d, want 0", got)
	}
}

func TestLen_AfterExtend(t *testing.T) {
	t.Parallel()
	c := Ctx{}
	c.Extend("x", ast.Sort{U: 0})
	if got := c.Len(); got != 1 {
		t.Errorf("Len() = %d, want 1", got)
	}

	c.Extend("y", ast.Sort{U: 1})
	if got := c.Len(); got != 2 {
		t.Errorf("Len() = %d, want 2", got)
	}

	c.Extend("z", ast.Sort{U: 2})
	if got := c.Len(); got != 3 {
		t.Errorf("Len() = %d, want 3", got)
	}
}

// ============================================================================
// Drop Tests
// ============================================================================

func TestDrop_EmptyContext(t *testing.T) {
	t.Parallel()
	c := Ctx{}
	dropped := c.Drop()
	if dropped.Len() != 0 {
		t.Errorf("Drop on empty context: Len() = %d, want 0", dropped.Len())
	}
}

func TestDrop_SingleBinding(t *testing.T) {
	t.Parallel()
	c := Ctx{}
	c.Extend("x", ast.Sort{U: 0})
	dropped := c.Drop()
	if dropped.Len() != 0 {
		t.Errorf("Drop single binding: Len() = %d, want 0", dropped.Len())
	}
}

func TestDrop_MultipleBindings(t *testing.T) {
	t.Parallel()
	c := Ctx{}
	c.Extend("x", ast.Sort{U: 0})
	c.Extend("y", ast.Sort{U: 1})
	c.Extend("z", ast.Sort{U: 2})

	dropped := c.Drop()
	if dropped.Len() != 2 {
		t.Errorf("Drop from 3: Len() = %d, want 2", dropped.Len())
	}

	// Verify x and y remain, z is gone
	// After drop, ix=0 should be y (was ix=1 before drop)
	if got, ok := dropped.LookupVar(0); !ok {
		t.Error("LookupVar(0) after drop should succeed")
	} else if s, ok := got.(ast.Sort); !ok || s.U != 1 {
		t.Errorf("LookupVar(0) = %v, want Sort{U:1}", got)
	}
}

func TestDrop_MultipleTimes(t *testing.T) {
	t.Parallel()
	c := Ctx{}
	c.Extend("a", ast.Sort{U: 0})
	c.Extend("b", ast.Sort{U: 1})
	c.Extend("c", ast.Sort{U: 2})

	// Drop once
	d1 := c.Drop()
	if d1.Len() != 2 {
		t.Errorf("First drop: Len() = %d, want 2", d1.Len())
	}

	// Drop again
	d2 := d1.Drop()
	if d2.Len() != 1 {
		t.Errorf("Second drop: Len() = %d, want 1", d2.Len())
	}

	// Drop to empty
	d3 := d2.Drop()
	if d3.Len() != 0 {
		t.Errorf("Third drop: Len() = %d, want 0", d3.Len())
	}
}

func TestDrop_Immutability(t *testing.T) {
	t.Parallel()
	c := Ctx{}
	c.Extend("x", ast.Sort{U: 0})
	c.Extend("y", ast.Sort{U: 1})

	original := c.Len()
	_ = c.Drop()

	// Original context should be unchanged
	if c.Len() != original {
		t.Errorf("Original Len() changed from %d to %d", original, c.Len())
	}
}

// ============================================================================
// LookupVar Edge Cases
// ============================================================================

func TestLookupVar_NegativeIndex(t *testing.T) {
	t.Parallel()
	c := Ctx{}
	c.Extend("x", ast.Sort{U: 0})

	if _, ok := c.LookupVar(-1); ok {
		t.Error("LookupVar(-1) should return false")
	}
	if _, ok := c.LookupVar(-100); ok {
		t.Error("LookupVar(-100) should return false")
	}
}

func TestLookupVar_EmptyContext(t *testing.T) {
	t.Parallel()
	c := Ctx{}

	if _, ok := c.LookupVar(0); ok {
		t.Error("LookupVar(0) on empty context should return false")
	}
}

func TestLookupVar_BoundaryIndex(t *testing.T) {
	t.Parallel()
	c := Ctx{}
	c.Extend("x", ast.Sort{U: 0})
	c.Extend("y", ast.Sort{U: 1})

	// Index at len-1 should succeed
	if _, ok := c.LookupVar(1); !ok {
		t.Error("LookupVar(len-1) should succeed")
	}

	// Index at len should fail
	if _, ok := c.LookupVar(2); ok {
		t.Error("LookupVar(len) should fail")
	}
}

// ============================================================================
// Extend Tests
// ============================================================================

func TestExtend_Multiple(t *testing.T) {
	t.Parallel()
	c := Ctx{}

	// Extend with different types
	c.Extend("a", ast.Sort{U: 0})
	c.Extend("b", ast.Global{Name: "T"})
	c.Extend("c", ast.Var{Ix: 0})

	if c.Len() != 3 {
		t.Errorf("Len() = %d, want 3", c.Len())
	}

	// Verify ordering (most recent first in de Bruijn)
	// c is ix=0, b is ix=1, a is ix=2
	if got, ok := c.LookupVar(0); !ok {
		t.Error("LookupVar(0) should succeed")
	} else if _, ok := got.(ast.Var); !ok {
		t.Errorf("LookupVar(0) = %T, want ast.Var", got)
	}

	if got, ok := c.LookupVar(1); !ok {
		t.Error("LookupVar(1) should succeed")
	} else if _, ok := got.(ast.Global); !ok {
		t.Errorf("LookupVar(1) = %T, want ast.Global", got)
	}

	if got, ok := c.LookupVar(2); !ok {
		t.Error("LookupVar(2) should succeed")
	} else if _, ok := got.(ast.Sort); !ok {
		t.Errorf("LookupVar(2) = %T, want ast.Sort", got)
	}
}

func TestExtend_DuplicateNames(t *testing.T) {
	t.Parallel()
	c := Ctx{}

	// Same name, different types - should both be stored
	c.Extend("x", ast.Sort{U: 0})
	c.Extend("x", ast.Sort{U: 1})

	if c.Len() != 2 {
		t.Errorf("Len() = %d, want 2", c.Len())
	}

	// Most recent x (ix=0) has U=1
	if got, ok := c.LookupVar(0); !ok {
		t.Error("LookupVar(0) should succeed")
	} else if s, ok := got.(ast.Sort); !ok || s.U != 1 {
		t.Errorf("LookupVar(0) = %v, want Sort{U:1}", got)
	}
}
