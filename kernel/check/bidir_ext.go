//go:build !cubical

package check

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
	tyctx "github.com/watchthelight/HypergraphGo/kernel/ctx"
)

// synthExtension is the non-cubical stub that returns false.
func synthExtension(c *Checker, context *tyctx.Ctx, span Span, term ast.Term) (ast.Term, *TypeError, bool) {
	return nil, nil, false
}

// checkExtension is the non-cubical stub that returns false.
func checkExtension(c *Checker, context *tyctx.Ctx, span Span, term ast.Term, expected ast.Term) (*TypeError, bool) {
	return nil, false
}
