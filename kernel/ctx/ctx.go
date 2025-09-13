package ctx

import (
	"github.com/watchthelight/hypergraphgo/internal/ast"
)

// Binding represents a binding in the context.
type Binding struct {
	Name string
	Ty   ast.Term
}

// Ctx represents the typing context.
type Ctx struct {
	Tele []Binding
}

// Len returns the length of the context.
func (c Ctx) Len() int {
	return len(c.Tele)
}

// LookupVar looks up the type of a de Bruijn variable.
// ix=0 is the most recent binding.
func (c Ctx) LookupVar(ix int) (ast.Term, bool) {
	if ix < 0 || ix >= len(c.Tele) {
		return nil, false
	}
	return c.Tele[len(c.Tele)-1-ix].Ty, true
}

// Extend adds a new binding to the context.
func (c *Ctx) Extend(name string, ty ast.Term) {
	c.Tele = append(c.Tele, Binding{Name: name, Ty: ty})
}

// Drop removes the most recent binding from the context.
func (c Ctx) Drop() Ctx {
	if len(c.Tele) == 0 {
		return c
	}
	return Ctx{Tele: c.Tele[:len(c.Tele)-1]}
}
