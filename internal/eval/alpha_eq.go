// Package eval implements Normalization by Evaluation (NbE) for the HoTT kernel.
// This file provides alpha-equality checking for AST terms.
package eval

import "github.com/watchthelight/HypergraphGo/internal/ast"

// AlphaEq checks if two terms are alpha-equal.
// Terms are alpha-equal if they have the same structure and all binders
// correspond correctly. Since we use de Bruijn indices, binder names are
// irrelevant and structurally identical terms are alpha-equal.
func AlphaEq(a, b ast.Term) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch ta := a.(type) {
	case ast.Var:
		tb, ok := b.(ast.Var)
		return ok && ta.Ix == tb.Ix

	case ast.Global:
		tb, ok := b.(ast.Global)
		return ok && ta.Name == tb.Name

	case ast.Sort:
		tb, ok := b.(ast.Sort)
		return ok && ta.U == tb.U

	case ast.Lam:
		tb, ok := b.(ast.Lam)
		// Binder names are irrelevant; only check body
		return ok && AlphaEq(ta.Body, tb.Body)

	case ast.App:
		tb, ok := b.(ast.App)
		return ok && AlphaEq(ta.T, tb.T) && AlphaEq(ta.U, tb.U)

	case ast.Pi:
		tb, ok := b.(ast.Pi)
		// Binder names are irrelevant
		return ok && AlphaEq(ta.A, tb.A) && AlphaEq(ta.B, tb.B)

	case ast.Sigma:
		tb, ok := b.(ast.Sigma)
		return ok && AlphaEq(ta.A, tb.A) && AlphaEq(ta.B, tb.B)

	case ast.Pair:
		tb, ok := b.(ast.Pair)
		return ok && AlphaEq(ta.Fst, tb.Fst) && AlphaEq(ta.Snd, tb.Snd)

	case ast.Fst:
		tb, ok := b.(ast.Fst)
		return ok && AlphaEq(ta.P, tb.P)

	case ast.Snd:
		tb, ok := b.(ast.Snd)
		return ok && AlphaEq(ta.P, tb.P)

	case ast.Let:
		tb, ok := b.(ast.Let)
		return ok && AlphaEq(ta.Val, tb.Val) && AlphaEq(ta.Body, tb.Body)

	case ast.Id:
		tb, ok := b.(ast.Id)
		return ok && AlphaEq(ta.A, tb.A) && AlphaEq(ta.X, tb.X) && AlphaEq(ta.Y, tb.Y)

	case ast.Refl:
		tb, ok := b.(ast.Refl)
		return ok && AlphaEq(ta.A, tb.A) && AlphaEq(ta.X, tb.X)

	case ast.J:
		tb, ok := b.(ast.J)
		return ok &&
			AlphaEq(ta.A, tb.A) &&
			AlphaEq(ta.C, tb.C) &&
			AlphaEq(ta.D, tb.D) &&
			AlphaEq(ta.X, tb.X) &&
			AlphaEq(ta.Y, tb.Y) &&
			AlphaEq(ta.P, tb.P)

	// --- Cubical Terms ---

	case ast.Interval:
		_, ok := b.(ast.Interval)
		return ok

	case ast.I0:
		_, ok := b.(ast.I0)
		return ok

	case ast.I1:
		_, ok := b.(ast.I1)
		return ok

	case ast.IVar:
		tb, ok := b.(ast.IVar)
		return ok && ta.Ix == tb.Ix

	case ast.Path:
		tb, ok := b.(ast.Path)
		return ok && AlphaEq(ta.A, tb.A) && AlphaEq(ta.X, tb.X) && AlphaEq(ta.Y, tb.Y)

	case ast.PathP:
		tb, ok := b.(ast.PathP)
		return ok && AlphaEq(ta.A, tb.A) && AlphaEq(ta.X, tb.X) && AlphaEq(ta.Y, tb.Y)

	case ast.PathLam:
		tb, ok := b.(ast.PathLam)
		// Binder names are irrelevant
		return ok && AlphaEq(ta.Body, tb.Body)

	case ast.PathApp:
		tb, ok := b.(ast.PathApp)
		return ok && AlphaEq(ta.P, tb.P) && AlphaEq(ta.R, tb.R)

	case ast.Transport:
		tb, ok := b.(ast.Transport)
		return ok && AlphaEq(ta.A, tb.A) && AlphaEq(ta.E, tb.E)

	// --- Face Formulas ---

	case ast.FaceTop:
		_, ok := b.(ast.FaceTop)
		return ok

	case ast.FaceBot:
		_, ok := b.(ast.FaceBot)
		return ok

	case ast.FaceEq:
		tb, ok := b.(ast.FaceEq)
		return ok && ta.IVar == tb.IVar && ta.IsOne == tb.IsOne

	case ast.FaceAnd:
		tb, ok := b.(ast.FaceAnd)
		return ok && alphaEqFace(ta.Left, tb.Left) && alphaEqFace(ta.Right, tb.Right)

	case ast.FaceOr:
		tb, ok := b.(ast.FaceOr)
		return ok && alphaEqFace(ta.Left, tb.Left) && alphaEqFace(ta.Right, tb.Right)

	// --- Partial Types ---

	case ast.Partial:
		tb, ok := b.(ast.Partial)
		return ok && alphaEqFace(ta.Phi, tb.Phi) && AlphaEq(ta.A, tb.A)

	case ast.System:
		tb, ok := b.(ast.System)
		if !ok || len(ta.Branches) != len(tb.Branches) {
			return false
		}
		for i := range ta.Branches {
			if !alphaEqFace(ta.Branches[i].Phi, tb.Branches[i].Phi) {
				return false
			}
			if !AlphaEq(ta.Branches[i].Term, tb.Branches[i].Term) {
				return false
			}
		}
		return true

	// --- Composition Operations ---

	case ast.Comp:
		tb, ok := b.(ast.Comp)
		return ok &&
			AlphaEq(ta.A, tb.A) &&
			alphaEqFace(ta.Phi, tb.Phi) &&
			AlphaEq(ta.Tube, tb.Tube) &&
			AlphaEq(ta.Base, tb.Base)

	case ast.HComp:
		tb, ok := b.(ast.HComp)
		return ok &&
			AlphaEq(ta.A, tb.A) &&
			alphaEqFace(ta.Phi, tb.Phi) &&
			AlphaEq(ta.Tube, tb.Tube) &&
			AlphaEq(ta.Base, tb.Base)

	case ast.Fill:
		tb, ok := b.(ast.Fill)
		return ok &&
			AlphaEq(ta.A, tb.A) &&
			alphaEqFace(ta.Phi, tb.Phi) &&
			AlphaEq(ta.Tube, tb.Tube) &&
			AlphaEq(ta.Base, tb.Base)

	// --- Glue Types ---

	case ast.Glue:
		tb, ok := b.(ast.Glue)
		if !ok || len(ta.System) != len(tb.System) {
			return false
		}
		if !AlphaEq(ta.A, tb.A) {
			return false
		}
		for i := range ta.System {
			if !alphaEqFace(ta.System[i].Phi, tb.System[i].Phi) {
				return false
			}
			if !AlphaEq(ta.System[i].T, tb.System[i].T) {
				return false
			}
			if !AlphaEq(ta.System[i].Equiv, tb.System[i].Equiv) {
				return false
			}
		}
		return true

	case ast.GlueElem:
		tb, ok := b.(ast.GlueElem)
		if !ok || len(ta.System) != len(tb.System) {
			return false
		}
		for i := range ta.System {
			if !alphaEqFace(ta.System[i].Phi, tb.System[i].Phi) {
				return false
			}
			if !AlphaEq(ta.System[i].Term, tb.System[i].Term) {
				return false
			}
		}
		return AlphaEq(ta.Base, tb.Base)

	case ast.Unglue:
		tb, ok := b.(ast.Unglue)
		return ok && AlphaEq(ta.Ty, tb.Ty) && AlphaEq(ta.G, tb.G)

	// --- Univalence ---

	case ast.UA:
		tb, ok := b.(ast.UA)
		return ok && AlphaEq(ta.A, tb.A) && AlphaEq(ta.B, tb.B) && AlphaEq(ta.Equiv, tb.Equiv)

	case ast.UABeta:
		tb, ok := b.(ast.UABeta)
		return ok && AlphaEq(ta.Equiv, tb.Equiv) && AlphaEq(ta.Arg, tb.Arg)

	// --- Higher Inductive Types ---

	case ast.HITApp:
		tb, ok := b.(ast.HITApp)
		if !ok || ta.HITName != tb.HITName || ta.Ctor != tb.Ctor {
			return false
		}
		if len(ta.Args) != len(tb.Args) || len(ta.IArgs) != len(tb.IArgs) {
			return false
		}
		for i := range ta.Args {
			if !AlphaEq(ta.Args[i], tb.Args[i]) {
				return false
			}
		}
		for i := range ta.IArgs {
			if !AlphaEq(ta.IArgs[i], tb.IArgs[i]) {
				return false
			}
		}
		return true

	default:
		return false
	}
}

// alphaEqFace checks if two face formulas are alpha-equal.
func alphaEqFace(a, b ast.Face) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch fa := a.(type) {
	case ast.FaceTop:
		_, ok := b.(ast.FaceTop)
		return ok

	case ast.FaceBot:
		_, ok := b.(ast.FaceBot)
		return ok

	case ast.FaceEq:
		fb, ok := b.(ast.FaceEq)
		return ok && fa.IVar == fb.IVar && fa.IsOne == fb.IsOne

	case ast.FaceAnd:
		fb, ok := b.(ast.FaceAnd)
		return ok && alphaEqFace(fa.Left, fb.Left) && alphaEqFace(fa.Right, fb.Right)

	case ast.FaceOr:
		fb, ok := b.(ast.FaceOr)
		return ok && alphaEqFace(fa.Left, fb.Left) && alphaEqFace(fa.Right, fb.Right)

	default:
		return false
	}
}
