package subst

import (
	"github.com/watchthelight/hypergraphgo/internal/ast"
)

// Shift shifts all free de Bruijn variables >= cutoff by d positions.
// Follows TAPL §6.2.1 conventions.
func Shift(d, cutoff int, t ast.Term) ast.Term {
	if t == nil {
		return nil
	}
	switch t := t.(type) {
	case ast.Var:
		if t.Ix >= cutoff {
			return ast.Var{Ix: t.Ix + d}
		}
		return t
	case ast.Sort:
		return t
	case ast.Global:
		return t
	case ast.Pi:
		return ast.Pi{
			Binder: t.Binder,
			A:      Shift(d, cutoff, t.A),
			B:      Shift(d, cutoff+1, t.B),
		}
	case ast.Lam:
		return ast.Lam{
			Binder: t.Binder,
			Ann:    Shift(d, cutoff, t.Ann),
			Body:   Shift(d, cutoff+1, t.Body),
		}
	case ast.App:
		return ast.App{
			T: Shift(d, cutoff, t.T),
			U: Shift(d, cutoff, t.U),
		}
	case ast.Sigma:
		return ast.Sigma{
			Binder: t.Binder,
			A:      Shift(d, cutoff, t.A),
			B:      Shift(d, cutoff+1, t.B),
		}
	case ast.Pair:
		return ast.Pair{
			Fst: Shift(d, cutoff, t.Fst),
			Snd: Shift(d, cutoff, t.Snd),
		}
	case ast.Fst:
		return ast.Fst{P: Shift(d, cutoff, t.P)}
	case ast.Snd:
		return ast.Snd{P: Shift(d, cutoff, t.P)}
	case ast.Let:
		return ast.Let{
			Binder: t.Binder,
			Ann:    Shift(d, cutoff, t.Ann),
			Val:    Shift(d, cutoff, t.Val),
			Body:   Shift(d, cutoff+1, t.Body),
		}
	default:
		panic("unknown term type")
	}
}

// Subst substitutes s for variable j in t, adjusting free variables.
// Follows TAPL §6.2.4 conventions.
func Subst(j int, s ast.Term, t ast.Term) ast.Term {
	if t == nil {
		return nil
	}
	switch t := t.(type) {
	case ast.Var:
		if t.Ix == j {
			return s
		} else if t.Ix > j {
			return ast.Var{Ix: t.Ix - 1}
		}
		return t
	case ast.Sort:
		return t
	case ast.Global:
		return t
	case ast.Pi:
		return ast.Pi{
			Binder: t.Binder,
			A:      Subst(j, s, t.A),
			B:      Subst(j+1, Shift(1, 0, s), t.B),
		}
	case ast.Lam:
		return ast.Lam{
			Binder: t.Binder,
			Ann:    Subst(j, s, t.Ann),
			Body:   Subst(j+1, Shift(1, 0, s), t.Body),
		}
	case ast.App:
		return ast.App{
			T: Subst(j, s, t.T),
			U: Subst(j, s, t.U),
		}
	case ast.Sigma:
		return ast.Sigma{
			Binder: t.Binder,
			A:      Subst(j, s, t.A),
			B:      Subst(j+1, Shift(1, 0, s), t.B),
		}
	case ast.Pair:
		return ast.Pair{
			Fst: Subst(j, s, t.Fst),
			Snd: Subst(j, s, t.Snd),
		}
	case ast.Fst:
		return ast.Fst{P: Subst(j, s, t.P)}
	case ast.Snd:
		return ast.Snd{P: Subst(j, s, t.P)}
	case ast.Let:
		return ast.Let{
			Binder: t.Binder,
			Ann:    Subst(j, s, t.Ann),
			Val:    Subst(j, s, t.Val),
			Body:   Subst(j+1, Shift(1, 0, s), t.Body),
		}
	default:
		panic("unknown term type")
	}
}
