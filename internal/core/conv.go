package core

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
)

// Env represents a typing environment for conversion checking.
// For now, this is a simple wrapper around eval.Env.
type Env struct {
	evalEnv *eval.Env
}

// NewEnv creates a new empty environment.
func NewEnv() *Env {
	return &Env{evalEnv: &eval.Env{Bindings: nil}}
}

// Extend adds a term to the environment by evaluating it.
func (e *Env) Extend(t ast.Term) *Env {
	val := eval.Eval(e.evalEnv, t)
	return &Env{evalEnv: e.evalEnv.Extend(val)}
}

// ConvOptions controls the behavior of definitional equality checking.
type ConvOptions struct {
	EnableEta bool // Enable η-equality for functions (Π) and pairs (Σ)
}

// Conv reports whether t and u are definitionally equal under env.
// If opts.EnableEta is true, use η-rules for functions (Pi) and pairs (Sigma).
//
// Implementation strategy:
// 1. Evaluate both terms to Values using NbE
// 2. Reify both Values back to normal forms
// 3. Apply η-expansion if enabled
// 4. Compare normal forms structurally
func Conv(env *Env, t, u ast.Term, opts ConvOptions) bool {
	// Handle nil environment
	if env == nil {
		env = NewEnv()
	}

	// Evaluate both terms using NbE
	valT := eval.Eval(env.evalEnv, t)
	valU := eval.Eval(env.evalEnv, u)

	// Reify to normal forms
	nfT := eval.Reify(valT)
	nfU := eval.Reify(valU)

	// Compare normal forms with η-equality if enabled
	if opts.EnableEta {
		return etaEqual(nfT, nfU)
	}

	// Structural comparison of normal forms
	return AlphaEq(nfT, nfU)
}

// etaEqual compares two terms with η-equality rules applied.
// This is a more sophisticated approach than simple η-expansion.
func etaEqual(a, b ast.Term) bool {
	// First try direct structural equality
	if AlphaEq(a, b) {
		return true
	}

	// Try η-equality for functions
	if etaEqualFunction(a, b) || etaEqualFunction(b, a) {
		return true
	}

	// Try η-equality for pairs
	if etaEqualPair(a, b) || etaEqualPair(b, a) {
		return true
	}

	return false
}

// etaEqualFunction checks if a neutral term is η-equal to a lambda.
// f ≡ \x. f x
func etaEqualFunction(neutral, lambda ast.Term) bool {
	lam, ok := lambda.(ast.Lam)
	if !ok {
		return false
	}

	// Check if lambda body is an application of the neutral to variable 0
	app, ok := lam.Body.(ast.App)
	if !ok {
		return false
	}

	// Check if the argument is variable 0
	if !AlphaEq(app.U, ast.Var{Ix: 0}) {
		return false
	}

	// Check if the function part is the neutral term (shifted by 1)
	shifted := shiftTerm(neutral, 1, 0)
	return AlphaEq(app.T, shifted)
}

// etaEqualPair checks if a neutral term is η-equal to a pair of its projections.
// p ≡ (fst p, snd p)
func etaEqualPair(neutral, pair ast.Term) bool {
	p, ok := pair.(ast.Pair)
	if !ok {
		return false
	}

	// Check if fst component is fst of neutral
	fstProj, ok := p.Fst.(ast.Fst)
	if !ok {
		return false
	}
	if !AlphaEq(fstProj.P, neutral) {
		return false
	}

	// Check if snd component is snd of neutral
	sndProj, ok := p.Snd.(ast.Snd)
	if !ok {
		return false
	}
	if !AlphaEq(sndProj.P, neutral) {
		return false
	}

	return true
}

// shiftTerm shifts de Bruijn indices in a term.
func shiftTerm(t ast.Term, d, cutoff int) ast.Term {
	if t == nil {
		return nil
	}
	switch tm := t.(type) {
	case ast.Var:
		if tm.Ix >= cutoff {
			return ast.Var{Ix: tm.Ix + d}
		}
		return tm
	case ast.Global:
		return tm
	case ast.Sort:
		return tm
	case ast.Lam:
		return ast.Lam{
			Binder: tm.Binder,
			Ann:    shiftTerm(tm.Ann, d, cutoff),
			Body:   shiftTerm(tm.Body, d, cutoff+1),
		}
	case ast.App:
		return ast.App{
			T: shiftTerm(tm.T, d, cutoff),
			U: shiftTerm(tm.U, d, cutoff),
		}
	case ast.Pi:
		return ast.Pi{
			Binder: tm.Binder,
			A:      shiftTerm(tm.A, d, cutoff),
			B:      shiftTerm(tm.B, d, cutoff+1),
		}
	case ast.Sigma:
		return ast.Sigma{
			Binder: tm.Binder,
			A:      shiftTerm(tm.A, d, cutoff),
			B:      shiftTerm(tm.B, d, cutoff+1),
		}
	case ast.Pair:
		return ast.Pair{
			Fst: shiftTerm(tm.Fst, d, cutoff),
			Snd: shiftTerm(tm.Snd, d, cutoff),
		}
	case ast.Fst:
		return ast.Fst{P: shiftTerm(tm.P, d, cutoff)}
	case ast.Snd:
		return ast.Snd{P: shiftTerm(tm.P, d, cutoff)}
	case ast.Let:
		return ast.Let{
			Binder: tm.Binder,
			Ann:    shiftTerm(tm.Ann, d, cutoff),
			Val:    shiftTerm(tm.Val, d, cutoff),
			Body:   shiftTerm(tm.Body, d, cutoff+1),
		}
	case ast.Id:
		return ast.Id{
			A: shiftTerm(tm.A, d, cutoff),
			X: shiftTerm(tm.X, d, cutoff),
			Y: shiftTerm(tm.Y, d, cutoff),
		}
	case ast.Refl:
		return ast.Refl{
			A: shiftTerm(tm.A, d, cutoff),
			X: shiftTerm(tm.X, d, cutoff),
		}
	case ast.J:
		return ast.J{
			A: shiftTerm(tm.A, d, cutoff),
			C: shiftTerm(tm.C, d, cutoff),
			D: shiftTerm(tm.D, d, cutoff),
			X: shiftTerm(tm.X, d, cutoff),
			Y: shiftTerm(tm.Y, d, cutoff),
			P: shiftTerm(tm.P, d, cutoff),
		}
	default:
		// Try extension handlers (e.g., cubical terms when built with -tags cubical)
		if result, ok := shiftTermExtension(t, d, cutoff); ok {
			return result
		}
		return t
	}
}

// AlphaEq compares two core terms modulo alpha (de Bruijn makes this structural).
func AlphaEq(a, b ast.Term) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

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
			return AlphaEq(a.Val, bb.Val) && AlphaEq(a.Body, bb.Body) &&
				((a.Ann == nil && bb.Ann == nil) ||
					(a.Ann != nil && bb.Ann != nil && AlphaEq(a.Ann, bb.Ann)))
		}
	case ast.Id:
		if bb, ok := b.(ast.Id); ok {
			return AlphaEq(a.A, bb.A) && AlphaEq(a.X, bb.X) && AlphaEq(a.Y, bb.Y)
		}
	case ast.Refl:
		if bb, ok := b.(ast.Refl); ok {
			return AlphaEq(a.A, bb.A) && AlphaEq(a.X, bb.X)
		}
	case ast.J:
		if bb, ok := b.(ast.J); ok {
			return AlphaEq(a.A, bb.A) && AlphaEq(a.C, bb.C) && AlphaEq(a.D, bb.D) &&
				AlphaEq(a.X, bb.X) && AlphaEq(a.Y, bb.Y) && AlphaEq(a.P, bb.P)
		}
	default:
		// Try extension handlers (e.g., cubical terms when built with -tags cubical)
		if result, ok := alphaEqExtension(a, b); ok {
			return result
		}
	}
	return false
}

// Legacy API compatibility - kept for existing tests
type EtaFlags struct{ Pi, Sigma bool }

// Conv (legacy) - wrapper for backward compatibility
func ConvLegacy(a, b ast.Term, flags EtaFlags) bool {
	env := NewEnv()
	opts := ConvOptions{EnableEta: flags.Pi || flags.Sigma}
	return Conv(env, a, b, opts)
}
