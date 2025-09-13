package core

import (
	"github.com/watchthelight/hypergraphgo/internal/ast"
	"github.com/watchthelight/hypergraphgo/internal/eval"
)

type EtaFlags struct{ Pi, Sigma bool }

// Conv reports whether terms a and b are definitionally equal under the given eta flags.
// Strategy: normalize via NbE then alpha-compare normal forms.
// (We keep eta-expansion in eval.Reify for functions/pairs behind flags later; for now,
// Conv simply compares the normal forms returned by eval.Normalize.)
func Conv(a, b ast.Term, _ EtaFlags) bool {
	nfa := eval.Normalize(a)
	nfb := eval.Normalize(b)
	return AlphaEq(nfa, nfb)
}

// AlphaEq compares two core terms modulo alpha (de Bruijn makes this structural).
func AlphaEq(a, b ast.Term) bool {
	switch a := a.(type) {
	case ast.Sort:
		if bb, ok := b.(ast.Sort); ok {
			return a.U == bb.U
		}
	case ast.Var:
		if bb, ok := b.(ast.Var); ok {
			return a.Ix == bb.Ix
		}
	case ast.Global:
		if bb, ok := b.(ast.Global); ok {
			return a.Name == bb.Name
		}
	case ast.Pi:
		if bb, ok := b.(ast.Pi); ok {
			return AlphaEq(a.A, bb.A) && AlphaEq(a.B, bb.B)
		}
	case ast.Lam:
		if bb, ok := b.(ast.Lam); ok {
			return AlphaEq(a.Body, bb.Body)
		}
	case ast.App:
		if bb, ok := b.(ast.App); ok {
			return AlphaEq(a.T, bb.T) && AlphaEq(a.U, bb.U)
		}
	case ast.Sigma:
		if bb, ok := b.(ast.Sigma); ok {
			return AlphaEq(a.A, bb.A) && AlphaEq(a.B, bb.B)
		}
	case ast.Pair:
		if bb, ok := b.(ast.Pair); ok {
			return AlphaEq(a.Fst, bb.Fst) && AlphaEq(a.Snd, bb.Snd)
		}
	case ast.Fst:
		if bb, ok := b.(ast.Fst); ok {
			return AlphaEq(a.P, bb.P)
		}
	case ast.Snd:
		if bb, ok := b.(ast.Snd); ok {
			return AlphaEq(a.P, bb.P)
		}
	case ast.Let:
		if bb, ok := b.(ast.Let); ok {
			return AlphaEq(a.Val, bb.Val) && AlphaEq(a.Body, bb.Body) && ((a.Ann == nil && bb.Ann == nil) || (a.Ann != nil && bb.Ann != nil && AlphaEq(a.Ann, bb.Ann)))
		}
	}
	return false
}
