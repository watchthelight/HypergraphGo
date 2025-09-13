package eval

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/kernel/subst"
)

// Normalize reduces a term to weak head normal form using beta-reduction and projections.
func Normalize(t ast.Term) ast.Term {
	switch t := t.(type) {
	case ast.App:
		f := Normalize(t.T)
		if lam, ok := f.(ast.Lam); ok {
			// beta reduction
			return Normalize(subst.Subst(0, Normalize(t.U), lam.Body))
		}
		return ast.App{T: f, U: Normalize(t.U)}
	case ast.Fst:
		p := Normalize(t.P)
		if pair, ok := p.(ast.Pair); ok {
			return Normalize(pair.Fst)
		}
		return ast.Fst{P: p}
	case ast.Snd:
		p := Normalize(t.P)
		if pair, ok := p.(ast.Pair); ok {
			return Normalize(pair.Snd)
		}
		return ast.Snd{P: p}
	default:
		// for other terms, assume WHNF
		return t
	}
}
