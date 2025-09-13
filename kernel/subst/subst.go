package subst

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// Shift shifts all free de Bruijn variables >= cutoff by d positions.
// Follows TAPL §6.2.1.
func Shift(d, cutoff int, t ast.Term) ast.Term {
	if t == nil {
		return nil
	}
	switch tm := t.(type) {
	case ast.Var:
		if tm.Ix >= cutoff {
			return ast.Var{Ix: tm.Ix + d}
		}
		return tm
	case ast.Sort:
		return tm
	case ast.Global:
		return tm
	case ast.Pi:
		return ast.Pi{
			Binder: tm.Binder,
			A:      Shift(d, cutoff, tm.A),
			B:      Shift(d, cutoff+1, tm.B),
		}
	case ast.Lam:
		return ast.Lam{
			Binder: tm.Binder,
			Ann:    Shift(d, cutoff, tm.Ann),
			Body:   Shift(d, cutoff+1, tm.Body),
		}
	case ast.App:
		return ast.App{
			T: Shift(d, cutoff, tm.T),
			U: Shift(d, cutoff, tm.U),
		}
	case ast.Sigma:
		return ast.Sigma{
			Binder: tm.Binder,
			A:      Shift(d, cutoff, tm.A),
			B:      Shift(d, cutoff+1, tm.B),
		}
	case ast.Pair:
		return ast.Pair{
			Fst: Shift(d, cutoff, tm.Fst),
			Snd: Shift(d, cutoff, tm.Snd),
		}
	case ast.Fst:
		return ast.Fst{P: Shift(d, cutoff, tm.P)}
	case ast.Snd:
		return ast.Snd{P: Shift(d, cutoff, tm.P)}
	case ast.Let:
		return ast.Let{
			Binder: tm.Binder,
			Ann:    Shift(d, cutoff, tm.Ann),
			Val:    Shift(d, cutoff, tm.Val),
			Body:   Shift(d, cutoff+1, tm.Body),
		}
	default:
		panic("unhandled term type in Shift")
	}
}

// Subst substitutes s for variable j in t, adjusting indices.
// Follows TAPL §6.2.2.
func Subst(j int, s ast.Term, t ast.Term) ast.Term {
	if t == nil {
		return nil
	}
	switch tm := t.(type) {
	case ast.Var:
		if tm.Ix == j {
			return s
		} else if tm.Ix > j {
			return ast.Var{Ix: tm.Ix - 1}
		}
		return tm
	case ast.Sort:
		return tm
	case ast.Global:
		return tm
	case ast.Pi:
		return ast.Pi{
			Binder: tm.Binder,
			A:      Subst(j, s, tm.A),
			B:      Subst(j+1, Shift(1, 0, s), tm.B),
		}
	case ast.Lam:
		return ast.Lam{
			Binder: tm.Binder,
			Ann:    Subst(j, s, tm.Ann),
			Body:   Subst(j+1, Shift(1, 0, s), tm.Body),
		}
	case ast.App:
		return ast.App{
			T: Subst(j, s, tm.T),
			U: Subst(j, s, tm.U),
		}
	case ast.Sigma:
		return ast.Sigma{
			Binder: tm.Binder,
			A:      Subst(j, s, tm.A),
			B:      Subst(j+1, Shift(1, 0, s), tm.B),
		}
	case ast.Pair:
		return ast.Pair{
			Fst: Subst(j, s, tm.Fst),
			Snd: Subst(j, s, tm.Snd),
		}
	case ast.Fst:
		return ast.Fst{P: Subst(j, s, tm.P)}
	case ast.Snd:
		return ast.Snd{P: Subst(j, s, tm.P)}
	case ast.Let:
		return ast.Let{
			Binder: tm.Binder,
			Ann:    Subst(j, s, tm.Ann),
			Val:    Subst(j, s, tm.Val),
			Body:   Subst(j+1, Shift(1, 0, s), tm.Body),
		}
	default:
		panic("unhandled term type in Subst")
	}
}
