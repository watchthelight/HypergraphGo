package subst

import (
	"github.com/watchthelight/hypergraphgo/internal/ast"
)

// Shift shifts all free de Bruijn variables >= cutoff by d.
// Follows TAPL §6.2.1.
func Shift(d, cutoff int, t ast.Term) ast.Term {
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
		a := Shift(d, cutoff, t.A)
		b := Shift(d, cutoff+1, t.B)
		return ast.Pi{Binder: t.Binder, A: a, B: b}
	case ast.Lam:
		ann := t.Ann
		if ann != nil {
			ann = Shift(d, cutoff, ann)
		}
		body := Shift(d, cutoff+1, t.Body)
		return ast.Lam{Binder: t.Binder, Ann: ann, Body: body}
	case ast.App:
		fun := Shift(d, cutoff, t.T)
		arg := Shift(d, cutoff, t.U)
		return ast.App{T: fun, U: arg}
	case ast.Sigma:
		a := Shift(d, cutoff, t.A)
		b := Shift(d, cutoff+1, t.B)
		return ast.Sigma{Binder: t.Binder, A: a, B: b}
	case ast.Pair:
		fst := Shift(d, cutoff, t.Fst)
		snd := Shift(d, cutoff, t.Snd)
		return ast.Pair{Fst: fst, Snd: snd}
	case ast.Fst:
		p := Shift(d, cutoff, t.P)
		return ast.Fst{P: p}
	case ast.Snd:
		p := Shift(d, cutoff, t.P)
		return ast.Snd{P: p}
	case ast.Let:
		ann := t.Ann
		if ann != nil {
			ann = Shift(d, cutoff, ann)
		}
		val := Shift(d, cutoff, t.Val)
		body := Shift(d, cutoff+1, t.Body)
		return ast.Let{Binder: t.Binder, Ann: ann, Val: val, Body: body}
	default:
		panic("unhandled term type in Shift")
	}
}

// Subst substitutes s for Var(j) in t, adjusting free variables.
// Follows TAPL §6.2.4.
func Subst(j int, s ast.Term, t ast.Term) ast.Term {
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
		a := Subst(j, s, t.A)
		b := Subst(j+1, Shift(1, 0, s), t.B)
		return ast.Pi{Binder: t.Binder, A: a, B: b}
	case ast.Lam:
		ann := t.Ann
		if ann != nil {
			ann = Subst(j, s, ann)
		}
		body := Subst(j+1, Shift(1, 0, s), t.Body)
		return ast.Lam{Binder: t.Binder, Ann: ann, Body: body}
	case ast.App:
		fun := Subst(j, s, t.T)
		arg := Subst(j, s, t.U)
		return ast.App{T: fun, U: arg}
	case ast.Sigma:
		a := Subst(j, s, t.A)
		b := Subst(j+1, Shift(1, 0, s), t.B)
		return ast.Sigma{Binder: t.Binder, A: a, B: b}
	case ast.Pair:
		fst := Subst(j, s, t.Fst)
		snd := Subst(j, s, t.Snd)
		return ast.Pair{Fst: fst, Snd: snd}
	case ast.Fst:
		p := Subst(j, s, t.P)
		return ast.Fst{P: p}
	case ast.Snd:
		p := Subst(j, s, t.P)
		return ast.Snd{P: p}
	case ast.Let:
		ann := t.Ann
		if ann != nil {
			ann = Subst(j, s, ann)
		}
		val := Subst(j, s, t.Val)
		body := Subst(j+1, Shift(1, 0, s), t.Body)
		return ast.Let{Binder: t.Binder, Ann: ann, Val: val, Body: body}
	default:
		panic("unhandled term type in Subst")
	}
}
