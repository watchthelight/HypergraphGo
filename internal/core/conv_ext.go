//go:build !cubical

package core

import "github.com/watchthelight/HypergraphGo/internal/ast"

// alphaEqExtension is the non-cubical stub that returns false.
func alphaEqExtension(a, b ast.Term) (bool, bool) {
	return false, false
}

// shiftTermExtension is the non-cubical stub that returns false.
func shiftTermExtension(t ast.Term, d, cutoff int) (ast.Term, bool) {
	return nil, false
}
