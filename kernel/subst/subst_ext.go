//go:build !cubical

package subst

import "github.com/watchthelight/HypergraphGo/internal/ast"

// shiftExtension is the non-cubical stub that returns false.
func shiftExtension(d, cutoff int, t ast.Term) (ast.Term, bool) {
	return nil, false
}

// substExtension is the non-cubical stub that returns false.
func substExtension(j int, s ast.Term, t ast.Term) (ast.Term, bool) {
	return nil, false
}
