//go:build cubical

package core

import "github.com/watchthelight/HypergraphGo/internal/ast"

// alphaEqExtension handles alpha-equality for cubical terms.
// Returns (result, true) if the terms were cubical, (false, false) otherwise.
func alphaEqExtension(a, b ast.Term) (bool, bool) {
	switch aa := a.(type) {
	case ast.Interval:
		if _, ok := b.(ast.Interval); ok {
			return true, true
		}
		return false, true

	case ast.I0:
		if _, ok := b.(ast.I0); ok {
			return true, true
		}
		return false, true

	case ast.I1:
		if _, ok := b.(ast.I1); ok {
			return true, true
		}
		return false, true

	case ast.IVar:
		if bb, ok := b.(ast.IVar); ok {
			return aa.Ix == bb.Ix, true
		}
		return false, true

	case ast.Path:
		if bb, ok := b.(ast.Path); ok {
			return AlphaEq(aa.A, bb.A) && AlphaEq(aa.X, bb.X) && AlphaEq(aa.Y, bb.Y), true
		}
		return false, true

	case ast.PathP:
		if bb, ok := b.(ast.PathP); ok {
			return AlphaEq(aa.A, bb.A) && AlphaEq(aa.X, bb.X) && AlphaEq(aa.Y, bb.Y), true
		}
		return false, true

	case ast.PathLam:
		if bb, ok := b.(ast.PathLam); ok {
			return AlphaEq(aa.Body, bb.Body), true
		}
		return false, true

	case ast.PathApp:
		if bb, ok := b.(ast.PathApp); ok {
			return AlphaEq(aa.P, bb.P) && AlphaEq(aa.R, bb.R), true
		}
		return false, true

	case ast.Transport:
		if bb, ok := b.(ast.Transport); ok {
			return AlphaEq(aa.A, bb.A) && AlphaEq(aa.E, bb.E), true
		}
		return false, true

	default:
		return false, false
	}
}

// shiftTermExtension handles shifting for cubical terms.
// Returns (result, true) if the term was cubical, (nil, false) otherwise.
func shiftTermExtension(t ast.Term, d, cutoff int) (ast.Term, bool) {
	switch tm := t.(type) {
	case ast.Interval, ast.I0, ast.I1:
		return tm, true // Constants, no shifting needed

	case ast.IVar:
		// IVar uses separate index space, no term shifting
		return tm, true

	case ast.Path:
		return ast.Path{
			A: shiftTerm(tm.A, d, cutoff),
			X: shiftTerm(tm.X, d, cutoff),
			Y: shiftTerm(tm.Y, d, cutoff),
		}, true

	case ast.PathP:
		// PathP doesn't bind term variables
		return ast.PathP{
			A: shiftTerm(tm.A, d, cutoff),
			X: shiftTerm(tm.X, d, cutoff),
			Y: shiftTerm(tm.Y, d, cutoff),
		}, true

	case ast.PathLam:
		// PathLam doesn't bind term variables
		return ast.PathLam{
			Binder: tm.Binder,
			Body:   shiftTerm(tm.Body, d, cutoff),
		}, true

	case ast.PathApp:
		return ast.PathApp{
			P: shiftTerm(tm.P, d, cutoff),
			R: shiftTerm(tm.R, d, cutoff),
		}, true

	case ast.Transport:
		// Transport doesn't bind term variables
		return ast.Transport{
			A: shiftTerm(tm.A, d, cutoff),
			E: shiftTerm(tm.E, d, cutoff),
		}, true

	default:
		return nil, false
	}
}
