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

	// --- Face Formulas ---

	case ast.FaceTop:
		_, ok := b.(ast.FaceTop)
		return ok, true

	case ast.FaceBot:
		_, ok := b.(ast.FaceBot)
		return ok, true

	case ast.FaceEq:
		if bb, ok := b.(ast.FaceEq); ok {
			return aa.IVar == bb.IVar && aa.IsOne == bb.IsOne, true
		}
		return false, true

	case ast.FaceAnd:
		if bb, ok := b.(ast.FaceAnd); ok {
			return alphaEqFace(aa.Left, bb.Left) && alphaEqFace(aa.Right, bb.Right), true
		}
		return false, true

	case ast.FaceOr:
		if bb, ok := b.(ast.FaceOr); ok {
			return alphaEqFace(aa.Left, bb.Left) && alphaEqFace(aa.Right, bb.Right), true
		}
		return false, true

	// --- Partial Types ---

	case ast.Partial:
		if bb, ok := b.(ast.Partial); ok {
			return alphaEqFace(aa.Phi, bb.Phi) && AlphaEq(aa.A, bb.A), true
		}
		return false, true

	case ast.System:
		if bb, ok := b.(ast.System); ok {
			if len(aa.Branches) != len(bb.Branches) {
				return false, true
			}
			for i := range aa.Branches {
				if !alphaEqFace(aa.Branches[i].Phi, bb.Branches[i].Phi) ||
					!AlphaEq(aa.Branches[i].Term, bb.Branches[i].Term) {
					return false, true
				}
			}
			return true, true
		}
		return false, true

	// --- Composition Operations ---

	case ast.Comp:
		if bb, ok := b.(ast.Comp); ok {
			return AlphaEq(aa.A, bb.A) &&
				alphaEqFace(aa.Phi, bb.Phi) &&
				AlphaEq(aa.Tube, bb.Tube) &&
				AlphaEq(aa.Base, bb.Base), true
		}
		return false, true

	case ast.HComp:
		if bb, ok := b.(ast.HComp); ok {
			return AlphaEq(aa.A, bb.A) &&
				alphaEqFace(aa.Phi, bb.Phi) &&
				AlphaEq(aa.Tube, bb.Tube) &&
				AlphaEq(aa.Base, bb.Base), true
		}
		return false, true

	case ast.Fill:
		if bb, ok := b.(ast.Fill); ok {
			return AlphaEq(aa.A, bb.A) &&
				alphaEqFace(aa.Phi, bb.Phi) &&
				AlphaEq(aa.Tube, bb.Tube) &&
				AlphaEq(aa.Base, bb.Base), true
		}
		return false, true

	// --- Glue Types ---

	case ast.Glue:
		if bb, ok := b.(ast.Glue); ok {
			if !AlphaEq(aa.A, bb.A) || len(aa.System) != len(bb.System) {
				return false, true
			}
			for i := range aa.System {
				if !alphaEqFace(aa.System[i].Phi, bb.System[i].Phi) ||
					!AlphaEq(aa.System[i].T, bb.System[i].T) ||
					!AlphaEq(aa.System[i].Equiv, bb.System[i].Equiv) {
					return false, true
				}
			}
			return true, true
		}
		return false, true

	case ast.GlueElem:
		if bb, ok := b.(ast.GlueElem); ok {
			if len(aa.System) != len(bb.System) {
				return false, true
			}
			for i := range aa.System {
				if !alphaEqFace(aa.System[i].Phi, bb.System[i].Phi) ||
					!AlphaEq(aa.System[i].Term, bb.System[i].Term) {
					return false, true
				}
			}
			return AlphaEq(aa.Base, bb.Base), true
		}
		return false, true

	case ast.Unglue:
		if bb, ok := b.(ast.Unglue); ok {
			return AlphaEq(aa.Ty, bb.Ty) && AlphaEq(aa.G, bb.G), true
		}
		return false, true

	// --- Univalence ---

	case ast.UA:
		if bb, ok := b.(ast.UA); ok {
			return AlphaEq(aa.A, bb.A) &&
				AlphaEq(aa.B, bb.B) &&
				AlphaEq(aa.Equiv, bb.Equiv), true
		}
		return false, true

	case ast.UABeta:
		if bb, ok := b.(ast.UABeta); ok {
			return AlphaEq(aa.Equiv, bb.Equiv) && AlphaEq(aa.Arg, bb.Arg), true
		}
		return false, true

	// --- Higher Inductive Types ---

	case ast.HITApp:
		if bb, ok := b.(ast.HITApp); ok {
			if aa.HITName != bb.HITName || aa.Ctor != bb.Ctor {
				return false, true
			}
			if len(aa.Args) != len(bb.Args) || len(aa.IArgs) != len(bb.IArgs) {
				return false, true
			}
			for i := range aa.Args {
				if !AlphaEq(aa.Args[i], bb.Args[i]) {
					return false, true
				}
			}
			for i := range aa.IArgs {
				if !AlphaEq(aa.IArgs[i], bb.IArgs[i]) {
					return false, true
				}
			}
			return true, true
		}
		return false, true

	default:
		return false, false
	}
}

// alphaEqFace checks alpha-equality for face formulas.
func alphaEqFace(a, b ast.Face) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch aa := a.(type) {
	case ast.FaceTop:
		_, ok := b.(ast.FaceTop)
		return ok
	case ast.FaceBot:
		_, ok := b.(ast.FaceBot)
		return ok
	case ast.FaceEq:
		if bb, ok := b.(ast.FaceEq); ok {
			return aa.IVar == bb.IVar && aa.IsOne == bb.IsOne
		}
		return false
	case ast.FaceAnd:
		if bb, ok := b.(ast.FaceAnd); ok {
			return alphaEqFace(aa.Left, bb.Left) && alphaEqFace(aa.Right, bb.Right)
		}
		return false
	case ast.FaceOr:
		if bb, ok := b.(ast.FaceOr); ok {
			return alphaEqFace(aa.Left, bb.Left) && alphaEqFace(aa.Right, bb.Right)
		}
		return false
	default:
		return false
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
