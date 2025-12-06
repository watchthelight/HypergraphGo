//go:build !cubical

package eval

import "github.com/watchthelight/HypergraphGo/internal/ast"

// tryEvalCubical is the non-cubical stub that returns false.
func tryEvalCubical(env *Env, t ast.Term) (Value, bool) {
	return nil, false
}

// tryReifyCubical is the non-cubical stub that returns false.
func tryReifyCubical(level int, v Value) (ast.Term, bool) {
	return nil, false
}
