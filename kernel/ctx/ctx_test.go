package ctx

import (
	"testing"

	"github.com/watchthelight/hypergraphgo/internal/ast"
)

func TestLookupVar(t *testing.T) {
	var c Ctx

	// Empty context
	if _, ok := c.LookupVar(0); ok {
		t.Error("Expected false for empty context")
	}

	// Add first binding
	c.Extend("x", ast.Sort{U: 0})
	if _, ok := c.LookupVar(0); !ok {
		t.Error("Expected true for ix=0")
	}

	// Add second binding
	c.Extend("y", ast.Sort{U: 1})
	if _, ok := c.LookupVar(0); !ok {
		t.Error("Expected true for ix=0 after extend")
	}
	if _, ok := c.LookupVar(1); !ok {
		t.Error("Expected true for ix=1")
	}
	if _, ok := c.LookupVar(2); ok {
		t.Error("Expected false for ix=2")
	}
}
