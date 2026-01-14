package eval

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// This file provides cached versions of NbE functions.
// These are drop-in replacements that use memoization for performance.

// EvalCached evaluates a term with caching support.
// If cache is nil, falls back to regular Eval.
func EvalCached(env *Env, t ast.Term, cache *Cache) Value {
	if cache == nil {
		return Eval(env, t)
	}

	if t == nil {
		return evalError("nil term")
	}
	if env == nil {
		env = &Env{Bindings: nil}
	}

	// Check cache first (using term and env pointer identity)
	if val, ok := cache.LookupEval(t, env); ok {
		return val
	}

	// Evaluate (mostly same as regular Eval, but with caching for recursive calls)
	var result Value

	switch tm := t.(type) {
	case ast.Var:
		result = env.Lookup(tm.Ix)

	case ast.Global:
		result = vGlobal(tm.Name)

	case ast.Sort:
		result = VSort{Level: int(tm.U)}

	case ast.Lam:
		result = VLam{Body: &Closure{Env: env, Term: tm.Body}}

	case ast.App:
		fun := EvalCached(env, tm.T, cache)
		arg := EvalCached(env, tm.U, cache)
		result = ApplyCached(fun, arg, cache)

	case ast.Pair:
		fst := EvalCached(env, tm.Fst, cache)
		snd := EvalCached(env, tm.Snd, cache)
		result = VPair{Fst: fst, Snd: snd}

	case ast.Fst:
		p := EvalCached(env, tm.P, cache)
		result = Fst(p)

	case ast.Snd:
		p := EvalCached(env, tm.P, cache)
		result = Snd(p)

	case ast.Pi:
		a := EvalCached(env, tm.A, cache)
		result = VPi{A: a, B: &Closure{Env: env, Term: tm.B}}

	case ast.Sigma:
		a := EvalCached(env, tm.A, cache)
		result = VSigma{A: a, B: &Closure{Env: env, Term: tm.B}}

	case ast.Let:
		val := EvalCached(env, tm.Val, cache)
		newEnv := env.Extend(val)
		result = EvalCached(newEnv, tm.Body, cache)

	case ast.Id:
		a := EvalCached(env, tm.A, cache)
		x := EvalCached(env, tm.X, cache)
		y := EvalCached(env, tm.Y, cache)
		result = VId{A: a, X: x, Y: y}

	case ast.Refl:
		a := EvalCached(env, tm.A, cache)
		x := EvalCached(env, tm.X, cache)
		result = VRefl{A: a, X: x}

	case ast.J:
		a := EvalCached(env, tm.A, cache)
		c := EvalCached(env, tm.C, cache)
		d := EvalCached(env, tm.D, cache)
		x := EvalCached(env, tm.X, cache)
		y := EvalCached(env, tm.Y, cache)
		p := EvalCached(env, tm.P, cache)
		result = evalJ(a, c, d, x, y, p)

	default:
		// Fall back to regular evaluation for extension types
		if val, ok := tryEvalCubical(env, t); ok {
			result = val
		} else {
			result = evalError("unknown term type")
		}
	}

	// Store in cache
	cache.StoreEval(t, env, result)
	return result
}

// ApplyCached performs function application with caching support.
func ApplyCached(fun Value, arg Value, cache *Cache) Value {
	switch f := fun.(type) {
	case VLam:
		// Beta reduction: evaluate body in extended environment
		newEnv := f.Body.Env.Extend(arg)
		return EvalCached(newEnv, f.Body.Term, cache)

	case VNeutral:
		// Extend the spine of the neutral term
		newSp := make([]Value, len(f.N.Sp)+1)
		copy(newSp, f.N.Sp)
		newSp[len(f.N.Sp)] = arg

		// Check for recursor computation rules
		if result := tryRecursorReduction(f.N.Head, newSp); result != nil {
			return result
		}

		return VNeutral{N: Neutral{Head: f.N.Head, Sp: newSp}}

	default:
		// Non-function applied to argument
		if DebugMode {
			panic("nbe: application to non-function")
		}
		head := Head{Glob: "error:bad_app"}
		return VNeutral{N: Neutral{Head: head, Sp: []Value{fun, arg}}}
	}
}

// ReifyCached converts a Value back to ast.Term.
// Note: Reification is not cached because Values may contain slices
// which are not hashable. The main performance benefit comes from
// caching evaluation, not reification.
func ReifyCached(v Value, cache *Cache) ast.Term {
	return Reify(v)
}

// EvalNBECached is a cached version of EvalNBE.
// It evaluates and reifies a term using NbE with memoization.
func EvalNBECached(t ast.Term, cache *Cache) ast.Term {
	if cache == nil {
		return EvalNBE(t)
	}
	env := &Env{Bindings: nil}
	val := EvalCached(env, t, cache)
	return ReifyCached(val, cache)
}

// WithCache creates a new cache, runs the provided function with it,
// and returns the result. The cache is automatically discarded after use.
// This is the recommended way to use caching for one-off operations.
func WithCache[T any](f func(*Cache) T) T {
	cache := NewDefaultCache()
	return f(cache)
}

// NormalizeWithCache normalizes a term using NbE with caching.
// This is a convenience function for common use cases.
func NormalizeWithCache(t ast.Term) ast.Term {
	return WithCache(func(c *Cache) ast.Term {
		return EvalNBECached(t, c)
	})
}
