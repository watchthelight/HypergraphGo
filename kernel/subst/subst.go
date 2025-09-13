package subst

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// Shift shifts all free variables >= cutoff by d.
// Follows TAPL §6.2.1.
func Shift(d, cutoff int, t ast.Term) ast.Term {
	return shift(d, cutoff, t)
}

// Subst substitutes s for Var(j) in t, adjusting indices.
// Follows TAPL §6.2.4.
func Subst(j int, s ast.Term, t ast.Term) ast.Term {
	return subst(j, s, t)
}

func shift(d, c int, t ast.Term) ast.Term {
	switch t := t.(type) {
	case ast.Var:
		if t.Ix >= c {
			return ast.Var{Ix: t.Ix + d}
		}
		return t
	case ast.Sort:
		return t
	case ast.Global:
		return t
	case ast.Pi:
		return ast.Pi{Binder: t.Binder, A: shift(d, c, t.A), B: shift(d, c+1, t.B)}
	case ast.Lam:
		ann := t.Ann
		if ann != nil {
			ann = shift(d, c, ann)
		}
		return ast.Lam{Binder: t.Binder, Ann: ann, Body: shift(d, c+1, t.Body)}
	case ast.App:
		return ast.App{T: shift(d, c, t.T), U: shift(d, c, t.U)}
	case ast.Sigma:
		return ast.Sigma{Binder: t.Binder, A: shift(d, c, t.A), B: shift(d, c+1, t.B)}
	case ast.Pair:
		return ast.Pair{Fst: shift(d, c, t.Fst), Snd: shift(d, c, t.Snd)}
	case ast.Fst:
		return ast.Fst{P: shift(d, c, t.P)}
	case ast.Snd:
		return ast.Snd{P: shift(d, c, t.P)}
	case ast.Let:
		ann := t.Ann
		if ann != nil {
			ann = shift(d, c, ann)
		}
		return ast.Let{Binder: t.Binder, Ann: ann, Val: shift(d, c, t.Val), Body: shift(d, c+1, t.Body)}
	default:
		panic("unknown term type")
	}
}

func subst(j int, s ast.Term, t ast.Term) ast.Term {
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
		return ast.Pi{Binder: t.Binder, A: subst(j, s, t.A), B: subst(j+1, shift(1, 0, s), t.B)}
	case ast.Lam:
		ann := t.Ann
		if ann != nil {
			ann = subst(j, s, ann)
		}
		return ast.Lam{Binder: t.Binder, Ann: ann, Body: subst(j+1, shift(1, 0, s), t.Body)}
	case ast.App:
		return ast.App{T: subst(j, s, t.T), U: subst(j, s, t.U)}
	case ast.Sigma:
		return ast.Sigma{Binder: t.Binder, A: subst(j, s, t.A), B: subst(j+1, shift(1, 0, s), t.B)}
	case ast.Pair:
		return ast.Pair{Fst: subst(j, s, t.Fst), Snd: subst(j, s, t.Snd)}
	case ast.Fst:
		return ast.Fst{P: subst(j, s, t.P)}
	case ast.Snd:
		return ast.Snd{P: subst(j, s, t.P)}
	case ast.Let:
		ann := t.Ann
		if ann != nil {
			ann = subst(j, s, ann)
		}
		return ast.Let{Binder: t.Binder, Ann: ann, Val: subst(j, s, t.Val), Body: subst(j+1, shift(1, 0, s), t.Body)}
	default:
		panic("unknown term type")
	}
}
