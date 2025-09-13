package ast

// Resolver turns Raw terms with names into Core terms with de Bruijn indices.
// On unbound RVar, Resolver looks for a Global; if not found, returns error.

import "fmt"

type scope []string

func (s scope) push(name string) scope { return append(s, name) }

func (s scope) index(name string) (int, bool) {
	// de Bruijn: last binder is index 0
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == name {
			return len(s) - 1 - i, true
		}
	}
	return 0, false
}

type Globals interface {
	// Has reports whether a global name exists.
	Has(name string) bool
}

// Resolve converts an RTerm to a core Term with respect to a scope and globals.
func Resolve(globals Globals, sc scope, r RTerm) (Term, error) {
	switch t := r.(type) {
	case RVar:
		if ix, ok := sc.index(t.Name); ok {
			return Var{Ix: ix}, nil
		}
		if globals != nil && globals.Has(t.Name) {
			return Global{Name: t.Name}, nil
		}
		return nil, fmt.Errorf("resolve: unbound variable %q", t.Name)
	case RGlobal:
		return Global{Name: t.Name}, nil
	case RSort:
		return Sort{U: t.U}, nil
	case RPi:
		A, err := Resolve(globals, sc, t.A)
		if err != nil {
			return nil, err
		}
		B, err := Resolve(globals, sc.push(t.Binder), t.B)
		if err != nil {
			return nil, err
		}
		return Pi{Binder: t.Binder, A: A, B: B}, nil
	case RLam:
		var ann Term
		var err error
		if t.Ann != nil {
			ann, err = Resolve(globals, sc, t.Ann)
			if err != nil {
				return nil, err
			}
		}
		body, err := Resolve(globals, sc.push(t.Binder), t.Body)
		if err != nil {
			return nil, err
		}
		return Lam{Binder: t.Binder, Ann: ann, Body: body}, nil
	case RApp:
		T, err := Resolve(globals, sc, t.T)
		if err != nil {
			return nil, err
		}
		U, err := Resolve(globals, sc, t.U)
		if err != nil {
			return nil, err
		}
		return App{T: T, U: U}, nil
	case RSigma:
		A, err := Resolve(globals, sc, t.A)
		if err != nil {
			return nil, err
		}
		B, err := Resolve(globals, sc.push(t.Binder), t.B)
		if err != nil {
			return nil, err
		}
		return Sigma{Binder: t.Binder, A: A, B: B}, nil
	case RPair:
		F, err := Resolve(globals, sc, t.Fst)
		if err != nil {
			return nil, err
		}
		S, err := Resolve(globals, sc, t.Snd)
		if err != nil {
			return nil, err
		}
		return Pair{Fst: F, Snd: S}, nil
	case RFst:
		P, err := Resolve(globals, sc, t.P)
		if err != nil {
			return nil, err
		}
		return Fst{P: P}, nil
	case RSnd:
		P, err := Resolve(globals, sc, t.P)
		if err != nil {
			return nil, err
		}
		return Snd{P: P}, nil
	case RLet:
		var ann Term
		var err error
		if t.Ann != nil {
			ann, err = Resolve(globals, sc, t.Ann)
			if err != nil {
				return nil, err
			}
		}
		val, err := Resolve(globals, sc, t.Val)
		if err != nil {
			return nil, err
		}
		body, err := Resolve(globals, sc.push(t.Binder), t.Body)
		if err != nil {
			return nil, err
		}
		return Let{Binder: t.Binder, Ann: ann, Val: val, Body: body}, nil
	default:
		return nil, fmt.Errorf("resolve: unknown raw term %T", t)
	}
}
