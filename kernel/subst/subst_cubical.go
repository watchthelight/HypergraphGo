//go:build cubical

package subst

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// IShift shifts all free interval variables >= cutoff by d positions.
// This operates in the interval variable namespace, separate from term variables.
func IShift(d, cutoff int, t ast.Term) ast.Term {
	if t == nil {
		return nil
	}
	switch tm := t.(type) {
	// Interval terms
	case ast.IVar:
		if tm.Ix >= cutoff {
			return ast.IVar{Ix: tm.Ix + d}
		}
		return tm
	case ast.I0, ast.I1, ast.Interval:
		return tm // Endpoints and interval type are constants

	// Path types - PathP and Transport bind interval variables
	case ast.Path:
		return ast.Path{
			A: IShift(d, cutoff, tm.A),
			X: IShift(d, cutoff, tm.X),
			Y: IShift(d, cutoff, tm.Y),
		}
	case ast.PathP:
		// A binds an interval variable
		return ast.PathP{
			A: IShift(d, cutoff+1, tm.A),
			X: IShift(d, cutoff, tm.X),
			Y: IShift(d, cutoff, tm.Y),
		}
	case ast.PathLam:
		// PathLam binds an interval variable
		return ast.PathLam{
			Binder: tm.Binder,
			Body:   IShift(d, cutoff+1, tm.Body),
		}
	case ast.PathApp:
		return ast.PathApp{
			P: IShift(d, cutoff, tm.P),
			R: IShift(d, cutoff, tm.R),
		}
	case ast.Transport:
		// A binds an interval variable
		return ast.Transport{
			A: IShift(d, cutoff+1, tm.A),
			E: IShift(d, cutoff, tm.E),
		}

	// Standard terms - recurse without changing cutoff (no interval binders)
	case ast.Var, ast.Sort, ast.Global:
		return tm
	case ast.Pi:
		return ast.Pi{
			Binder: tm.Binder,
			A:      IShift(d, cutoff, tm.A),
			B:      IShift(d, cutoff, tm.B), // Pi doesn't bind interval vars
		}
	case ast.Lam:
		return ast.Lam{
			Binder: tm.Binder,
			Ann:    IShift(d, cutoff, tm.Ann),
			Body:   IShift(d, cutoff, tm.Body),
		}
	case ast.App:
		return ast.App{
			T: IShift(d, cutoff, tm.T),
			U: IShift(d, cutoff, tm.U),
		}
	case ast.Sigma:
		return ast.Sigma{
			Binder: tm.Binder,
			A:      IShift(d, cutoff, tm.A),
			B:      IShift(d, cutoff, tm.B),
		}
	case ast.Pair:
		return ast.Pair{
			Fst: IShift(d, cutoff, tm.Fst),
			Snd: IShift(d, cutoff, tm.Snd),
		}
	case ast.Fst:
		return ast.Fst{P: IShift(d, cutoff, tm.P)}
	case ast.Snd:
		return ast.Snd{P: IShift(d, cutoff, tm.P)}
	case ast.Let:
		return ast.Let{
			Binder: tm.Binder,
			Ann:    IShift(d, cutoff, tm.Ann),
			Val:    IShift(d, cutoff, tm.Val),
			Body:   IShift(d, cutoff, tm.Body),
		}
	case ast.Id:
		return ast.Id{
			A: IShift(d, cutoff, tm.A),
			X: IShift(d, cutoff, tm.X),
			Y: IShift(d, cutoff, tm.Y),
		}
	case ast.Refl:
		return ast.Refl{
			A: IShift(d, cutoff, tm.A),
			X: IShift(d, cutoff, tm.X),
		}
	case ast.J:
		return ast.J{
			A: IShift(d, cutoff, tm.A),
			C: IShift(d, cutoff, tm.C),
			D: IShift(d, cutoff, tm.D),
			X: IShift(d, cutoff, tm.X),
			Y: IShift(d, cutoff, tm.Y),
			P: IShift(d, cutoff, tm.P),
		}
	default:
		return t
	}
}

// ISubst substitutes interval term s for interval variable j in t.
// This operates in the interval variable namespace.
func ISubst(j int, s ast.Term, t ast.Term) ast.Term {
	if t == nil {
		return nil
	}
	switch tm := t.(type) {
	// Interval terms
	case ast.IVar:
		if tm.Ix == j {
			return s
		} else if tm.Ix > j {
			return ast.IVar{Ix: tm.Ix - 1}
		}
		return tm
	case ast.I0, ast.I1, ast.Interval:
		return tm

	// Path types - PathP, PathLam, and Transport bind interval variables
	case ast.Path:
		return ast.Path{
			A: ISubst(j, s, tm.A),
			X: ISubst(j, s, tm.X),
			Y: ISubst(j, s, tm.Y),
		}
	case ast.PathP:
		return ast.PathP{
			A: ISubst(j+1, IShift(1, 0, s), tm.A),
			X: ISubst(j, s, tm.X),
			Y: ISubst(j, s, tm.Y),
		}
	case ast.PathLam:
		return ast.PathLam{
			Binder: tm.Binder,
			Body:   ISubst(j+1, IShift(1, 0, s), tm.Body),
		}
	case ast.PathApp:
		return ast.PathApp{
			P: ISubst(j, s, tm.P),
			R: ISubst(j, s, tm.R),
		}
	case ast.Transport:
		return ast.Transport{
			A: ISubst(j+1, IShift(1, 0, s), tm.A),
			E: ISubst(j, s, tm.E),
		}

	// Standard terms - recurse without changing j (no interval binders)
	case ast.Var, ast.Sort, ast.Global:
		return tm
	case ast.Pi:
		return ast.Pi{
			Binder: tm.Binder,
			A:      ISubst(j, s, tm.A),
			B:      ISubst(j, s, tm.B),
		}
	case ast.Lam:
		return ast.Lam{
			Binder: tm.Binder,
			Ann:    ISubst(j, s, tm.Ann),
			Body:   ISubst(j, s, tm.Body),
		}
	case ast.App:
		return ast.App{
			T: ISubst(j, s, tm.T),
			U: ISubst(j, s, tm.U),
		}
	case ast.Sigma:
		return ast.Sigma{
			Binder: tm.Binder,
			A:      ISubst(j, s, tm.A),
			B:      ISubst(j, s, tm.B),
		}
	case ast.Pair:
		return ast.Pair{
			Fst: ISubst(j, s, tm.Fst),
			Snd: ISubst(j, s, tm.Snd),
		}
	case ast.Fst:
		return ast.Fst{P: ISubst(j, s, tm.P)}
	case ast.Snd:
		return ast.Snd{P: ISubst(j, s, tm.P)}
	case ast.Let:
		return ast.Let{
			Binder: tm.Binder,
			Ann:    ISubst(j, s, tm.Ann),
			Val:    ISubst(j, s, tm.Val),
			Body:   ISubst(j, s, tm.Body),
		}
	case ast.Id:
		return ast.Id{
			A: ISubst(j, s, tm.A),
			X: ISubst(j, s, tm.X),
			Y: ISubst(j, s, tm.Y),
		}
	case ast.Refl:
		return ast.Refl{
			A: ISubst(j, s, tm.A),
			X: ISubst(j, s, tm.X),
		}
	case ast.J:
		return ast.J{
			A: ISubst(j, s, tm.A),
			C: ISubst(j, s, tm.C),
			D: ISubst(j, s, tm.D),
			X: ISubst(j, s, tm.X),
			Y: ISubst(j, s, tm.Y),
			P: ISubst(j, s, tm.P),
		}
	default:
		return t
	}
}

// shiftExtension extends Shift to handle cubical term types.
// Called from the Shift function when cubical is enabled.
func shiftExtension(d, cutoff int, t ast.Term) (ast.Term, bool) {
	switch tm := t.(type) {
	case ast.Interval, ast.I0, ast.I1, ast.IVar:
		return tm, true // Interval terms have no term variables
	case ast.Path:
		return ast.Path{
			A: Shift(d, cutoff, tm.A),
			X: Shift(d, cutoff, tm.X),
			Y: Shift(d, cutoff, tm.Y),
		}, true
	case ast.PathP:
		// PathP doesn't bind term variables, only interval variables
		return ast.PathP{
			A: Shift(d, cutoff, tm.A),
			X: Shift(d, cutoff, tm.X),
			Y: Shift(d, cutoff, tm.Y),
		}, true
	case ast.PathLam:
		// PathLam doesn't bind term variables
		return ast.PathLam{
			Binder: tm.Binder,
			Body:   Shift(d, cutoff, tm.Body),
		}, true
	case ast.PathApp:
		return ast.PathApp{
			P: Shift(d, cutoff, tm.P),
			R: Shift(d, cutoff, tm.R),
		}, true
	case ast.Transport:
		return ast.Transport{
			A: Shift(d, cutoff, tm.A),
			E: Shift(d, cutoff, tm.E),
		}, true
	default:
		return nil, false
	}
}

// substExtension extends Subst to handle cubical term types.
// Called from the Subst function when cubical is enabled.
func substExtension(j int, s ast.Term, t ast.Term) (ast.Term, bool) {
	switch tm := t.(type) {
	case ast.Interval, ast.I0, ast.I1, ast.IVar:
		return tm, true // Interval terms have no term variables
	case ast.Path:
		return ast.Path{
			A: Subst(j, s, tm.A),
			X: Subst(j, s, tm.X),
			Y: Subst(j, s, tm.Y),
		}, true
	case ast.PathP:
		return ast.PathP{
			A: Subst(j, s, tm.A),
			X: Subst(j, s, tm.X),
			Y: Subst(j, s, tm.Y),
		}, true
	case ast.PathLam:
		return ast.PathLam{
			Binder: tm.Binder,
			Body:   Subst(j, s, tm.Body),
		}, true
	case ast.PathApp:
		return ast.PathApp{
			P: Subst(j, s, tm.P),
			R: Subst(j, s, tm.R),
		}, true
	case ast.Transport:
		return ast.Transport{
			A: Subst(j, s, tm.A),
			E: Subst(j, s, tm.E),
		}, true
	default:
		return nil, false
	}
}
